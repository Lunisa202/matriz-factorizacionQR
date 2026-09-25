import swaggerUi from 'swagger-ui-express';
import { config } from './config.js';

/**
 * Especificación OpenAPI 3.0 de api-node (definición central, sin JSDoc disperso).
 * Se sirve en `/docs` (Swagger UI) y `/docs.json` (spec cruda).
 * `servers: [{ url: '/' }]` hace que "Try it out" apunte al host actual:
 * funciona igual en localhost, Docker y Render sin cambiar nada.
 */
export const swaggerSpec = {
  openapi: '3.0.3',
  info: {
    title: 'Reto técnico — API Node (Express 5)',
    version: '1.0.0',
    description:
      'Recibe matrices Q/R de la API de Go y calcula estadísticas: ' +
      'máximo, mínimo, promedio, suma total y verificación de matriz diagonal.',
  },
  servers: [{ url: '/', description: 'Este servidor (Try it out funciona aquí mismo)' }],
  components: {
    securitySchemes: {
      bearerAuth: { type: 'http', scheme: 'bearer', bearerFormat: 'JWT' },
    },
    schemas: {
      Matrix: {
        type: 'array',
        description: 'Matriz rectangular: array de arrays de números finitos.',
        items: { type: 'array', items: { type: 'number' } },
        example: [
          [1, 0],
          [0, 1],
        ],
      },
      StatsRequest: {
        type: 'object',
        description: 'Lo que envía la API de Go (o variantes para pruebas directas).',
        properties: {
          q: { $ref: '#/components/schemas/Matrix' },
          r: { $ref: '#/components/schemas/Matrix' },
          matrices: { type: 'array', items: { $ref: '#/components/schemas/Matrix' } },
          matrix: { $ref: '#/components/schemas/Matrix' },
        },
        example: {
          q: [
            [1, 0],
            [0, 1],
          ],
          r: [
            [5, 0],
            [0, 3],
          ],
        },
      },
      StatsResponse: {
        type: 'object',
        properties: {
          max: { type: 'number', example: 5 },
          min: { type: 'number', example: 0 },
          average: { type: 'number', example: 1.25 },
          sum: { type: 'number', example: 10 },
          count: { type: 'integer', example: 8 },
          isDiagonal: { type: 'boolean', example: true },
          diagonalMatrix: { type: 'string', nullable: true, example: 'q' },
          matricesAnalyzed: { type: 'array', items: { type: 'string' }, example: ['q', 'r'] },
        },
      },
      TokenRequest: {
        type: 'object',
        properties: { user: { type: 'string', example: 'demo' } },
      },
      TokenResponse: {
        type: 'object',
        properties: { token: { type: 'string' } },
      },
      Error: {
        type: 'object',
        properties: {
          error: { type: 'string', example: 'invalid or expired token' },
          code: { type: 'integer', example: 401 },
        },
      },
    },
  },
  paths: {
    '/health': {
      get: {
        summary: 'Salud del servicio',
        responses: {
          200: {
            description: 'OK',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  example: { status: 'ok', service: 'api-node' },
                },
              },
            },
          },
        },
      },
    },
    '/api/auth/token': {
      post: {
        summary: 'Emite un JWT demo (solo si ENABLE_DEMO_TOKEN=true)',
        requestBody: {
          content: { 'application/json': { schema: { $ref: '#/components/schemas/TokenRequest' } } },
        },
        responses: {
          200: {
            description: 'Token emitido',
            content: { 'application/json': { schema: { $ref: '#/components/schemas/TokenResponse' } } },
          },
          404: {
            description: 'Emisor demo desactivado',
            content: { 'application/json': { schema: { $ref: '#/components/schemas/Error' } } },
          },
        },
      },
    },
    '/api/stats': {
      post: {
        summary: 'Calcula estadísticas sobre las matrices recibidas',
        security: [{ bearerAuth: [] }],
        requestBody: {
          required: true,
          content: { 'application/json': { schema: { $ref: '#/components/schemas/StatsRequest' } } },
        },
        responses: {
          200: {
            description: 'Estadísticas calculadas',
            content: { 'application/json': { schema: { $ref: '#/components/schemas/StatsResponse' } } },
          },
          400: {
            description: 'Body inválido',
            content: { 'application/json': { schema: { $ref: '#/components/schemas/Error' } } },
          },
          401: {
            description: 'Sin token o inválido',
            content: { 'application/json': { schema: { $ref: '#/components/schemas/Error' } } },
          },
          413: {
            description: 'Excede MAX_MATRIX_ELEMENTS',
            content: { 'application/json': { schema: { $ref: '#/components/schemas/Error' } } },
          },
          429: {
            description: 'Rate limit excedido',
            content: { 'application/json': { schema: { $ref: '#/components/schemas/Error' } } },
          },
        },
      },
    },
  },
};

export function mountSwagger(app) {
  if (config.env === 'test') return; // no montar UI durante vitest
  app.get('/docs.json', (_req, res) => res.json(swaggerSpec));
  app.use('/docs', swaggerUi.serve, swaggerUi.setup(swaggerSpec));
}
