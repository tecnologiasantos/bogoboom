import { z } from 'zod';

const schema = z.object({
  NODE_ENV: z.string().default('development'),
  PORT: z.coerce.number().default(4000),
  DATABASE_URL: z.string().min(1),
  AUTH_MODE: z.enum(['dev', 'oidc']).default('dev'),
  JWT_SECRET: z.string().min(32).optional(),
  OIDC_JWKS_URL: z.string().url().optional(),
  OIDC_ISSUER: z.string().optional(),
  OIDC_AUDIENCE: z.string().optional(),
  PAYMENT_WEBHOOK_SECRET: z.string().min(16).optional(),
  CORS_ORIGIN: z.string().default('http://localhost:3000'),
});

export type Config = z.infer<typeof schema>;

export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
  const c = schema.parse(env);
  if (c.AUTH_MODE === 'dev') {
    if (c.NODE_ENV === 'production') throw new Error('AUTH_MODE=dev is not allowed in production');
    if (!c.JWT_SECRET) throw new Error('JWT_SECRET (>=32 chars) is required in dev auth mode');
  } else if (!c.OIDC_JWKS_URL || !c.OIDC_ISSUER) {
    throw new Error('OIDC_JWKS_URL and OIDC_ISSUER are required when AUTH_MODE=oidc');
  }
  return c;
}
