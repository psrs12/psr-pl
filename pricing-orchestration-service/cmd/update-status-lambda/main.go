package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"pricing-orchestration-service/internal/pricing"
)

// applicationManagementBaseURL points at application-management-service's
// internal (not applicant-facing) endpoints — secured at the network/IAM
// layer, same pattern as its existing workflow status-update endpoint.
var applicationManagementBaseURL = os.Getenv("APPLICATION_MANAGEMENT_BASE_URL")

type updateStatusRequest struct {
	Status string `json:"status"`
}

func handle(ctx context.Context, input pricing.WorkflowState) (pricing.WorkflowState, error) {
	if input.Decision == nil {
		return input, errors.New("update-application-status requires a decision")
	}
	if applicationManagementBaseURL == "" {
		return input, errors.New("APPLICATION_MANAGEMENT_BASE_URL is not configured")
	}

	payload, err := json.Marshal(updateStatusRequest{Status: input.Decision.Outcome})
	if err != nil {
		return input, fmt.Errorf("encoding status update: %w", err)
	}

	url := fmt.Sprintf("%s/internal/applications/%s/status", applicationManagementBaseURL, input.ApplicationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return input, fmt.Errorf("building status update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return input, fmt.Errorf("calling application-management-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return input, fmt.Errorf("application-management-service rejected status update: %d", resp.StatusCode)
	}
	return input, nil
}

func main() {
	lambda.Start(handle)
}
