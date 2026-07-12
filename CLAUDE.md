# CLAUDE.md

# Enterprise Go Architecture Standards

This document defines the architectural standards, coding conventions, and implementation guidelines for this project. Claude should always follow these standards when generating code, designing APIs, creating packages, or proposing architectural changes.

---

# Architecture Principles

The application follows a lightweight, enterprise-grade Go architecture based on business capabilities.

## Primary Goals

- Keep the architecture simple.
- Prefer readability over cleverness.
- Avoid unnecessary abstractions.
- Use explicit dependencies.
- Follow Go idioms.
- Business logic must remain independent of infrastructure.

---
**Techinical Stack**
- **Languages**: Go (see go.mod for the version currenlty in use)
- **Frameworks**: Gin web framework, AWS Lambda (serverless)
- **Database**: DynamoDB(pl-application,pl-audit)
- **Architecture**: Service-oriented with handler -> service-> persitance layer
- **Deployment**: AWS Fargate via Jenkins(Bogiefile)


# Application Architecture

The application follows the following flow:

```
HTTP / REST API

        │

     Handler

        │

     Service

        │

   persitance

        │
    
    Integration 

 
```

Responsibilities:

## Handler

Responsible for:

- HTTP request parsing
- Request validation
- Authentication/Authorization
- Response generation
- Error mapping
- Calling services

Handlers should NEVER contain business logic.

---

## Service

Responsible for:

- Business rules
- Workflow orchestration
- Calling repositories
- Calling external services
- Publishing events
- Domain validation

Services should never know about HTTP.

---

## Repository

Responsible for:

- Reading from DynamoDB
- Writing to DynamoDB
- Query operations
- Persistence mapping

Repositories should never contain business logic.

---

# Package Organization

Packages are organized by business capability.

Example:

```
internal/

    application/

        handler.go
        service.go
        repository.go
        model.go
        validator.go

    customer/

        handler.go
        service.go
        repository.go
        model.go

    offer/

    pricing/

    documents/

    workflow/
```

DO NOT organize packages like:

```
controllers/
services/
repositories/
models/
```

Business capabilities own their code.

---

# Dependency Rules

Dependencies always flow downward.

```
Handler

↓

Service

↓

Repository

↓

Infrastructure
```

Business logic must never depend on:

- HTTP framework
- DynamoDB SDK
- AWS SDK
- Lambda
- Step Functions
- Fargate

Business logic should be testable without AWS.

---

# Dependency Injection

Use constructor injection only.

Example:

```go
service := NewApplicationService(
    repository,
    workflowClient,
    logger,
)
```

Do not use:

- Global variables
- Service locators
- Singletons

---

# Context

Every public function accepts context.Context.

Example:

```go
func (s *ApplicationService) CreateApplication(
    ctx context.Context,
    request CreateApplicationRequest,
) error
```

Never ignore context.

---

# Interfaces

Interfaces belong to consumers.

Keep interfaces small.

Good:

```go
type ApplicationReader interface {
    GetByID(ctx context.Context, id string) (*Application, error)
}

type ApplicationWriter interface {
    Save(ctx context.Context, app *Application) error
}
```

Avoid large interfaces.

---

# Error Handling

Always return errors.

Wrap errors.

Example:

```go
return fmt.Errorf("saving application: %w", err)
```

Never panic except during startup.

---

# Logging

Use structured logging.

Every log entry should include:

- requestId
- applicationId
- customerId (if available)
- workflowId (if available)

Never log:

- SSN
- PII
- Authentication tokens
- Secrets

---

# AWS Architecture

The application is deployed on AWS.

Primary AWS services:

- API Gateway
- Lambda
- Step Functions
- Fargate
- DynamoDB
- CloudWatch
- IAM
- EventBridge
- S3 (documents if applicable)
- Secrets Manager

---

# Workflow Architecture

Business workflows are orchestrated using AWS Step Functions.

Rules:

