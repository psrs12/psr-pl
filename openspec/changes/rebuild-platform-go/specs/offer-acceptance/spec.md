## ADDED Requirements

### Requirement: Declarations must be listed before e-signature
`offer-acceptance-service` SHALL provide the applicant with the set of declarations they must acknowledge before e-signing, including which are required.

#### Scenario: Retrieving declarations
- **WHEN** an applicant requests the declarations for their application
- **THEN** the service returns the fixed set of declarations, each marked required or not

### Requirement: E-signature requires every required declaration
The service SHALL reject an e-signature attempt that does not include every required declaration id. Partial acceptance SHALL NOT be treated as a valid signature.

#### Scenario: Missing a required declaration
- **WHEN** an applicant attempts to e-sign without accepting all required declarations
- **THEN** the service rejects the request

#### Scenario: All required declarations accepted
- **WHEN** an applicant e-signs having accepted every required declaration
- **THEN** the service records the e-signature

### Requirement: E-signature is immutable and single-use
Once recorded, an e-signature record SHALL NOT be updated or deleted. A second e-signature attempt for the same application SHALL be rejected.

#### Scenario: Attempt to e-sign twice
- **WHEN** an application that has already been e-signed is e-signed again
- **THEN** the service rejects the second attempt

### Requirement: E-signature advances the application status
On successfully recording an e-signature, the service SHALL advance the linked application's status to DOCUMENTS_REQUIRED via `application-management-service`'s internal status-update endpoint.

#### Scenario: Successful e-signature
- **WHEN** an e-signature is successfully recorded
- **THEN** the service calls `application-management-service` to set the application's status to DOCUMENTS_REQUIRED
