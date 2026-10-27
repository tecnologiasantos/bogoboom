'use client';
import { useCallback, useEffect, useState } from 'react';
import { api, money } from '@/lib/api';
import { Shell, Msg, msg } from '@/components/Shell';

export default function Newcomer() {
  const [f, setF] = useState({ first_name: '', full_name: '', phone: '', city_of_contract: 'Bogotá', arrival_date: '', budget_max: '', bedrooms: '', furnished: '', districts: '', allow_contact: false });
  const [err, setErr] = useState<string | null>(null);
  const [ok, setOk] = useState<string | null>(null);
  const [matches, setMatches] = useState<any[]>([]);
  const [search, setSearch] = useState({ city: 'Bogotá', max_price: '', type: '' });
  const [results, setResults] = useState<{ total: number; items: any[] }>({ total: 0, items: [] });
  const [favs, setFavs] = useState<Set<string>>(new Set());
  const set = (k: string, v: unknown) => setF((p) => ({ ...p, [k]: v }));

  const refresh = useCallback(async () => {
    try {
      const p = await api('/newcomer/profile');
      setF({ first_name: p.first_name, full_name: p.full_name ?? '', phone: p.phone ?? '', city_of_contract: p.city_of_contract, arrival_date: p.arrival_date ?? '',
        budget_max: p.budget_max ?? '', bedrooms: p.bedrooms ?? '', furnished: p.furnished == null ? '' : String(p.furnished), districts: (p.preferred_districts ?? []).join(', '), allow_contact: p.allow_contact });
      setMatches(await api('/newcomer/matches'));
    } catch { /* no profile yet */ }
    setFavs(new Set((await api('/favorites')).map((x: any) => x.id)));
  }, []);
  useEffect(() => { refresh(); }, [refresh]);

  const save = async (e: React.FormEvent) => {
    e.preventDefault(); setErr(null); setOk(null);
    try {
      await api('/newcomer/profile', { method: 'PUT', body: {
        first_name: f.first_name, full_name: f.full_name || undefined, phone: f.phone || undefined, city_of_contract: f.city_of_contract,
        arrival_date: f.arrival_date || null, budget_max: f.budget_max ? Number(f.budget_max) : null, bedrooms: f.bedrooms === '' ? null : Number(f.bedrooms),
        furnished: f.furnished === '' ? null : f.furnished === 'true', preferred_districts: f.districts.split(',').map((x) => x.trim()).filter(Boolean), allow_contact: f.allow_contact } });
      setOk('Profile saved'); refresh();
    } catch (x) { setErr(msg(x)); }
  };
  const find = async (e?: React.FormEvent) => {
    e?.preventDefault(); setErr(null);
    const qs = new URLSearchParams(Object.entries(search).filter(([, v]) => v) as [string, string][]);
    try { setResults(await api(`/properties?${qs}`)); } catch (x) { setErr(msg(x)); }
  };
  const toggle = async (id: string) => {
    await api(`/favorites/${id}`, { method: favs.has(id) ? 'DELETE' : 'PUT' });
    setFavs((s) => { const n = new Set(s); n.has(id) ? n.delete(id) : n.add(id); return n; });
  };
  const Card = ({ p, score }: { p: any; score?: number }) => (
    <div className="card"><strong>{p.title}</strong> {score != null && <span className="badge ok">{score}% match</span>} {p.agency_verified && <span className="badge ok">Verified agency</span>}
      <div className="muted">{p.type} · {p.bedrooms ?? '?'} bd · {p.furnished ? 'furnished' : 'unfurnished'} · {p.district ?? p.city}</div>
      <div><b>{money(p.price, p.currency)}</b> /month</div>
      <button className="btn sec" aria-pressed={favs.has(p.id)} onClick={() => toggle(p.id)}>{favs.has(p.id) ? '♥ Saved' : '♡ Save'}</button></div>
  );
  return (
    <Shell role="newcomer" title="Newcomer">
      <h1>Your relocation profile</h1>
      <form className="card" onSubmit={save}>
        <div className="grid">
          <div><label htmlFor="fn">First name</label><input id="fn" required value={f.first_name} onChange={(e) => set('first_name', e.target.value)} /></div>
          <div><label htmlFor="ful">Full name (shared only after you agree)</label><input id="ful" value={f.full_name} onChange={(e) => set('full_name', e.target.value)} /></div>
          <div><label htmlFor="ph">Phone / WhatsApp</label><input id="ph" value={f.phone} onChange={(e) => set('phone', e.target.value)} /></div>
          <div><label htmlFor="c">City of contract</label><input id="c" required value={f.city_of_contract} onChange={(e) => set('city_of_contract', e.target.value)} /></div>
          <div><label htmlFor="ad">Arrival date</label><input id="ad" type="date" value={f.arrival_date} onChange={(e) => set('arrival_date', e.target.value)} /></div>
          <div><label htmlFor="b">Max budget / month (COP)</label><input id="b" type="number" min="0" value={f.budget_max} onChange={(e) => set('budget_max', e.target.value)} /></div>
          <div><label htmlFor="bd">Bedrooms</label><input id="bd" type="number" min="0" value={f.bedrooms} onChange={(e) => set('bedrooms', e.target.value)} /></div>
          <div><label htmlFor="fu">Furnished</label><select id="fu" value={f.furnished} onChange={(e) => set('furnished', e.target.value)}><option value="">No preference</option><option value="true">Yes</option><option value="false">No</option></select></div>
          <div><label htmlFor="di">Preferred districts (comma separated)</label><input id="di" value={f.districts} onChange={(e) => set('districts', e.target.value)} /></div>
        </div>
        <label style={{ display: 'flex', gap: 8, alignItems: 'center', fontSize: '1rem', color: 'inherit' }}>
          <input type="checkbox" style={{ width: 'auto' }} checked={f.allow_contact} onChange={(e) => set('allow_contact', e.target.checked)} />
          Real estate agencies may contact me (I can change this at any time)
        </label>
        <p><button className="btn">Save profile</button></p><Msg error={err} ok={ok} />
      </form>

      {matches.length > 0 && <><h2>Best matches for you</h2><div className="grid">{matches.slice(0, 6).map((p) => <Card key={p.id} p={p} score={p.score} />)}</div></>}

      <h2>Search homes</h2>
      <form className="card" onSubmit={find}>
        <div className="grid">
          <div><label htmlFor="sc">City</label><input id="sc" value={search.city} onChange={(e) => setSearch({ ...search, city: e.target.value })} /></div>
          <div><label htmlFor="sp">Max price</label><input id="sp" type="number" value={search.max_price} onChange={(e) => setSearch({ ...search, max_price: e.target.value })} /></div>
          <div><label htmlFor="st">Type</label><select id="st" value={search.type} onChange={(e) => setSearch({ ...search, type: e.target.value })}><option value="">Any</option><option>apartment</option><option>studio</option><option>room</option><option>house</option></select></div>
        </div>
        <p><button className="btn">Search</button> <span className="muted">{results.total} results</span></p>
      </form>
      <div className="grid">{results.items.map((p) => <Card key={p.id} p={p} />)}</div>
    </Shell>
  );
}
