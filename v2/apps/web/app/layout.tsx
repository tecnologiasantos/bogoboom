import './globals.css';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { Consent } from '@/components/Consent';

export const metadata: Metadata = {
  title: 'Bogoboom | Welcome to Bogotá',
  description: 'Bogoboom connects newcomers, HR teams and verified real estate agencies for a smooth relocation. Launching in Bogotá, Colombia.',
  openGraph: { title: 'Bogoboom — your smart start in Bogotá', type: 'website' },
  icons: { icon: '/mira-avatar.png' },
};

export default function Root({ children }: { children: ReactNode }) {
  return <html lang="en"><body>{children}<Consent /></body></html>;
}
