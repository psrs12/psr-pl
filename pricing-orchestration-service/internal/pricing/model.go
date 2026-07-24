package pricing

// Address is the applicant's current address, as needed to identify them
// to a credit bureau (same shape as application-management-service's own
// Address, duplicated here since this is a different service's payload).
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postcode"`
}

// WorkflowInput is the payload application-management-service passes when
// starting the Step Functions execution. Includes the applicant identity
// fields a real tri-bureau/alternative-data credit pull needs to look up a
// person's record at all — FirstName, LastName, DateOfBirth, SSN, Address
// — alongside the financial figures pricing itself uses. An earlier
// version of this payload deliberately excluded all of these on the
// assumption pricing only needed loan-request figures; that assumption
// didn't survive contact with how bureau pulls actually work (see
// rebuild-platform-go/design.md's reversed decision and the resulting
// Step-Functions-execution-history PII risk it now accepts).
type WorkflowInput struct {
	ApplicationID           string  `json:"applicationId"`
	FirstName               string  `json:"firstName"`
	LastName                string  `json:"lastName"`
	DateOfBirth             string  `json:"dateOfBirth"`
	SSN                     string  `json:"ssn"`
	Address                 Address `json:"address"`
	RequestedAmountCents    int64   `json:"requestedAmountCents"`
	RequestedTermMonths     int     `json:"requestedTermMonths"`
	AnnualIncomeCents       int64   `json:"annualIncomeCents"`
	MonthlyObligationsCents int64   `json:"monthlyObligationsCents"`
	// RequestID traces this execution back to the HTTP request that
	// started it (see application-management-service's WorkflowInput).
	RequestID string `json:"requestId,omitempty"`
}

// CreditSummary is the simulated output of a credit bureau pull. There is
// no real bureau integration here — SoftPull and HardPull both produce
// one of these from a deterministic stand-in, clearly marked as such.
type CreditSummary struct {
	Score  int    `json:"score"`
	Source string `json:"source"`
}

// PricedOffer is one set of loan terms: the terms actually being offered,
// which may differ from what the applicant requested (e.g. a lower amount
// if debt-to-income doesn't support the full ask).
type PricedOffer struct {
	AmountCents   int64   `json:"amountCents"`
	TermMonths    int     `json:"termMonths"`
	APRPercentage float64 `json:"aprPercentage"`
}

// OfferOption is one of several priced offers presented to the applicant
// to choose from (see PriceOptions) -- e.g. the requested terms, a
// lower-payment/longer-term variant, and a faster-payoff/shorter-term
// variant. ID is stable within one pricing run and is what the applicant's
// selection (and the eventual Decide call) refers back to.
type OfferOption struct {
	ID    string `json:"offerId"`
	Label string `json:"label"`
	PricedOffer
}

// Decision is decision-routing's terminal output.
type Decision struct {
	Outcome string `json:"outcome"` // APPROVED | DECLINED | REFERRED
	Reason  string `json:"reason"`
}

// OfferSelectionResult is AwaitOfferSelection's output once the applicant
// has responded: the consent decision, and (when an offer was presented)
// which one they selected, echoed back so DecisionRouting doesn't need a
// second round trip to look it up.
type OfferSelectionResult struct {
	ConsentGiven  bool         `json:"consentGiven"`
	SelectedOffer *OfferOption `json:"selectedOffer,omitempty"`
}

// WorkflowState is the accumulated execution data as it flows through the
// state machine: each state adds its own key via ResultPath rather than
// replacing the whole payload, so later states (and workflow-status-service
// reading execution history) can see everything that happened so far.
type WorkflowState struct {
	WorkflowInput
	SoftPull       *CreditSummary        `json:"softPull,omitempty"`
	PricedOffers   []OfferOption         `json:"pricedOffers,omitempty"`
	OfferSelection *OfferSelectionResult `json:"offerSelection,omitempty"`
	HardPull       *CreditSummary        `json:"hardPull,omitempty"`
	Decision       *Decision             `json:"decision,omitempty"`
}
