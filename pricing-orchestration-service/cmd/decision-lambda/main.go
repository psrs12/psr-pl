package main

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/lambda"

	"pricing-orchestration-service/internal/pricing"
)

func handle(ctx context.Context, input pricing.WorkflowState) (pricing.Decision, error) {
	if input.PricedOffer == nil || input.HardPull == nil {
		return pricing.Decision{}, errors.New("decision requires both a priced offer and a hard pull result")
	}
	return pricing.Decide(*input.PricedOffer, *input.HardPull), nil
}

func main() {
	lambda.Start(handle)
}
