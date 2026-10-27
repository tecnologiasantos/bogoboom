import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { boot, newcomerBody, propertyBody } from './helpers.js';

let t: Awaited<ReturnType<typeof boot>>;
beforeAll(async () => { t = await boot(); });
afterAll(async () => { await t.close(); });

let seq = 0;
const uniq = () => `Ciudad${++seq}x${Date.now() % 100000}`;

async function setupMarket(city = uniq()) {
  const hr = await t.as('hr');
  expect((await hr.call('POST', '/hr/company', { name: 'Acme BPO', industry: 'BPO', city: 'Bogotá', country: 'CO', full_name: 'Hanna HR' })).statusCode).toBe(201);
  const agencyA = await t.as('agency');
  await agencyA.call('POST', '/agency', { name: 'Casas Bogotá', city: 'Bogotá', country: 'CO', full_name: 'Ana' });
  expect((await agencyA.call('POST', '/agency/properties', propertyBody({ city }))).statusCode).toBe(201);
  const agencyB = await t.as('agency');
  await agencyB.call('POST', '/agency', { name: 'Medellín Homes', city: 'Medellín', country: 'CO', full_name: 'Beto' });
  await agencyB.call('POST', '/agency/properties', propertyBody({ city: city + '-other', district: 'Poblado' }));
  return { hr, agencyA, agencyB, city };
}

describe('auth & roles', () => {
  it('rejects missing/invalid tokens and wrong roles', async () => {
    expect((await t.app.inject({ method: 'GET', url: '/me' })).statusCode).toBe(401);
    expect((await t.app.inject({ method: 'GET', url: '/me', headers: { authorization: 'Bearer nope' } })).statusCode).toBe(401);
    const nc = await t.as('newcomer');
    expect((await nc.call('GET', '/agency/leads')).statusCode).toBe(403);
    expect((await nc.call('POST', '/admin/agencies/00000000-0000-0000-0000-000000000000/verify')).statusCode).toBe(403);
  });
  it('cannot self-register as admin', async () => {
    const r = await t.app.inject({ method: 'POST', url: '/auth/dev-login', payload: { email: 'x@y.co', role: 'admin' } });
    expect(r.statusCode).toBe(400);
  });
});

