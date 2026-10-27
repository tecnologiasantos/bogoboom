import pg from 'pg';
import { migrate } from '../src/migrate.js';

export default async function setup() {
  const url = process.env.TEST_DATABASE_URL;
  if (!url) throw new Error('Set TEST_DATABASE_URL (a throw-away Postgres database; it is WIPED on every run)');
  const db = new pg.Pool({ connectionString: url });
  await db.query('DROP SCHEMA public CASCADE; CREATE SCHEMA public;');
  await migrate(db);
  await db.end();
}
