package offeracceptance

import "time"

// Declaration is a disclosure/agreement the applicant must acknowledge
// before e-signing. Content here is a placeholder — not real legal text —
// same status as pricing-orchestration-service's credit-scoring formulas:
// structurally correct, not production copy.
type Declaration struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Required bool   `json:"required"`
}

// EsignRecord is what this service persists once the applicant accepts
// all required declarations. Create/Get only — no update, no delete: an
// e-signature is a one-time, immutable act.
type EsignRecord struct {
	ApplicationID          string    `json:"applicationId" dynamodbav:"applicationId"`
	AcceptedDeclarationIDs []string  `json:"acceptedDeclarationIds" dynamodbav:"acceptedDeclarationIds"`
	SignedAt               time.Time `json:"signedAt" dynamodbav:"signedAt"`
}
