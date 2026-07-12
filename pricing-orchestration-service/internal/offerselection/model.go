package offerselection

// Status tracks whether the applicant has acted on the priced offer yet.
type Status string

const (
	StatusPendingSelection Status = "PENDING_SELECTION"
	StatusConfirmed        Status = "CONFIRMED"
)

// SelectedOffer is the pause-point record present-offer-lambda writes when
// the workflow reaches AwaitOfferSelection, and this service's API reads/
// updates while the applicant reviews and confirms. TaskToken is never
// exposed over JSON — it's purely internal, needed to resume the paused
// Step Functions execution via SendTaskSuccess.
type SelectedOffer struct {
	ApplicationID string  `json:"applicationId" dynamodbav:"applicationId"`
	AmountCents   int64   `json:"amountCents" dynamodbav:"amountCents"`
	TermMonths    int     `json:"termMonths" dynamodbav:"termMonths"`
	APRPercentage float64 `json:"aprPercentage" dynamodbav:"aprPercentage"`
	Status        Status  `json:"status" dynamodbav:"status"`
	ConsentGiven  bool    `json:"consentGiven" dynamodbav:"consentGiven"`
	TaskToken     string  `json:"-" dynamodbav:"taskToken"`
}
