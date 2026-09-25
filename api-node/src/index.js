import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import morgan from 'morgan';
import rateLimit from 'express-rate-limit';
import { config, validateConfig } from './config.js';
import { jwtMiddleware } from './middleware/auth.js';
import statsRoutes from './routes/stats.js';
import authRoutes from './routes/auth.js';
import { errorHandler, notFound } from './utils/errors.js';
import { mountSwagger } from './swagger.js';

// Fail-fast: no arrancar con configuración insegura en producción.
validateConfig();

const app = express();

// Helmet: cabeceras de seguridad. La CSP por defecto bloquearía el JS/CSS
// inline que sirve Swagger UI, así que se relaja SOLO lo necesario para /docs
// (scripts y estilos inline same-origin). Resto de directivas intactas.
app.use(
  helmet({
    contentSecurityPolicy: {
      directives: {
        ...helmet.contentSecurityPolicy.getDefaultDirectives(),
        'script-src': ["'self'", "'unsafe-inline'"],
        'style-src': ["'self'", "'unsafe-inline'", 'https:'],
        'img-src': ["'self'", 'data:', 'https:'],
      },
    },
  }),
);
// CORS restringible por env (en local '*' para Postman/browser).
app.use(cors({ origin: config.corsOrigin }));
app.use(express.json({ limit: config.bodyLimit }));
app.use(morgan('dev'));

// Health público y SIN rate-limit (probes de Docker/Render).
app.get('/health', (_req, res) => res.json({ status: 'ok', service: 'api-node' }));

// Documentación interactiva: Swagger UI en /docs, spec cruda en /docs.json.
// Públicas y sin rate-limit para que los evaluadores prueben sin fricción.
mountSwagger(app);

// Rate limit solo para la API (defensa anti fuerza-bruta/DoS).
const apiLimiter = rateLimit({
  windowMs: config.rateLimitWindowMs,
  max: config.rateLimitMax,
  standardHeaders: 'draft-8',
  legacyHeaders: false,
  message: { error: 'too many requests, try again later', code: 429 },
});
app.use('/api', apiLimiter);

// Rutas públicas de autenticación (emisión de token demo, con guard).
app.use('/api/auth', authRoutes);

// Rutas protegidas con JWT compartido.
app.use('/api/stats', jwtMiddleware, statsRoutes);

app.use(notFound);
app.use(errorHandler);

app.listen(config.port, () => {
  console.log(`api-node listening on :${config.port} [${config.env}]`);
});
