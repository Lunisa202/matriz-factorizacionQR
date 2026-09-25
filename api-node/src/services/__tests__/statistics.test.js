import { describe, it, expect } from 'vitest';
import { computeStatistics, isDiagonalMatrix } from '../statistics.js';

describe('isDiagonalMatrix', () => {
  it('returns true for an identity matrix', () => {
    expect(
      isDiagonalMatrix([
        [1, 0],
        [0, 1],
      ]),
    ).toBe(true);
  });

  it('returns false when an off-diagonal element is non-zero', () => {
    expect(
      isDiagonalMatrix([
        [1, 2],
        [0, 1],
      ]),
    ).toBe(false);
  });

  it('returns false for non-square matrices', () => {
    expect(
      isDiagonalMatrix([
        [1, 0, 0],
        [0, 1, 0],
      ]),
    ).toBe(false);
  });

  it('tolerates floating-point noise from QR (1e-9)', () => {
    expect(
      isDiagonalMatrix([
        [1, 1e-10],
        [0, 1],
      ]),
    ).toBe(true);
  });
});

describe('computeStatistics', () => {
  it('computes max/min/average/sum/count over all matrices', () => {
    const stats = computeStatistics([
      { name: 'q', data: [[1, 0], [0, 1]] },
      { name: 'r', data: [[5, 0], [0, 3]] },
    ]);
    expect(stats).toMatchObject({
      max: 5,
      min: 0,
      sum: 10,
      count: 8,
      average: 1.25,
    });
  });

  it('reports isDiagonal=true with the first diagonal matrix name', () => {
    const stats = computeStatistics([
      { name: 'q', data: [[1, 2], [3, 4]] }, // not diagonal
      { name: 'r', data: [[5, 0], [0, 3]] }, // diagonal
    ]);
    expect(stats.isDiagonal).toBe(true);
    expect(stats.diagonalMatrix).toBe('r');
  });

  it('reports isDiagonal=false when no matrix is diagonal', () => {
    const stats = computeStatistics([{ name: 'matrix', data: [[1, 2], [3, 4]] }]);
    expect(stats.isDiagonal).toBe(false);
    expect(stats.diagonalMatrix).toBeNull();
  });
});
