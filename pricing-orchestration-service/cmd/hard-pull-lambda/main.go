package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"pricing-orchestration-service/internal/pricing"
)

func handle(ctx context.Context, input pricing.WorkflowState) (pricing.CreditSummary, error) {
	return pricing.SimulatedCreditPull(input.ApplicationID, "HARD_PULL"), nil
}

func main() {
	lambda.Start(handle)
}
