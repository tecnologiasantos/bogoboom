'use client';
import { Suspense, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { api, setToken, type Role } from '@/lib/api';
import { Msg, msg } from '@/components/Shell';

function Form() {
  const r = useRouter();
  const q = useSearchParams().get('role');
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<Role>(q === 'hr' || q === 'agency' ? q : 'newcomer');
  const [err, setErr] = useState<string | null>(null);
  const go = async (e: React.FormEvent) => {
    e.preventDefault(); setErr(null);
    try { const d = await api('/auth/dev-login', { method: 'POST', body: { email, role } }); setToken(d.token); r.push(`/${d.user.role}`); }
    catch (x) { setErr(msg(x)); }
  };
  return (
    <main style={{ maxWidth: 420 }}>
      <h1>Log in</h1>
      <p className="muted">Development sign-in (e-mail only). Production uses Amazon Cognito / Google — see issue “Replace password auth with Cognito”.</p>
      <form className="card" onSubmit={go}>
        <label htmlFor="e">E-mail</label><input id="e" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
        <label htmlFor="r">I am a…</label>
        <select id="r" value={role} onChange={(e) => setRole(e.target.value as Role)}>
          <option value="newcomer">Newcomer</option><option value="hr">HR department</option><option value="agency">Real estate agency</option>
        </select>
        <p><button className="btn">Continue</button></p><Msg error={err} />
      </form>
    </main>
  );
}
export default function Login() { return <Suspense><Form /></Suspense>; }
