// Navigation config keys match services/application-management-service's
// ApplicationStatus enum exactly (not the narrative names in
// docs/architecture/002-application-state-machine.md, which is missing
// READY_FOR_SUBMISSION and CONSENT_CAPTURED and documents STARTED, a value
// that does not exist in the backend enum). COMPLIANCE_HOLD is intentionally
// absent: it is not a real backend status today (see reconciliation notes in
// openspec/changes/configurable-navigation-flow/design.md).

/**
 * @typedef {
 *   | { kind: 'spinner', label: string, description: string, group: 'processing' | 'under-review' }
 *   | { kind: 'web-component', tag: string, scriptUrlKey: string, apiBaseUrlKey: string, attributes: string[], containerWidth: number, completionEvent?: string }
 *   | { kind: 'static-block', block: 'declined' | 'cancelled-expired' | 'post-acceptance', label?: string, message?: string }
 * } ScreenDescriptor
 */

const PROCESSING_SPINNER = (label, description) => ({
  kind: 'spinner',
  label,
  description,
  group: 'processing',
});

const UNDER_REVIEW_SPINNER = (label, description) => ({
  kind: 'spinner',
  label,
  description,
  group: 'under-review',
});

const POST_ACCEPTANCE_BLOCK = (label, message) => ({
  kind: 'static-block',
  block: 'post-acceptance',
  label,
  message,
});

/** @type {Record<string, ScreenDescriptor>} */
export const NAVIGATION_CONFIG = {
  CREATED: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  IN_PROGRESS: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  READY_FOR_SUBMISSION: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  SUBMITTED: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  PROCESSING: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  SOFT_PULL_PENDING: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  PRICING_PENDING: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  CONSENT_CAPTURED: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  HARD_PULL_PENDING: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),
  DECISION_PENDING: PROCESSING_SPINNER('Application Under Review', 'Your application is being processed. This usually takes a few minutes.\nThis page will update automatically.'),

  OFFER_PENDING: {
    kind: 'web-component',
    tag: 'pricing-offer-selector',
    scriptUrlKey: 'pricingOffersUiJs',
    apiBaseUrlKey: 'pricing',
    attributes: ['application-id', 'api-base-url', 'session-token'],
    containerWidth: 760,
    completionEvent: 'offer-confirmed',
  },

  APPROVED: {
    kind: 'web-component',
    tag: 'offer-acceptance-flow',
    scriptUrlKey: 'offerAcceptanceUiJs',
    apiBaseUrlKey: 'offerAcceptance',
    attributes: ['application-id', 'api-base-url', 'session-token', 'pricing-api-base-url'],
    containerWidth: 760,
    completionEvent: 'offer-accepted',
  },

  DOCUMENTS_REQUIRED: {
    kind: 'web-component',
    tag: 'document-upload-manager',
    scriptUrlKey: 'documentManagementUiJs',
    apiBaseUrlKey: 'document',
    attributes: ['application-id', 'api-base-url', 'session-token'],
    containerWidth: 800,
  },

  UNDERWRITING: UNDER_REVIEW_SPINNER('Under Review by Our Team', 'Your application is being reviewed by our underwriting team.\nYou will be notified once a decision has been made.'),
  REFERRED: UNDER_REVIEW_SPINNER('Under Review by Our Team', 'Your application is being reviewed by our underwriting team.\nYou will be notified once a decision has been made.'),

  DECLINED: {
    kind: 'static-block',
    block: 'declined',
  },

  CANCELLED: { kind: 'static-block', block: 'cancelled-expired', label: 'Cancelled' },
  EXPIRED: { kind: 'static-block', block: 'cancelled-expired', label: 'Expired' },

  OFFER_ACCEPTED: POST_ACCEPTANCE_BLOCK('Offer Accepted', 'Your offer has been accepted. We are now completing final verification and collecting your disbursement details before your loan can be funded.'),
  FUNDING_PENDING: POST_ACCEPTANCE_BLOCK('Funding in Progress', 'Your loan funds are being prepared and will be disbursed shortly.'),
  FUNDED: POST_ACCEPTANCE_BLOCK('Funded', 'Your loan has been funded. The funds have been sent to your nominated account.'),
  COMPLETED: POST_ACCEPTANCE_BLOCK('Completed', 'Your loan application is complete. Thank you for choosing us.'),
};

export const PROCESSING_STATES = new Set(
  Object.entries(NAVIGATION_CONFIG)
    .filter(([, d]) => d.kind === 'spinner' && d.group === 'processing')
    .map(([status]) => status)
);

export const UNDER_REVIEW_STATES = new Set(
  Object.entries(NAVIGATION_CONFIG)
    .filter(([, d]) => d.kind === 'spinner' && d.group === 'under-review')
    .map(([status]) => status)
);
