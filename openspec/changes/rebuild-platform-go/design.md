## Context

`psr-personal-loan` (a separate, unmodified repository) documents a working personal loan acquisition platform: six Java/Spring Boot microservices, DDD + Hexagonal architecture, Postgres per service, Kafka choreography, a 15-state application lifecycle, and five compliance gates. That repository is the functional reference and stays untouched — it is someone's own architecture, not something this change alters.

Corporate direction requires a parallel, independent rebuild in `psr-pl` on this org's enterprise Go standards (`CLAUDE.md`): Go, DynamoDB, Fargate/Lambda, EventBridge/Step Functions, Handler→Service→Repository layering instead of Hexagonal. Service boundaries, applicant-facing behavior, and the lifecycle must be preserved; the internal architecture and infrastructure are being deliberately replaced.

This design covers the platform-wide architectural mapping and the one capability explored in depth so far, `application-management-service`'s intake flow. The other four standalone services keep their documented responsibilities but are not designed in detail here — that is follow-on work.

## Goals / Non-Goals

**Goals:**
- Map each of the five standalone existing services to a Go/DynamoDB/Fargate implementation with equivalent external behavior (APIs, lifecycle states, compliance gates).
- Replace Kafka choreography with EventBridge + Step Functions, explicitly designing how the five compliance gates get sequenced under central orchestration instead of independent reaction.
- Fully specify `application-management-service`'s intake behavior: entry paths, field/prefill rules, persistence model, and its interaction with the external OfferLog system (via an internal adapter) for offer validation and status transitions.

**Non-Goals:**
- Changing anything in `psr-personal-loan`.
- Changing the UI layer (`application-management-ui`, `pricing-offers-ui`, `document-management-ui`) — these are consumed as-is.
- Detailed DynamoDB table/GSI design for `pricing-orchestration-service`, `offer-acceptance-service`, `document-service`, or `compliance-orchestration-service` — captured as open questions/follow-on work, not designed here.
- Consolidating the five services into a single deployable. They remain five independently deployed Go/Fargate services; CLAUDE.md's `internal/` package-by-capability example happens to name-match the existing capabilities but is not a signal to merge them.
- Standing up `invitation-service` as a deployable — see the decision below.

## Decisions

**Five services, not six, not a monolith.** Each of `application-management-service`, `pricing-orchestration-service`, `offer-acceptance-service`, `document-service`, and `compliance-orchestration-service` becomes its own Go module/deployable on Fargate/ECS, each with Handler→Service→Repository layering and its own DynamoDB table(s). Rejected alternative: a single Go modular monolith using CLAUDE.md's `internal/<capability>` layout — tempting because the package names coincidentally match the service list, but rejected because it would merge independently-scaled, independently-owned deployables and contradicts "keep all microservices intact."

**`invitation-service` is not stood up as a standalone deployable.** Its documented responsibility — retrieve/validate an offer and update its status — is entirely a facade over the external OfferLog system: it owns no data of its own (OfferLog is system of record for both offer content and status), and `application-management-service` is currently its only consumer. Standing up a separate Fargate task, DynamoDB table, and deployment pipeline for a two-call proxy contradicts CLAUDE.md's "avoid unnecessary abstractions." Instead, this is implemented as an internal adapter package (`OfferValidator` / `OfferStatusUpdater`) inside `application-management-service`. Rejected alternative: build it standalone to match the reference platform's service list exactly — rejected because preserving service boundaries is valuable where a service owns data or has multiple consumers, neither of which holds here; matching the org chart isn't a reason on its own. Extraction trigger: the moment a second service needs to read or write offer data directly, pull this adapter out into its own service — the interface boundary is already shaped for that.

**EventBridge + Step Functions replace Kafka choreography.** Cross-service domain events (e.g., "application submitted", "decision made") move to EventBridge. The acquisition workflow's sequencing — today implicit in which events each service chooses to consume — becomes an explicit Step Functions state machine that orchestrates: submission → compliance Gate 1 (AML pre-screen) → soft pull → pricing → Gate 2 (FCRA consent) → hard pull → decision → (approval path: documents → Gate 4 TILA audit → e-sign → Gate 5 pre-funding AML → funding) or (decline path: Gate 3 adverse action). Rejected alternative: keep Kafka running alongside the AWS-native stack — rejected because CLAUDE.md's AWS service list doesn't include Kafka and running two messaging backbones adds operational surface for no clear benefit once the platform is Go/AWS-native anyway.

**DynamoDB per service, access-pattern-driven.** Each service owns its own table(s); no shared/global table across services (preserves the "per-service data ownership" property Postgres-per-service had). Rejected alternative: one shared single-table design across all six services — rejected because it would blur the service ownership boundaries CLAUDE.md and the existing platform both treat as inviolable.

