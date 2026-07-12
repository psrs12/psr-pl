## Why

Corporate direction requires the personal loan acquisition platform to be rebuilt on this org's enterprise Go standards (see `CLAUDE.md`) — Go services, DynamoDB, Fargate/Step Functions/EventBridge — rather than the existing Java/Spring/Postgres/Kafka/Hexagonal implementation. The existing implementation (documented in the separate, untouched `psr-personal-loan` repository) remains the functional source of truth: this change preserves its service boundaries, domain behavior, and application lifecycle while replacing the technology stack and internal architecture style end to end.

## What Changes

- Introduce a five-service platform in `psr-pl`, mirroring the existing service boundaries with one deliberate deviation: `application-management-service`, `pricing-orchestration-service`, `offer-acceptance-service`, `document-service`, `compliance-orchestration-service`. `invitation-service` is not stood up as a standalone deployable — see the `invitation-service` decision below.
- Rebuild each service in Go using Handler → Service → Repository layering, packages organized by business capability, and small consumer-owned interfaces. **BREAKING**: drops Hexagonal ports-and-adapters as the internal architecture style.
- Replace Postgres-per-service with DynamoDB-per-service. **BREAKING**: relational/audit queries (e.g. compliance gate lookups) must be redesigned around access patterns and GSIs; no joins, no scans.
- Replace Spring Boot long-running JVM processes with Fargate/ECS containers per service; reserve Lambda for lightweight glue (validation, notifications, event handlers) per CLAUDE.md's Lambda-vs-Fargate guidance.
- Replace Kafka choreography with EventBridge (domain events) + Step Functions (central workflow orchestration). **BREAKING**: this is a behavioral shift from choreography (each service reacts independently) to orchestration (a Step Function centrally sequences the acquisition workflow, including the five compliance gates); event ownership and gate-sequencing rules need to be re-established under the new model.
- Define `application-management-service`'s acquisition-intake behavior in Go/DynamoDB terms:
  - Three application entry paths (invitation, direct, partner), with invitation-path name/address prefilled from the validated offer and non-editable; all other fields collected fresh on every path.
  - Email and phone are mandatory on all paths; a fixed set of applicant fields (employment, income, other income, rental expense, SSN, DOB, citizenship status) is required before submission regardless of path.
  - No draft/resume capability — the application is not persisted until submission (all-or-nothing, single session). No local cache of offer "in use" status is kept; checks against OfferLog are always live.
  - Application records support create, update, and retrieve only — no delete, to satisfy recordkeeping/adverse-action retention requirements.
  - Offer economics (term, amount, APR) are snapshotted onto the Application record at submission time as embedded attributes, since `pricing-orchestration-service` needs them later to honor the invited offer. Offer *status* (active/used) is never duplicated locally — OfferLog remains the sole source of truth, referenced from Application only by id.
  - A configurable inactivity-expiry sweep reverts an application's linked offer to `ACTIVE` in OfferLog when the applicant stalls after submission (exact lifecycle stage(s) monitored is an open design question, see design.md).
- **`invitation-service` is not a standalone service.** Its entire scope — retrieve/validate an offer from the external OfferLog system, and update the offer's status in OfferLog — is implemented as an internal adapter package inside `application-management-service`, since that is currently OfferLog's only consumer and the adapter owns no data of its own (OfferLog is the system of record for both offer content and offer status). If a second consumer of offer data emerges (e.g. `pricing-orchestration-service` needing direct offer status), extract this adapter into a standalone service at that point — the interface boundary (`OfferValidator` / `OfferStatusUpdater`) is already shaped for that extraction.
- UI layer is explicitly unchanged: `application-management-ui`, `pricing-offers-ui`, `document-management-ui` continue as currently documented (Vite/React/TS/Tailwind, web components, micro-frontend shell) and are out of scope for this change.

## Capabilities

### New Capabilities
- `application-intake`: Applicant-facing application creation across invitation, direct, and partner entry paths — field requirements, prefill/editability rules, and submission gating, owned by `application-management-service`.
- `offer-lifecycle`: Offer validation, status transitions (active/used), and inactivity-driven reversion. OfferLog (external) is the system of record; the requirements in this capability constrain the internal OfferLog-adapter package inside `application-management-service`, not a standalone service.
- `platform-service-architecture`: Cross-cutting architectural requirements applying to the five standalone services — Go layering, DynamoDB access-pattern design, Fargate/Lambda compute placement, and EventBridge/Step-Functions integration.

### Modified Capabilities
- None — no existing `openspec/specs/` capabilities exist in `psr-pl` yet; this change establishes the first ones.

## Impact

- **Affected systems**: Five backend services in `psr-pl` (new implementations); no changes to `psr-personal-loan` (reference-only) or to any UI codebase.
- **Affected infrastructure**: New DynamoDB tables (per service), Fargate/ECS services, EventBridge event bus and rules, Step Functions state machines, Lambda functions for lightweight tasks and the inactivity-expiry sweep. No separate infrastructure is provisioned for `invitation-service`, since it is not a standalone deployable.
- **Dependencies**: `application-management-service` depends on the external OfferLog system directly (via its internal adapter) and hands off to `pricing-orchestration-service` at submission. `compliance-orchestration-service`'s five gates (AML pre-screen, FCRA consent, adverse action, TILA disclosure audit, pre-funding AML re-check) depend on the new Step Functions orchestration model.
- **Out of scope**: Any change to `psr-personal-loan`; any change to the three frontend UIs.
