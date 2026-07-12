import { describe, it, expect } from 'vitest';
import { NAVIGATION_CONFIG } from './navigationConfig.js';
import { SCREEN_REGISTRY } from './screenRegistry.js';
import { STATIC_BLOCK_REGISTRY } from './staticBlockRegistry.js';

describe('screenRegistry completeness', () => {
  it('has a container for every descriptor kind used in NAVIGATION_CONFIG', () => {
    for (const descriptor of Object.values(NAVIGATION_CONFIG)) {
      expect(SCREEN_REGISTRY[descriptor.kind]).toBeDefined();
    }
  });

  it('has a static-block registry entry for every static-block descriptor', () => {
    for (const descriptor of Object.values(NAVIGATION_CONFIG)) {
      if (descriptor.kind === 'static-block') {
        expect(STATIC_BLOCK_REGISTRY[descriptor.block]).toBeDefined();
      }
    }
  });
});
