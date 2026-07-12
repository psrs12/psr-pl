import { describe, it, expect } from 'vitest';
import { resolveScreen, UnmappedApplicationStatusError } from './resolveScreen.js';
import { NAVIGATION_CONFIG } from './navigationConfig.js';

// Mirrors services/application-management-service ApplicationStatus enum exactly.
const CANONICAL_APPLICATION_STATUSES = [
  'CREATED', 'IN_PROGRESS', 'READY_FOR_SUBMISSION', 'SUBMITTED', 'PROCESSING',
  'SOFT_PULL_PENDING', 'PRICING_PENDING', 'OFFER_PENDING', 'CONSENT_CAPTURED',
  'HARD_PULL_PENDING', 'DECISION_PENDING', 'APPROVED', 'DECLINED', 'REFERRED',
  'DOCUMENTS_REQUIRED', 'OFFER_ACCEPTED', 'UNDERWRITING', 'FUNDING_PENDING',
  'FUNDED', 'COMPLETED', 'CANCELLED', 'EXPIRED',
];

describe('resolveScreen', () => {
  it('covers every canonical ApplicationStatus value', () => {
    for (const status of CANONICAL_APPLICATION_STATUSES) {
      expect(() => resolveScreen(status)).not.toThrow();
    }
  });

  it('has no config entries outside the canonical status set', () => {
    for (const status of Object.keys(NAVIGATION_CONFIG)) {
      expect(CANONICAL_APPLICATION_STATUSES).toContain(status);
    }
  });

  it('returns the exact configured descriptor for a known status', () => {
    expect(resolveScreen('OFFER_PENDING')).toBe(NAVIGATION_CONFIG.OFFER_PENDING);
    expect(resolveScreen('DOCUMENTS_REQUIRED')).toBe(NAVIGATION_CONFIG.DOCUMENTS_REQUIRED);
  });

  it('throws UnmappedApplicationStatusError for an unmapped status', () => {
    expect(() => resolveScreen('NOT_A_REAL_STATUS')).toThrow(UnmappedApplicationStatusError);
  });

  it('does not accept the UI-invented COMPLIANCE_HOLD status', () => {
    expect(() => resolveScreen('COMPLIANCE_HOLD')).toThrow(UnmappedApplicationStatusError);
  });
});
