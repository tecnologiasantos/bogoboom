import type { FastifyPluginAsync } from 'fastify';
import { z } from 'zod';
import type { Db } from '../db.js';
import { tx, audit } from '../db.js';
import type { Auth } from '../auth.js';
import { AppError, notFound } from '../errors.js';
import { propertySchema, uuid } from '../schemas.js';
import { matchProperty } from '../matching.js';

const FREE_CREDITS = 5;

const LEAD_SELECT = `
  SELECT n.id, nt.match_score, nt.created_at AS notified_at, n.city_of_contract, date_trunc('month', n.arrival_date)::date AS arrival_month,
         n.budget_min, n.budget_max, n.currency, n.bedrooms, n.furnished, n.preferred_districts, n.allow_contact, n.stage,
         co.industry AS company_industry, co.name AS company_name, (cu.id IS NOT NULL) AS unlocked,
         k.full_name, k.email, k.phone
    FROM notifications nt
    JOIN newcomers n ON n.id = nt.newcomer_id AND n.deleted_at IS NULL
    LEFT JOIN companies co ON co.id = n.company_id
    LEFT JOIN contact_unlocks cu ON cu.agency_id = nt.agency_id AND cu.newcomer_id = n.id
    LEFT JOIN newcomer_contacts k ON k.newcomer_id = n.id
   WHERE nt.agency_id = $1 AND nt.type = 'match'`;

/** Strips every identifying field. This is the ONLY place lead rows are serialised for agencies. */
function blurred(r: any) {
  return {
    id: r.id, match_score: r.match_score, notified_at: r.notified_at, city: r.city_of_contract, arrival_month: r.arrival_month,
    budget_min: r.budget_min, budget_max: r.budget_max, currency: r.currency, bedrooms: r.bedrooms, furnished: r.furnished,
    preferred_districts: r.preferred_districts, company_industry: r.company_industry, stage: r.stage,
    contactable: r.allow_contact, unlocked: r.unlocked,
  };
}
const withContact = (r: any) => ({
  ...blurred(r), contact: { full_name: r.full_name, email: r.email, phone: r.phone }, company_name: r.company_name,
});

