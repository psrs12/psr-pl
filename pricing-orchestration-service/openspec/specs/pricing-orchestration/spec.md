# Pricing Orchestration

Capability spec for `pricing-orchestration-service`: soft pull, pricing, hard pull, and decision, implemented as Step Functions-invoked Lambda/Fargate steps. Consumed here from the platform-wide `rebuild-platform-go` change; this file is this service's own copy of record going forward.

## Requirements

### Requirement: No standalone API
`pricing-orchestration-service` SHALL NOT expose an always-on REST API. Its soft-pull, pricing-calculation, hard-pull, and decision-routing steps SHALL be implemented as individual Lambda functions and, for the CPU-bound pricing calculation, a Fargate task, each invoked directly by a Step Functions state machine.

#### Scenario: No HTTP server exists for this capability
- **WHEN** the platform is deployed
- **THEN** there is no long-running Fargate/ECS service fronting pricing-orchestration's steps — only the Lambda functions, the Fargate task definition, and the state machine

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

### Requirement: Sequential soft pull, pricing, hard pull, decision
The state machine SHALL execute soft-pull, pricing-calculation, hard-pull, and decision-routing in that order, with each step's output added to (not replacing) the accumulated execution data.

#### Scenario: Each step's output is preserved
- **WHEN** the workflow completes
- **THEN** the execution's final data includes the soft-pull result, the priced offer, the hard-pull result, and the decision outcome, all present simultaneously

### Requirement: Decision routing updates the application record
On reaching a terminal decision (APPROVED, DECLINED, or REFERRED), the workflow SHALL update the linked application's status in `application-management-service` to match.

#### Scenario: Workflow reaches a decision
- **WHEN** decision-routing produces an outcome
- **THEN** the workflow calls `application-management-service`'s internal status-update endpoint with that outcome before completing
