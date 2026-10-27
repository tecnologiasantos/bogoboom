import { createHmac } from 'node:crypto';
import { buildApp } from '../src/app.js';
import { createPool } from '../src/db.js';
import { loadConfig } from '../src/config.js';

export const SECRET = 'test-webhook-secret-123456';
export async function boot() {
  const cfg = loadConfig({
    NODE_ENV: 'test', DATABASE_URL: process.env.TEST_DATABASE_URL!, AUTH_MODE: 'dev',
    JWT_SECRET: 'x'.repeat(40), PAYMENT_WEBHOOK_SECRET: SECRET,
  } as any);
  const db = createPool(cfg.DATABASE_URL);
  const app = await buildApp(cfg, db);
  let n = 0;
  const as = async (role: 'newcomer' | 'hr' | 'agency', email = `u${++n}-${Date.now()}@test.dev`) => {
    const r = await app.inject({ method: 'POST', url: '/auth/dev-login', payload: { email, role } });
    const { token } = r.json();
    const h = { authorization: `Bearer ${token}` };
    const call = (method: any, url: string, payload?: any, headers: Record<string, string> = {}) =>
      app.inject({ method, url, payload, headers: { ...h, ...headers } });
    return { email, call, token };
  };
  const sign = (body: string) => createHmac('sha256', SECRET).update(body).digest('hex');
  return { app, db, as, sign, close: async () => { await app.close(); await db.end(); } };
}

export const newcomerBody = (over: Record<string, unknown> = {}) => ({
  first_name: 'Maria', full_name: 'Maria Secreta Gomez', email: `maria-${Math.random().toString(36).slice(2)}@secret.example`,
  phone: '+57 300 123 4567', city_of_contract: 'Bogotá', arrival_date: '2026-12-01',
  budget_max: 2_000_000, currency: 'COP', bedrooms: 1, furnished: true, preferred_districts: ['Chapinero'], ...over,
});
export const propertyBody = (over: Record<string, unknown> = {}) => ({
  title: 'Furnished studio in Chapinero', price: 1_800_000, currency: 'COP', type: 'studio', bedrooms: 1, furnished: true,
  city: 'Bogota', district: 'Chapinero', available_from: '2026-11-25', ...over,
});
