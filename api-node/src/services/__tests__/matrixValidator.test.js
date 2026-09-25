import { describe, it, expect } from 'vitest';
import { extractMatrices } from '../matrixValidator.js';

describe('extractMatrices', () => {
  it('accepts the { q, r } payload sent by the Go API', () => {
    const out = extractMatrices({ q: [[1, 0]], r: [[2, 0]] });
    expect(out.map((m) => m.name)).toEqual(['q', 'r']);
  });

  it('accepts the { matrices } and { matrix } variants', () => {
    expect(extractMatrices({ matrices: [[[1]]] }).map((m) => m.name)).toEqual(['matrices[0]']);
    expect(extractMatrices({ matrix: [[1]] }).map((m) => m.name)).toEqual(['matrix']);
  });

  it('rejects non-rectangular matrices with 400', () => {
    try {
      extractMatrices({ matrix: [[1, 2], [3]] });
      expect.unreachable();
    } catch (err) {
      expect(err.status).toBe(400);
    }
  });

  it('rejects non-numeric values with 400', () => {
    try {
      extractMatrices({ matrix: [['a']] });
      expect.unreachable();
    } catch (err) {
      expect(err.status).toBe(400);
    }
  });

  it('rejects empty bodies with 400', () => {
    try {
      extractMatrices({});
      expect.unreachable();
    } catch (err) {
      expect(err.status).toBe(400);
    }
  });

  it('rejects oversized payloads with 413 (anti-DoS)', () => {
    // Default limit is 10_000 elements: 101x101 = 10_201 exceeds it.
    const big = Array.from({ length: 101 }, () => new Array(101).fill(1));
    try {
      extractMatrices({ matrix: big });
      expect.unreachable();
    } catch (err) {
      expect(err.status).toBe(413);
    }
  });
});
