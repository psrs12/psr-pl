package document

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ApplicationStatusUpdater is a small, consumer-owned interface over
// application-management-service's internal status-update endpoint — the
// same one offer-acceptance-service's Service and pricing-orchestration-
// service's update-status-lambda call.
type ApplicationStatusUpdater interface {
	UpdateStatus(ctx context.Context, applicationID, status string) error
}

type applicationManagementClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewApplicationManagementClient(baseURL string, httpClient *http.Client) *applicationManagementClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &applicationManagementClient{baseURL: baseURL, httpClient: httpClient}
}

func (c *applicationManagementClient) UpdateStatus(ctx context.Context, applicationID, status string) error {
	payload, err := json.Marshal(map[string]string{"status": status})
	if err != nil {
		return fmt.Errorf("encoding status update: %w", err)
	}

	url := fmt.Sprintf("%s/internal/applications/%s/status", c.baseURL, applicationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("building status update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if id := requestIDFromContext(ctx); id != "" {
		req.Header.Set(RequestIDHeader, id)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling application-management-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("application-management-service rejected status update: %d", resp.StatusCode)
	}
	return nil
}
