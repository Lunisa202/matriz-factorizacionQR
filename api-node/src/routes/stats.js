import express from 'express';
import { extractMatrices } from '../services/matrixValidator.js';
import { computeStatistics } from '../services/statistics.js';

const router = express.Router();

/**
 * POST /api/stats (se monta con jwtMiddleware en index.js).
 * Recibe las matrices procesadas por la API de Go y retorna estadísticas.
 *
 * Body aceptado: { "q": [...], "r": [...] } (lo que envía Go),
 * o { "matrices": [...] } o { "matrix": [...] } para pruebas directas.
 *
 * Nota Express 5: los errores lanzados en handlers async llegan solos al
 * error-handler; aquí el handler es sync y delega con next(err) igualmente.
 */
router.post('/', (req, res, next) => {
  try {
    const matrices = extractMatrices(req.body);
    const stats = computeStatistics(matrices);
    res.json({
      ...stats,
      matricesAnalyzed: matrices.map((m) => m.name),
    });
  } catch (err) {
    next(err);
  }
});

export default router;
