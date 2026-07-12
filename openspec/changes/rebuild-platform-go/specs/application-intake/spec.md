## ADDED Requirements

### Requirement: Entry paths
The system SHALL support three application entry paths: invitation id, direct-to-portal, and partner-referred.

#### Scenario: Applicant arrives with an invitation id
- **WHEN** an applicant starts an application using an invitation id
- **THEN** the system validates the invitation against `invitation-service` before proceeding

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
- **WHEN** `invitation-service` reports the offer as invalid, expired, or already used
- **THEN** the system continues the application as a direct applicant, with name and address collected as editable fields

### Requirement: Mandatory contact fields on every path
The system SHALL require email address and phone number on all three entry paths.

#### Scenario: Submission blocked without contact fields
- **WHEN** an applicant attempts to submit without a phone number or email address
- **THEN** the system rejects the submission

### Requirement: Required fields before submission
The system SHALL require the following fields, regardless of entry path, before an application can be submitted: name, address, phone number, email address, employment details, annual income, other income, rental expense, SSN, date of birth, and citizenship status.

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
- **THEN** it queries `invitation-service` directly; it does not consult any locally cached record

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
The system SHALL NOT store or cache the offer's status (e.g., active, used) on the application record. `invitation-service` SHALL remain the sole source of truth for offer status at all times.

#### Scenario: Reading offer status
- **WHEN** any part of the system needs to know an offer's current status
- **THEN** it queries `invitation-service` directly rather than reading a locally stored status value

### Requirement: Offer status transition on submission
The system SHALL mark the linked offer as used in `invitation-service` when an invitation-path application is submitted.

#### Scenario: Successful submission marks the offer used
- **WHEN** an invitation-path application is submitted successfully
- **THEN** the system calls `invitation-service` to transition the offer's status to used

### Requirement: Offer details are not shown to the applicant
The system SHALL NOT display offer term, amount, or APR to the applicant during the intake flow.

#### Scenario: Applicant views the application form
- **WHEN** an applicant on the invitation path is filling out the application
- **THEN** the offer's term, amount, and APR are not rendered anywhere in the intake UI
