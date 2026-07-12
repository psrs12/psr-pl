package offerselection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SessionValidator is a small, consumer-owned interface over
// application-management-service's internal session-validate endpoint —
// the same one workflow-status-service uses. Session ownership stays
// entirely in application-management-service; this service never caches
// or duplicates it.
type SessionValidator interface {
	ValidateSession(ctx context.Context, token, applicationID string) (bool, error)
}

type applicationManagementSessionValidator struct {
	baseURL    string
	httpClient *http.Client
}

func NewApplicationManagementSessionValidator(baseURL string, httpClient *http.Client) *applicationManagementSessionValidator {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &applicationManagementSessionValidator{baseURL: baseURL, httpClient: httpClient}
}

func (c *applicationManagementSessionValidator) ValidateSession(ctx context.Context, token, applicationID string) (bool, error) {
	payload, err := json.Marshal(map[string]string{"token": token, "applicationId": applicationID})
	if err != nil {
		return false, fmt.Errorf("encoding session validation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/sessions/validate", bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("building session validation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
