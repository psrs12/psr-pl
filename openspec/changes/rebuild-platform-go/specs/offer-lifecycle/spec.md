## ADDED Requirements

### Requirement: Offer is single-use
The external OfferLog system SHALL ensure an offer can be attached to, and consumed by, at most one submitted application.

#### Scenario: Second attempt to use an already-used offer
- **WHEN** a caller attempts to validate or consume an offer whose status is already used
- **THEN** OfferLog reports the offer as unavailable rather than returning valid offer details

### Requirement: Offer status is authoritative only in OfferLog
OfferLog SHALL be the sole system of record for an offer's status (active, used). No service in this platform SHALL persist a copy of offer status.

#### Scenario: Consuming service checks offer status
- **WHEN** any part of the platform needs to know whether an offer is active or used
- **THEN** it queries OfferLog directly (via the OfferLog adapter in `application-management-service`) rather than relying on a locally cached value

### Requirement: OfferLog adapter for status transitions
`application-management-service` SHALL implement an internal adapter (not a standalone service) that transitions an offer's status in OfferLog at defined points in the acquisition workflow, including at minimum: marking an offer used, and reverting an offer to active. This adapter SHALL own no offer data of its own — it is a pass-through to OfferLog.

#### Scenario: Mark offer used
- **WHEN** an invitation-path application is submitted
- **THEN** `application-management-service` calls OfferLog, via its internal adapter, to transition the offer's status to used

#### Scenario: Revert offer to active
- **WHEN** the inactivity-expiry sweep determines a submitted application has stalled
- **THEN** the sweep calls OfferLog, via the same adapter, to transition the linked offer's status back to active

#### Scenario: A second consumer needs offer data
- **WHEN** a service other than `application-management-service` needs to read or write offer data
- **THEN** the adapter is extracted into a standalone `invitation-service` rather than every consumer independently integrating with OfferLog

### Requirement: Inactivity-driven offer reversion
The system SHALL run a configurable, scheduled sweep that identifies submitted applications with no applicant action for a configurable number of days and reverts their linked offer to active status in OfferLog.

#### Scenario: Application stalls past the configurable threshold
- **WHEN** a submitted, offer-linked application has had no applicant action for the configured number of days
- **THEN** the sweep marks the application expired and reverts the linked offer to active in OfferLog

#### Scenario: Application progresses within the threshold
- **WHEN** a submitted, offer-linked application has recent applicant action within the configured threshold
- **THEN** the sweep does not alter the application or offer status
