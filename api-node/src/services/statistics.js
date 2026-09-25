/**
 * Cálculo de estadísticas (SRP: este módulo SOLO calcula;
 * la validación/extracción vive en `matrixValidator.js`).
 */

const EPS = 1e-9; // tolerancia para comparar con cero (errores de punto flotante de QR)

/** ¿La matriz es diagonal? Solo aplica a cuadradas; fuera de la diagonal todo debe ser ~0. */
export function isDiagonalMatrix(m) {
  if (m.length === 0 || m.length !== m[0].length) return false;
  for (let i = 0; i < m.length; i++) {
    for (let j = 0; j < m[i].length; j++) {
      if (i !== j && Math.abs(m[i][j]) > EPS) return false;
    }
  }
  return true;
}

/**
 * Calcula: máximo global, mínimo global, promedio, suma total y
 * si ALGUNA matriz es diagonal.
 */
export function computeStatistics(matrices) {
  let max = -Infinity;
  let min = Infinity;
  let sum = 0;
  let count = 0;
  let isDiagonal = false;
  let diagonalMatrix = null;

  for (const { name, data } of matrices) {
    if (isDiagonalMatrix(data)) {
      isDiagonal = true;
      if (!diagonalMatrix) diagonalMatrix = name;
    }
    for (const row of data) {
      for (const v of row) {
        if (v > max) max = v;
        if (v < min) min = v;
        sum += v;
        count += 1;
      }
    }
  }

  return {
    max,
    min,
    average: count === 0 ? 0 : sum / count,
    sum,
    count,
    isDiagonal,
    diagonalMatrix,
  };
}
