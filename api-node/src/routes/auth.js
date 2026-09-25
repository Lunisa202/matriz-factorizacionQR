import express from 'express';
import jwt from 'jsonwebtoken';
import { config } from '../config.js';
import { HttpError } from '../utils/errors.js';

const router = express.Router();

/**
 * POST /api/auth/token (público SOLO si ENABLE_DEMO_TOKEN=true).
 * Emite un JWT firmado con el secreto compartido, incluyendo issuer/audience.
 * En producción: desactivar (ENABLE_DEMO_TOKEN=false) y usar login real.
 */
router.post('/token', (req, res, next) => {
  if (!config.enableDemoToken) {
    return next(new HttpError(404, 'not found'));
  }
  const user = (req.body && req.body.user) || 'demo';
  const token = jwt.sign({ sub: user }, config.jwtSecret, {
    expiresIn: config.jwtExpiresIn,
    issuer: config.jwtIssuer,
    audience: config.jwtAudience,
  });
  res.json({ token });
});

export default router;
