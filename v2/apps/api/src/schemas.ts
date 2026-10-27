import { z } from 'zod';

export const uuid = z.string().uuid();
const money = z.coerce.number().nonnegative().max(1e10);
const date = z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'YYYY-MM-DD');
const currency = z.string().length(3).toUpperCase().default('COP');

export const needsSchema = z.object({
  city_of_contract: z.string().trim().min(2).max(80),
  arrival_date: date.nullish(),
  budget_min: money.nullish(),
  budget_max: money.nullish(),
  currency,
  bedrooms: z.coerce.number().int().min(0).max(10).nullish(),
  furnished: z.boolean().nullish(),
  preferred_districts: z.array(z.string().trim().min(1).max(60)).max(20).default([]),
});

export const newcomerProfileSchema = needsSchema.extend({
  first_name: z.string().trim().min(1).max(80),
  full_name: z.string().trim().max(160).optional(),
  phone: z.string().trim().max(40).optional(),
  allow_contact: z.boolean(),
});

export const hrNewcomerSchema = needsSchema.extend({
  first_name: z.string().trim().min(1).max(80),
  full_name: z.string().trim().max(160).optional(),
  email: z.string().email().max(200),
  phone: z.string().trim().max(40).optional(),
});

export const propertySchema = z.object({
  title: z.string().trim().min(3).max(160),
  description: z.string().max(5000).optional(),
  price: money.refine((v) => v > 0, 'must be > 0'),
  currency,
  type: z.enum(['apartment', 'room', 'house', 'studio']),
  bedrooms: z.coerce.number().int().min(0).max(20).nullish(),
  furnished: z.boolean().default(false),
  city: z.string().trim().min(2).max(80),
  district: z.string().trim().max(80).nullish(),
  address: z.string().trim().max(200).nullish(),
  lat: z.number().min(-90).max(90).nullish(),
  lng: z.number().min(-180).max(180).nullish(),
  available_from: date.nullish(),
});

export const stageSchema = z.object({
  stage: z.enum(['invited', 'profile_done', 'searching', 'visiting', 'contract_signed', 'moved_in', 'cancelled']),
});

export const toBool = (v: unknown) => (typeof v === 'string' ? v === 'true' : v);
