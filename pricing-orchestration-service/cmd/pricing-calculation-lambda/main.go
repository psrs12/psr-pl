// pricing-calculation-lambda is a LocalStack-only substitute for
// cmd/pricing-fargate. This LocalStack deployment's license doesn't
// include the ECS API, so the state machine's PricingCalculation step
// can't use ecs:runTask there; this runs the identical
// pricing.PriceOptions(...) logic as an ordinary lambda:invoke Task
// instead of an ecs:runTask.waitForTaskToken Fargate task. The real,
// production ASL (pricing-orchestration-service/statemachine/
// definition.asl.json) is untouched and still Fargate-based — this
// function is only referenced by the separate LocalStack-only state
// machine definition in deploy/cloudformation/psr-pl-localstack.yaml.
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

func handle(ctx context.Context, state pricing.WorkflowState) ([]pricing.OfferOption, error) {
	log := logger.With("requestId", state.RequestID, "applicationId", state.ApplicationID)
	log.Info("pricing-calculation-lambda invoked")

	if state.SoftPull == nil {
		err := errors.New("pricing requires a completed soft pull")
		log.Error("pricing-calculation-lambda failed", "error", err)
		return nil, err
	}

	offers := pricing.PriceOptions(state.WorkflowInput, *state.SoftPull)
	log.Info("pricing-calculation-lambda succeeded", "offerCount", len(offers))
	return offers, nil
}

func main() {
	lambda.Start(handle)
}
