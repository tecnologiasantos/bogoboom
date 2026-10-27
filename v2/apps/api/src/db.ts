import pg from 'pg';

pg.types.setTypeParser(1082, (v: string) => v); // DATE -> 'YYYY-MM-DD' (no timezone shifts)

export type Db = pg.Pool;
export type Q = Pick<pg.Pool | pg.PoolClient, 'query'>;

export function createPool(url: string): Db {
  return new pg.Pool({ connectionString: url, max: 10 });
}

export async function tx<T>(db: Db, fn: (c: pg.PoolClient) => Promise<T>): Promise<T> {
  const c = await db.connect();
  try {
    await c.query('BEGIN');
    const r = await fn(c);
    await c.query('COMMIT');
    return r;
  } catch (e) {
    await c.query('ROLLBACK').catch(() => {});
    throw e;
  } finally {
    c.release();
  }
}

export async function audit(q: Q, userId: string | null, action: string, table?: string, recordId?: string, changes?: unknown) {
  await q.query(
    'INSERT INTO audit_logs (user_id, action, table_name, record_id, changes) VALUES ($1,$2,$3,$4,$5)',
    [userId, action, table ?? null, recordId ?? null, changes ? JSON.stringify(changes) : null],
  );
}
