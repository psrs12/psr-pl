#!/usr/bin/env bash
# MiniStack deployment for psr-pl (ministackorg/ministack, MIT-licensed).
# Partial migration from an earlier LocalStack-based deployment: MiniStack
# implements real ECS (genuine ecs:RunTask Docker container execution,
# verified via a standalone smoke test) so the 5 backend services and
# pricing-fargate run as real ECS tasks now, not plain `docker run`
# containers, and the state machine registers the real, unmodified
# pricing-orchestration-service/statemachine/definition.asl.json. VpcLink
# and Cloud Map DNS resolution do NOT work on MiniStack (verified, not
# assumed -- VpcLink's API 404s, Cloud Map records don't resolve inside
# real containers), so API Gateway integrations and service-to-service
# calls still use host.docker.internal:<port> directly, same as the prior
# deployment.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ENDPOINT="http://localhost:4566"
REGION="us-east-1"
STACK_NAME="psr-pl"
TEMPLATE="$ROOT/deploy/cloudformation/psr-pl-localstack.yaml"
LAMBDA_BUCKET="psr-pl-lambda-artifacts"
BUILD_DIR="$ROOT/deploy/localstack/.build"

APP_MGMT_PORT=8081
OFFERLOG_MOCK_PORT=9090
PRICING_PORT=8082
WORKFLOW_PORT=8086
OFFER_ACCEPTANCE_PORT=8085
DOCUMENT_PORT=8084

# UI ports. MiniStack has no S3 static-website-hosting mode at all
# (confirmed live: the s3-website.* hostname just returns a plain bucket
# listing, not index/error document behavior) -- so unlike the earlier
# LocalStack deployment, the 4 UIs are nginx containers on fixed,
# published host ports instead of S3 buckets. Still public/unauthenticated
# (no login wall at the hosting layer), just not via S3.
APP_MGMT_UI_PORT=3000
PRICING_OFFERS_UI_PORT=3001
DOCUMENT_MANAGEMENT_UI_PORT=3002
OFFER_ACCEPTANCE_UI_PORT=3003

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION="$REGION"
# Newer AWS CLI defaults to CRC64NVME checksums on S3 uploads, which
# MiniStack's S3 implementation doesn't support (bundles SHA256/SHA1/CRC32
# only, not the native CRC64NVME/CRC32C deps) -- verified via a failed
# `aws s3 cp` before adding this.
export AWS_REQUEST_CHECKSUM_CALCULATION=when_required

aw() { aws --endpoint-url="$ENDPOINT" --region "$REGION" "$@"; }

echo "==> Building pricing Lambda binaries (linux/arm64, provided.al2023)"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"
for fn in soft-pull-lambda hard-pull-lambda decision-lambda present-offer-lambda update-status-lambda; do
  echo "  - $fn"
  (cd "$ROOT/pricing-orchestration-service" && \
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o "$BUILD_DIR/$fn/bootstrap" "./cmd/$fn")
  (cd "$BUILD_DIR/$fn" && zip -q "../$fn.zip" bootstrap)
done

echo "==> Creating Lambda artifact bucket ($LAMBDA_BUCKET) and uploading zips"
aw s3 mb "s3://$LAMBDA_BUCKET" >/dev/null 2>&1 || true
for fn in soft-pull-lambda hard-pull-lambda decision-lambda present-offer-lambda update-status-lambda; do
  aw s3 cp "$BUILD_DIR/$fn.zip" "s3://$LAMBDA_BUCKET/$fn.zip" >/dev/null
done

echo "==> Building backend Docker images (incl. pricing-fargate -- real ECS task now, not a Lambda substitute)"
docker build -t psr-pl/application-management-service:latest "$ROOT/application-management-service"
docker build -f "$ROOT/application-management-service/cmd/offerlog-mock/Dockerfile" \
  -t psr-pl/offerlog-mock:latest "$ROOT/application-management-service"
docker build -t psr-pl/pricing-orchestration-service:latest "$ROOT/pricing-orchestration-service"
docker build -f "$ROOT/pricing-orchestration-service/cmd/pricing-fargate/Dockerfile" \
  -t psr-pl/pricing-fargate:latest "$ROOT/pricing-orchestration-service"
docker build -t psr-pl/workflow-status-service:latest "$ROOT/workflow-status-service"
docker build -t psr-pl/offer-acceptance-service:latest "$ROOT/offer-acceptance-service"
docker build -t psr-pl/document-service:latest "$ROOT/document-service"