export const agencyRoutes = (db: Db, auth: Auth): FastifyPluginAsync => async (app) => {
  const only = { preHandler: auth.requireRole('agency') };

  async function myAgencyId(userId: string) {
    const r = (await db.query('SELECT agency_id FROM agency_profiles WHERE user_id=$1', [userId])).rows[0];
    if (!r) throw new AppError(409, 'Create your agency first', 'agency_required');
    return r.agency_id as string;
  }

  app.post('/agency', only, async (req, reply) => {
    const b = z.object({ name: z.string().trim().min(2).max(120), city: z.string().trim().max(80).optional(),
      country: z.string().length(2).toUpperCase().optional(), description: z.string().max(2000).optional(),
      full_name: z.string().trim().min(2).max(120) }).parse(req.body);
    if ((await db.query('SELECT 1 FROM agency_profiles WHERE user_id=$1', [req.user!.id])).rowCount) throw new AppError(409, 'Agency already created', 'exists');
    const id = await tx(db, async (c) => {
      const aid = (await c.query('INSERT INTO agencies (name, city, country, description) VALUES ($1,$2,$3,$4) RETURNING id',
        [b.name, b.city ?? null, b.country ?? null, b.description ?? null])).rows[0].id;
      await c.query('INSERT INTO agency_profiles (user_id, agency_id, full_name) VALUES ($1,$2,$3)', [req.user!.id, aid, b.full_name]);
      return aid;
    });
    return reply.code(201).send({ id });
  });

  app.get('/agency', only, async (req) => {
    const id = await myAgencyId(req.user!.id);
    const a = (await db.query(
      `SELECT id, name, city, country, description, verified_at, plan, plan_active_until, contact_credits, credits_period FROM agencies WHERE id=$1`, [id])).rows[0];
    const thisMonth = new Date().toISOString().slice(0, 7);
    const periodOk = String(a.credits_period instanceof Date ? a.credits_period.toISOString() : a.credits_period).slice(0, 7) === thisMonth;
    const planActive = a.plan !== 'free' && a.plan_active_until && new Date(a.plan_active_until) > new Date();
    return { ...a, plan_active: !!planActive, credits_remaining: periodOk ? a.contact_credits : FREE_CREDITS };
  });

  // ---------- properties ----------
  app.post('/agency/properties', only, async (req, reply) => {
    const d = propertySchema.parse(req.body);
    const aid = await myAgencyId(req.user!.id);
    const id = await tx(db, async (c) => {
      const pid = (await c.query(
        `INSERT INTO properties (agency_id,title,description,price,currency,type,bedrooms,furnished,city,district,address,lat,lng,available_from)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`,
        [aid, d.title, d.description ?? null, d.price, d.currency, d.type, d.bedrooms ?? null, d.furnished, d.city, d.district ?? null,
          d.address ?? null, d.lat ?? null, d.lng ?? null, d.available_from ?? null])).rows[0].id;
      await matchProperty(c, pid);
      return pid;
    });
    return reply.code(201).send({ id });
  });

  app.get('/agency/properties', only, async (req) => {
    const aid = await myAgencyId(req.user!.id);
    return (await db.query('SELECT * FROM properties WHERE agency_id=$1 ORDER BY created_at DESC', [aid])).rows;
  });

  app.patch('/agency/properties/:id', only, async (req) => {
    const id = uuid.parse((req.params as any).id);
    const aid = await myAgencyId(req.user!.id);
    const d = propertySchema.partial().extend({ is_active: z.boolean().optional() }).parse(req.body);
    const cols = ['title', 'description', 'price', 'currency', 'type', 'bedrooms', 'furnished', 'city', 'district', 'address', 'lat', 'lng', 'available_from', 'is_active'] as const;
    const sets: string[] = [], vals: unknown[] = [id, aid];
    for (const k of cols) if (k in d) { vals.push((d as any)[k] ?? null); sets.push(`${k}=$${vals.length}`); }
    if (!sets.length) throw new AppError(400, 'Nothing to update', 'empty');
    const r = await db.query(`UPDATE properties SET ${sets.join(', ')}, updated_at=now() WHERE id=$1 AND agency_id=$2 RETURNING id`, vals);
    if (!r.rowCount) throw notFound('Property');
    await matchProperty(db, id);
    return { id };
  });

  app.delete('/agency/properties/:id', only, async (req) => {
    const id = uuid.parse((req.params as any).id);
    const aid = await myAgencyId(req.user!.id);
    const r = await db.query('UPDATE properties SET is_active=false, updated_at=now() WHERE id=$1 AND agency_id=$2', [id, aid]);
    if (!r.rowCount) throw notFound('Property');
    return { ok: true };
  });

  // ---------- leads ----------
  app.get('/agency/leads', only, async (req) => {
    const aid = await myAgencyId(req.user!.id);
    const { min_score } = z.object({ min_score: z.coerce.number().min(0).max(100).default(0) }).parse(req.query);
    const rows = (await db.query(`${LEAD_SELECT} AND nt.match_score >= $2 ORDER BY nt.match_score DESC, nt.created_at DESC LIMIT 100`, [aid, min_score])).rows;
    return rows.map(blurred); // never contains contact fields
  });

  app.get('/agency/leads/:id', only, async (req) => {
    const id = uuid.parse((req.params as any).id);
    const aid = await myAgencyId(req.user!.id);
    const r = (await db.query(`${LEAD_SELECT} AND n.id = $2`, [aid, id])).rows[0];
    if (!r) throw notFound('Lead');
    if (r.unlocked && r.allow_contact) {
      await audit(db, req.user!.id, 'agency.contact_viewed', 'newcomers', id);
      return withContact(r);
    }
    return blurred(r);
  });

  app.post('/agency/leads/:id/unlock', { ...only, config: { rateLimit: { max: 30, timeWindow: '1 minute' } } }, async (req) => {
    const id = uuid.parse((req.params as any).id);
    const aid = await myAgencyId(req.user!.id);
    return tx(db, async (c) => {
      const a = (await c.query('SELECT * FROM agencies WHERE id=$1 FOR UPDATE', [aid])).rows[0];
      // the lead must have been matched to THIS agency (prevents id enumeration)
      const lead = (await c.query(`${LEAD_SELECT} AND n.id = $2`, [aid, id])).rows[0];
      if (!lead) throw notFound('Lead');
      if (!lead.allow_contact) throw new AppError(403, 'This person has not agreed to be contacted by agencies', 'contact_not_allowed');
      const planActive = a.plan !== 'free' && a.plan_active_until && new Date(a.plan_active_until) > new Date();
      let credits = a.contact_credits as number;
      const month = new Date().toISOString().slice(0, 7) + '-01';
      const period = String(a.credits_period instanceof Date ? a.credits_period.toISOString() : a.credits_period).slice(0, 10);
      if (period.slice(0, 7) !== month.slice(0, 7)) {
        credits = FREE_CREDITS;
        await c.query('UPDATE agencies SET contact_credits=$2, credits_period=$3 WHERE id=$1', [aid, credits, month]);
      }
      if (!lead.unlocked) {
        let usedCredit = false;
        if (!planActive) {
          if (credits <= 0) throw new AppError(402, 'No contact credits left this month. Upgrade to Pro for unlimited contacts.', 'payment_required');
          credits -= 1; usedCredit = true;
          await c.query('UPDATE agencies SET contact_credits=$2 WHERE id=$1', [aid, credits]);
        }
        await c.query('INSERT INTO contact_unlocks (agency_id, newcomer_id, unlocked_by, used_credit) VALUES ($1,$2,$3,$4)', [aid, id, req.user!.id, usedCredit]);
        await audit(c, req.user!.id, 'agency.contact_unlocked', 'newcomers', id, { usedCredit });
      }
      return { ...withContact({ ...lead, unlocked: true }), credits_remaining: planActive ? null : credits };
    });
  });

  app.get('/agency/notifications', only, async (req) => {
    const aid = await myAgencyId(req.user!.id);
    return (await db.query(
      `SELECT id, type, match_score, message, read_at, created_at FROM notifications WHERE agency_id=$1 ORDER BY created_at DESC LIMIT 50`, [aid])).rows;
  });
};
