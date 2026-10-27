'use client';
import { useEffect, useState } from 'react';

const GTM = process.env.NEXT_PUBLIC_GTM_ID; // e.g. GTM-XXXXXXX — loaded ONLY after consent
export function Consent() {
  const [show, setShow] = useState(false);
  const load = () => {
    if (!GTM || (window as any).__gtm) return; (window as any).__gtm = 1;
    const w = window as any; w.dataLayer = w.dataLayer || []; w.dataLayer.push({ 'gtm.start': Date.now(), event: 'gtm.js' });
    const s = document.createElement('script'); s.async = true; s.src = `https://www.googletagmanager.com/gtm.js?id=${GTM}`; document.head.appendChild(s);
  };
  useEffect(() => {
    let c: string | null = null; try { c = localStorage.getItem('bb_consent'); } catch {}
    if (c === 'granted') load(); else if (!c && GTM) setShow(true);
  }, []);
  const set = (v: string) => { try { localStorage.setItem('bb_consent', v); } catch {} setShow(false); if (v === 'granted') load(); };
  if (!show) return null;
  return (
    <div role="dialog" aria-label="Cookie consent" style={{ position: 'fixed', left: 16, right: 16, bottom: 16, maxWidth: 560, margin: 'auto', background: '#111', color: '#fff', padding: 16, borderRadius: 8, display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'center', zIndex: 50 }}>
      <span>We use analytics cookies to improve Bogoboom.</span>
      <button className="btn" onClick={() => set('granted')}>Accept</button>
      <button className="btn sec" style={{ color: '#fff', borderColor: '#fff' }} onClick={() => set('denied')}>Reject</button>
    </div>
  );
}