describe('lead privacy (server-side blur)', () => {
  it('never leaks contact data to agencies that have not unlocked', async () => {
    const { hr, agencyA, agencyB, city } = await setupMarket();
    const body = newcomerBody({ city_of_contract: city });
    expect((await hr.call('POST', '/hr/newcomers', body)).statusCode).toBe(201);

    const list = await agencyA.call('GET', '/agency/leads');
    expect(list.statusCode).toBe(200);
    const leads = list.json();
    expect(leads.length).toBe(1);
    expect(leads[0].match_score).toBeGreaterThanOrEqual(90);
    const raw = list.body;
    for (const secret of [body.email, body.first_name, body.full_name, body.phone, 'Acme BPO']) expect(raw).not.toContain(secret);
    expect(raw).not.toMatch(/"email"|"phone"|"full_name"|"contact"/);

    const detail = await agencyA.call('GET', `/agency/leads/${leads[0].id}`);
    for (const secret of [body.email, body.full_name, body.phone, 'Acme BPO']) expect(detail.body).not.toContain(secret);

    // an agency in another city never sees (or unlocks) it
    expect((await agencyB.call('GET', '/agency/leads')).json()).toEqual([]);
    expect((await agencyB.call('POST', `/agency/leads/${leads[0].id}/unlock`)).statusCode).toBe(404);
  });

  it('unlock uses one credit, is idempotent, then returns contact', async () => {
    const { hr, agencyA, city } = await setupMarket();
    const body = newcomerBody({ city_of_contract: city });
    await hr.call('POST', '/hr/newcomers', body);
    const id = (await agencyA.call('GET', '/agency/leads')).json()[0].id;
    const u1 = await agencyA.call('POST', `/agency/leads/${id}/unlock`);
    expect(u1.statusCode).toBe(200);
    expect(u1.json().contact.email).toBe(body.email.toLowerCase());
    expect(u1.json().credits_remaining).toBe(4);
    const u2 = await agencyA.call('POST', `/agency/leads/${id}/unlock`);
    expect(u2.json().credits_remaining).toBe(4);
    expect((await agencyA.call('GET', '/agency')).json().credits_remaining).toBe(4);
    const audit = await t.db.query("SELECT count(*)::int n FROM audit_logs WHERE action='agency.contact_unlocked'");
    expect(audit.rows[0].n).toBeGreaterThan(0);
  });

  it('blocks unlock after free credits, until a paid plan webhook arrives', async () => {
    const { hr, agencyA, city } = await setupMarket();
    for (let i = 0; i < 7; i++) await hr.call('POST', '/hr/newcomers', newcomerBody({ city_of_contract: city }));
    const leads = (await agencyA.call('GET', '/agency/leads')).json();
    expect(leads.length).toBe(7);
    for (let i = 0; i < 5; i++) expect((await agencyA.call('POST', `/agency/leads/${leads[i].id}/unlock`)).statusCode).toBe(200);
    const blocked = await agencyA.call('POST', `/agency/leads/${leads[5].id}/unlock`);
    expect(blocked.statusCode).toBe(402);
    expect(blocked.body).not.toMatch(/@secret\.example/);

    const agencyId = (await agencyA.call('GET', '/agency')).json().id;
    const payload = JSON.stringify({ agency_id: agencyId, plan: 'pro', status: 'active', provider: 'stripe', provider_subscription_id: 'sub_123',
      current_period_end: new Date(Date.now() + 30 * 864e5).toISOString() });
    const bad = await t.app.inject({ method: 'POST', url: '/webhooks/payments', payload, headers: { 'content-type': 'application/json', 'x-signature': 'ab'.repeat(32) } });
    expect(bad.statusCode).toBe(401);
    for (let i = 0; i < 2; i++) { // idempotent
      const ok = await t.app.inject({ method: 'POST', url: '/webhooks/payments', payload, headers: { 'content-type': 'application/json', 'x-signature': t.sign(payload) } });
      expect(ok.statusCode).toBe(200);
    }
    expect((await t.db.query("SELECT count(*)::int n FROM subscriptions WHERE provider_subscription_id='sub_123'")).rows[0].n).toBe(1);
    const after = await agencyA.call('POST', `/agency/leads/${leads[5].id}/unlock`);
    expect(after.statusCode).toBe(200);
    expect(after.json().credits_remaining).toBeNull();
  });

  it('respects the newcomer opting out of agency contact', async () => {
    const { hr, agencyA, city } = await setupMarket();
    const email = `claim-${Date.now()}@secret.example`;
    await hr.call('POST', '/hr/newcomers', newcomerBody({ email, city_of_contract: city }));
    const id = (await agencyA.call('GET', '/agency/leads')).json()[0].id;
    expect((await agencyA.call('POST', `/agency/leads/${id}/unlock`)).statusCode).toBe(200);

    // newcomer claims the profile created by HR and withdraws consent
    const nc = await t.as('newcomer', email);
    const save = await nc.call('PUT', '/newcomer/profile', { ...newcomerBody({ email, city_of_contract: city }), allow_contact: false });
    expect(save.statusCode).toBe(200);
    const detail = await agencyA.call('GET', `/agency/leads/${id}`);
    expect(detail.json().contact).toBeUndefined();
    expect(detail.body).not.toContain(email);
    expect(detail.json().contactable).toBe(false);
  });
});

describe('matching & notifications', () => {
  it('new listing created after the newcomer also produces a lead', async () => {
    const cali = uniq();
    const hr = await t.as('hr');
    await hr.call('POST', '/hr/company', { name: 'Late Co', full_name: 'Hana HR' });
    await hr.call('POST', '/hr/newcomers', newcomerBody({ city_of_contract: cali }));
    const ag = await t.as('agency');
    await ag.call('POST', '/agency', { name: 'Cali Casas', full_name: 'Carla' });
    expect((await ag.call('GET', '/agency/leads')).json()).toEqual([]);
    await ag.call('POST', '/agency/properties', propertyBody({ city: cali, district: 'Granada' }));
    expect((await ag.call('GET', '/agency/leads')).json().length).toBe(1);
  });
  it('moving a newcomer to moved_in withdraws the lead', async () => {
    const { hr, agencyA, city } = await setupMarket();
    const id = (await hr.call('POST', '/hr/newcomers', newcomerBody({ city_of_contract: city }))).json().id;
    expect((await agencyA.call('GET', '/agency/leads')).json().length).toBe(1);
    expect((await hr.call('PATCH', `/hr/newcomers/${id}/stage`, { stage: 'moved_in' })).statusCode).toBe(200);
    expect((await agencyA.call('GET', '/agency/leads')).json().length).toBe(0);
  });
});

