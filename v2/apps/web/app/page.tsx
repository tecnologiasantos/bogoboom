import Link from 'next/link';
import Image from 'next/image';

const personas = [
  { role: 'newcomer', title: "I'm a Newcomer", text: 'Register for free, find verified long-term housing before you land.', cta: 'Get started' },
  { role: 'hr', title: "I'm from HR", text: 'Register the people you hire and track their relocation in one dashboard.', cta: 'Register HR' },
  { role: 'agency', title: "I'm a Real Estate Agency", text: 'List properties for free and receive matched, qualified leads. 5 free contacts every month.', cta: 'Join now' },
];

export default function Home() {
  return (
    <>
      <header className="site"><div className="in"><strong>Bogoboom</strong><Link href="/login">Login</Link></div></header>
      <main>
        <section style={{ textAlign: 'center', padding: '2rem 0' }}>
          <h1>Your smart start in Bogotá — and soon, anywhere</h1>
          <p style={{ maxWidth: 640, margin: '0 auto' }}>No more hotel stays. Bogoboom connects newcomers, HR teams and verified real estate agencies so people land in a real home.</p>
        </section>
        <div className="grid">
          {personas.map((p) => (
            <div className="card" key={p.role}><h2>{p.title}</h2><p className="muted">{p.text}</p><Link className="btn" href={`/login?role=${p.role}`}>{p.cta}</Link></div>
          ))}
        </div>
        <div className="card" style={{ display: 'flex', gap: 16, alignItems: 'center', flexWrap: 'wrap' }}>
          <Image src="/mira-avatar.png" width={96} height={96} alt="MIRA, the Bogoboom assistant mascot wearing a headset and a Colombian-flag shirt" style={{ borderRadius: '50%' }} />
          <div><h2 style={{ margin: 0 }}>Meet MIRA</h2><p className="muted" style={{ margin: 0 }}>Your relocation guide — coming soon in English, Spanish and Portuguese.</p></div>
        </div>
        <footer className="muted" style={{ textAlign: 'center', marginTop: 32 }}>
          © {new Date().getFullYear()} Bogoboom · a Tecnologia Santos project · <a href="mailto:support@tecnologiasantos.com">support@tecnologiasantos.com</a>
        </footer>
      </main>
    </>
  );
}
