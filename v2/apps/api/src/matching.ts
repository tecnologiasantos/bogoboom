import type { Q } from './db.js';

export interface NewcomerNeeds {
  city: string;
  arrival_date: string | null;
  budget_min: number | null;
  budget_max: number | null;
  bedrooms: number | null;
  furnished: boolean | null;
  districts: string[];
}
export interface PropertyFacts {
  city: string;
  district: string | null;
  price: number;
  bedrooms: number | null;
  furnished: boolean;
  available_from: string | null;
}

export const NOTIFY_THRESHOLD = 50;
const norm = (s: string | null | undefined) =>
  (s ?? '').normalize('NFD').replace(/[̀-ͯ]/g, '').trim().toLowerCase();
const DAY = 86_400_000;

/**
 * Rule-based match score 0–100 (see docs/matching.md).
 * Hard rejects (score 0): different city, or price > 125% of the max budget.
 * Weights: budget 35 · availability 20 · bedrooms 15 · furnished 15 · district 15.
 */
export function scoreMatch(n: NewcomerNeeds, p: PropertyFacts): number {
  if (norm(n.city) !== norm(p.city)) return 0;
  let score = 0;

  // budget (35)
  if (n.budget_max == null) score += 25;
  else if (p.price > n.budget_max * 1.25) return 0;
  else if (p.price <= n.budget_max) score += 35;
  else if (p.price <= n.budget_max * 1.1) score += 15;
  else score += 5;

  // availability (20)
  if (!n.arrival_date || !p.available_from) score += 20;
  else {
    const delta = (Date.parse(p.available_from) - Date.parse(n.arrival_date)) / DAY;
    if (delta <= 14) score += 20;
    else if (delta <= 30) score += 8;
  }

  // bedrooms (15)
  if (n.bedrooms == null || p.bedrooms == null) score += 15;
  else if (p.bedrooms >= n.bedrooms) score += 15;
  else if (p.bedrooms === n.bedrooms - 1) score += 5;

  // furnished (15)
  if (n.furnished == null) score += 15;
  else if (n.furnished === p.furnished) score += 15;
  else if (!n.furnished && p.furnished) score += 10;

  // district (15)
  if (n.districts.length === 0) score += 8;
  else if (p.district && n.districts.map(norm).includes(norm(p.district))) score += 15;

  return Math.min(100, score);
}

const num = (v: unknown) => (v == null ? null : Number(v));
const iso = (v: unknown) => (v == null ? null : v instanceof Date ? v.toISOString().slice(0, 10) : String(v).slice(0, 10));

export function rowToNeeds(r: any): NewcomerNeeds {
  return {
    city: r.city_of_contract, arrival_date: iso(r.arrival_date),
    budget_min: num(r.budget_min), budget_max: num(r.budget_max),
    bedrooms: r.bedrooms, furnished: r.furnished, districts: r.preferred_districts ?? [],
  };
}
export function rowToProperty(r: any): PropertyFacts {
  return { city: r.city, district: r.district, price: Number(r.price), bedrooms: r.bedrooms, furnished: r.furnished, available_from: iso(r.available_from) };
}

const ACTIVE_STAGES = ['invited', 'profile_done', 'searching', 'visiting'];

/** Recompute agency notifications for one newcomer (after create/update). */
export async function matchNewcomer(q: Q, newcomerId: string) {
  const n = (await q.query('SELECT * FROM newcomers WHERE id=$1 AND deleted_at IS NULL', [newcomerId])).rows[0];
  if (!n) return;
  const needs = rowToNeeds(n);
  const props = (await q.query(
    `SELECT * FROM properties WHERE is_active AND lower(unaccent(city)) = lower(unaccent($1))`, [n.city_of_contract])).rows;
  const best = new Map<string, number>();
  if (ACTIVE_STAGES.includes(n.stage)) {
    for (const p of props) {
      const s = scoreMatch(needs, rowToProperty(p));
      if (s > (best.get(p.agency_id) ?? 0)) best.set(p.agency_id, s);
    }
  }
  const keep: string[] = [];
  for (const [agencyId, s] of best) {
    if (s < NOTIFY_THRESHOLD) continue;
    keep.push(agencyId);
    await q.query(
      `INSERT INTO notifications (agency_id, newcomer_id, type, match_score, message)
       VALUES ($1,$2,'match',$3,$4)
       ON CONFLICT (agency_id, newcomer_id, type) DO UPDATE SET match_score = EXCLUDED.match_score, message = EXCLUDED.message`,
      [agencyId, newcomerId, s, `New ${s}% match in ${n.city_of_contract}`]);
  }
  await q.query(
    `DELETE FROM notifications WHERE newcomer_id=$1 AND type='match' AND NOT (agency_id = ANY($2::uuid[]))`,
    [newcomerId, keep]);
}

/** Notify one agency about newcomers that match a (new/updated) property. */
export async function matchProperty(q: Q, propertyId: string) {
  const p = (await q.query('SELECT * FROM properties WHERE id=$1 AND is_active', [propertyId])).rows[0];
  if (!p) return;
  const facts = rowToProperty(p);
  const ns = (await q.query(
    `SELECT * FROM newcomers WHERE deleted_at IS NULL AND stage = ANY($2::relocation_stage[])
       AND lower(unaccent(city_of_contract)) = lower(unaccent($1))`, [p.city, ACTIVE_STAGES])).rows;
  for (const n of ns) {
    const s = scoreMatch(rowToNeeds(n), facts);
    if (s < NOTIFY_THRESHOLD) continue;
    await q.query(
      `INSERT INTO notifications (agency_id, newcomer_id, type, match_score, message)
       VALUES ($1,$2,'match',$3,$4)
       ON CONFLICT (agency_id, newcomer_id, type) DO UPDATE
         SET match_score = GREATEST(notifications.match_score, EXCLUDED.match_score)`,
      [p.agency_id, n.id, s, `New ${s}% match in ${p.city}`]);
  }
}
