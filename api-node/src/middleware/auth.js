import jwt from 'jsonwebtoken';
import { HttpError } from '../utils/errors.js';
import { config } from '../config.js';

/**
 * Middleware JWT: valida "Authorization: Bearer <token>" con el secreto
 * compartido (el mismo JWT_SECRET de la API de Go para interoperabilidad).
 * Además exige `issuer` y `audience` para que tokens de otros sistemas
 * (aunque usen el mismo secreto) no sean aceptados aquí.
 */
export function jwtMiddleware(req, _res, next) {
  const auth = req.headers.authorization || '';

  if (!auth) return next(new HttpError(401, 'missing Authorization header'));

  const [scheme, token] = auth.split(' ');
  if (!/^bearer$/i.test(scheme) || !token) {
    return next(new HttpError(401, "invalid Authorization format, expected 'Bearer <token>'"));
  }

  try {
    const claims = jwt.verify(token, config.jwtSecret, {
      issuer: config.jwtIssuer,
      audience: config.jwtAudience,
    });
    req.user = claims;
    next();
  } catch {
    next(new HttpError(401, 'invalid or expired token'));
  }
}
