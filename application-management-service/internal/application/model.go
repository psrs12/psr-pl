package application

import "time"

// EntryPath is derived from the X-Channel-ID header the UI sends: ITA for
// the invitation path, WEB for direct. There is no channel value for the
// partner path yet — the UI doesn't support it.
type EntryPath string

const (
	EntryPathInvitation EntryPath = "INVITATION"
	EntryPathDirect     EntryPath = "DIRECT"
	EntryPathPartner    EntryPath = "PARTNER"
)

// Status matches services/application-management-service's ApplicationStatus
// enum exactly, per application-management-ui/src/navigation/navigationConfig.js
// (the UI's screen routing is keyed off these exact values).
type Status string

const (
	StatusCreated            Status = "CREATED"
	StatusInProgress         Status = "IN_PROGRESS"
	StatusReadyForSubmission Status = "READY_FOR_SUBMISSION"
	StatusSubmitted          Status = "SUBMITTED"
	StatusProcessing         Status = "PROCESSING"
	StatusSoftPullPending    Status = "SOFT_PULL_PENDING"
	StatusPricingPending     Status = "PRICING_PENDING"
	StatusConsentCaptured    Status = "CONSENT_CAPTURED"
	StatusHardPullPending    Status = "HARD_PULL_PENDING"
	StatusDecisionPending    Status = "DECISION_PENDING"
	StatusOfferPending       Status = "OFFER_PENDING"
	StatusApproved           Status = "APPROVED"
	StatusDocumentsRequired  Status = "DOCUMENTS_REQUIRED"
	StatusUnderwriting       Status = "UNDERWRITING"
	StatusReferred           Status = "REFERRED"
	StatusDeclined           Status = "DECLINED"
	StatusCancelled          Status = "CANCELLED"
	StatusExpired            Status = "EXPIRED"
	StatusOfferAccepted      Status = "OFFER_ACCEPTED"
	StatusFundingPending     Status = "FUNDING_PENDING"
	StatusFunded             Status = "FUNDED"
	StatusCompleted          Status = "COMPLETED"
)

type Address struct {
	Line1      string `json:"line1" dynamodbav:"line1"`
	Line2      string `json:"line2,omitempty" dynamodbav:"line2,omitempty"`
	City       string `json:"city" dynamodbav:"city"`
	State      string `json:"state" dynamodbav:"state"`
	PostalCode string `json:"postcode" dynamodbav:"postalCode"`
}

// Offer is a snapshot of offer economics taken at submission time.
// Offer status (active/used) is never stored here — OfferLog remains the
// sole source of truth for status, referenced only by IntakeID.
type Offer struct {
	IntakeID        string  `json:"intakeId" dynamodbav:"intakeId"`
	OfferID         string  `json:"offerId" dynamodbav:"offerId"`
	CampaignOfferID string  `json:"campaignOfferId,omitempty" dynamodbav:"campaignOfferId,omitempty"`
	TermMonths      int     `json:"termMonths" dynamodbav:"termMonths"`
	AmountCents     int64   `json:"amountCents" dynamodbav:"amountCents"`
	APRPercentage   float64 `json:"aprPercentage" dynamodbav:"aprPercentage"`
}

type Employment struct {
	EmployerName     string `json:"employerName,omitempty" dynamodbav:"employerName,omitempty"`
	EmploymentStatus string `json:"employmentStatus" dynamodbav:"employmentStatus"`
}

// Applicant carries every field required before submission. Name and
// Address are prefilled and non-editable on the invitation path; every
// other field is always collected fresh regardless of entry path.
type Applicant struct {
	FirstName               string     `json:"firstName" dynamodbav:"firstName"`
	LastName                string     `json:"lastName" dynamodbav:"lastName"`
	Address                 Address    `json:"address" dynamodbav:"address"`
	Phone                   string     `json:"phone" dynamodbav:"phone"`
	Email                   string     `json:"email" dynamodbav:"email"`
	Employment              Employment `json:"employment" dynamodbav:"employment"`
	AnnualIncomeCents       int64      `json:"annualIncomeCents" dynamodbav:"annualIncomeCents"`
	MonthlyObligationsCents int64      `json:"monthlyObligationsCents" dynamodbav:"monthlyObligationsCents"`
	SSN                     string     `json:"-" dynamodbav:"ssn"`
	SSNVerificationToken    string     `json:"-" dynamodbav:"ssnVerificationToken,omitempty"`
	DateOfBirth             string     `json:"dateOfBirth" dynamodbav:"dateOfBirth"` // ISO 8601 date
	CitizenshipStatus       string     `json:"citizenshipStatus" dynamodbav:"citizenshipStatus"`
	RequestedAmountCents    int64      `json:"requestedAmountCents" dynamodbav:"requestedAmountCents"`
	RequestedTermMonths     int        `json:"requestedTermMonths" dynamodbav:"requestedTermMonths"`
	LoanPurpose             string     `json:"loanPurpose" dynamodbav:"loanPurpose"`
}

// Application is the only aggregate this service persists. It is created
// once, at submission — there is no pre-submission draft record.
type Application struct {
	ID           string    `json:"applicationId" dynamodbav:"id"`
	EntryPath    EntryPath `json:"entryPath" dynamodbav:"entryPath"`
	PartnerID    string    `json:"partnerId,omitempty" dynamodbav:"partnerId,omitempty"`
	Applicant    Applicant `json:"applicant" dynamodbav:"applicant"`
	Offer        *Offer    `json:"offer,omitempty" dynamodbav:"offer,omitempty"`
	Status       Status    `json:"applicationStatus" dynamodbav:"status"`
	SubmittedAt  time.Time `json:"submittedAt" dynamodbav:"submittedAt"`
	LastActionAt time.Time `json:"lastActionAt" dynamodbav:"lastActionAt"`
}
