.PHONY: run-all stop-all tail-logs

LOG_DIR := logs

## run-all: start every backend service in the order documented in README.md,
## each with its output redirected to logs/<service>.log instead of your
## terminal. Use `make tail-logs` to follow all of them at once, or
## `tail -f logs/<service>.log | jq .` for just one (they're slog JSON).
run-all:
	@mkdir -p $(LOG_DIR)
	@echo "Registering pricing-orchestration-service's state machine (Step Functions Local :8083)..."
	@$(MAKE) -C pricing-orchestration-service statemachine-up > $(LOG_DIR)/stepfunctions-setup.log 2>&1
	@echo "Starting application-management-service (:8081)..."
	@( $(MAKE) -C application-management-service run > $(LOG_DIR)/application-management-service.log 2>&1 & echo $$! > $(LOG_DIR)/application-management-service.pid )
	@sleep 2
	@echo "Starting pricing-orchestration-service (:8082)..."
	@( $(MAKE) -C pricing-orchestration-service run > $(LOG_DIR)/pricing-orchestration-service.log 2>&1 & echo $$! > $(LOG_DIR)/pricing-orchestration-service.pid )
	@echo "Starting workflow-status-service (:8086)..."
	@( cd workflow-status-service && \
		APPLICATION_MANAGEMENT_BASE_URL=http://localhost:8081 \
		STEPFUNCTIONS_ENDPOINT_URL=http://localhost:8083 \
		AWS_ACCESS_KEY_ID=local AWS_SECRET_ACCESS_KEY=local AWS_DEFAULT_REGION=us-east-1 \
		PORT=8086 go run ./cmd/server > ../$(LOG_DIR)/workflow-status-service.log 2>&1 & echo $$! > $(LOG_DIR)/workflow-status-service.pid )
	@echo "Starting offer-acceptance-service (:8085)..."
	@( $(MAKE) -C offer-acceptance-service run > $(LOG_DIR)/offer-acceptance-service.log 2>&1 & echo $$! > $(LOG_DIR)/offer-acceptance-service.pid )
	@sleep 2
	@echo ""
	@echo "All services started. Logs: $(LOG_DIR)/*.log -- run 'make tail-logs' to follow them, 'make stop-all' to stop everything."

## tail-logs: follow every service's log at once, prefixed by filename.
tail-logs:
	@tail -f $(LOG_DIR)/*.log

## stop-all: tear down everything run-all started (services, DynamoDB Local,
## Step Functions Local) via each service's own `down` target, plus
## workflow-status-service's port directly since it has no Makefile of its own.
stop-all:
	@-$(MAKE) -C application-management-service down
	@-$(MAKE) -C pricing-orchestration-service down
	@-$(MAKE) -C offer-acceptance-service down
	@-kill -9 $$(lsof -ti tcp:8086) 2>/dev/null || true
	@echo "Stopped."
