package workflowstatus

// Status mirrors the subset of application-management-service's
// ApplicationStatus enum this service is able to derive from a Step
// Functions execution. It intentionally doesn't cover intake-only states
// (CREATED, IN_PROGRESS, ...) since those never reach a workflow
// execution at all.
type Status string

const (
	StatusProcessing      Status = "PROCESSING"
	StatusSoftPullPending Status = "SOFT_PULL_PENDING"
	StatusPricingPending  Status = "PRICING_PENDING"
	StatusOfferPending    Status = "OFFER_PENDING"
	StatusHardPullPending Status = "HARD_PULL_PENDING"
	StatusDecisionPending Status = "DECISION_PENDING"
	StatusApproved        Status = "APPROVED"
	StatusDeclined        Status = "DECLINED"
	StatusReferred        Status = "REFERRED"
	StatusError           Status = "ERROR"
)

type NextStep struct {
	Action      string `json:"action"`
	Description string `json:"description"`
}

// WorkflowStatus is what the applicant-facing endpoint returns — richer
// than a bare status label, per the decision to surface explicit
// next-steps rather than making the UI infer them from a status enum.
type WorkflowStatus struct {
	ApplicationID string     `json:"applicationId"`
	Status        Status     `json:"status"`
	CurrentStep   string     `json:"currentStep,omitempty"`
	NextSteps     []NextStep `json:"nextSteps"`
}
