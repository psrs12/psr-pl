package document

import "time"

// RequiredDocument is one document the applicant must submit before
// underwriting can proceed. Content here is a placeholder — not a real
// document-collection policy — same status as offer-acceptance-service's
// declaration text and pricing-orchestration-service's credit formulas:
// structurally correct, not production content. There is no real file
// upload/storage (no S3 integration) here either — Upload just records
// that a given document id was submitted, which is enough to exercise
// the DOCUMENTS_REQUIRED -> OFFER_ACCEPTED transition end to end.
type RequiredDocument struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// Status tracks whether every required document has been submitted yet.
type Status string

const (
	StatusPending  Status = "PENDING"
	StatusComplete Status = "COMPLETE"
)

// Record is the per-application document-submission state this service
// owns. Unlike offer-acceptance-service's EsignRecord (create-only,
// immutable), this is mutable — documents arrive one at a time, so the
// record is upserted as each one is submitted.
type Record struct {
	ApplicationID        string     `json:"applicationId" dynamodbav:"applicationId"`
	SubmittedDocumentIDs []string   `json:"submittedDocumentIds" dynamodbav:"submittedDocumentIds"`
	Status               Status     `json:"status" dynamodbav:"status"`
	CompletedAt          *time.Time `json:"completedAt,omitempty" dynamodbav:"completedAt,omitempty"`
}
