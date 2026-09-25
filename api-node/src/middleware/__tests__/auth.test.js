import { describe, it, expect } from 'vitest';
import jwt from 'jsonwebtoken';
import { jwtMiddleware } from '../auth.js';
import { config } from '../../config.js';

function runMiddleware(headers = {}) {
  return new Promise((resolve) => {
    const req = { headers };
    const next = (err) => resolve({ req, err });
    jwtMiddleware(req, {}, next);
  });
}

function sign(payload = { sub: 'demo' }, options = {}) {
  return jwt.sign(payload, config.jwtSecret, {
    expiresIn: '2h',
    issuer: config.jwtIssuer,
    audience: config.jwtAudience,
    ...options,
  });
}

describe('jwtMiddleware', () => {
  it('calls next() without error and attaches claims for a valid token', async () => {
    const { req, err } = await runMiddleware({ authorization: `Bearer ${sign()}` });
    expect(err).toBeUndefined();
    expect(req.user.sub).toBe('demo');
  });

  it('rejects missing Authorization header with 401', async () => {
    const { err } = await runMiddleware({});
    expect(err?.status).toBe(401);
  });

  it('rejects malformed scheme with 401', async () => {
    const { err } = await runMiddleware({ authorization: 'Token abc' });
    expect(err?.status).toBe(401);
  });

  it('rejects expired tokens with 401', async () => {
    const expired = sign({}, { expiresIn: '-1s' });
    const { err } = await runMiddleware({ authorization: `Bearer ${expired}` });
    expect(err?.status).toBe(401);
  });

  it('rejects tokens with wrong audience with 401', async () => {
    const foreign = sign({}, { audience: 'other-system' });
    const { err } = await runMiddleware({ authorization: `Bearer ${foreign}` });
    expect(err?.status).toBe(401);
  });
});
