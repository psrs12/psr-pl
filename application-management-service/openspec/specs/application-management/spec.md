# Application Management

Capability spec for `application-management-service`: applicant-facing intake across all entry paths, and the Application record this service persists. Consumed here from the platform-wide `rebuild-platform-go` change; this file is this service's own copy of record going forward.

Offer retrieval and offer *status* (active/used) are owned entirely by the external OfferLog system, reached through an internal adapter package (`internal/application/offerclient.go`) rather than a separate `invitation-service` — see the `offer-lifecycle` capability for why. This service only ever holds a reference to an offer, never its status.

## Requirements

### Requirement: Entry paths
The system SHALL support three application entry paths: invitation id, direct-to-portal, and partner-referred.

#### Scenario: Applicant arrives with an invitation id
- **WHEN** an applicant starts an application using an invitation id
- **THEN** the system validates the invitation against OfferLog, via the internal adapter, before proceeding

#### Scenario: Applicant arrives directly
- **WHEN** an applicant starts an application with no invitation id and no partner reference
- **THEN** the system starts a blank application with no prefilled fields

#### Scenario: Applicant arrives through a partner
- **WHEN** an applicant starts an application via a partner referral
- **THEN** the system starts the application attributed to that partner with no prefilled fields

### Requirement: Invitation-path identity binding
For the invitation entry path, the system SHALL prefill name and address from the validated offer and SHALL NOT allow the applicant to edit them.

#### Scenario: Name and address are locked on the invitation path
- **WHEN** an applicant has a validly-resolved invitation offer
- **THEN** the application form displays the offer's name and address as read-only fields

### Requirement: Offer validation failure degrades to direct application
The system SHALL allow an applicant whose invitation offer fails validation to continue as a direct applicant rather than terminating the flow.

#### Scenario: Invitation id does not resolve to a valid, unused, unexpired offer
- **WHEN** OfferLog reports the offer as invalid, expired, or already used
- **THEN** the system continues the application as a direct applicant, with name and address collected as editable fields

### Requirement: Mandatory contact fields on every path
The system SHALL require email address and phone number on all three entry paths.

#### Scenario: Submission blocked without contact fields
- **WHEN** an applicant attempts to submit without a phone number or email address
- **THEN** the system rejects the submission

### Requirement: Required fields before submission
The system SHALL require the following fields, regardless of entry path, before an application can be submitted: name, address, phone number, email address, employment status, annual income, monthly obligations, SSN, date of birth, citizenship status, requested loan amount, requested loan term, and loan purpose.

### Requirement: Applicant-specified loan parameters
The system SHALL allow the applicant to specify their own requested loan amount, term, and purpose on every entry path, independent of any offer's pre-approved ceiling. These are persisted on the Application record distinctly from the offer economics snapshot.

#### Scenario: Invitation-path applicant requests different terms than the offer
- **WHEN** an applicant on the invitation path submits a requested amount or term that differs from the linked offer's amount or term
- **THEN** the system persists both the applicant's requested amount/term and the offer's own amount/term without reconciling them

#### Scenario: Submission blocked with incomplete required fields
- **WHEN** an applicant attempts to submit with any required field missing
- **THEN** the system rejects the submission and identifies the missing fields

#### Scenario: All required fields present
- **WHEN** an applicant has supplied every required field, including a resolved invitation offer's prefilled fields where applicable
- **THEN** the system allows submission

### Requirement: No persistence before submission
The system SHALL NOT persist any application data before submission. The intake flow SHALL be single-session with no draft-resume capability.

#### Scenario: Applicant abandons the flow before submitting
- **WHEN** an applicant leaves the application flow without submitting
- **THEN** no application record exists in the system for that attempt

#### Scenario: Offer-in-use check during intake
- **WHEN** the system needs to determine whether an invitation offer is already in use
- **THEN** it queries OfferLog directly via the internal adapter; it does not consult any locally cached record

### Requirement: Application persistence is create/update/retrieve only
The system SHALL support creating, updating, and retrieving application records. The system SHALL NOT support deleting an application record once it has been persisted (i.e., once submitted).

#### Scenario: Attempt to delete a submitted application
- **WHEN** any caller attempts to delete a persisted application record
- **THEN** the system rejects the operation

### Requirement: Offer economics snapshot at submission
At submission, for applications on the invitation path, the system SHALL persist the offer's term, amount, and APR as attributes on the application record, along with a reference to the offer (offer id / invitation id / customer reference id).

#### Scenario: Invitation-path application is submitted
- **WHEN** an applicant on the invitation path submits their application
- **THEN** the persisted application record includes the offer term, amount, APR, and offer reference id

#### Scenario: Direct or partner-path application is submitted
- **WHEN** an applicant on the direct or partner path submits their application
- **THEN** the persisted application record includes no offer economics or offer reference, since no offer exists

### Requirement: Offer status is never persisted locally
The system SHALL NOT store or cache the offer's status (e.g., active, used) on the application record. OfferLog SHALL remain the sole source of truth for offer status at all times.

#### Scenario: Reading offer status
- **WHEN** any part of the system needs to know an offer's current status
- **THEN** it queries OfferLog directly via the internal adapter rather than reading a locally stored status value

### Requirement: Offer status transition on submission
The system SHALL mark the linked offer as used in OfferLog when an invitation-path application is submitted.

#### Scenario: Successful submission marks the offer used
- **WHEN** an invitation-path application is submitted successfully
- **THEN** the system calls OfferLog, via the internal adapter, to transition the offer's status to used

### Requirement: Offer details are not shown to the applicant
The system SHALL NOT display offer term, amount, or APR to the applicant during the intake flow.

#### Scenario: Applicant views the application form
- **WHEN** an applicant on the invitation path is filling out the application
- **THEN** the offer's term, amount, and APR are not rendered anywhere in the intake UI

### Requirement: Self-service login after submission
The system SHALL let an applicant re-access their submitted application by verifying application id, the last 4 digits of SSN, and date of birth, issuing a short-lived session token on success.

#### Scenario: Successful login
- **WHEN** an applicant submits their application id, last 4 SSN digits, and date of birth, and all three match a persisted application
- **THEN** the system issues a session token valid for a limited time (30 minutes)

#### Scenario: Failed login
- **WHEN** any of application id, last 4 SSN digits, or date of birth do not match a persisted application
- **THEN** the system rejects the login attempt without indicating which field was wrong

### Requirement: Session-authenticated status retrieval
The system SHALL require a valid, non-expired session token bound to the requested application id before returning that application's status.

#### Scenario: Valid session
- **WHEN** a request to retrieve an application's status includes a valid session token for that application id
- **THEN** the system returns the application's current status

#### Scenario: Missing or expired session
- **WHEN** a request to retrieve an application's status has no session token, an expired token, or a token bound to a different application id
- **THEN** the system rejects the request so the caller can re-authenticate
