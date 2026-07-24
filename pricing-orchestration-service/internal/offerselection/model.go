package offerselection

import "pricing-orchestration-service/internal/pricing"

// Status tracks whether the applicant has acted on the priced offer yet.
type Status string

const (
	StatusPendingSelection Status = "PENDING_SELECTION"
	StatusConfirmed        Status = "CONFIRMED"
)

// SelectedOffer is the pause-point record present-offer-lambda writes when
// the workflow reaches AwaitOfferSelection, and this service's API reads/
// updates while the applicant reviews and confirms. Offers is every
// option presented; SelectedOfferID is populated on Confirm. TaskToken is
// never exposed over JSON — it's purely internal, needed to resume the
// paused Step Functions execution via SendTaskSuccess.
type SelectedOffer struct {
	ApplicationID   string                `json:"applicationId" dynamodbav:"applicationId"`
	Offers          []pricing.OfferOption `json:"offers" dynamodbav:"offers"`
	SelectedOfferID string                `json:"selectedOfferId,omitempty" dynamodbav:"selectedOfferId,omitempty"`
	Status          Status                `json:"status" dynamodbav:"status"`
	ConsentGiven    bool                  `json:"consentGiven" dynamodbav:"consentGiven"`
	TaskToken       string                `json:"-" dynamodbav:"taskToken"`
}
