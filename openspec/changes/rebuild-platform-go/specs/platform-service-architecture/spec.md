## ADDED Requirements

### Requirement: Service boundaries preserved
The platform SHALL consist of five independently deployed Go services, matching the existing documented service boundaries with one deliberate exception: `application-management-service`, `pricing-orchestration-service`, `offer-acceptance-service`, `document-service`, and `compliance-orchestration-service`. The platform SHALL NOT consolidate these into a single deployable.

#### Scenario: Deploying a single service
- **WHEN** a change is made to one service's business logic
- **THEN** it can be built, tested, and deployed independently of the other four services

### Requirement: Standalone services own their data or have multiple consumers
A capability is only stood up as its own deployable service if it owns persistent data of its own, or is consumed by more than one other service. A capability that is a stateless facade over an external system, with a single internal consumer, SHALL instead be implemented as an internal adapter package inside its sole consumer.

#### Scenario: invitation-service is not standalone
- **WHEN** evaluating whether to deploy `invitation-service` as its own service
- **THEN** it is implemented instead as an internal adapter inside `application-management-service`, since it owns no data (OfferLog is system of record) and has a single consumer

#### Scenario: A second consumer appears
- **WHEN** a second service needs to read or write data currently reached only through an internal adapter
- **THEN** the adapter is extracted into its own standalone service before that second integration is built

### Requirement: Layered architecture per service
Each service SHALL follow Handler → Service → Repository layering, organized into packages by business capability, using constructor injection and small consumer-owned interfaces.

#### Scenario: Business logic has no HTTP or AWS SDK dependency
- **WHEN** a service's Service-layer code is unit tested
- **THEN** it runs without any HTTP framework, DynamoDB SDK, or AWS SDK dependency

### Requirement: DynamoDB per service
Each service SHALL own its own DynamoDB table(s), designed around that service's access patterns. No table SHALL be shared across services.

#### Scenario: Cross-service data access
- **WHEN** one service needs data owned by another service
- **THEN** it calls that service's API rather than reading its DynamoDB table directly

### Requirement: Compute placement
Each of the five standalone services SHALL run as a long-running Fargate/ECS task for its synchronous REST API. Lambda SHALL be used only for short-lived, event-triggered work (e.g., notifications, validation steps, scheduled sweeps), not for hosting a service's primary API.

#### Scenario: Primary REST API compute
- **WHEN** a service's REST API is deployed
- **THEN** it runs as a Fargate/ECS task, not as an API-Gateway-fronted Lambda function

#### Scenario: Scheduled inactivity sweep
- **WHEN** the inactivity-expiry sweep runs
- **THEN** it executes as a Lambda function triggered on a schedule, not as part of a long-running service task

### Requirement: Cross-service integration via EventBridge and Step Functions
Domain events between services SHALL be carried on EventBridge. Multi-step business workflows spanning services, including the acquisition workflow's compliance gates, SHALL be centrally orchestrated by Step Functions rather than left to independent per-service event reaction.

#### Scenario: Domain event published
- **WHEN** a service completes a state transition another service needs to know about
- **THEN** it publishes a domain event to EventBridge rather than calling the other service synchronously or using Kafka

#### Scenario: Acquisition workflow sequencing
- **WHEN** an application moves from submission through decisioning
- **THEN** a Step Functions state machine drives the sequence of steps, including invoking `compliance-orchestration-service`'s gates in order
