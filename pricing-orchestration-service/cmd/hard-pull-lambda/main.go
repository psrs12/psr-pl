package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"pricing-orchestration-service/internal/pricing"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func handle(ctx context.Context, input pricing.WorkflowState) (pricing.CreditSummary, error) {
	log := logger.With("requestId", input.RequestID, "applicationId", input.ApplicationID)
	log.Info("hard-pull-lambda invoked")
	result := pricing.SimulatedCreditPull(input.ApplicationID, "HARD_PULL")
	log.Info("hard-pull-lambda succeeded", "score", result.Score)
	return result, nil
}

func main() {
	lambda.Start(handle)
}
