package workflowstatus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

// ExecutionReader is a small, consumer-owned interface over the two Step
// Functions read APIs this service needs. It never writes to the state
// machine — only application-management-service and the workflow's own
// Lambda steps do that.
type ExecutionReader interface {
	// CurrentState returns the name of the most recently entered state
	// for a still-running execution.
	CurrentState(ctx context.Context, executionArn string) (string, error)
	// Outcome returns the execution's status (RUNNING/SUCCEEDED/FAILED/...)
	// and, if SUCCEEDED, the terminal decision outcome from its output.
	Outcome(ctx context.Context, executionArn string) (status types.ExecutionStatus, decisionOutcome string, err error)
}

type sfnExecutionReader struct {
	client *sfn.Client
}

func NewSFNExecutionReader(client *sfn.Client) *sfnExecutionReader {
	return &sfnExecutionReader{client: client}
}

func (r *sfnExecutionReader) CurrentState(ctx context.Context, executionArn string) (string, error) {
	out, err := r.client.GetExecutionHistory(ctx, &sfn.GetExecutionHistoryInput{
		ExecutionArn: &executionArn,
		ReverseOrder: true,
		MaxResults:   20,
	})
	if err != nil {
		return "", fmt.Errorf("getting execution history: %w", err)
	}

	for _, event := range out.Events {
		if event.Type == types.HistoryEventTypeTaskStateEntered && event.StateEnteredEventDetails != nil {
			return *event.StateEnteredEventDetails.Name, nil
		}
	}
	return "", nil
}

func (r *sfnExecutionReader) Outcome(ctx context.Context, executionArn string) (types.ExecutionStatus, string, error) {
	out, err := r.client.DescribeExecution(ctx, &sfn.DescribeExecutionInput{ExecutionArn: &executionArn})
	if err != nil {
		return "", "", fmt.Errorf("describing execution: %w", err)
	}

	if out.Status != types.ExecutionStatusSucceeded || out.Output == nil {
		return out.Status, "", nil
	}

	var result struct {
		Decision struct {
			Outcome string `json:"outcome"`
		} `json:"decision"`
	}
	if err := json.Unmarshal([]byte(*out.Output), &result); err != nil {
		return out.Status, "", fmt.Errorf("parsing execution output: %w", err)
	}
	return out.Status, result.Decision.Outcome, nil
}
