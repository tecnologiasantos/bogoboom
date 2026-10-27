import type { FastifyPluginAsync } from 'fastify';
import { z } from 'zod';
import type { Db, Q } from '../db.js';
import { tx, audit } from '../db.js';
import type { Auth } from '../auth.js';
import { AppError, notFound } from '../errors.js';
import { hrNewcomerSchema, stageSchema, uuid } from '../schemas.js';
import { matchNewcomer } from '../matching.js';
import { parseCsv } from '../csv.js';

type NewcomerInput = z.infer<typeof hrNewcomerSchema>;

async function insertNewcomer(c: Q, company: { id: string; allow: boolean }, userId: string, d: NewcomerInput) {
  const dup = await c.query(
    `SELECT 1 FROM newcomers n JOIN newcomer_contacts k ON k.newcomer_id=n.id
      WHERE n.company_id=$1 AND n.deleted_at IS NULL AND lower(k.email)=lower($2)`, [company.id, d.email]);
  if (dup.rowCount) throw new AppError(409, 'A newcomer with this e-mail already exists in your company', 'duplicate');
  const id = (await c.query(
    `INSERT INTO newcomers (company_id, created_by, first_name, city_of_contract, arrival_date, budget_min, budget_max, currency,
       bedrooms, furnished, preferred_districts, allow_contact)
     VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
    [company.id, userId, d.first_name, d.city_of_contract, d.arrival_date ?? null, d.budget_min ?? null, d.budget_max ?? null,
      d.currency, d.bedrooms ?? null, d.furnished ?? null, d.preferred_districts, company.allow])).rows[0].id as string;
  await c.query('INSERT INTO newcomer_contacts (newcomer_id, full_name, email, phone) VALUES ($1,$2,$3,$4)',
    [id, d.full_name ?? null, d.email.toLowerCase(), d.phone ?? null]);
  await matchNewcomer(c, id);
  return id;
}

const yes = (s: string) => ['true', 'yes', 'sí', 'si', 'sim', '1'].includes(s.trim().toLowerCase());
const blank = (s: string | undefined) => (s === undefined || s.trim() === '' ? undefined : s.trim());

export const hrRoutes = (db: Db, auth: Auth): FastifyPluginAsync => async (app) => {
  const only = { preHandler: auth.requireRole('hr') };

  async function myCompany(userId: string) {
    const r = (await db.query(
      `SELECT c.id, c.name, c.allow_agency_contact AS allow FROM hr_profiles h JOIN companies c ON c.id=h.company_id WHERE h.user_id=$1`, [userId])).rows[0];
    if (!r) throw new AppError(409, 'Create your company first', 'company_required');
    return r as { id: string; name: string; allow: boolean };
  }

  app.post('/hr/company', only, async (req, reply) => {
    const b = z.object({
      name: z.string().trim().min(2).max(120), industry: z.string().trim().max(80).optional(),
      city: z.string().trim().max(80).optional(), country: z.string().length(2).toUpperCase().optional(),
      allow_agency_contact: z.boolean().default(true), full_name: z.string().trim().min(2).max(120),
    }).parse(req.body);
    const exists = await db.query('SELECT 1 FROM hr_profiles WHERE user_id=$1', [req.user!.id]);
    if (exists.rowCount) throw new AppError(409, 'Company already created', 'exists');
    const id = await tx(db, async (c) => {
      const cid = (await c.query(
        'INSERT INTO companies (name, industry, city, country, allow_agency_contact) VALUES ($1,$2,$3,$4,$5) RETURNING id',
        [b.name, b.industry ?? null, b.city ?? null, b.country ?? null, b.allow_agency_contact])).rows[0].id;
      await c.query('INSERT INTO hr_profiles (user_id, company_id, full_name) VALUES ($1,$2,$3)', [req.user!.id, cid, b.full_name]);
      return cid;
    });
    return reply.code(201).send({ id });
  });

  app.get('/hr/company', only, async (req) => myCompany(req.user!.id));

  app.post('/hr/newcomers', only, async (req, reply) => {
    const d = hrNewcomerSchema.parse(req.body);
    const co = await myCompany(req.user!.id);
    const id = await tx(db, async (c) => {
      const nid = await insertNewcomer(c, co, req.user!.id, d);
      await audit(c, req.user!.id, 'hr.newcomer_created', 'newcomers', nid);
      return nid;
    });
    return reply.code(201).send({ id });
  });

  // CSV: header row required. Columns: first_name,full_name,email,phone,city_of_contract,arrival_date,
  // budget_min,budget_max,currency,bedrooms,furnished,preferred_districts (separated by ';')
  app.post('/hr/newcomers/import', { ...only, bodyLimit: 1024 * 1024 }, async (req, reply) => {
    if (typeof req.body !== 'string') throw new AppError(415, 'Send the CSV with Content-Type: text/csv', 'csv_required');
    const co = await myCompany(req.user!.id);
    const rows = parseCsv(req.body);
    if (rows.length < 2) throw new AppError(400, 'CSV needs a header and at least one row', 'empty_csv');
    if (rows.length > 1001) throw new AppError(400, 'Maximum 1000 rows per import', 'too_many_rows');
    const header = rows[0]!.map((h) => h.trim().toLowerCase());
    for (const need of ['first_name', 'email', 'city_of_contract']) {
      if (!header.includes(need)) throw new AppError(400, `Missing column: ${need}`, 'bad_header');
    }
    const errors: { row: number; error: string }[] = [];
    let created = 0;
    for (let i = 1; i < rows.length; i++) {
      const cell = (name: string) => blank(rows[i]![header.indexOf(name)]);
      const parsed = hrNewcomerSchema.safeParse({
        first_name: cell('first_name'), full_name: cell('full_name'), email: cell('email'), phone: cell('phone'),
        city_of_contract: cell('city_of_contract'), arrival_date: cell('arrival_date'),
        budget_min: cell('budget_min'), budget_max: cell('budget_max'), currency: cell('currency'),
        bedrooms: cell('bedrooms'), furnished: cell('furnished') === undefined ? undefined : yes(cell('furnished')!),
        preferred_districts: (cell('preferred_districts') ?? '').split(';').map((x) => x.trim()).filter(Boolean),
      });
      if (!parsed.success) { errors.push({ row: i + 1, error: parsed.error.issues.map((x) => `${x.path.join('.')}: ${x.message}`).join('; ') }); continue; }
      try {
        await tx(db, (c) => insertNewcomer(c, co, req.user!.id, parsed.data));
        created++;
      } catch (e) {
        errors.push({ row: i + 1, error: e instanceof AppError ? e.message : 'could not import row' });
      }
    }
    await audit(db, req.user!.id, 'hr.csv_import', 'newcomers', undefined, { created, errors: errors.length });
    return reply.code(errors.length && !created ? 422 : 200).send({ created, errors });
  });

  app.get('/hr/newcomers', only, async (req) => {
    const co = await myCompany(req.user!.id);
    const { stage } = z.object({ stage: stageSchema.shape.stage.optional() }).parse(req.query);
    const rows = (await db.query(
      `SELECT n.id, n.first_name, k.full_name, k.email, k.phone, n.city_of_contract, n.arrival_date, n.stage, n.allow_contact,
              n.budget_min, n.budget_max, n.currency, n.bedrooms, n.furnished, n.preferred_districts,
              (n.user_id IS NOT NULL) AS claimed,
              (SELECT count(*) FROM notifications x WHERE x.newcomer_id=n.id AND x.type='match') AS interested_agencies
         FROM newcomers n LEFT JOIN newcomer_contacts k ON k.newcomer_id=n.id
        WHERE n.company_id=$1 AND n.deleted_at IS NULL AND ($2::relocation_stage IS NULL OR n.stage=$2)
        ORDER BY n.arrival_date NULLS LAST, n.created_at`, [co.id, stage ?? null])).rows;
    const summary: Record<string, number> = {};
    for (const r of rows) summary[r.stage] = (summary[r.stage] ?? 0) + 1;
    return { summary, items: rows };
  });

  app.patch('/hr/newcomers/:id/stage', only, async (req) => {
    const id = uuid.parse((req.params as any).id);
    const { stage } = stageSchema.parse(req.body);
    const co = await myCompany(req.user!.id);
    const r = await db.query(
      `UPDATE newcomers SET stage=$3, updated_at=now() WHERE id=$1 AND company_id=$2 AND deleted_at IS NULL RETURNING id`, [id, co.id, stage]);
    if (!r.rowCount) throw notFound('Newcomer');
    await audit(db, req.user!.id, 'hr.stage_changed', 'newcomers', id, { stage });
    await matchNewcomer(db, id); // moving to a final stage withdraws the lead from agencies
    return { id, stage };
  });
};
