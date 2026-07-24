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
// machine. Includes the applicant identity fields (first name, last name,
// DOB, SSN, address) a real tri-bureau/alternative-data credit pull needs
// to identify the applicant, alongside the financial figures pricing uses
// — see pricing-orchestration-service's own WorkflowInput and
// rebuild-platform-go/design.md's reversed "no PII in workflow input"
// decision. This does mean SSN-grade PII now lives in Step Functions
// execution history for the life of the execution — an accepted,
// documented trade-off, not an oversight.
type WorkflowInput struct {
	ApplicationID           string  `json:"applicationId"`
	FirstName               string  `json:"firstName"`
	LastName                string  `json:"lastName"`
	DateOfBirth             string  `json:"dateOfBirth"`
	SSN                     string  `json:"ssn"`
	Address                 Address `json:"address"`
	RequestedAmountCents    int64   `json:"requestedAmountCents"`
	RequestedTermMonths     int     `json:"requestedTermMonths"`
	AnnualIncomeCents       int64   `json:"annualIncomeCents"`
	MonthlyObligationsCents int64   `json:"monthlyObligationsCents"`
	// RequestID is the submit request's X-Request-Id, carried through so the
	// whole workflow execution (and every Lambda/Fargate step it invokes)
	// can be traced back to the original HTTP request that started it.
	RequestID string `json:"requestId,omitempty"`
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