describe('HR', () => {
  it('imports CSV with per-row error report and dedupes', async () => {
    const hr = await t.as('hr');
    await hr.call('POST', '/hr/company', { name: 'CSV Co', full_name: 'Hana HR' });
    const csv = [
      'first_name,full_name,email,phone,city_of_contract,arrival_date,budget_max,currency,bedrooms,furnished,preferred_districts',
      'Ana,"Ana, de la Cruz",ana@csv.example,,Bogotá,2026-12-10,2500000,COP,2,yes,Chapinero;Usaquén',
      'Bob,,not-an-email,,Bogotá,2026-12-10,,,,,',
      'Cid,,cid@csv.example,,Bogotá,2026-13-45,,,,,',
      'Dan,,ana@csv.example,,Bogotá,,,,,,',
    ].join('\r\n');
    const r = await hr.call('POST', '/hr/newcomers/import', csv, { 'content-type': 'text/csv' });
    expect(r.statusCode).toBe(200);
    const out = r.json();
    expect(out.created).toBe(1);
    expect(out.errors.map((e: any) => e.row)).toEqual([3, 4, 5]);
    const list = (await hr.call('GET', '/hr/newcomers')).json();
    expect(list.items[0].full_name).toBe('Ana, de la Cruz');
    expect(list.items[0].preferred_districts).toEqual(['Chapinero', 'Usaquén']);
  });
  it("HR cannot touch another company's newcomers", async () => {
    const a = await t.as('hr'); await a.call('POST', '/hr/company', { name: 'Alpha Co', full_name: 'Alice' });
    const b = await t.as('hr'); await b.call('POST', '/hr/company', { name: 'Beta Co', full_name: 'Bruno' });
    const id = (await a.call('POST', '/hr/newcomers', newcomerBody())).json().id;
    expect((await b.call('PATCH', `/hr/newcomers/${id}/stage`, { stage: 'cancelled' })).statusCode).toBe(404);
    expect((await b.call('GET', '/hr/newcomers')).json().items).toEqual([]);
  });
});

describe('newcomer & search', () => {
  it('searches with filters, saves favorites once, sees matches', async () => {
    const { agencyA, city } = await setupMarket();
    const nc = await t.as('newcomer');
    expect((await nc.call('PUT', '/newcomer/profile', { ...newcomerBody({ city_of_contract: city }), allow_contact: true })).statusCode).toBe(200);
    const s = (await nc.call('GET', `/properties?city=${city}&furnished=true&max_price=2000000&type=studio`)).json();
    expect(s.total).toBeGreaterThanOrEqual(1);
    expect(JSON.stringify(s)).not.toMatch(/"email"|"phone"/);
    const pid = s.items[0].id;
    await nc.call('PUT', `/favorites/${pid}`); await nc.call('PUT', `/favorites/${pid}`);
    expect((await nc.call('GET', '/favorites')).json().length).toBe(1);
    expect((await nc.call('GET', '/newcomer/matches')).json().length).toBeGreaterThanOrEqual(1);
    expect((await nc.call('GET', '/properties?type=castle')).statusCode).toBe(400);
    // another agency cannot edit this agency's listing
    const other = await t.as('agency'); await other.call('POST', '/agency', { name: 'Other', full_name: 'Olga' });
    expect((await other.call('PATCH', `/agency/properties/${pid}`, { price: 1 })).statusCode).toBe(404);
    expect((await agencyA.call('PATCH', `/agency/properties/${pid}`, { price: 1_700_000 })).statusCode).toBe(200);
  });
});
