'use client';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useEffect, useState, type ReactNode } from 'react';
import { api, getToken, setToken, type Role } from '@/lib/api';

export function Shell({ role, title, children }: { role: Role; title: string; children: ReactNode }) {
  const r = useRouter();
  const [ready, setReady] = useState(false);
  useEffect(() => {
    if (!getToken()) { r.replace('/login'); return; }
    api('/me').then((m) => { if (m.user.role !== role) r.replace('/login'); else setReady(true); }).catch(() => { setToken(null); r.replace('/login'); });
  }, [r, role]);
  return (
    <>
      <a className="skip" href="#main">Skip to content</a>
      <header className="site"><div className="in">
        <Link href="/">Bogoboom</Link><span>{title}</span>
        <button className="btn sec" style={{ color: '#fff', borderColor: '#fff' }} onClick={() => { setToken(null); r.push('/'); }}>Log out</button>
      </div></header>
      <main id="main">{ready ? children : <p className="muted">Loading…</p>}</main>
    </>
  );
}

export function Msg({ error, ok }: { error?: string | null; ok?: string | null }) {
  return <>{error && <p role="alert" className="err">{error}</p>}{ok && <p role="status">{ok}</p>}</>;
}
export const msg = (e: unknown) => (e instanceof Error ? e.message : 'Something went wrong');
