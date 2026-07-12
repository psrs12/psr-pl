package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/google/uuid"
)

// WorkflowStarter is a small, consumer-owned interface: this service only
// ever starts the pricing-orchestration workflow, it never reads its
// state back (that's workflow-status-service's job, reached independently
// by the UI — see internal/applications/{id}/execution below).
type WorkflowStarter interface {
	Start(ctx context.Context, input WorkflowInput) (executionArn string, err error)
}

// WorkflowInput is the payload handed to the pricing-orchestration state
// machine. Deliberately minimal — financial figures only, no name, SSN,
// DOB, or address, since this payload lives in Step Functions execution
// history for the life of the execution.
type WorkflowInput struct {
	ApplicationID           string `json:"applicationId"`
	RequestedAmountCents    int64  `json:"requestedAmountCents"`
	RequestedTermMonths     int    `json:"requestedTermMonths"`
	AnnualIncomeCents       int64  `json:"annualIncomeCents"`
	MonthlyObligationsCents int64  `json:"monthlyObligationsCents"`
}

type sfnWorkflowStarter struct {
	client          *sfn.Client
	stateMachineARN string
}

func NewSFNWorkflowStarter(client *sfn.Client, stateMachineARN string) *sfnWorkflowStarter {
	return &sfnWorkflowStarter{client: client, stateMachineARN: stateMachineARN}
}

func (s *sfnWorkflowStarter) Start(ctx context.Context, input WorkflowInput) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("encoding workflow input: %w", err)
	}

	name := "app-" + input.ApplicationID + "-" + uuid.NewString()[:8]
	out, err := s.client.StartExecution(ctx, &sfn.StartExecutionInput{
		StateMachineArn: &s.stateMachineARN,
		Name:            &name,
		Input:           stringPtr(string(payload)),
	})
	if err != nil {
		return "", fmt.Errorf("starting execution: %w", err)
	}
	return *out.ExecutionArn, nil
}

func stringPtr(s string) *string { return &s }
