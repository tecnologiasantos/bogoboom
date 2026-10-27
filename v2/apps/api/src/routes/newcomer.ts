import type { FastifyPluginAsync } from 'fastify';
import type { Db } from '../db.js';
import { tx, audit } from '../db.js';
import type { Auth } from '../auth.js';
import { newcomerProfileSchema, uuid } from '../schemas.js';
import { matchNewcomer, rowToNeeds, rowToProperty, scoreMatch, NOTIFY_THRESHOLD } from '../matching.js';
import { AppError, notFound } from '../errors.js';

export const newcomerRoutes = (db: Db, auth: Auth): FastifyPluginAsync => async (app) => {
  const only = { preHandler: auth.requireRole('newcomer') };

  app.put('/newcomer/profile', only, async (req) => {
    const b = newcomerProfileSchema.parse(req.body);
    const u = req.user!;
    const id = await tx(db, async (c) => {
      let row = (await c.query('SELECT id, stage FROM newcomers WHERE user_id=$1 AND deleted_at IS NULL', [u.id])).rows[0];
      if (!row) {
        // claim a profile that HR created for this e-mail (invite flow)
        row = (await c.query(
          `SELECT n.id, n.stage FROM newcomers n JOIN newcomer_contacts k ON k.newcomer_id = n.id
            WHERE n.user_id IS NULL AND n.deleted_at IS NULL AND lower(k.email) = lower($1)
            ORDER BY n.created_at LIMIT 1 FOR UPDATE OF n`, [u.email])).rows[0];
        if (row) await c.query('UPDATE newcomers SET user_id=$2 WHERE id=$1', [row.id, u.id]);
      }
      const vals = [b.first_name, b.city_of_contract, b.arrival_date ?? null, b.budget_min ?? null, b.budget_max ?? null,
        b.currency, b.bedrooms ?? null, b.furnished ?? null, b.preferred_districts, b.allow_contact];
      let nid: string;
      if (row) {
        nid = row.id;
        await c.query(
          `UPDATE newcomers SET first_name=$2, city_of_contract=$3, arrival_date=$4, budget_min=$5, budget_max=$6, currency=$7,
             bedrooms=$8, furnished=$9, preferred_districts=$10, allow_contact=$11, updated_at=now(),
             stage = CASE WHEN stage='invited' THEN 'profile_done'::relocation_stage ELSE stage END
           WHERE id=$1`, [nid, ...vals]);
      } else {
        nid = (await c.query(
          `INSERT INTO newcomers (user_id, created_by, stage, first_name, city_of_contract, arrival_date, budget_min, budget_max,
             currency, bedrooms, furnished, preferred_districts, allow_contact)
           VALUES ($1,$1,'profile_done',$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`, [u.id, ...vals])).rows[0].id;
      }
      await c.query(
        `INSERT INTO newcomer_contacts (newcomer_id, full_name, email, phone) VALUES ($1,$2,$3,$4)
         ON CONFLICT (newcomer_id) DO UPDATE SET full_name=EXCLUDED.full_name, email=EXCLUDED.email, phone=EXCLUDED.phone`,
        [nid, b.full_name ?? null, u.email, b.phone ?? null]);
      await c.query('INSERT INTO consents (user_id, purpose, granted, policy_version) VALUES ($1,$2,$3,$4)',
        [u.id, 'share_with_agencies', b.allow_contact, '2026-10']);
      await audit(c, u.id, 'newcomer.profile_saved', 'newcomers', nid);
      await matchNewcomer(c, nid);
      return nid;
    });
    return { id };
  });

  app.get('/newcomer/profile', only, async (req) => {
    const r = (await db.query(
      `SELECT n.*, k.full_name, k.phone FROM newcomers n LEFT JOIN newcomer_contacts k ON k.newcomer_id=n.id
        WHERE n.user_id=$1 AND n.deleted_at IS NULL`, [req.user!.id])).rows[0];
    if (!r) throw notFound('Profile');
    return r;
  });

  app.get('/newcomer/matches', only, async (req) => {
    const n = (await db.query('SELECT * FROM newcomers WHERE user_id=$1 AND deleted_at IS NULL', [req.user!.id])).rows[0];
    if (!n) throw new AppError(409, 'Complete your profile first', 'profile_required');
    const needs = rowToNeeds(n);
    const props = (await db.query(
      `SELECT p.*, a.name AS agency_name, (a.verified_at IS NOT NULL) AS agency_verified
         FROM properties p JOIN agencies a ON a.id=p.agency_id
        WHERE p.is_active AND lower(unaccent(p.city)) = lower(unaccent($1))`, [n.city_of_contract])).rows;
    return props.map((p) => ({ ...p, score: scoreMatch(needs, rowToProperty(p)) }))
      .filter((p) => p.score >= NOTIFY_THRESHOLD).sort((a, b) => b.score - a.score).slice(0, 50);
  });

  app.get('/favorites', only, async (req) => {
    const n = (await db.query('SELECT id FROM newcomers WHERE user_id=$1', [req.user!.id])).rows[0];
    if (!n) return [];
    return (await db.query(
      `SELECT p.* FROM favorites f JOIN properties p ON p.id=f.property_id
        WHERE f.newcomer_id=$1 ORDER BY f.created_at DESC`, [n.id])).rows;
  });

  app.put('/favorites/:propertyId', only, async (req) => {
    const pid = uuid.parse((req.params as any).propertyId);
    const n = (await db.query('SELECT id FROM newcomers WHERE user_id=$1', [req.user!.id])).rows[0];
    if (!n) throw new AppError(409, 'Complete your profile first', 'profile_required');
    const p = await db.query('SELECT 1 FROM properties WHERE id=$1 AND is_active', [pid]);
    if (!p.rowCount) throw notFound('Property');
    await db.query('INSERT INTO favorites (newcomer_id, property_id) VALUES ($1,$2) ON CONFLICT DO NOTHING', [n.id, pid]);
    return { ok: true };
  });

  app.delete('/favorites/:propertyId', only, async (req) => {
    const pid = uuid.parse((req.params as any).propertyId);
    await db.query('DELETE FROM favorites WHERE property_id=$2 AND newcomer_id IN (SELECT id FROM newcomers WHERE user_id=$1)', [req.user!.id, pid]);
    return { ok: true };
  });
};