**Fargate for the services themselves; Lambda for glue.** Each of the six services runs as a long-running Fargate/ECS task behind the API layer, since they are synchronous REST APIs equivalent to today's always-on Spring Boot services. Lambda is reserved for short-lived, event-triggered work: the inactivity-expiry sweep, notification sends, lightweight validation steps invoked from Step Functions.

**`application-management-service` persists Application only, never Offer status.**
- Application supports create, update, retrieve — no delete (recordkeeping/adverse-action retention).
- Nothing is persisted pre-submission: the intake flow is stateless/single-session until submit, so there is no draft state to clean up and no local record to check for "offer already in use" — that check always calls OfferLog live via the internal adapter.
- At submission, offer economics (term, amount, APR, and the offer/invitation/customer-reference id) are snapshotted as embedded attributes on the Application item — needed later by `pricing-orchestration-service` to honor the invited offer, and simple to model since nothing else independently updates those attributes.
- Offer *status* is deliberately never stored on Application, even as a cache — OfferLog is the sole source of truth, always queried live, to avoid the two-sources-of-truth drift that a cached status field would introduce (e.g., an expiry sweep flips the canonical status but a stale local copy doesn't). Application only carries a reference id. Status transitions are driven through a small consumer-owned interface (`OfferStatusUpdater`) called from `application-management-service` at defined trigger points: mark `USED` on submission; revert to `ACTIVE` on the inactivity-expiry sweep.

**Invitation-path identity binding.** Name and address are sourced from the validated offer and rendered non-editable specifically to bind the applicant to the invited identity — an editable name/address on the invitation path would defeat the purpose of requiring an invitation at all. This applies only to the invitation path; direct and partner paths collect and allow editing of every field.

**Offer-validation failure degrades gracefully.** An invalid/expired/already-used offer does not hard-stop the applicant — they continue on the direct-application path instead, re-entering any fields the offer would otherwise have prefilled.

## Risks / Trade-offs

- [Choreography → orchestration is a behavioral change, not a swap] Centralizing sequencing in Step Functions changes failure/retry semantics for the five compliance gates and puts sequencing ownership in one place instead of each gate enforcing its own trigger conditions. → Design the Step Function's gate sequencing explicitly as part of `compliance-orchestration-service`'s own design pass (flagged as open question below); do not assume today's Kafka consumer logic ports over unchanged.
- [DynamoDB access-pattern redesign for audit/compliance queries] Relational queries like "all applications currently in `COMPLIANCE_HOLD`" have no direct DynamoDB equivalent without deliberate GSI design. → Treat this as required design work for `compliance-orchestration-service`, not an implementation detail to improvise later.
- [Live OfferLog dependency during intake] Every offer-in-use check and status transition is a live call, with no local fallback/cache by design. → Acceptable given OfferLog is the sole source of truth by design; if this becomes an availability concern, revisit as a deliberate trade-off change, not a workaround bolted onto `application-management-service`.
- [Folding the OfferLog adapter into `application-management-service` couples its release cycle to a second consumer's needs] If `pricing-orchestration-service` later needs offer data, its access is mediated through `application-management-service` (or the adapter must be extracted first) rather than calling OfferLog directly. → Treat "second consumer appears" as the explicit trigger to extract the adapter into its own service before wiring that second consumer up, not something to defer indefinitely.
- [Embedded offer attributes vs. separate item] Snapshotting offer economics as embedded attributes assumes no future need to independently update just the offer portion of an Application item. → If that assumption breaks (e.g., pricing needs to write back adjusted terms), revisit splitting into a separate item under the same partition key.

## Migration Plan

Not applicable in the traditional sense — `psr-pl` is a new, independent codebase; there is no live system to migrate off of, and `psr-personal-loan` is not being decommissioned or cut over. Rollout is standard net-new service delivery (build, deploy, cut traffic to `psr-pl` when each service is ready), sequenced service-by-service starting with `application-management-service` since the other four depend on its behavior.

## Open Questions

- Which lifecycle stage(s) does the inactivity-expiry sweep actually watch (e.g. `DOCUMENTS_REQUIRED`, `OFFER_ACCEPTED` awaiting e-sign, both, others)? Not yet decided.
- Does withdrawal or decline (in addition to expiry) also revert the linked offer to `ACTIVE` in OfferLog, or only expiry?
- Does `compliance-orchestration-service`'s gate logic live inside the Step Function definition itself, or inside the service (invoked as a task from the Step Function)?
- GSI/access-pattern design for `compliance-orchestration-service`'s audit and status-lookup queries is unstarted.
- DynamoDB and event-schema design for `pricing-orchestration-service`, `offer-acceptance-service`, and `document-service` has not been explored yet.
- `pl-audit` (the immutable audit-event table CLAUDE.md's DynamoDB rules carve out) is not yet designed or implemented for `application-management-service` — what events it captures, write pattern, and retention are still open. Deliberately deferred, not an oversight.
