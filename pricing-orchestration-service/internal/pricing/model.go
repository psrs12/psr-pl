package pricing

// WorkflowInput is the payload application-management-service passes when
// starting the Step Functions execution. Deliberately minimal — no name,
// SSN, DOB, or address: pricing needs financial figures and a bureau
// reference only, not full applicant PII, since this payload lives in
// Step Functions execution history (see rebuild-platform-go/design.md's
// risk note on what's acceptable to pass through workflow state).
type WorkflowInput struct {
	ApplicationID           string `json:"applicationId"`
	RequestedAmountCents    int64  `json:"requestedAmountCents"`
	RequestedTermMonths     int    `json:"requestedTermMonths"`
	AnnualIncomeCents       int64  `json:"annualIncomeCents"`
	MonthlyObligationsCents int64  `json:"monthlyObligationsCents"`
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

// PricedOffer is pricing-calculation's output: the terms actually being
// offered, which may differ from what the applicant requested (e.g. a
// lower amount if debt-to-income doesn't support the full ask).
type PricedOffer struct {
	AmountCents   int64   `json:"amountCents"`
	TermMonths    int     `json:"termMonths"`
	APRPercentage float64 `json:"aprPercentage"`
}

// Decision is decision-routing's terminal output.
type Decision struct {
	Outcome string `json:"outcome"` // APPROVED | DECLINED | REFERRED
	Reason  string `json:"reason"`
}

// WorkflowState is the accumulated execution data as it flows through the
// state machine: each state adds its own key via ResultPath rather than
// replacing the whole payload, so later states (and workflow-status-service
// reading execution history) can see everything that happened so far.
type WorkflowState struct {
	WorkflowInput
	SoftPull    *CreditSummary `json:"softPull,omitempty"`
	PricedOffer *PricedOffer   `json:"pricedOffer,omitempty"`
	HardPull    *CreditSummary `json:"hardPull,omitempty"`
	Decision    *Decision      `json:"decision,omitempty"`
}
