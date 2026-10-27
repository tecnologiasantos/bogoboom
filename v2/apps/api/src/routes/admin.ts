import type { FastifyPluginAsync } from 'fastify';
import type { Db } from '../db.js';
import { audit } from '../db.js';
import type { Auth } from '../auth.js';
import { notFound } from '../errors.js';
import { uuid } from '../schemas.js';

export const adminRoutes = (db: Db, auth: Auth): FastifyPluginAsync => async (app) => {
  const only = { preHandler: auth.requireRole('admin') };
  app.get('/admin/agencies/pending', only, async () =>
    (await db.query('SELECT id, name, city, country, created_at FROM agencies WHERE verified_at IS NULL ORDER BY created_at')).rows);
  app.post('/admin/agencies/:id/verify', only, async (req) => {
    const id = uuid.parse((req.params as any).id);
    const r = await db.query('UPDATE agencies SET verified_at=now() WHERE id=$1 RETURNING id', [id]);
    if (!r.rowCount) throw notFound('Agency');
    await audit(db, req.user!.id, 'admin.agency_verified', 'agencies', id);
    return { id, verified: true };
  });
};
