# Workflow Status

Capability spec for `workflow-status-service`: applicant-facing status/next-steps derived live from a Step Functions execution, with no state of its own. Consumed here from the platform-wide `rebuild-platform-go` change; this file is this service's own copy of record going forward.

## Requirements

### Requirement: Session-authenticated status retrieval
`workflow-status-service` SHALL require a valid session token, authenticated via `application-management-service`, before returning an application's workflow status.

#### Scenario: Missing or invalid session token
- **WHEN** a status request has no session token, or a token that does not validate for the requested application id
- **THEN** the service rejects the request

#### Scenario: Valid session token
- **WHEN** a status request includes a session token that validates for the requested application id
- **THEN** the service proceeds to derive and return the status

### Requirement: Status derived live from the workflow execution
`workflow-status-service` SHALL NOT persist its own copy of an application's status. Every status request SHALL derive the response from the linked Step Functions execution's current state, read live.

#### Scenario: Status request
- **WHEN** an authenticated status request is received
- **THEN** the service looks up the execution ARN from `application-management-service` and queries Step Functions directly, rather than reading any locally stored status value

### Requirement: Structured next-steps, not a bare status label
The status response SHALL include a structured list of next steps (action and description), not only a status label.

#### Scenario: Application is mid-workflow
- **WHEN** the linked execution is still running
- **THEN** the response includes a status reflecting the current step and a next-steps list describing what the applicant should expect

#### Scenario: Application has reached a decision
- **WHEN** the linked execution has completed with an APPROVED, DECLINED, or REFERRED outcome
- **THEN** the response's next-steps reflect that outcome (e.g. "select your offer" for APPROVED)

#### Scenario: Application is awaiting offer selection
- **WHEN** the linked execution is paused at the offer-selection step
- **THEN** the response's status is OFFER_PENDING

### Requirement: No duplicated execution or session state
`workflow-status-service` SHALL NOT independently store the applicationId-to-execution mapping or any session data. Both SHALL be looked up from `application-management-service` on every request.

#### Scenario: Service restart
- **WHEN** `workflow-status-service` restarts with no persistent storage of its own
- **THEN** it continues to serve correct status responses, since it holds no state that could be lost
