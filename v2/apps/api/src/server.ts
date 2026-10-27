import { loadConfig } from './config.js';
import { createPool } from './db.js';
import { buildApp } from './app.js';

const cfg = loadConfig();
const db = createPool(cfg.DATABASE_URL);
const app = await buildApp(cfg, db, { logger: true });
await app.listen({ port: cfg.PORT, host: '0.0.0.0' });
for (const sig of ['SIGINT', 'SIGTERM']) process.on(sig, async () => { await app.close(); await db.end(); process.exit(0); });
