package workflowstatus

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

var ErrUnauthorized = errors.New("invalid session")

type Service struct {
	applications ApplicationClient
	executions   ExecutionReader
}

func NewService(applications ApplicationClient, executions ExecutionReader) *Service {
	return &Service{applications: applications, executions: executions}
}

// GetStatus authenticates the applicant's session (owned entirely by
// application-management-service — this service never persists a session
// of its own), looks up the execution linked to the application, and
// translates its current state into an applicant-facing status.
func (s *Service) GetStatus(ctx context.Context, token, applicationID string) (*WorkflowStatus, error) {
	valid, err := s.applications.ValidateSession(ctx, token, applicationID)
	if err != nil {
		return nil, fmt.Errorf("validating session: %w", err)
	}
	if !valid {
		return nil, ErrUnauthorized
	}

	executionArn, err := s.applications.ExecutionARN(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("looking up execution: %w", err)
	}

	status, outcome, err := s.executions.Outcome(ctx, executionArn)
	if err != nil {
		return nil, fmt.Errorf("reading execution outcome: %w", err)
	}

	switch status {
	case types.ExecutionStatusSucceeded:
		result := forOutcome(applicationID, outcome)
		return &result, nil
	case types.ExecutionStatusRunning:
		currentState, err := s.executions.CurrentState(ctx, executionArn)
		if err != nil {
			return nil, fmt.Errorf("reading current state: %w", err)
		}
		result := forRunningState(applicationID, currentState)
		return &result, nil
	default: // FAILED, TIMED_OUT, ABORTED
		result := forError(applicationID)
		return &result, nil
	}
}
