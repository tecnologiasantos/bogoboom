import { InjectionToken } from '@angular/core';

export interface RuntimeConfig {
  /** Base URL of the API, e.g. https://api.bogoboom.com/api/v1 */
  apiUrl: string;
  auth0: { domain: string; clientId: string; audience: string };
}

export const RUNTIME_CONFIG = new InjectionToken<RuntimeConfig>('RUNTIME_CONFIG');

/** Checks the shape of config.json so a typo fails at start-up with a clear message. */
export function parseRuntimeConfig(raw: unknown): RuntimeConfig {
  const c = raw as Partial<RuntimeConfig> | null;
  const ok =
    !!c &&
    typeof c.apiUrl === 'string' &&
    /^https?:\/\//.test(c.apiUrl) &&
    !!c.auth0 &&
    typeof c.auth0.domain === 'string' &&
    typeof c.auth0.clientId === 'string' &&
    typeof c.auth0.audience === 'string';
  if (!ok) {
    throw new Error('Invalid /config.json: expected apiUrl and auth0.{domain,clientId,audience}');
  }
  return { apiUrl: c.apiUrl!.replace(/\/+$/, ''), auth0: c.auth0! };
}

export async function loadRuntimeConfig(): Promise<RuntimeConfig> {
  const res = await fetch('/config.json', { cache: 'no-store' });
  if (!res.ok) throw new Error(`/config.json answered ${res.status}`);
  return parseRuntimeConfig(await res.json());
}
