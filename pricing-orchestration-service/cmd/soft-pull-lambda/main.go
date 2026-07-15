package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"pricing-orchestration-service/internal/pricing"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func handle(ctx context.Context, input pricing.WorkflowInput) (pricing.CreditSummary, error) {
	log := logger.With("requestId", input.RequestID, "applicationId", input.ApplicationID)
	log.Info("soft-pull-lambda invoked")
	result := pricing.SimulatedCreditPull(input.ApplicationID, "SOFT_PULL")
	log.Info("soft-pull-lambda succeeded", "score", result.Score)
	return result, nil
}

func main() {
	lambda.Start(handle)
}