echo "==> Deploying CloudFormation stack ($STACK_NAME)"
aw cloudformation deploy \
  --template-file "$TEMPLATE" \
  --stack-name "$STACK_NAME" \
  --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
  --parameter-overrides LambdaArtifactBucket="$LAMBDA_BUCKET" \
    AppMgmtHostPort="$APP_MGMT_PORT" OfferlogMockHostPort="$OFFERLOG_MOCK_PORT" \
    PricingHostPort="$PRICING_PORT" WorkflowHostPort="$WORKFLOW_PORT" \
    OfferAcceptanceHostPort="$OFFER_ACCEPTANCE_PORT" DocumentHostPort="$DOCUMENT_PORT"

# The Lambda zips always land at the same S3 key, so CloudFormation sees
# no property diff on the Function resources and skips redeploying code
# on updates -- force it explicitly every run so a rebuilt Lambda binary
# actually takes effect. (Carried over from the LocalStack deployment;
# re-verify this is still needed on MiniStack rather than assuming.)
echo "==> Force-refreshing Lambda code (same S3 key every run -- CFN alone may not detect the change)"
for fn in soft-pull-lambda hard-pull-lambda decision-lambda present-offer-lambda update-status-lambda; do
  aw lambda update-function-code --function-name "$fn" --s3-bucket "$LAMBDA_BUCKET" --s3-key "$fn.zip" >/dev/null
done

# MiniStack's CloudFormation didn't pick up an IntegrationUri change on
# `cloudformation deploy` (confirmed: the resource stayed at its old value
# after a successful "Successfully created/updated stack") -- same class
# of update-detection gap as the Lambda code issue above, just for a
# different resource type. Force it explicitly too.
echo "==> Force-refreshing API Gateway v2 integration URIs (CFN alone did not apply a confirmed change)"
force_refresh_integration() {
  api_id="$1"; port="$2"
  integration_id=$(aw apigatewayv2 get-integrations --api-id "$api_id" --query 'Items[0].IntegrationId' --output text)
  aw apigatewayv2 update-integration --api-id "$api_id" --integration-id "$integration_id" \
    --integration-uri "http://host.docker.internal:$port" >/dev/null
}
APP_MGMT_API_ID=$(aw cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='AppMgmtApiUrl'].OutputValue" --output text | sed -E 's#http://([^.]+)\..*#\1#')
PRICING_API_ID=$(aw cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='PricingApiUrl'].OutputValue" --output text | sed -E 's#http://([^.]+)\..*#\1#')
WORKFLOW_API_ID=$(aw cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='WorkflowApiUrl'].OutputValue" --output text | sed -E 's#http://([^.]+)\..*#\1#')
OFFER_ACCEPTANCE_API_ID=$(aw cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='OfferAcceptanceApiUrl'].OutputValue" --output text | sed -E 's#http://([^.]+)\..*#\1#')
DOCUMENT_API_ID=$(aw cloudformation describe-stacks --stack-name "$STACK_NAME" --query "Stacks[0].Outputs[?OutputKey=='DocumentApiUrl'].OutputValue" --output text | sed -E 's#http://([^.]+)\..*#\1#')
force_refresh_integration "$APP_MGMT_API_ID" "$APP_MGMT_PORT"
force_refresh_integration "$PRICING_API_ID" "$PRICING_PORT"
force_refresh_integration "$WORKFLOW_API_ID" "$WORKFLOW_PORT"
force_refresh_integration "$OFFER_ACCEPTANCE_API_ID" "$OFFER_ACCEPTANCE_PORT"
force_refresh_integration "$DOCUMENT_API_ID" "$DOCUMENT_PORT"

# MiniStack's HTTP API v2 rejects CORS preflight (OPTIONS) with a hard
# 403 "CORS not configured" unless CorsConfiguration is explicitly set on
# the Api resource -- real AWS just proxies OPTIONS through the ANY route
# like any other method when CORS isn't configured, so this is a
# MiniStack-specific behavior, not something the backend's own CORS
# middleware can work around. The template sets CorsConfiguration, but
# confirm it actually took (same update-detection caution as above).
echo "==> Confirming/force-refreshing API Gateway v2 CORS config"
for api_id in "$APP_MGMT_API_ID" "$PRICING_API_ID" "$WORKFLOW_API_ID" "$OFFER_ACCEPTANCE_API_ID" "$DOCUMENT_API_ID"; do
  aw apigatewayv2 update-api --api-id "$api_id" --cors-configuration 'AllowOrigins=*,AllowMethods=*,AllowHeaders=*' >/dev/null
