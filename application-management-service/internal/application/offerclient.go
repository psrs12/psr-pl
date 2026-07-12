package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrOfferUnavailable is returned when OfferLog reports an offer as
// invalid, expired, or already used. Callers degrade to the direct path
// rather than treating this as a hard failure.
var ErrOfferUnavailable = fmt.Errorf("offer unavailable")

// OfferValidator and OfferStatusUpdater are small, consumer-owned
// interfaces (not a hexagonal port) satisfied by offerLogClient. Offer
// status is never cached locally — every read and write goes straight to
// OfferLog. ValidateOffer is called twice in the lifecycle of an
// invitation-path application: once with the raw invitation token (when
// the applicant opens the link), and again with the intakeId OfferLog
// issued from that first call (at submission, to re-validate live and
// snapshot final terms).
type OfferValidator interface {
	ValidateOffer(ctx context.Context, reference string) (*Offer, Applicant, error)
}

type OfferStatusUpdater interface {
	MarkUsed(ctx context.Context, intakeID string) error
	RevertActive(ctx context.Context, intakeID string) error
}

type offerLogClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOfferLogClient(baseURL string, httpClient *http.Client) *offerLogClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &offerLogClient{baseURL: baseURL, httpClient: httpClient}
}

type validateInvitationRequest struct {
	Token string `json:"token"`
}

type validateInvitationResponse struct {
	IntakeID        string  `json:"intakeId"`
	OfferID         string  `json:"offerId"`
	CampaignOfferID string  `json:"campaignOfferId"`
	TermMonths      int     `json:"requestedTermMonths"`
	AmountCents     int64   `json:"requestedAmountCents"`
	APRPercentage   float64 `json:"aprPercentage"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	Address         Address `json:"address"`
	Status          string  `json:"status"`
}

func (c *offerLogClient) ValidateOffer(ctx context.Context, reference string) (*Offer, Applicant, error) {
	url := fmt.Sprintf("%s/invitations/validate", c.baseURL)
	payload, err := json.Marshal(validateInvitationRequest{Token: reference})
	if err != nil {
		return nil, Applicant{}, fmt.Errorf("encoding offer validation request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, Applicant{}, fmt.Errorf("building offer validation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, Applicant{}, fmt.Errorf("calling OfferLog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusConflict {
		return nil, Applicant{}, ErrOfferUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return nil, Applicant{}, fmt.Errorf("OfferLog returned status %d", resp.StatusCode)
	}

	var body validateInvitationResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, Applicant{}, fmt.Errorf("decoding offer validation response: %w", err)
	}
	if body.Status != "ACTIVE" {
		return nil, Applicant{}, ErrOfferUnavailable
	}

	offer := &Offer{
		IntakeID:        body.IntakeID,
		OfferID:         body.OfferID,
		CampaignOfferID: body.CampaignOfferID,
		TermMonths:      body.TermMonths,
		AmountCents:     body.AmountCents,
		APRPercentage:   body.APRPercentage,
	}
	prefill := Applicant{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Address:   body.Address,
	}
	return offer, prefill, nil
}

type offerStatusRequest struct {
	Status string `json:"status"`
}

func (c *offerLogClient) MarkUsed(ctx context.Context, intakeID string) error {
	return c.updateStatus(ctx, intakeID, "USED")
}

func (c *offerLogClient) RevertActive(ctx context.Context, intakeID string) error {
	return c.updateStatus(ctx, intakeID, "ACTIVE")
}

func (c *offerLogClient) updateStatus(ctx context.Context, intakeID, status string) error {
	url := fmt.Sprintf("%s/invitations/%s/status", c.baseURL, intakeID)
	payload, err := json.Marshal(offerStatusRequest{Status: status})
	if err != nil {
		return fmt.Errorf("encoding offer status request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("building offer status request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling OfferLog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OfferLog rejected status update: %d", resp.StatusCode)
	}
	return nil
}
