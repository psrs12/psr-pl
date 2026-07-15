// present-offer-lambda runs as the AwaitOfferSelection state's
// waitForTaskToken task. It does not call SendTaskSuccess itself — that
// happens later, when the applicant confirms via pricing-orchestration-
// service's POST .../selected-offer/confirm endpoint, which is what
// actually resumes the paused execution. This Lambda's only job is to
// hand the task token and the priced offer to that API so it has
// something to resume later.
package main

import (
	"bytes"
	"context"
	"encoding/json"
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

var (
	pricingOrchestrationBaseURL = os.Getenv("PRICING_ORCHESTRATION_BASE_URL")
	logger                      = slog.New(slog.NewJSONHandler(os.Stdout, nil))
)

type input struct {
	TaskToken     string              `json:"taskToken"`
	ApplicationID string              `json:"applicationId"`
	PricedOffer   pricing.PricedOffer `json:"pricedOffer"`
	RequestID     string              `json:"requestId"`
}

type presentRequest struct {
	ApplicationID string  `json:"applicationId"`
	TaskToken     string  `json:"taskToken"`
	AmountCents   int64   `json:"amountCents"`
	TermMonths    int     `json:"termMonths"`
	APRPercentage float64 `json:"aprPercentage"`
}

func handle(ctx context.Context, in input) (map[string]string, error) {
	log := logger.With("requestId", in.RequestID, "applicationId", in.ApplicationID)
	log.Info("present-offer-lambda invoked")

	if pricingOrchestrationBaseURL == "" {
		err := fmt.Errorf("PRICING_ORCHESTRATION_BASE_URL is not configured")
		log.Error("present-offer-lambda failed", "error", err)
		return nil, err
	}

	payload, err := json.Marshal(presentRequest{
		ApplicationID: in.ApplicationID,
		TaskToken:     in.TaskToken,
		AmountCents:   in.PricedOffer.AmountCents,
		TermMonths:    in.PricedOffer.TermMonths,
		APRPercentage: in.PricedOffer.APRPercentage,
	})
	if err != nil {
		return nil, fmt.Errorf("encoding present-offer request: %w", err)
	}

	url := pricingOrchestrationBaseURL + "/internal/applications/" + in.ApplicationID + "/selected-offer"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("building present-offer request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if in.RequestID != "" {
		req.Header.Set(requestIDHeader, in.RequestID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("present-offer-lambda failed", "error", err)
		return nil, fmt.Errorf("calling pricing-orchestration-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("pricing-orchestration-service rejected present-offer: %d", resp.StatusCode)
		log.Error("present-offer-lambda failed", "error", err)
		return nil, err
	}
	log.Info("present-offer-lambda succeeded")
	return map[string]string{"applicationId": in.ApplicationID}, nil
}

func main() {
	lambda.Start(handle)
}
