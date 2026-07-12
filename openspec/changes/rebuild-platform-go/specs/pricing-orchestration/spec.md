## ADDED Requirements

### Requirement: No standalone API for automated steps
`pricing-orchestration-service` SHALL NOT expose an always-on REST API for its soft-pull, pricing-calculation, hard-pull, and decision-routing steps. These SHALL be implemented as individual Lambda functions and, for the CPU-bound pricing calculation, a Fargate task, each invoked directly by a Step Functions state machine. The one exception is the offer-selection API — see the requirements below — which exists because that step is not automated.

#### Scenario: No HTTP server fronts the automated steps
- **WHEN** the platform is deployed
- **THEN** there is no long-running Fargate/ECS service fronting soft-pull, pricing-calculation, hard-pull, or decision-routing — only the Lambda functions, the Fargate task definition, and the state machine

### Requirement: Workflow started on application submission
`application-management-service` SHALL start the pricing-orchestration state machine execution immediately after persisting a submitted application, and SHALL persist the resulting execution identifier on the Application record.

#### Scenario: Submission starts the workflow
- **WHEN** an application is successfully persisted at submission
- **THEN** a new Step Functions execution is started for that application, and its execution ARN is stored on the Application record

### Requirement: Workflow input excludes applicant PII
The payload passed to the pricing-orchestration state machine SHALL contain only the financial figures needed for pricing (requested amount, requested term, annual income, monthly obligations) and the application id. It SHALL NOT contain name, SSN, date of birth, or address.

#### Scenario: Starting an execution
- **WHEN** `application-management-service` starts a pricing-orchestration execution
- **THEN** the execution input contains no applicant PII fields

### Requirement: Sequential soft pull, pricing, offer selection, hard pull, decision
The state machine SHALL execute soft-pull, pricing-calculation, offer-selection, hard-pull, and decision-routing in that order, with each step's output added to (not replacing) the accumulated execution data.

#### Scenario: Each step's output is preserved
- **WHEN** the workflow completes
- **THEN** the execution's final data includes the soft-pull result, the priced offer, the offer-selection outcome, the hard-pull result, and the decision outcome, all present simultaneously

### Requirement: Workflow pauses for applicant offer selection and hard-pull consent
After pricing and before any hard pull, the state machine SHALL pause and wait for the applicant to select the offer and give explicit hard-pull consent. It SHALL NOT proceed to a hard pull automatically.

#### Scenario: Pricing completes
- **WHEN** pricing-calculation produces a priced offer
- **THEN** the workflow pauses (does not invoke hard-pull) until the applicant's selection is received

### Requirement: No hard pull without explicit consent
The state machine SHALL NOT invoke the hard-pull step unless the applicant's response includes explicit, affirmative hard-pull consent. Declining consent (or declining the offer) SHALL route directly to a declined outcome without a hard pull.

#### Scenario: Consent given
- **WHEN** the applicant confirms the offer with consent
- **THEN** the workflow proceeds to hard-pull

#### Scenario: Consent withheld
- **WHEN** the applicant declines consent, or declines the offer entirely
- **THEN** the workflow routes to a declined outcome and updates the application status accordingly, without ever invoking hard-pull

### Requirement: Offer-selection API is session-authenticated
`pricing-orchestration-service`'s offer-selection endpoints (retrieve the presented offer, confirm selection and consent) SHALL require a valid session token, authenticated via `application-management-service`, scoped to the requesting application id.

#### Scenario: Missing or invalid session token
- **WHEN** a request to the offer-selection API has no session token, or a token that does not validate for the requested application id
- **THEN** the request is rejected

### Requirement: Offer confirmation is retryable on failure
If resuming the paused workflow execution fails, the offer-selection record SHALL remain in a state that allows the applicant to retry — it SHALL NOT be marked confirmed unless the resume genuinely succeeded.

#### Scenario: Workflow resume fails
- **WHEN** an applicant confirms their offer but the underlying workflow resume call fails
- **THEN** the offer-selection record stays pending selection, and a subsequent confirmation attempt is not rejected as already-confirmed

### Requirement: Decision routing updates the application record
On reaching a terminal decision (APPROVED, DECLINED, or REFERRED), the workflow SHALL update the linked application's status in `application-management-service` to match.

#### Scenario: Workflow reaches a decision
- **WHEN** decision-routing (or the consent-declined path) produces an outcome
- **THEN** the workflow calls `application-management-service`'s internal status-update endpoint with that outcome before completing
