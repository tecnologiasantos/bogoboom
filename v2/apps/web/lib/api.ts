export const API = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:4000';
export type Role = 'newcomer' | 'hr' | 'agency';

const KEY = 'bb_token';
export const getToken = () => { try { return localStorage.getItem(KEY); } catch { return null; } };
export const setToken = (t: string | null) => { try { t ? localStorage.setItem(KEY, t) : localStorage.removeItem(KEY); } catch {} };

export class ApiError extends Error { constructor(public status: number, public code: string, message: string) { super(message); } }

export async function api<T = any>(path: string, init: { method?: string; body?: unknown; csv?: string } = {}): Promise<T> {
  const headers: Record<string, string> = {};
  const t = getToken(); if (t) headers.authorization = `Bearer ${t}`;
  let body: string | undefined;
  if (init.csv !== undefined) { headers['content-type'] = 'text/csv'; body = init.csv; }
  else if (init.body !== undefined) { headers['content-type'] = 'application/json'; body = JSON.stringify(init.body); }
  const r = await fetch(`${API}${path}`, { method: init.method ?? 'GET', headers, body });
  const data = await r.json().catch(() => ({}));
  if (!r.ok && r.status !== 422) throw new ApiError(r.status, data.error ?? 'error', data.message ?? data.issues?.map((i: any) => `${i.path}: ${i.message}`).join('; ') ?? r.statusText);
  return data as T;
}

export const money = (n: number | string | null | undefined, cur = 'COP') =>
  n == null ? '—' : new Intl.NumberFormat('es-CO', { style: 'currency', currency: cur, maximumFractionDigits: 0 }).format(Number(n));
