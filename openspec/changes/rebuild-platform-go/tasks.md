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

- [x] 5.1 Design the Step Functions state machine for the pricing portion of the acquisition workflow (submission → soft pull → pricing → hard pull → decision → update status). Compliance gates and the funding tail are not yet part of this state machine — see 5.2.
- [ ] 5.2 Resolve the open design question: whether `compliance-orchestration-service`'s gate logic lives in the state machine or inside the service as invoked tasks. Still open — not addressed by the pricing-orchestration work.
- [ ] 5.3 Define the EventBridge domain events each service publishes/consumes, replacing today's Kafka topics one-for-one where behavior must be preserved. Not started — the submission → pricing-workflow handoff uses a direct `StartExecution` call instead (see design.md decision), not EventBridge.

## 6. Deferred service design (tracked, not detailed in this change)

- [x] 6.1 ~~Design `pricing-orchestration-service`'s DynamoDB access patterns~~ — superseded: `pricing-orchestration-service` has no DynamoDB table at all (Step-Functions-native, no persistence of its own; see design.md decision). Its consumption of the offer economics on the Application record happens via the workflow input `application-management-service` passes at `StartExecution`, not a data lookup.
- [x] 6.2 Design `offer-acceptance-service`'s DynamoDB access patterns and e-sign integration — see section 11 below; built, not just designed.
- [ ] 6.3 Design `document-service`'s DynamoDB access patterns and document storage (S3) integration.
- [ ] 6.4 Design `compliance-orchestration-service`'s DynamoDB access patterns for the five compliance gates, including audit/status-lookup GSIs.

## 7. Verification

- [x] 7.1 Unit test `application-management-service`'s Service layer with no AWS/HTTP dependencies (mock the `OfferValidator`/`OfferStatusUpdater` adapter only).
- [x] 7.2 Repository integration tests against a local DynamoDB instance for `application-management-service` — done manually via DynamoDB Local + `curl`/`aws dynamodb get-item` (create, session issuance, no-delete); not yet automated as `go test` integration tests.
- [ ] 7.3 API test covering all three entry paths, including the offer-validation-failure-degrades-to-direct scenario. Partner path still untested (no UI support for it).
- [ ] 7.4 Workflow integration test for the inactivity-expiry sweep reverting an offer to active. Sweep itself is unimplemented (see 4.1-4.5).

## 8. pricing-orchestration-service

- [x] 8.1 Implement `internal/pricing`'s business logic (credit-pull simulation, pricing calculation, decision routing) as pure, unit-tested Go — no AWS dependency. Explicitly a placeholder for a real bureau/underwriting integration, not production pricing logic.
- [x] 8.2 Implement soft-pull, hard-pull, and decision-routing as individual Lambda functions (`cmd/soft-pull-lambda`, `cmd/hard-pull-lambda`, `cmd/decision-lambda`).
- [x] 8.3 Implement pricing-calculation as a Fargate task (`cmd/pricing-fargate`) using the `ecs:runTask.waitForTaskToken` pattern (`SendTaskSuccess`/`SendTaskFailure`), since `.sync` alone has no return-value channel.
- [x] 8.4 Implement `update-status-lambda`, which calls `application-management-service`'s internal status-update endpoint with the terminal decision.
- [x] 8.5 Write the Step Functions state machine definition (`statemachine/definition.asl.json`) chaining all five steps with `ResultPath`-based data accumulation.
- [x] 8.6 Verify the state machine registers and runs against Step Functions Local; verify `workflow-status-service`'s execution-history parsing against a real (if Lambda-less) execution. Full mocked-response control-flow testing (AWS's `SFN_MOCK_CONFIG` feature) was attempted but did not work in this sandbox — flagged in design.md as something to revisit, not silently dropped.
- [ ] 8.7 Deploy and dry-run against real (or better-emulated) Lambda/Fargate — not done; see design.md's risk note.

## 9. workflow-status-service