- Step Functions own workflow orchestration.
- Services initiate workflows.
- Business logic remains inside services and workflow tasks.
- Long-running processes should never be handled synchronously.

Example:

```
Create Application

↓

Persist Application

↓

Start Step Function

↓

Workflow

    ↓

 Lambda

    ↓

 Lambda

    ↓

 Fargate

    ↓

 Lambda

↓

Update Application Status
```

---

# Lambda Usage

Use Lambda for:

- Lightweight business tasks
- Integrations
- Validation
- Notifications
- Event processing

Avoid:

- Long-running computation
- Large file processing
- Stateful workloads

---

# Fargate Usage

Use Fargate for:

- Long-running processes
- CPU-intensive tasks
- Large document processing
- ML inference
- Batch workloads
- Complex business processing

Never use Lambda for workloads better suited to containers.

---

# DynamoDB

DynamoDB is the system of record.

Rules:

- Single-table design is preferred unless requirements justify otherwise.
- Access patterns drive schema design.
- Avoid scans.
- Use Query operations.
- Design GSIs intentionally.
- Use optimistic locking where appropriate.
- Store immutable audit events separately when needed.

Table naming:

- `pl-application` — operational application/domain data.
- `pl-audit` — immutable audit events, written but never updated in place.

Repositories encapsulate all DynamoDB access.

Business logic must never construct DynamoDB expressions.

---

# API Design

REST APIs should follow:

```
POST    /applications

GET     /applications/{id}

PUT     /applications/{id}

DELETE  /applications/{id}
```

Use nouns.

Avoid verbs in URLs.

---

# Validation

Validation occurs at the handler.

Business validation occurs inside services.

Repository validation is limited to persistence constraints.

---

# Models

Separate models by responsibility.

```
HTTP Request

↓

DTO

↓

Domain Model

↓

Persistence Model
```

Do not reuse HTTP models inside repositories.

---

# Configuration

Configuration comes from:

- Environment Variables
- AWS Parameter Store
- AWS Secrets Manager

Never hardcode:

- URLs
- Credentials
- ARNs
- Table names

---

# Testing

Preferred testing pyramid:

- Unit Tests
- Repository Integration Tests
- API Tests
- Workflow Integration Tests

Mock only external dependencies.

Do not mock business logic.

---

# Observability

Every service must provide:

- Structured Logging
- CloudWatch Logs
- OpenTelemetry Tracing
- Metrics
- Health Endpoint
- Readiness Endpoint

---

# Security

Never:

- Log secrets
- Log JWTs
- Log OAuth tokens
- Log customer PII

Always use IAM roles.

Never embed credentials.

---

# Project Structure

```
cmd/

internal/

    application/

    customer/

    offer/

    pricing/

    workflow/

pkg/

configs/

deploy/

docs/

scripts/

test/

go.mod
```

---

# Coding Guidelines

Prefer:

- Small functions
- Explicit code
- Constructor injection
- Composition
- Small interfaces
- Context propagation
- Error wrapping

Avoid:

- Reflection
- Overengineering
- Global state
- Hidden dependencies
- Large utility packages
- Clever abstractions

---

# AI Code Generation Rules

When generating code:

1. Always generate business-oriented packages.
2. Follow Handler → Service → Repository architecture.
3. Keep handlers thin.
4. Keep services focused on business logic.
5. Keep repositories focused on persistence.
6. Generate constructor injection.
7. Accept context.Context in all public methods.
8. Wrap errors with context.
9. Prefer Go standard library when possible.
10. Generate production-quality, idiomatic Go code.
11. Design DynamoDB access around access patterns, not relational thinking.
12. Use Step Functions for orchestration, Lambda for short-lived tasks, and Fargate for long-running or compute-intensive workloads.
13. Keep AWS-specific code isolated in infrastructure components.
14. Favor maintainability, readability, and simplicity over unnecessary abstraction.

Claude should always optimize for enterprise maintainability, scalability, observability, and testability while adhering to idiomatic Go practices.