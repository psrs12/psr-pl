.PHONY: build vet fmt test stepfunctions-up stepfunctions-down statemachine-create statemachine-up run-happy-path run-declined-path describe

GO := go

STEPFUNCTIONS_CONTAINER := stepfunctions-local
STEPFUNCTIONS_PORT      := 8083
STEPFUNCTIONS_ENDPOINT  := http://localhost:$(STEPFUNCTIONS_PORT)
STATE_MACHINE_NAME      := PricingOrchestration
STATE_MACHINE_DEFINITION_FILE := statemachine/definition.asl.json
MOCK_CONFIG_FILE               := statemachine/local-mock-config.json

AWS_LOCAL_ENV := AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local AWS_DEFAULT_REGION=us-east-1

build:
	$(GO) build ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

test:
	$(GO) test ./...

## stepfunctions-up: start AWS Step Functions Local (idempotent), with the
## mock-response config mounted so states can be tested without real
## Lambda/Fargate. See statemachine/local-mock-config.json.
stepfunctions-up:
	@docker inspect -f '{{.State.Running}}' $(STEPFUNCTIONS_CONTAINER) >/dev/null 2>&1 && \
		echo "$(STEPFUNCTIONS_CONTAINER) already running" || \
		docker run -d --name $(STEPFUNCTIONS_CONTAINER) -p $(STEPFUNCTIONS_PORT):8083 \
			-v $(CURDIR)/$(MOCK_CONFIG_FILE):/home/StepFunctionsLocal/MockConfigFile.json \
			-e SFN_MOCK_CONFIG=/home/StepFunctionsLocal/MockConfigFile.json \
			amazon/aws-stepfunctions-local
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		curl -s -o /dev/null $(STEPFUNCTIONS_ENDPOINT) && break; \
		sleep 1; \
	done

## stepfunctions-down: stop and remove the local Step Functions container.
stepfunctions-down:
	@docker rm -f $(STEPFUNCTIONS_CONTAINER) >/dev/null 2>&1 || true

## statemachine-create: register definition.asl.json with placeholder
## Lambda/Fargate ARNs (they're never actually invoked when the execution
## name selects a mocked test case) — idempotent.
statemachine-create: stepfunctions-up
	@$(AWS_LOCAL_ENV) aws stepfunctions create-state-machine \
		--endpoint-url $(STEPFUNCTIONS_ENDPOINT) \
		--name $(STATE_MACHINE_NAME) \
		--definition "$$(sed \
			-e 's#$${SoftPullLambdaArn}#arn:aws:lambda:us-east-1:000000000000:function:soft-pull#' \
			-e 's#$${HardPullLambdaArn}#arn:aws:lambda:us-east-1:000000000000:function:hard-pull#' \
			-e 's#$${DecisionLambdaArn}#arn:aws:lambda:us-east-1:000000000000:function:decision#' \
			-e 's#$${UpdateStatusLambdaArn}#arn:aws:lambda:us-east-1:000000000000:function:update-status#' \
			-e 's#$${EcsClusterArn}#arn:aws:ecs:us-east-1:000000000000:cluster/local#' \
			-e 's#$${PricingFargateTaskDefinitionArn}#arn:aws:ecs:us-east-1:000000000000:task-definition/pricing-fargate#' \
			$(STATE_MACHINE_DEFINITION_FILE))" \
		--role-arn "arn:aws:iam::000000000000:role/local-unused" >/dev/null 2>&1 || true

## statemachine-up: stepfunctions-up + statemachine-create in one step.
statemachine-up: statemachine-create

## run-happy-path: execute the HappyPath mocked test case (soft pull 720,
## hard pull 710 -> APPROVED) and print the final status/output.
run-happy-path: statemachine-up
	@ARN=$$($(AWS_LOCAL_ENV) aws stepfunctions list-state-machines --endpoint-url $(STEPFUNCTIONS_ENDPOINT) --query "stateMachines[?name=='$(STATE_MACHINE_NAME)'].stateMachineArn" --output text); \
	EXEC_ARN=$$($(AWS_LOCAL_ENV) aws stepfunctions start-execution --endpoint-url $(STEPFUNCTIONS_ENDPOINT) --state-machine-arn "$$ARN" --name "HappyPath" --input '{"applicationId":"app-happy","requestedAmountCents":1500000,"requestedTermMonths":60,"annualIncomeCents":8500000,"monthlyObligationsCents":30000}' --query executionArn --output text); \
	sleep 1; \
	$(AWS_LOCAL_ENV) aws stepfunctions describe-execution --endpoint-url $(STEPFUNCTIONS_ENDPOINT) --execution-arn "$$EXEC_ARN"

## run-declined-path: execute the DeclinedPath mocked test case (soft pull
## 550, hard pull 540 -> DECLINED) and print the final status/output.
run-declined-path: statemachine-up
	@ARN=$$($(AWS_LOCAL_ENV) aws stepfunctions list-state-machines --endpoint-url $(STEPFUNCTIONS_ENDPOINT) --query "stateMachines[?name=='$(STATE_MACHINE_NAME)'].stateMachineArn" --output text); \
	EXEC_ARN=$$($(AWS_LOCAL_ENV) aws stepfunctions start-execution --endpoint-url $(STEPFUNCTIONS_ENDPOINT) --state-machine-arn "$$ARN" --name "DeclinedPath" --input '{"applicationId":"app-declined","requestedAmountCents":1500000,"requestedTermMonths":60,"annualIncomeCents":8500000,"monthlyObligationsCents":30000}' --query executionArn --output text); \
	sleep 1; \
	$(AWS_LOCAL_ENV) aws stepfunctions describe-execution --endpoint-url $(STEPFUNCTIONS_ENDPOINT) --execution-arn "$$EXEC_ARN"
