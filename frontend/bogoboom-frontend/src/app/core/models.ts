export type Role = 'newcomer' | 'hr' | 'agency' | 'admin';

export interface Me {
  isRegistered: boolean;
  email: string;
  id?: string;
  role?: Role;
  locale?: string;
}

export type Stage =
  | 'invited' | 'profile_done' | 'searching' | 'visiting' | 'contract_signed' | 'moved_in' | 'cancelled';

export interface Newcomer {
  id: string;
  firstName: string;
  cityOfContract: string;
  arrivalDate?: string;
  budgetMin?: number;
  budgetMax?: number;
  currency: string;
  bedrooms?: number;
  furnished?: boolean;
  preferredDistricts: string[];
  stage: Stage;
  allowContact: boolean;
  createdAt: string;
}

export interface Contact { fullName: string; email: string; phone: string }

/** What an agency sees: no contact data until it is unlocked. */
export interface Lead extends Newcomer {
  matchScore: number;
  unlocked: boolean;
  contact?: Contact;
}

export type PropertyType = 'apartment' | 'room' | 'house' | 'studio';

export interface Property {
  id: string;
  agencyId: string;
  title: string;
  description: string;
  price: number;
  currency: string;
  type: PropertyType;
  bedrooms?: number;
  furnished: boolean;
  city: string;
  district: string;
  address: string;
  availableFrom?: string;
  verified: boolean;
  isActive: boolean;
}

export interface PropertyInput {
  title: string;
  description: string;
  price: number;
  currency: string;
  type: PropertyType;
  bedrooms?: number | null;
  furnished: boolean;
  city: string;
  district: string;
  address: string;
  availableFrom?: string | null;
  isActive: boolean;
}

export interface Agency {
  id: string;
  name: string;
  city: string;
  country: string;
  verified: boolean;
  contactCredits: number;
  plan: 'free' | 'pro' | 'enterprise';
  planActiveUntil?: string;
}

export interface AgencyNotification {
  id: string;
  newcomerId: string;
  matchScore: number;
  message: string;
  readAt?: string;
  createdAt: string;
}

export interface PartnerInquiry {
  id: string;
  fullName: string;
  email: string;
  phone: string;
  country: string;
  city: string;
  collaborationType: string;
  createdAt: string;
}

/** Body of every API error. `details` maps a field to its problem on validation errors. */
export interface ApiError { error: true; message: string; details?: Record<string, string> }
