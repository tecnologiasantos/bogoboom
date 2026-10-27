import type { FastifyPluginAsync } from 'fastify';
import { z } from 'zod';
import type { Db } from '../db.js';
import type { Auth } from '../auth.js';
import { notFound } from '../errors.js';
import { uuid } from '../schemas.js';

const q = z.object({
  city: z.string().trim().max(80).optional(),
  district: z.string().trim().max(80).optional(),
  min_price: z.coerce.number().optional(),
  max_price: z.coerce.number().optional(),
  type: z.enum(['apartment', 'room', 'house', 'studio']).optional(),
  min_bedrooms: z.coerce.number().int().optional(),
  furnished: z.enum(['true', 'false']).optional(),
  available_before: z.string().regex(/^\d{4}-\d{2}-\d{2}$/).optional(),
  limit: z.coerce.number().int().min(1).max(50).default(20),
  offset: z.coerce.number().int().min(0).default(0),
});

export const propertyRoutes = (db: Db, auth: Auth): FastifyPluginAsync => async (app) => {
  app.get('/properties', { preHandler: auth.authenticate }, async (req) => {
    const f = q.parse(req.query);
    const w: string[] = ['p.is_active'], v: unknown[] = [];
    const add = (sql: string, val: unknown) => { v.push(val); w.push(sql.replace('?', `$${v.length}`)); };
    if (f.city) add('lower(unaccent(p.city)) = lower(unaccent(?))', f.city);
    if (f.district) add('lower(unaccent(p.district)) = lower(unaccent(?))', f.district);
    if (f.min_price != null) add('p.price >= ?', f.min_price);
    if (f.max_price != null) add('p.price <= ?', f.max_price);
    if (f.type) add('p.type = ?', f.type);
    if (f.min_bedrooms != null) add('p.bedrooms >= ?', f.min_bedrooms);
    if (f.furnished) add('p.furnished = ?', f.furnished === 'true');
    if (f.available_before) add('(p.available_from IS NULL OR p.available_from <= ?)', f.available_before);
    v.push(f.limit, f.offset);
    const rows = (await db.query(
      `SELECT p.id, p.title, p.description, p.price, p.currency, p.type, p.bedrooms, p.furnished, p.city, p.district,
              p.lat, p.lng, p.available_from, p.created_at, a.id AS agency_id, a.name AS agency_name,
              (a.verified_at IS NOT NULL) AS agency_verified, count(*) OVER() AS total
         FROM properties p JOIN agencies a ON a.id=p.agency_id
        WHERE ${w.join(' AND ')} ORDER BY (a.verified_at IS NOT NULL) DESC, p.created_at DESC
        LIMIT $${v.length - 1} OFFSET $${v.length}`, v)).rows;
    return { total: rows[0] ? Number(rows[0].total) : 0, items: rows.map(({ total, ...r }) => r) };
  });

  app.get('/properties/:id', { preHandler: auth.authenticate }, async (req) => {
    const id = uuid.parse((req.params as any).id);
    const r = (await db.query(
      `SELECT p.*, a.name AS agency_name, (a.verified_at IS NOT NULL) AS agency_verified
         FROM properties p JOIN agencies a ON a.id=p.agency_id WHERE p.id=$1 AND p.is_active`, [id])).rows[0];
    if (!r) throw notFound('Property');
    return r;
  });
};
