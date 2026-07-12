.PHONY: build vet fmt test dynamodb-up dynamodb-down dynamodb-table offerlog-mock-up offerlog-mock-down dev-env down run stop

GO := go

DYNAMODB_CONTAINER := dynamodb-local
DYNAMODB_PORT      := 8000
DYNAMODB_ENDPOINT  := http://localhost:$(DYNAMODB_PORT)
TABLE_NAME         := pl-application
PORT               := 8081
OFFERLOG_MOCK_PORT := 9090
OFFERLOG_BASE_URL  := http://localhost:$(OFFERLOG_MOCK_PORT)
OFFERLOG_MOCK_PID_FILE := /tmp/psr-pl-offerlog-mock.pid
OFFERLOG_MOCK_LOG      := /tmp/psr-pl-offerlog-mock.log

# pricing-orchestration-service's Step Functions Local instance. Start it
# separately first: `cd ../pricing-orchestration-service && make statemachine-up`.
STEPFUNCTIONS_ENDPOINT_URL := http://localhost:8083
PRICING_STATE_MACHINE_ARN  := arn:aws:states:us-east-1:123456789012:stateMachine:PricingOrchestration

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

## dynamodb-table: create pl-application with TTL enabled (idempotent).
dynamodb-table:
	@$(AWS_LOCAL_ENV) aws dynamodb create-table \
		--endpoint-url $(DYNAMODB_ENDPOINT) \
		--table-name $(TABLE_NAME) \
		--attribute-definitions AttributeName=id,AttributeType=S \
		--key-schema AttributeName=id,KeyType=HASH \
		--billing-mode PAY_PER_REQUEST >/dev/null 2>&1 || true
	@$(AWS_LOCAL_ENV) aws dynamodb update-time-to-live \
		--endpoint-url $(DYNAMODB_ENDPOINT) \
		--table-name $(TABLE_NAME) \
		--time-to-live-specification "Enabled=true,AttributeName=ttl" >/dev/null 2>&1 || true

## offerlog-mock-up: start the local OfferLog stand-in (idempotent).
## Fixtures: VALID-INVITE-123 (active), USED-INVITE-456 (already used),
## any other VALID-* token mints a fresh active offer on first use.
offerlog-mock-up:
	@curl -sf $(OFFERLOG_BASE_URL)/healthz >/dev/null 2>&1 && \
		echo "offerlog-mock already running" || \
		( OFFERLOG_MOCK_PORT=$(OFFERLOG_MOCK_PORT) nohup $(GO) run ./cmd/offerlog-mock > $(OFFERLOG_MOCK_LOG) 2>&1 & \
		  echo $$! > $(OFFERLOG_MOCK_PID_FILE) )
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		curl -sf $(OFFERLOG_BASE_URL)/healthz >/dev/null 2>&1 && break; \
		sleep 1; \
	done

## offerlog-mock-down: stop the local OfferLog stand-in. Kills by port, not
## the PID file — `go run` execs a child binary, so the PID file only ever
## has the wrapper's PID, and killing just that can leave the real process
## holding the port.
offerlog-mock-down:
	@kill -9 $$(lsof -ti tcp:$(OFFERLOG_MOCK_PORT)) 2>/dev/null || true
	@rm -f $(OFFERLOG_MOCK_PID_FILE)

## stop: free the service's own port (in case a previous `make run` was
## killed without letting it shut down cleanly, e.g. `go run`'s child
## process outliving a killed parent).
stop:
	@kill -9 $$(lsof -ti tcp:$(PORT)) 2>/dev/null || true

## dev-env: bring up DynamoDB Local, ensure the table exists, and start
## the OfferLog mock.
dev-env: dynamodb-up dynamodb-table offerlog-mock-up

## down: tear down everything dev-env (and run) started.
down: dynamodb-down offerlog-mock-down stop

## run: run the service against local DynamoDB and the OfferLog mock.
## Requires dev-env.
run: dev-env
	APPLICATION_TABLE_NAME=$(TABLE_NAME) \
	OFFERLOG_BASE_URL=$(OFFERLOG_BASE_URL) \
	DYNAMODB_ENDPOINT_URL=$(DYNAMODB_ENDPOINT) \
	STEPFUNCTIONS_ENDPOINT_URL=$(STEPFUNCTIONS_ENDPOINT_URL) \
	PRICING_STATE_MACHINE_ARN=$(PRICING_STATE_MACHINE_ARN) \
	PORT=$(PORT) \
	$(AWS_LOCAL_ENV) \
	$(GO) run ./cmd/server
