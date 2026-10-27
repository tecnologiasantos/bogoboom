'use client';
import { useCallback, useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { Shell, Msg, msg } from '@/components/Shell';

const STAGES = ['invited', 'profile_done', 'searching', 'visiting', 'contract_signed', 'moved_in', 'cancelled'];
const TEMPLATE = 'first_name,full_name,email,phone,city_of_contract,arrival_date,budget_max,currency,bedrooms,furnished,preferred_districts\nAna,Ana Pérez,ana@example.com,,Bogotá,2026-12-01,2500000,COP,1,yes,Chapinero;Usaquén\n';

export default function HR() {
  const [company, setCompany] = useState<any>(null);
  const [data, setData] = useState<{ summary: Record<string, number>; items: any[] }>({ summary: {}, items: [] });
  const [err, setErr] = useState<string | null>(null);
  const [ok, setOk] = useState<string | null>(null);
  const [co, setCo] = useState({ name: '', industry: '', full_name: '' });
  const [n, setN] = useState({ first_name: '', email: '', city_of_contract: 'Bogotá', arrival_date: '', budget_max: '' });

  const load = useCallback(async () => {
    try { setCompany(await api('/hr/company')); setData(await api('/hr/newcomers')); } catch (x: any) { if (x.code !== 'company_required') setErr(msg(x)); }
  }, []);
  useEffect(() => { load(); }, [load]);

  const run = (fn: () => Promise<string | void>) => async (e?: React.FormEvent) => {
    e?.preventDefault(); setErr(null); setOk(null);
    try { const m = await fn(); if (m) setOk(m); await load(); } catch (x) { setErr(msg(x)); }
  };
  const createCompany = run(async () => { await api('/hr/company', { method: 'POST', body: { name: co.name, industry: co.industry || undefined, full_name: co.full_name } }); });
  const add = run(async () => {
    await api('/hr/newcomers', { method: 'POST', body: { first_name: n.first_name, email: n.email, city_of_contract: n.city_of_contract, arrival_date: n.arrival_date || null, budget_max: n.budget_max ? Number(n.budget_max) : null } });
    setN({ ...n, first_name: '', email: '' }); return 'Newcomer added';
  });
  const upload = async (file: File | undefined) => {
    if (!file) return;
    await run(async () => {
      const r = await api('/hr/newcomers/import', { method: 'POST', csv: await file.text() });
      return `Imported ${r.created} · ${r.errors.length} errors${r.errors.length ? ': ' + r.errors.slice(0, 5).map((x: any) => `row ${x.row} (${x.error})`).join('; ') : ''}`;
    })();
  };
  const setStage = async (id: string, s: string) => { await api(`/hr/newcomers/${id}/stage`, { method: 'PATCH', body: { stage: s } }).catch((x) => setErr(msg(x))); load(); };

  return (
    <Shell role="hr" title="HR dashboard">
      {!company ? (
        <form className="card" onSubmit={createCompany}>
          <h1>Create your company</h1>
          <label htmlFor="cn">Company name</label><input id="cn" required value={co.name} onChange={(e) => setCo({ ...co, name: e.target.value })} />
          <label htmlFor="ci">Industry</label><input id="ci" value={co.industry} onChange={(e) => setCo({ ...co, industry: e.target.value })} />
          <label htmlFor="cf">Your name</label><input id="cf" required value={co.full_name} onChange={(e) => setCo({ ...co, full_name: e.target.value })} />
          <p><button className="btn">Create</button></p><Msg error={err} />
        </form>
      ) : (
        <>
          <h1>{company.name} — relocations</h1>
          <p>{STAGES.map((s) => data.summary[s] ? <span key={s} className="badge" style={{ marginRight: 6 }}>{s.replace('_', ' ')}: {data.summary[s]}</span> : null)}</p>
          <Msg error={err} ok={ok} />
          <div className="grid">
            <form className="card" onSubmit={add}>
              <h2>Add a newcomer</h2>
              <label htmlFor="nf">First name</label><input id="nf" required value={n.first_name} onChange={(e) => setN({ ...n, first_name: e.target.value })} />
              <label htmlFor="ne">E-mail</label><input id="ne" type="email" required value={n.email} onChange={(e) => setN({ ...n, email: e.target.value })} />
              <label htmlFor="nc">City</label><input id="nc" required value={n.city_of_contract} onChange={(e) => setN({ ...n, city_of_contract: e.target.value })} />
              <label htmlFor="na">Arrival date</label><input id="na" type="date" value={n.arrival_date} onChange={(e) => setN({ ...n, arrival_date: e.target.value })} />
              <label htmlFor="nb">Max budget (COP)</label><input id="nb" type="number" value={n.budget_max} onChange={(e) => setN({ ...n, budget_max: e.target.value })} />
              <p><button className="btn">Add</button></p>
            </form>
            <div className="card">
              <h2>Import CSV</h2>
              <p className="muted">Up to 1000 rows. Invalid rows are reported, valid ones are imported.</p>
              <input type="file" accept=".csv,text/csv" aria-label="CSV file" onChange={(e) => upload(e.target.files?.[0])} />
              <p><a download="bogoboom-newcomers-template.csv" href={`data:text/csv;charset=utf-8,${encodeURIComponent(TEMPLATE)}`}>Download template</a></p>
            </div>
          </div>
          <div className="card" style={{ overflowX: 'auto' }}>
            <table><thead><tr><th>Name</th><th>E-mail</th><th>City</th><th>Arrival</th><th>Stage</th><th>Agencies interested</th><th>Account</th></tr></thead>
              <tbody>{data.items.map((x) => (
                <tr key={x.id}><td>{x.full_name ?? x.first_name}</td><td>{x.email}</td><td>{x.city_of_contract}</td><td>{x.arrival_date ?? '—'}</td>
                  <td><select aria-label={`Stage for ${x.first_name}`} value={x.stage} onChange={(e) => setStage(x.id, e.target.value)}>{STAGES.map((s) => <option key={s}>{s}</option>)}</select></td>
                  <td>{x.interested_agencies}</td><td>{x.claimed ? <span className="badge ok">active</span> : <span className="badge warn">invited</span>}</td></tr>))}
              </tbody></table>
            {data.items.length === 0 && <p className="muted">No newcomers yet.</p>}
          </div>
        </>
      )}
    </Shell>
  );
}