done

# MiniStack's CloudFormation doesn't do in-place updates for
# AWS::ApiGatewayV2::Api at all -- confirmed by inspecting `apigatewayv2
# get-apis` after several `cloudformation deploy` runs against this same
# stack: each run minted a fresh set of 5 Api resources (new ApiIds) and
# left the previous generation registered but untracked by the stack,
# rather than updating them in place. The orphans still work (they proxy
# to the same host.docker.internal ports), which is exactly what made
# this easy to miss -- but they pile up forever otherwise. Delete any API
# with one of these 5 names whose id isn't in the current stack outputs.
echo "==> Cleaning up orphaned API Gateway v2 APIs from prior deploys"
aw apigatewayv2 get-apis --query 'Items[].{id:ApiId,name:Name}' --output json | python3 -c "
import json, sys
current_ids = {'$APP_MGMT_API_ID', '$PRICING_API_ID', '$WORKFLOW_API_ID', '$OFFER_ACCEPTANCE_API_ID', '$DOCUMENT_API_ID'}
managed_names = {'application-management-api', 'pricing-orchestration-api', 'workflow-status-api', 'offer-acceptance-api', 'document-api'}
for item in json.load(sys.stdin):
    if item['name'] in managed_names and item['id'] not in current_ids:
        print(item['id'])
" > "$BUILD_DIR/orphan-apis.txt"
if [ -s "$BUILD_DIR/orphan-apis.txt" ]; then
  while IFS= read -r orphan_id; do
    aw apigatewayv2 delete-api --api-id "$orphan_id" >/dev/null 2>&1 && echo "  - deleted orphaned API $orphan_id"
  done < "$BUILD_DIR/orphan-apis.txt"
else
  echo "  - none found"
fi

echo "==> Reading stack outputs"
OUTPUTS_JSON=$(aw cloudformation describe-stacks --stack-name "$STACK_NAME" --query 'Stacks[0].Outputs' --output json)
out() { echo "$OUTPUTS_JSON" | python3 -c "import json,sys; d={o['OutputKey']:o['OutputValue'] for o in json.load(sys.stdin)}; print(d['$1'])"; }

APP_MGMT_API_URL=$(out AppMgmtApiUrl)
PRICING_API_URL=$(out PricingApiUrl)
WORKFLOW_API_URL=$(out WorkflowApiUrl)
OFFER_ACCEPTANCE_API_URL=$(out OfferAcceptanceApiUrl)
DOCUMENT_API_URL=$(out DocumentApiUrl)
STATE_MACHINE_ARN=$(out StateMachineArn)

APP_MGMT_UI_URL="http://localhost:$APP_MGMT_UI_PORT"
PRICING_OFFERS_UI_URL="http://localhost:$PRICING_OFFERS_UI_PORT"
OFFER_ACCEPTANCE_UI_URL="http://localhost:$OFFER_ACCEPTANCE_UI_PORT"
DOCUMENT_MANAGEMENT_UI_URL="http://localhost:$DOCUMENT_MANAGEMENT_UI_PORT"

echo "    application-management-service API: $APP_MGMT_API_URL"
echo "    pricing-orchestration-service API:   $PRICING_API_URL"
echo "    workflow-status-service API:         $WORKFLOW_API_URL"
echo "    offer-acceptance-service API:        $OFFER_ACCEPTANCE_API_URL"
echo "    document-service API:                $DOCUMENT_API_URL"
echo "    state machine:                       $STATE_MACHINE_ARN"

echo "==> Waiting for ECS services to report RUNNING (real ecs:RunTask containers now, not docker run)"
for svc in application-management-service pricing-orchestration-service workflow-status-service offer-acceptance-service document-service; do
  for i in $(seq 1 30); do
    RUNNING=$(aw ecs describe-services --cluster psr-pl-cluster --services "$svc" --query 'services[0].runningCount' --output text 2>/dev/null || echo 0)
    [ "$RUNNING" = "1" ] && { echo "  - $svc RUNNING"; break; }
    sleep 2
  done
done

