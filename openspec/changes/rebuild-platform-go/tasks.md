## 1. Platform scaffolding

- [ ] 1.1 Create the five Go service modules (`application-management-service`, `pricing-orchestration-service`, `offer-acceptance-service`, `document-service`, `compliance-orchestration-service`), each with Handler → Service → Repository package layout per CLAUDE.md. (`invitation-service` is not a separate module — see group 2.)
- [ ] 1.2 Define shared conventions used across services: structured logging fields (requestId, applicationId, customerId, workflowId), error-wrapping style, health/readiness endpoint pattern.
- [ ] 1.3 Provision one DynamoDB table per service (starting with `application-management-service`; the remaining three tables can follow once those services get their own detailed design). No table is provisioned for `invitation-service`, since it is not standalone and owns no data.
- [ ] 1.4 Provision Fargate/ECS task definitions for each service; confirm API Gateway/ALB routing.
- [ ] 1.5 Provision the EventBridge event bus and initial domain event schema registry.

## 2. OfferLog adapter (internal to application-management-service, not a standalone service)

- [x] 2.1 Implement offer retrieval/validation by invitation id, returning offer term, amount, APR, customer reference id, name, and address.
- [x] 2.2 Offer status query (active/used) is folded into 2.1 — every validation call reflects OfferLog's current status; there is no separate cached read path since none is kept.
- [x] 2.3 Implement the offer status update interface: mark used, revert to active.
- [ ] 2.4 Single-writer-safety for "an offer can only be marked used once" is OfferLog's own responsibility as the external system of record — not implemented in this platform. Confirm OfferLog actually provides this guarantee before relying on it.

## 3. application-management-service: intake

- [x] 3.1 Implement the three entry paths (invitation, direct, partner) in the Handler/Service layers.
- [x] 3.2 Implement invitation-path prefill: fetch offer via the internal OfferLog adapter, render name/address as non-editable.
- [x] 3.3 Implement graceful degradation to direct-path behavior when offer validation fails (invalid, expired, already used).
- [x] 3.4 Implement field validation for the full required-field set across all paths, with phone and email always mandatory. Superseded by the real `application-management-ui` contract: `otherIncome`/`rentalExpense` were replaced by a single `monthlyObligations` field, and `requestedAmount`/`requestedTermMonths`/`loanPurpose` (applicant-chosen, independent of any offer ceiling) were added — see the service's own spec.md for the current field list.
- [x] 3.10 Integrate with the existing `application-management-ui` (brought into the repo): reconciled routes/DTOs to its real contract (`/api/v1/application-management/...`, `X-Channel-ID` header, flat address fields on submit, nested address on invitation-validate), added a citizenship field to the UI (previously hardcoded), fixed a UI bug where `monthlyObligations` was collected but never sent, and implemented the self-service login/session capability (`POST /applications/login`, session-token auth on `GET /applications/{id}/status`) the UI already depended on.
- [x] 3.5 Implement submission as the sole persistence trigger — no draft/create-before-submit path.
- [x] 3.6 Implement the Application repository: create, update, retrieve only (no delete operation exposed at any layer).
- [x] 3.7 On submission of an invitation-path application, snapshot offer term/amount/APR and the offer reference id onto the Application record, and call OfferLog (via the internal adapter) to mark the offer used.
- [x] 3.8 Ensure the Service layer never persists or caches offer status locally; all status reads go directly to OfferLog via the internal adapter.
- [x] 3.9 Confirm no offer economics (term, amount, APR) are rendered in any intake API response consumed by the UI.

## 4. Inactivity-expiry sweep

- [ ] 4.1 Resolve the open design question: which lifecycle stage(s) the sweep monitors (design.md open question).
- [ ] 4.2 Resolve the open design question: whether withdrawal/decline also reverts the offer, or only expiry.
- [ ] 4.3 Implement the scheduled Lambda that identifies stalled, offer-linked applications past the configurable threshold.
- [ ] 4.4 Implement the sweep's call into OfferLog (via the internal adapter) to revert the offer to active, and the corresponding Application status update to expired.
- [ ] 4.5 Make the inactivity threshold configurable (Parameter Store), not hardcoded.

## 5. Cross-service orchestration

- [ ] 5.1 Design the Step Functions state machine for the acquisition workflow (submission → compliance gates → pricing → decision → funding), referencing the lifecycle in proposal.md.
- [ ] 5.2 Resolve the open design question: whether `compliance-orchestration-service`'s gate logic lives in the state machine or inside the service as invoked tasks.
- [ ] 5.3 Define the EventBridge domain events each service publishes/consumes, replacing today's Kafka topics one-for-one where behavior must be preserved.

## 6. Deferred service design (tracked, not detailed in this change)

- [ ] 6.1 Design `pricing-orchestration-service`'s DynamoDB access patterns and its consumption of the offer economics snapshotted on Application.
- [ ] 6.2 Design `offer-acceptance-service`'s DynamoDB access patterns and e-sign integration.
- [ ] 6.3 Design `document-service`'s DynamoDB access patterns and document storage (S3) integration.
- [ ] 6.4 Design `compliance-orchestration-service`'s DynamoDB access patterns for the five compliance gates, including audit/status-lookup GSIs.

## 7. Verification

- [ ] 7.1 Unit test `application-management-service`'s Service layer with no AWS/HTTP dependencies (mock the `OfferValidator`/`OfferStatusUpdater` adapter only).
- [ ] 7.2 Repository integration tests against a local DynamoDB instance for `application-management-service`.
- [ ] 7.3 API test covering all three entry paths, including the offer-validation-failure-degrades-to-direct scenario.
- [ ] 7.4 Workflow integration test for the inactivity-expiry sweep reverting an offer to active.
