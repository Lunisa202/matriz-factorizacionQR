/**
 * Configuración centralizada (única fuente de verdad).
 *
 * Toda variable de entorno se lee AQUÍ y solo aquí (principio DRY):
 * ningún otro módulo accede a `process.env` directamente.
 * Al arrancar, `validate()` obliga a usar un secreto real en producción.
 */

const INSECURE_DEFAULT_SECRET = 'supersecret-reto-tecnico';

function num(value, fallback) {
  const n = Number(value);
  return Number.isFinite(n) && n > 0 ? n : fallback;
}

function bool(value, fallback) {
  if (value === undefined) return fallback;
  return /^(1|true|yes)$/i.test(String(value).trim());
}

export const config = {
  env: process.env.NODE_ENV || 'development',
  port: num(process.env.PORT, 3002),
  jwtSecret: process.env.JWT_SECRET || INSECURE_DEFAULT_SECRET,
  jwtIssuer: process.env.JWT_ISSUER || 'reto-tecnico',
  jwtAudience: process.env.JWT_AUDIENCE || 'reto-tecnico-apis',
  jwtExpiresIn: process.env.JWT_EXPIRES_IN || '2h',
  // CORS: en local se permite todo; en producción se restringe por env.
  corsOrigin: process.env.CORS_ORIGIN || '*',
  bodyLimit: process.env.BODY_LIMIT || '1mb',
  // Rate limit aplicado solo a /api/* (el /health queda libre para probes).
  rateLimitWindowMs: num(process.env.RATE_LIMIT_WINDOW_MS, 60_000),
  rateLimitMax: num(process.env.RATE_LIMIT_MAX, 100),
  // Tope anti-DoS: Gram-Schmidt es O(m·n²); una matriz gigante quemaría CPU.
  maxMatrixElements: num(process.env.MAX_MATRIX_ELEMENTS, 10_000),
  // Permite desactivar el emisor de tokens demo (p. ej. en despliegues públicos).
  enableDemoToken: bool(process.env.ENABLE_DEMO_TOKEN, true),
};

/**
 * Falla rápido (fail-fast) si la configuración es insegura para producción:
 * mejor no arrancar que arrancar con el secreto de ejemplo.
 */
export function validateConfig() {
  if (config.env === 'production' && config.jwtSecret === INSECURE_DEFAULT_SECRET) {
    throw new Error(
      'Refusing to start in production with the default JWT_SECRET. ' +
        'Set a strong JWT_SECRET env var (e.g. `openssl rand -hex 32`).',
    );
  }
}