echo "==> Waiting for backend containers to become healthy (host-published ports)"
for i in $(seq 1 30); do
  ok=1
  for port in "$APP_MGMT_PORT" "$PRICING_PORT" "$WORKFLOW_PORT" "$OFFER_ACCEPTANCE_PORT" "$DOCUMENT_PORT"; do
    curl -sf "http://localhost:$port/healthz" >/dev/null 2>&1 || ok=0
  done
  [ "$ok" = "1" ] && break
  sleep 1
done

echo "==> Building UI Docker images (pricing-offers-ui, offer-acceptance-ui, document-management-ui)"
docker build -t psr-pl/pricing-offers-ui:latest "$ROOT/pricing-offers-ui"
docker build -t psr-pl/offer-acceptance-ui:latest "$ROOT/offer-acceptance-ui"
docker build -t psr-pl/document-management-ui:latest "$ROOT/document-management-ui"

echo "==> Building application-management-ui (needs the other 3 UIs' URLs baked in)"
docker build -t psr-pl/application-management-ui:latest "$ROOT/application-management-ui" \
  --build-arg VITE_APP_MANAGEMENT_API_URL="$APP_MGMT_API_URL/api/v1/application-management" \
  --build-arg VITE_WORKFLOW_STATUS_API_URL="$WORKFLOW_API_URL/api/v1/workflow-status" \
  --build-arg VITE_PRICING_API_URL="$PRICING_API_URL/api/v1/pricing-orchestration" \
  --build-arg VITE_OFFER_ACCEPTANCE_API_URL="$OFFER_ACCEPTANCE_API_URL/api/v1/offer-acceptance" \
  --build-arg VITE_DOCUMENT_API_URL="$DOCUMENT_API_URL/api/v1/document" \
  --build-arg VITE_PRICING_OFFERS_UI_JS_URL="$PRICING_OFFERS_UI_URL/pricing-offer-selector.iife.js" \
  --build-arg VITE_OFFER_ACCEPTANCE_UI_JS_URL="$OFFER_ACCEPTANCE_UI_URL/offer-acceptance-flow.iife.js" \
  --build-arg VITE_DOCUMENT_MANAGEMENT_UI_JS_URL="$DOCUMENT_MANAGEMENT_UI_URL/document-upload-manager.iife.js"

echo "==> Starting UI containers (plain docker run, public/unauthenticated -- see template notes on why not S3)"
for name in psr-pl-application-management-ui psr-pl-pricing-offers-ui psr-pl-offer-acceptance-ui psr-pl-document-management-ui; do
  docker rm -f "$name" >/dev/null 2>&1 || true
done
docker run -d --name psr-pl-application-management-ui -p "$APP_MGMT_UI_PORT:80" psr-pl/application-management-ui:latest
docker run -d --name psr-pl-pricing-offers-ui -p "$PRICING_OFFERS_UI_PORT:80" psr-pl/pricing-offers-ui:latest
docker run -d --name psr-pl-offer-acceptance-ui -p "$OFFER_ACCEPTANCE_UI_PORT:80" psr-pl/offer-acceptance-ui:latest
docker run -d --name psr-pl-document-management-ui -p "$DOCUMENT_MANAGEMENT_UI_PORT:80" psr-pl/document-management-ui:latest

echo "==> Waiting for UI containers to become healthy"
for i in $(seq 1 30); do
  ok=1
  for port in "$APP_MGMT_UI_PORT" "$PRICING_OFFERS_UI_PORT" "$OFFER_ACCEPTANCE_UI_PORT" "$DOCUMENT_MANAGEMENT_UI_PORT"; do
    curl -sf "http://localhost:$port/" >/dev/null 2>&1 || ok=0
  done
  [ "$ok" = "1" ] && break
  sleep 1
done

echo ""
echo "==> Done."
echo "    application-management-ui: $APP_MGMT_UI_URL"
echo "    pricing-offers-ui:         $PRICING_OFFERS_UI_URL"
echo "    offer-acceptance-ui:      $OFFER_ACCEPTANCE_UI_URL"
echo "    document-management-ui:   $DOCUMENT_MANAGEMENT_UI_URL"
echo "    application-management-service API: $APP_MGMT_API_URL"
echo "    pricing-orchestration-service API:   $PRICING_API_URL"
echo "    workflow-status-service API:         $WORKFLOW_API_URL"
echo "    offer-acceptance-service API:        $OFFER_ACCEPTANCE_API_URL"
echo "    document-service API:                $DOCUMENT_API_URL"
