import { describe, it, expect } from 'vitest';
import { swaggerSpec } from './swagger.js';

describe('swaggerSpec', () => {
  it('is a valid OpenAPI 3.x document', () => {
    expect(swaggerSpec.openapi).toMatch(/^3\./);
    expect(swaggerSpec.info.title).toBeTruthy();
  });

  it('documents every public route', () => {
    expect(Object.keys(swaggerSpec.paths).sort()).toEqual([
      '/api/auth/token',
      '/api/stats',
      '/health',
    ]);
  });

  it('marks /api/stats as Bearer-protected (Authorize button works)', () => {
    const op = swaggerSpec.paths['/api/stats'].post;
    expect(op.security).toEqual([{ bearerAuth: [] }]);
    expect(swaggerSpec.components.securitySchemes.bearerAuth).toMatchObject({
      type: 'http',
      scheme: 'bearer',
    });
  });

  it('uses a relative server so Try it out works on any host', () => {
    expect(swaggerSpec.servers).toEqual([
      { url: '/', description: expect.any(String) },
    ]);
  });
});
