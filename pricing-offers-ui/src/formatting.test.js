import { describe, it, expect } from 'vitest';
import { centsToDollarString } from './formatting.js';

describe('centsToDollarString', () => {
  it('formats whole dollars with two decimal places', () => {
    expect(centsToDollarString(1500000)).toBe('$15,000.00');
  });

  it('formats cents correctly', () => {
    expect(centsToDollarString(150050)).toBe('$1,500.50');
  });

  it('formats zero', () => {
    expect(centsToDollarString(0)).toBe('$0.00');
  });
});
