// Demo data for local development (Bogotá). Idempotent-ish: skips if demo HR user exists.
import { createPool } from './db.js';
import { migrate } from './migrate.js';
import { matchNewcomer, matchProperty } from './matching.js';

const url = process.env.DATABASE_URL;
if (!url) throw new Error('DATABASE_URL required');
if (process.env.NODE_ENV === 'production') throw new Error('Refusing to seed production');
const db = createPool(url);
await migrate(db);

if ((await db.query("SELECT 1 FROM users WHERE auth_sub='dev:hr@demo.test'")).rowCount) { console.log('already seeded'); await db.end(); process.exit(0); }

const user = async (email: string, role: string) =>
  (await db.query('INSERT INTO users (auth_sub, email, role) VALUES ($1,$2,$3) RETURNING id', [`dev:${email}`, email, role])).rows[0].id as string;

const hrU = await user('hr@demo.test', 'hr');
const co = (await db.query("INSERT INTO companies (name, industry, city, country) VALUES ('Demo BPO Colombia','BPO','Bogotá','CO') RETURNING id")).rows[0].id;
await db.query("INSERT INTO hr_profiles (user_id, company_id, full_name) VALUES ($1,$2,'Helena Demo')", [hrU, co]);

const agU = await user('agency@demo.test', 'agency');
const ag = (await db.query("INSERT INTO agencies (name, city, country, verified_at) VALUES ('Casas Chapinero','Bogotá','CO', now()) RETURNING id")).rows[0].id;
await db.query("INSERT INTO agency_profiles (user_id, agency_id, full_name) VALUES ($1,$2,'Andrés Demo')", [agU, ag]);
await user('newcomer@demo.test', 'newcomer');

const props = [
  ['Furnished studio near Parque 93', 'studio', 1_900_000, 1, true, 'Chapinero'],
  ['2BR apartment in Usaquén', 'apartment', 3_200_000, 2, false, 'Usaquén'],
  ['Room in shared house, Teusaquillo', 'room', 1_100_000, 1, true, 'Teusaquillo'],
  ['Modern 1BR, Zona T', 'apartment', 2_600_000, 1, true, 'Chapinero'],
];
for (const [title, type, price, beds, furn, district] of props) {
  const id = (await db.query(
    `INSERT INTO properties (agency_id,title,price,type,bedrooms,furnished,city,district,available_from)
     VALUES ($1,$2,$3,$4,$5,$6,'Bogotá',$7, current_date + 10) RETURNING id`, [ag, title, price, type, beds, furn, district])).rows[0].id;
  await matchProperty(db, id);
}
const people: [string, string, number, number, boolean][] = [['Lucía', 'lucia@demo.test', 2_200_000, 1, true], ['Tomás', 'tomas@demo.test', 3_500_000, 2, false], ['Priya', 'priya@demo.test', 1_300_000, 1, true]];
for (const [name, email, max, beds, furn] of people) {
  const id = (await db.query(
    `INSERT INTO newcomers (company_id, created_by, first_name, city_of_contract, arrival_date, budget_max, bedrooms, furnished, allow_contact, preferred_districts)
     VALUES ($1,$2,$3,'Bogotá', current_date + 30, $4,$5,$6,true, ARRAY['Chapinero']) RETURNING id`, [co, hrU, name, max, beds, furn])).rows[0].id;
  await db.query('INSERT INTO newcomer_contacts (newcomer_id, full_name, email) VALUES ($1,$2,$3)', [id, `${name} Demo`, email]);
  await matchNewcomer(db, id);
}
console.log('Seeded. Log in (dev mode) as hr@demo.test (hr), agency@demo.test (agency), newcomer@demo.test (newcomer).');
await db.end();
