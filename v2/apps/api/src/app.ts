import Fastify from 'fastify';
import cors from '@fastify/cors';
import helmet from '@fastify/helmet';
import rateLimit from '@fastify/rate-limit';
import { ZodError } from 'zod';
import type { Config } from './config.js';
import type { Db } from './db.js';
import { makeAuth } from './auth.js';
import { AppError } from './errors.js';
import { authRoutes } from './routes/auth.js';
import { newcomerRoutes } from './routes/newcomer.js';
import { propertyRoutes } from './routes/properties.js';
import { hrRoutes } from './routes/hr.js';
import { agencyRoutes } from './routes/agency.js';
import { adminRoutes } from './routes/admin.js';
import { webhookRoutes } from './routes/webhooks.js';

export async function buildApp(cfg: Config, db: Db, opts: { logger?: boolean } = {}) {
  const app = Fastify({
    logger: opts.logger ? { redact: ['req.headers.authorization', 'req.headers["x-signature"]'] } : false,
    bodyLimit: 2 * 1024 * 1024,
  });
  const auth = makeAuth(cfg, db);

  await app.register(helmet);
  await app.register(cors, { origin: cfg.CORS_ORIGIN.split(','), methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'] });
  await app.register(rateLimit, { max: 300, timeWindow: '1 minute' });
  app.addContentTypeParser('text/csv', { parseAs: 'string' }, (_req, body, done) => done(null, body));

  app.setErrorHandler((err: any, _req, reply) => {
    if (err instanceof ZodError) return reply.code(400).send({ error: 'validation_error', issues: err.issues.map((i) => ({ path: i.path.join('.'), message: i.message })) });
    if (err instanceof AppError) return reply.code(err.status).send({ error: err.code, message: err.message });
    if (err.statusCode && err.statusCode < 500) return reply.code(err.statusCode).send({ error: err.code ?? 'bad_request', message: err.message });
    app.log.error(err);
    return reply.code(500).send({ error: 'internal_error', message: 'Internal server error' });
  });

  app.get('/health', async () => ({ ok: true }));
  await app.register(authRoutes(auth, cfg));
  await app.register(newcomerRoutes(db, auth));
  await app.register(propertyRoutes(db, auth));
  await app.register(hrRoutes(db, auth));
  await app.register(agencyRoutes(db, auth));
  await app.register(adminRoutes(db, auth));
  await app.register(webhookRoutes(db, cfg));
  return app;
}
