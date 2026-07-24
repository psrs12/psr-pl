package workflowstatus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ApplicationClient is a small, consumer-owned interface over
// application-management-service's internal (not applicant-facing)
// endpoints. Session ownership and the applicationId->executionArn
// mapping both stay in application-management-service — this service
// holds neither itself, consistent with the platform's "no duplicated
// source of truth" pattern (see OfferLog's offer-status handling).
type ApplicationClient interface {
	ExecutionARN(ctx context.Context, applicationID string) (string, error)
	ValidateSession(ctx context.Context, token, applicationID string) (bool, error)
	CurrentStatus(ctx context.Context, applicationID string) (string, error)
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

func (c *applicationManagementClient) ExecutionARN(ctx context.Context, applicationID string) (string, error) {
	url := fmt.Sprintf("%s/internal/applications/%s/execution", c.baseURL, applicationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("building execution lookup request: %w", err)
	}
	if id := requestIDFromContext(ctx); id != "" {
		req.Header.Set(RequestIDHeader, id)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling application-management-service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("application-management-service returned status %d", resp.StatusCode)
	}

	var body struct {
		ExecutionARN string `json:"executionArn"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decoding execution lookup response: %w", err)
	}
	return body.ExecutionARN, nil
}

// CurrentStatus fetches application-management-service's live status for
// the application -- the authoritative value once a Step Functions
// execution has gone terminal (see comment on the service-layer caller).
func (c *applicationManagementClient) CurrentStatus(ctx context.Context, applicationID string) (string, error) {
	url := fmt.Sprintf("%s/internal/applications/%s/status", c.baseURL, applicationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("building status lookup request: %w", err)
	}
	if id := requestIDFromContext(ctx); id != "" {
		req.Header.Set(RequestIDHeader, id)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling application-management-service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("application-management-service returned status %d", resp.StatusCode)
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decoding status lookup response: %w", err)
	}
	return body.Status, nil
}

func (c *applicationManagementClient) ValidateSession(ctx context.Context, token, applicationID string) (bool, error) {
	url := fmt.Sprintf("%s/internal/sessions/validate", c.baseURL)
	payload, err := json.Marshal(map[string]string{"token": token, "applicationId": applicationID})
	if err != nil {
		return false, fmt.Errorf("encoding session validation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("building session validation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if id := requestIDFromContext(ctx); id != "" {
		req.Header.Set(RequestIDHeader, id)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("calling application-management-service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var body struct {
		Valid bool `json:"valid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("decoding session validation response: %w", err)
	}
	return body.Valid, nil
}
