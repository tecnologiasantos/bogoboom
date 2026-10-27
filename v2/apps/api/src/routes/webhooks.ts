import { createHmac, timingSafeEqual } from 'node:crypto';
import type { FastifyPluginAsync } from 'fastify';
import { z } from 'zod';
import type { Config } from '../config.js';
import type { Db } from '../db.js';
import { tx, audit } from '../db.js';
import { AppError } from '../errors.js';

const body = z.object({
  agency_id: z.string().uuid(),
  plan: z.enum(['pro', 'enterprise']),
  status: z.enum(['trialing', 'active', 'past_due', 'cancelled', 'expired']),
  provider: z.enum(['stripe', 'mercadopago']),
  provider_subscription_id: z.string().min(3).max(120),
  current_period_end: z.string().datetime(),
});

/** Payment provider events (verified with HMAC-SHA256 over the raw body in `x-signature`). Idempotent. */
export const webhookRoutes = (db: Db, cfg: Config): FastifyPluginAsync => async (app) => {
  app.register(async (scope) => {
    scope.addContentTypeParser('application/json', { parseAs: 'buffer' }, (_r, b, done) => done(null, b));
    scope.post('/webhooks/payments', async (req) => {
      if (!cfg.PAYMENT_WEBHOOK_SECRET) throw new AppError(503, 'Webhook not configured', 'not_configured');
      const raw = req.body as Buffer;
      const given = Buffer.from(String(req.headers['x-signature'] ?? ''), 'hex');
      const want = createHmac('sha256', cfg.PAYMENT_WEBHOOK_SECRET).update(raw).digest();
      if (given.length !== want.length || !timingSafeEqual(given, want)) throw new AppError(401, 'Bad signature', 'bad_signature');
      const e = body.parse(JSON.parse(raw.toString('utf8')));
      await tx(db, async (c) => {
        const ok = await c.query('SELECT 1 FROM agencies WHERE id=$1', [e.agency_id]);
        if (!ok.rowCount) throw new AppError(404, 'Agency not found', 'not_found');
        await c.query(
          `INSERT INTO subscriptions (agency_id, plan, status, provider, provider_subscription_id, current_period_end)
           VALUES ($1,$2,$3,$4,$5,$6)
           ON CONFLICT (provider_subscription_id) DO UPDATE SET status=EXCLUDED.status, plan=EXCLUDED.plan, current_period_end=EXCLUDED.current_period_end`,
          [e.agency_id, e.plan, e.status, e.provider, e.provider_subscription_id, e.current_period_end]);
        const on = e.status === 'active' || e.status === 'trialing';
        await c.query('UPDATE agencies SET plan=$2, plan_active_until=$3 WHERE id=$1',
          [e.agency_id, on ? e.plan : 'free', on ? e.current_period_end : null]);
        await audit(c, null, 'payments.webhook', 'agencies', e.agency_id, { status: e.status, plan: e.plan });
      });
      return { ok: true };
    });
  });
};
