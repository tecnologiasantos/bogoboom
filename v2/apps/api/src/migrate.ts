import { readdirSync, readFileSync, existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import type { Db } from './db.js';
import { createPool } from './db.js';

const here = dirname(fileURLToPath(import.meta.url));

function migrationsDir() {
  for (const p of [join(here, '../../../db/migrations'), join(here, 'db/migrations'), join(here, '../db/migrations')]) {
    if (existsSync(p)) return p;
  }
  throw new Error('db/migrations not found');
}

export async function migrate(db: Db, log: (m: string) => void = () => {}) {
  await db.query('CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())');
  const dir = migrationsDir();
  const done = new Set((await db.query('SELECT name FROM schema_migrations')).rows.map((r) => r.name as string));
  for (const f of readdirSync(dir).filter((x) => x.endsWith('.sql')).sort()) {
    if (done.has(f)) continue;
    const c = await db.connect();
    try {
      await c.query('BEGIN');
      await c.query(readFileSync(join(dir, f), 'utf8'));
      await c.query('INSERT INTO schema_migrations (name) VALUES ($1)', [f]);
      await c.query('COMMIT');
      log(`applied ${f}`);
    } catch (e) {
      await c.query('ROLLBACK');
      throw new Error(`migration ${f} failed: ${(e as Error).message}`);
    } finally {
      c.release();
    }
  }
}

if (process.argv[1] && fileURLToPath(import.meta.url) === process.argv[1]) {
  const url = process.env.DATABASE_URL;
  if (!url) throw new Error('DATABASE_URL required');
  const db = createPool(url);
  migrate(db, console.log).then(() => db.end()).catch((e) => { console.error(e); process.exit(1); });
}
