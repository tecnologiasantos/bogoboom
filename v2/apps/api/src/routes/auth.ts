import type { FastifyPluginAsync } from 'fastify';
import { z } from 'zod';
import type { Auth } from '../auth.js';
import type { Config } from '../config.js';

export const authRoutes = (auth: Auth, cfg: Config): FastifyPluginAsync => async (app) => {
  if (cfg.AUTH_MODE === 'dev') {
    app.post('/auth/dev-login', { config: { rateLimit: { max: cfg.NODE_ENV === 'test' ? 10_000 : 30, timeWindow: '1 minute' } } }, async (req) => {
      const b = z.object({ email: z.string().email(), role: z.enum(['newcomer', 'hr', 'agency']) }).parse(req.body);
      return auth.devLogin(b.email, b.role);
    });
  }
  app.get('/me', { preHandler: auth.authenticate }, async (req) => ({ user: req.user }));
};
