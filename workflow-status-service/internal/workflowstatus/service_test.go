package workflowstatus

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

type fakeApplicationClient struct {
	executionArn string
	sessionValid bool
}

func (f *fakeApplicationClient) ExecutionARN(ctx context.Context, applicationID string) (string, error) {
	return f.executionArn, nil
}

func (f *fakeApplicationClient) ValidateSession(ctx context.Context, token, applicationID string) (bool, error) {
	return f.sessionValid, nil
}

type fakeExecutionReader struct {
	status       types.ExecutionStatus
	outcome      string
	currentState string
}

func (f *fakeExecutionReader) CurrentState(ctx context.Context, executionArn string) (string, error) {
	return f.currentState, nil
}

func (f *fakeExecutionReader) Outcome(ctx context.Context, executionArn string) (types.ExecutionStatus, string, error) {
	return f.status, f.outcome, nil
}

func TestGetStatusRejectsInvalidSession(t *testing.T) {
	svc := NewService(&fakeApplicationClient{sessionValid: false}, &fakeExecutionReader{})

	_, err := svc.GetStatus(context.Background(), "bad-token", "app-1")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGetStatusRunningStateTranslatesToNextSteps(t *testing.T) {
	svc := NewService(
		&fakeApplicationClient{sessionValid: true, executionArn: "arn:exec:1"},
		&fakeExecutionReader{status: types.ExecutionStatusRunning, currentState: "PricingCalculation"},
	)

	status, err := svc.GetStatus(context.Background(), "token", "app-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != StatusPricingPending {
		t.Errorf("Status = %s, want %s", status.Status, StatusPricingPending)
	}
	if status.CurrentStep != "PricingCalculation" {
		t.Errorf("CurrentStep = %s, want PricingCalculation", status.CurrentStep)
	}
	if len(status.NextSteps) == 0 {
		t.Error("expected at least one next step while running")
	}
}

func TestGetStatusApprovedOutcome(t *testing.T) {
	svc := NewService(
		&fakeApplicationClient{sessionValid: true, executionArn: "arn:exec:1"},
		&fakeExecutionReader{status: types.ExecutionStatusSucceeded, outcome: "APPROVED"},
	)

	status, err := svc.GetStatus(context.Background(), "token", "app-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != StatusApproved {
		t.Errorf("Status = %s, want %s", status.Status, StatusApproved)
	}
	if status.NextSteps[0].Action != "SELECT_OFFER" {
		t.Errorf("NextSteps[0].Action = %s, want SELECT_OFFER", status.NextSteps[0].Action)
	}
}

func TestGetStatusFailedExecutionReturnsError(t *testing.T) {
	svc := NewService(
		&fakeApplicationClient{sessionValid: true, executionArn: "arn:exec:1"},
		&fakeExecutionReader{status: types.ExecutionStatusFailed},
	)

	status, err := svc.GetStatus(context.Background(), "token", "app-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Status != StatusError {
		t.Errorf("Status = %s, want %s", status.Status, StatusError)
	}
}
