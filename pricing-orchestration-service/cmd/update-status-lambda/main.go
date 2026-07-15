package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"pricing-orchestration-service/internal/pricing"
)

// requestIDHeader matches the header name application-management-service and
// pricing-orchestration-service's REST handlers use, so a requestId can be
// grepped across every service's logs for one request.
const requestIDHeader = "X-Request-Id"

// applicationManagementBaseURL points at application-management-service's
// internal (not applicant-facing) endpoints — secured at the network/IAM
// layer, same pattern as its existing workflow status-update endpoint.
var (
	applicationManagementBaseURL = os.Getenv("APPLICATION_MANAGEMENT_BASE_URL")
	logger                       = slog.New(slog.NewJSONHandler(os.Stdout, nil))
)

type updateStatusRequest struct {
	Status string `json:"status"`
}

func handle(ctx context.Context, input pricing.WorkflowState) (pricing.WorkflowState, error) {
	log := logger.With("requestId", input.RequestID, "applicationId", input.ApplicationID)
	log.Info("update-status-lambda invoked")

	if input.Decision == nil {
		err := errors.New("update-application-status requires a decision")
		log.Error("update-status-lambda failed", "error", err)
		return input, err
	}
	if applicationManagementBaseURL == "" {
		err := errors.New("APPLICATION_MANAGEMENT_BASE_URL is not configured")
		log.Error("update-status-lambda failed", "error", err)
		return input, err
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
	if input.RequestID != "" {
		req.Header.Set(requestIDHeader, input.RequestID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("update-status-lambda failed", "error", err)
		return input, fmt.Errorf("calling application-management-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("application-management-service rejected status update: %d", resp.StatusCode)
		log.Error("update-status-lambda failed", "error", err)
		return input, err
	}
	log.Info("update-status-lambda succeeded")
	return input, nil
}

func main() {
	lambda.Start(handle)
}
