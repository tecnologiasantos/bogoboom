import type { FastifyReply, FastifyRequest } from 'fastify';
import { SignJWT, jwtVerify, createRemoteJWKSet } from 'jose';
import type { Config } from './config.js';
import type { Db } from './db.js';
import { AppError } from './errors.js';

export type Role = 'newcomer' | 'hr' | 'agency' | 'admin';
export interface AuthUser { id: string; role: Role; email: string; sub: string }
declare module 'fastify' { interface FastifyRequest { user?: AuthUser } }

const SELF_ROLES = ['newcomer', 'hr', 'agency'] as const;

export function makeAuth(cfg: Config, db: Db) {
  const secret = cfg.JWT_SECRET ? new TextEncoder().encode(cfg.JWT_SECRET) : null;
  const jwks = cfg.AUTH_MODE === 'oidc' ? createRemoteJWKSet(new URL(cfg.OIDC_JWKS_URL!)) : null;

  async function upsertUser(sub: string, email: string, role: string | undefined): Promise<AuthUser> {
    const found = (await db.query('SELECT id, role, email FROM users WHERE auth_sub=$1 AND deleted_at IS NULL', [sub])).rows[0];
    if (found) return { id: found.id, role: found.role, email: found.email, sub };
    if (!role || !(SELF_ROLES as readonly string[]).includes(role)) throw new AppError(403, 'Unknown or disallowed role', 'bad_role');
    const r = await db.query(
      `INSERT INTO users (auth_sub, email, role) VALUES ($1,$2,$3)
       ON CONFLICT (auth_sub) DO UPDATE SET updated_at = now() RETURNING id, role, email`, [sub, email, role]);
    return { id: r.rows[0].id, role: r.rows[0].role, email: r.rows[0].email, sub };
  }

  /** DEV ONLY: email-based login that issues an HS256 token. Disabled in oidc mode / production. */
  async function devLogin(email: string, role: string) {
    if (cfg.AUTH_MODE !== 'dev') throw new AppError(404, 'Not found');
    const user = await upsertUser(`dev:${email.toLowerCase()}`, email.toLowerCase(), role);
    const token = await new SignJWT({ sub: user.sub }).setProtectedHeader({ alg: 'HS256' })
      .setIssuedAt().setExpirationTime('12h').sign(secret!);
    return { token, user };
  }

  async function authenticate(req: FastifyRequest) {
    const h = req.headers.authorization;
    if (!h?.startsWith('Bearer ')) throw new AppError(401, 'Missing bearer token', 'unauthenticated');
    const token = h.slice(7);
    try {
      if (cfg.AUTH_MODE === 'dev') {
        const { payload } = await jwtVerify(token, secret!, { algorithms: ['HS256'] });
        const u = (await db.query('SELECT id, role, email FROM users WHERE auth_sub=$1 AND deleted_at IS NULL', [payload.sub])).rows[0];
        if (!u) throw new Error('unknown user');
        req.user = { id: u.id, role: u.role, email: u.email, sub: String(payload.sub) };
      } else {
        const { payload } = await jwtVerify(token, jwks!, {
          issuer: cfg.OIDC_ISSUER, audience: cfg.OIDC_AUDIENCE, algorithms: ['RS256'],
        });
        const role = (payload['custom:role'] ?? payload['role']) as string | undefined;
        req.user = await upsertUser(String(payload.sub), String(payload.email ?? ''), role);
      }
    } catch (e) {
      if (e instanceof AppError) throw e;
      throw new AppError(401, "Invalid or expired token", "unauthenticated");
    }
  }

  const requireRole = (...roles: Role[]) => async (req: FastifyRequest, _reply: FastifyReply) => {
    await authenticate(req);
    if (!roles.includes(req.user!.role)) throw new AppError(403, 'Insufficient role', 'forbidden');
  };

  return { devLogin, authenticate, requireRole };
}
export type Auth = ReturnType<typeof makeAuth>;
