import { HttpError } from '../utils/errors.js';
import { config } from '../config.js';

/**
 * Validación de matrices (SRP: este módulo SOLO valida/extrae,
 * el cálculo vive en `statistics.js`).
 *
 * Formatos aceptados (flexible para que Go y Postman funcionen):
 *   { "q": [...], "r": [...] }  <- lo que envía la API de Go
 *   { "matrices": [[...],[...]] }
 *   { "matrix": [[...]] }
 */
export function extractMatrices(body) {
  if (!body || typeof body !== 'object') {
    throw new HttpError(400, 'invalid JSON body, expected {"q": [...], "r": [...]}');
  }
  const out = [];
  if (Array.isArray(body.q)) out.push({ name: 'q', data: body.q });
  if (Array.isArray(body.r)) out.push({ name: 'r', data: body.r });
  if (Array.isArray(body.matrices)) {
    body.matrices.forEach((m, i) => out.push({ name: `matrices[${i}]`, data: m }));
  }
  if (Array.isArray(body.matrix)) out.push({ name: 'matrix', data: body.matrix });

  if (out.length === 0) {
    throw new HttpError(400, 'no matrices found, send {"q": [...], "r": [...]}');
  }

  let totalElements = 0;
  for (const { name, data } of out) {
    if (data.length === 0) throw new HttpError(400, `matrix "${name}" must have at least one row`);
    const cols = data[0].length;
    if (!cols) throw new HttpError(400, `matrix "${name}" must have at least one column`);
    data.forEach((row, i) => {
      if (!Array.isArray(row) || row.length !== cols) {
        throw new HttpError(400, `matrix "${name}" must be rectangular (row ${i} mismatch)`);
      }
      row.forEach((v) => {
        if (typeof v !== 'number' || !Number.isFinite(v)) {
          throw new HttpError(400, `matrix "${name}" must contain only finite numbers`);
        }
      });
    });
    totalElements += data.length * cols;
  }

  // Tope anti-DoS (ver `config.maxMatrixElements`): responde 413, no 500.
  if (totalElements > config.maxMatrixElements) {
    throw new HttpError(
      413,
      `payload too large: ${totalElements} elements exceed the limit of ${config.maxMatrixElements}`,
    );
  }
  return out;
}
