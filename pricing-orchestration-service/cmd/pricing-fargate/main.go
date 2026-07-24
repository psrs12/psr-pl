// pricing-fargate is the CPU-bound pricing-calculation step, run as a
// Fargate/ECS task rather than a Lambda per CLAUDE.md's Lambda-vs-Fargate
// guidance. Step Functions' ECS RunTask.waitForTaskToken integration
// starts this container with the state's input and a task token in its
// environment; the container computes the result and reports it back via
// SendTaskSuccess/SendTaskFailure rather than returning a value directly
// (ECS RunTask has no return-value channel of its own).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"

	"pricing-orchestration-service/internal/pricing"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	if err := run(); err != nil {
		logger.Error("pricing-fargate failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	taskInput := os.Getenv("TASK_INPUT")
	taskToken := os.Getenv("TASK_TOKEN")
	if taskInput == "" || taskToken == "" {
		return fmt.Errorf("TASK_INPUT and TASK_TOKEN must both be set")
	}

	ctx := context.Background()
	client, err := newSFNClient(ctx)
	if err != nil {
		return fmt.Errorf("creating Step Functions client: %w", err)
	}

	var state pricing.WorkflowState
	if err := json.Unmarshal([]byte(taskInput), &state); err != nil {
		return sendFailure(ctx, client, taskToken, "InvalidInput", err.Error())
	}

	log := logger.With("requestId", state.RequestID, "applicationId", state.ApplicationID)
	log.Info("pricing-fargate invoked")

	if state.SoftPull == nil {
		log.Error("pricing-fargate failed", "error", "missing soft pull")
		return sendFailure(ctx, client, taskToken, "MissingSoftPull", "pricing requires a completed soft pull")
	}

	offers := pricing.PriceOptions(state.WorkflowInput, *state.SoftPull)

	output, err := json.Marshal(offers)
	if err != nil {
		return sendFailure(ctx, client, taskToken, "EncodingError", err.Error())
	}

	_, err = client.SendTaskSuccess(ctx, &sfn.SendTaskSuccessInput{
		TaskToken: &taskToken,
		Output:    stringPtr(string(output)),
	})
	if err != nil {
		return fmt.Errorf("sending task success: %w", err)
	}
	log.Info("pricing-fargate succeeded")
	return nil
}

func sendFailure(ctx context.Context, client *sfn.Client, taskToken, cause, errMsg string) error {
	_, err := client.SendTaskFailure(ctx, &sfn.SendTaskFailureInput{
		TaskToken: &taskToken,
		Error:     stringPtr(cause),
		Cause:     stringPtr(errMsg),
	})
	return err
}

func newSFNClient(ctx context.Context) (*sfn.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	var opts []func(*sfn.Options)
	if endpoint := os.Getenv("STEPFUNCTIONS_ENDPOINT_URL"); endpoint != "" {
		opts = append(opts, func(o *sfn.Options) { o.BaseEndpoint = &endpoint })
	}
	return sfn.NewFromConfig(cfg, opts...), nil
}

func stringPtr(s string) *string { return &s }
