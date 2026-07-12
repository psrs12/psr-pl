import { NAVIGATION_CONFIG } from './navigationConfig.js';

export class UnmappedApplicationStatusError extends Error {
  constructor(applicationStatus) {
    super(`No navigation config entry for applicationStatus "${applicationStatus}"`);
    this.name = 'UnmappedApplicationStatusError';
    this.applicationStatus = applicationStatus;
  }
}

/**
 * @param {string} applicationStatus
 * @returns {import('./navigationConfig.js').ScreenDescriptor}
 */
export function resolveScreen(applicationStatus) {
  const descriptor = NAVIGATION_CONFIG[applicationStatus];
  if (!descriptor) {
    throw new UnmappedApplicationStatusError(applicationStatus);
  }
  return descriptor;
}
