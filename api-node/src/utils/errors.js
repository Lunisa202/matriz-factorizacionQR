/**
 * Utilidades de error centralizado.
 *
 * Patrón: los servicios lanzan `HttpError(status, message)` y este
 * middleware final los convierte en JSON `{ error, code }`.
 * Así ningún handler repite lógica de formato de errores (DRY).
 */

export class HttpError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

// eslint-disable-next-line no-unused-vars
export function errorHandler(err, _req, res, _next) {
  const status = err.status || 500;
  // No filtrar detalles internos en 500 (evita fuga de información).
  const message = status === 500 ? 'internal server error' : err.message || 'internal server error';
  res.status(status).json({ error: message, code: status });
}

export function notFound(_req, res) {
  res.status(404).json({ error: 'not found', code: 404 });
}