- [x] 9.1 Implement `ExecutionReader` (DescribeExecution/GetExecutionHistory) and verify it against a real Step Functions Local execution — proved the current-state-parsing logic matches real API response shapes, not just assumptions.
- [x] 9.2 Implement `ApplicationClient` calling `application-management-service`'s two new internal endpoints (session validate, execution lookup) — no duplicated session or execution-mapping state in this service.
- [x] 9.3 Implement running-state → status/next-steps and terminal-outcome → status/next-steps translation (`translate.go`), unit-tested with fakes for all four outcome branches (unauthorized, running, approved, failed).
- [x] 9.4 Wire `application-management-service`: add `GET /internal/applications/{id}/execution` and `POST /internal/sessions/validate`, and have `Submit()` call `StartExecution` and persist the `executionArn`.
- [x] 9.5 UI integration: `application-management-ui`'s `client.js` (`getApplication`) now calls `workflow-status-service` (`VITE_WORKFLOW_STATUS_API_URL`, new env var across `.env.local`/`.env.example`/`.env.production`), not `application-management-service`'s own status endpoint. Verified live: network log confirmed the status call actually hits `:8086` (workflow-status-service), and the response correctly rendered.
- [x] 9.6 Fix the `ERROR` status gap found while doing 9.5: `navigationConfig.js` had no entry for it, so a failed/timed-out/aborted execution fell through to the generic "Processing" spinner instead of a real error. Added an `error` static-block kind + `ErrorScreen.jsx`; verified with a route-mocked browser render and an added test.

## 10. pricing-orchestration-service: offer-selection API

- [x] 10.1 Insert `AwaitOfferSelection` (`lambda:invoke.waitForTaskToken`) between `PricingCalculation` and `HardPullRequest`, and a `ConsentGivenCheck` Choice state so a hard pull cannot happen without recorded consent (`ConsentDeclined` routes straight to a declined outcome).
- [x] 10.2 Implement `present-offer-lambda`, which stores the task token + priced offer via a new internal endpoint rather than resuming the execution itself.
- [x] 10.3 Implement the offer-selection API (`GET .../selected-offer`, `POST .../selected-offer/confirm`) with its own small DynamoDB table (`pl-pricing-offer`) — the one deliberate exception to "no standalone API," justified because this step now has two real consumers (the paused workflow, and `pricing-offers-ui`).
- [x] 10.4 Add session authentication to the offer-selection API. Found missing during live verification (not code review) — every other applicant-facing endpoint in the platform validates a session, this one initially didn't. Fixed with the same `SessionValidator` → `application-management-service` pattern `workflow-status-service` uses; verified 401/401/200 for no-token/wrong-token/real-token.
- [x] 10.5 Fix `Confirm`'s write ordering: resume the workflow *before* persisting `CONFIRMED`, not after. Found via live testing — the original ordering left a record permanently stuck as confirmed-but-not-resumed when the resume call failed. Regression-tested with a fake resumer that fails then succeeds on retry.
- [x] 10.6 Add `AwaitOfferSelection`/`ConsentGivenCheck`/`ConsentDeclined` to `workflow-status-service`'s state-to-status map (`OFFER_PENDING`, `DECISION_PENDING`).
- [ ] 10.7 Real end-to-end pause-and-resume through a genuine (not fabricated) task token — not verified; requires the still-unresolved Fargate execution gap (see design.md) to produce a real paused execution first.

## 11. offer-acceptance-service

- [x] 11.1 Implement declarations (fixed placeholder set) and e-signature capture (`Repository`: create/get only, no update/delete).
- [x] 11.2 Enforce all-required-declarations-accepted before an e-signature is recorded; reject partial acceptance and double-signing. Unit-tested for both.
- [x] 11.3 On successful e-signature, call `application-management-service`'s internal status-update endpoint to advance status to `DOCUMENTS_REQUIRED`.
- [ ] 11.4 Live end-to-end verification (submit → approve → e-sign → confirm status advances) — not done this session; only unit-tested with fakes.

## 12. pricing-offers-ui

- [x] 12.1 Implement `<pricing-offer-selector>` as a hand-rolled custom element (no React — see design.md) built via Vite library mode to a single IIFE, matching how `application-management-ui` already expected to load it.
- [x] 12.2 Client-side hard-pull-consent enforcement (confirm button no-ops with a visible error until the box is checked) in addition to the server-side `ConsentGivenCheck` gate.
- [x] 12.3 Verified live in a real browser against the real offer-selection API: offer loads with real fetched data, consent-required validation blocks submission, and a genuine confirm failure (fabricated task token) is surfaced as a visible error rather than hanging.
