'use client';
import { useCallback, useEffect, useState } from 'react';
import { api, money } from '@/lib/api';
import { Shell, Msg, msg } from '@/components/Shell';

export default function Agency() {
  const [agency, setAgency] = useState<any>(null);
  const [leads, setLeads] = useState<any[]>([]);
  const [props, setProps] = useState<any[]>([]);
  const [contacts, setContacts] = useState<Record<string, any>>({});
  const [err, setErr] = useState<string | null>(null);
  const [form, setForm] = useState({ name: '', full_name: '', city: 'Bogotá' });
  const [p, setP] = useState({ title: '', price: '', type: 'apartment', bedrooms: '1', furnished: false, city: 'Bogotá', district: '', available_from: '' });

  const load = useCallback(async () => {
    try {
      setAgency(await api('/agency')); setLeads(await api('/agency/leads')); setProps(await api('/agency/properties'));
    } catch (x: any) { if (x.code !== 'agency_required') setErr(msg(x)); }
  }, []);
  useEffect(() => { load(); }, [load]);

  const create = async (e: React.FormEvent) => { e.preventDefault(); setErr(null); try { await api('/agency', { method: 'POST', body: { name: form.name, full_name: form.full_name, city: form.city } }); load(); } catch (x) { setErr(msg(x)); } };
  const addProp = async (e: React.FormEvent) => {
    e.preventDefault(); setErr(null);
    try {
      await api('/agency/properties', { method: 'POST', body: { title: p.title, price: Number(p.price), type: p.type, bedrooms: Number(p.bedrooms), furnished: p.furnished, city: p.city, district: p.district || null, available_from: p.available_from || null } });
      setP({ ...p, title: '', price: '' }); load();
    } catch (x) { setErr(msg(x)); }
  };
  const unlock = async (id: string) => {
    setErr(null);
    try { const r = await api(`/agency/leads/${id}/unlock`, { method: 'POST' }); setContacts((c) => ({ ...c, [id]: r })); load(); }
    catch (x: any) { setErr(x.status === 402 ? 'No free contacts left this month — upgrade to Pro for unlimited contacts.' : msg(x)); }
  };

  return (
    <Shell role="agency" title="Agency">
      {!agency ? (
        <form className="card" onSubmit={create}>
          <h1>Register your agency (free)</h1>
          <label htmlFor="an">Agency name</label><input id="an" required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          <label htmlFor="af">Your name</label><input id="af" required value={form.full_name} onChange={(e) => setForm({ ...form, full_name: e.target.value })} />
          <label htmlFor="ac">City</label><input id="ac" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
          <p><button className="btn">Create agency</button></p><Msg error={err} />
        </form>
      ) : (
        <>
          <h1>{agency.name} {agency.verified_at ? <span className="badge ok">Verified</span> : <span className="badge warn">Verification pending</span>}</h1>
          <p>Plan: <b>{agency.plan_active ? agency.plan : 'Free'}</b> · {agency.plan_active ? 'unlimited contacts' : `${agency.credits_remaining} free contacts left this month`}</p>
          <Msg error={err} />

          <h2>Matched leads</h2>
          <div className="card" style={{ overflowX: 'auto' }}>
            <table><thead><tr><th>Match</th><th>Who</th><th>City / arrival</th><th>Budget</th><th>Needs</th><th></th></tr></thead>
              <tbody>{leads.map((l) => {
                const c = contacts[l.id];
                return (
                  <tr key={l.id}>
                    <td><span className="badge ok">{l.match_score}%</span></td>
                    <td>{c ? <><b>{c.contact.full_name ?? '—'}</b><br />{c.contact.email}<br />{c.contact.phone}<br /><span className="muted">{c.company_name}</span></> : <span className="muted">Newcomer{l.company_industry ? ` · ${l.company_industry} company` : ''} <span className="blur" aria-hidden>Hidden Name</span></span>}</td>
                    <td>{l.city}<br /><span className="muted">{l.arrival_month?.slice(0, 7)}</span></td>
                    <td>{money(l.budget_max, l.currency)}</td>
                    <td>{l.bedrooms ?? '?'} bd · {l.furnished ? 'furnished' : ''} {l.preferred_districts?.join(', ')}</td>
                    <td>{c ? <span className="badge ok">Unlocked</span> : l.contactable ? <button className="btn" onClick={() => unlock(l.id)}>Unlock contact</button> : <span className="muted">Not contactable</span>}</td>
                  </tr>);
              })}</tbody></table>
            {leads.length === 0 && <p className="muted">No leads yet — add listings and matching newcomers will appear here.</p>}
          </div>

          <h2>Your listings</h2>
          <div className="grid">
            <form className="card" onSubmit={addProp}>
              <h3>Add a property</h3>
              <label htmlFor="pt">Title</label><input id="pt" required value={p.title} onChange={(e) => setP({ ...p, title: e.target.value })} />
              <label htmlFor="pp">Monthly price (COP)</label><input id="pp" type="number" min="1" required value={p.price} onChange={(e) => setP({ ...p, price: e.target.value })} />
              <label htmlFor="py">Type</label><select id="py" value={p.type} onChange={(e) => setP({ ...p, type: e.target.value })}><option>apartment</option><option>studio</option><option>room</option><option>house</option></select>
              <label htmlFor="pb">Bedrooms</label><input id="pb" type="number" min="0" value={p.bedrooms} onChange={(e) => setP({ ...p, bedrooms: e.target.value })} />
              <label htmlFor="pc">City</label><input id="pc" required value={p.city} onChange={(e) => setP({ ...p, city: e.target.value })} />
              <label htmlFor="pd">District</label><input id="pd" value={p.district} onChange={(e) => setP({ ...p, district: e.target.value })} />
              <label htmlFor="pa">Available from</label><input id="pa" type="date" value={p.available_from} onChange={(e) => setP({ ...p, available_from: e.target.value })} />
              <label style={{ display: 'flex', gap: 8, alignItems: 'center', fontSize: '1rem', color: 'inherit' }}><input type="checkbox" style={{ width: 'auto' }} checked={p.furnished} onChange={(e) => setP({ ...p, furnished: e.target.checked })} />Furnished</label>
              <p><button className="btn">Publish</button></p>
            </form>
            <div className="card"><h3>Published ({props.filter((x) => x.is_active).length})</h3>
              <ul>{props.map((x) => <li key={x.id}>{x.title} — {money(x.price, x.currency)} {!x.is_active && <span className="badge">paused</span>}</li>)}</ul></div>
          </div>
        </>
      )}
    </Shell>
  );
}
