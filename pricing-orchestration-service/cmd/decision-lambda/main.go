package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"pricing-orchestration-service/internal/pricing"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func handle(ctx context.Context, input pricing.WorkflowState) (pricing.Decision, error) {
	log := logger.With("requestId", input.RequestID, "applicationId", input.ApplicationID)
	log.Info("decision-lambda invoked")
	if input.PricedOffer == nil || input.HardPull == nil {
		err := errors.New("decision requires both a priced offer and a hard pull result")
		log.Error("decision-lambda failed", "error", err)
		return pricing.Decision{}, err
	}
	decision := pricing.Decide(*input.PricedOffer, *input.HardPull)
	log.Info("decision-lambda succeeded", "outcome", decision.Outcome)
	return decision, nil
}

func main() {
	lambda.Start(handle)
}
