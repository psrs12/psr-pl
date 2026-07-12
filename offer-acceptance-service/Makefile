.PHONY: build vet fmt test dynamodb-up dynamodb-down dynamodb-table dev-env down run stop

GO := go

DYNAMODB_CONTAINER := dynamodb-local
DYNAMODB_PORT       := 8000
DYNAMODB_ENDPOINT   := http://localhost:$(DYNAMODB_PORT)
TABLE_NAME          := pl-offer-acceptance
PORT                := 8085
APPLICATION_MANAGEMENT_BASE_URL := http://localhost:8081

AWS_LOCAL_ENV := AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local AWS_DEFAULT_REGION=us-east-1

build:
	$(GO) build ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

test:
	$(GO) test ./...

## dynamodb-up: start DynamoDB Local (idempotent — no-op if already running).
dynamodb-up:
	@docker inspect -f '{{.State.Running}}' $(DYNAMODB_CONTAINER) >/dev/null 2>&1 && \
		echo "$(DYNAMODB_CONTAINER) already running" || \
		docker run -d --name $(DYNAMODB_CONTAINER) -p $(DYNAMODB_PORT):8000 \
			amazon/dynamodb-local:latest -jar DynamoDBLocal.jar -sharedDb -inMemory
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		curl -s -o /dev/null $(DYNAMODB_ENDPOINT) && break; \
		sleep 1; \
	done

## dynamodb-down: stop and remove the local DynamoDB container.
dynamodb-down:
	@docker rm -f $(DYNAMODB_CONTAINER) >/dev/null 2>&1 || true

## dynamodb-table: create pl-offer-acceptance (idempotent).
dynamodb-table:
	@$(AWS_LOCAL_ENV) aws dynamodb create-table \
		--endpoint-url $(DYNAMODB_ENDPOINT) \
		--table-name $(TABLE_NAME) \
		--attribute-definitions AttributeName=applicationId,AttributeType=S \
		--key-schema AttributeName=applicationId,KeyType=HASH \
		--billing-mode PAY_PER_REQUEST >/dev/null 2>&1 || true

## stop: free the server's own port.
stop:
	@kill -9 $$(lsof -ti tcp:$(PORT)) 2>/dev/null || true

## dev-env: bring up DynamoDB Local and ensure the table exists.
dev-env: dynamodb-up dynamodb-table

## down: tear down everything dev-env (and run) started.
down: dynamodb-down stop

## run: run the service against local DynamoDB. Requires
## application-management-service to be running separately.
run: dev-env
	OFFER_ACCEPTANCE_TABLE_NAME=$(TABLE_NAME) \
	APPLICATION_MANAGEMENT_BASE_URL=$(APPLICATION_MANAGEMENT_BASE_URL) \
	DYNAMODB_ENDPOINT_URL=$(DYNAMODB_ENDPOINT) \
	PORT=$(PORT) \
	$(AWS_LOCAL_ENV) \
	$(GO) run ./cmd/server
