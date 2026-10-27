import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { RUNTIME_CONFIG } from './config';
import {
  Agency, AgencyNotification, ApiError, Lead, Me, Newcomer, PartnerInquiry, Property, PropertyInput,
} from './models';

/** An API failure with the server's message and its per-field problems. */
export class ApiFailure extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly details: Record<string, string> = {},
  ) {
    super(message);
  }
}

/** Turns any thrown value into text that is safe to show. */
export function describeError(e: unknown): string {
  if (e instanceof ApiFailure) {
    const fields = Object.entries(e.details).map(([k, v]) => `${k}: ${v}`);
    return fields.length ? `${e.message} (${fields.join('; ')})` : e.message;
  }
  return 'Something went wrong. Please try again.';
}

@Injectable({ providedIn: 'root' })
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly base = inject(RUNTIME_CONFIG).apiUrl;

  private async call<T>(method: string, path: string, body?: unknown): Promise<T> {
    try {
      return await firstValueFrom(this.http.request<T>(method, this.base + path, { body }));
    } catch (e) {
      if (e instanceof HttpErrorResponse) {
        const b = e.error as Partial<ApiError> | null;
        const message =
          b?.message ??
          (e.status === 0 ? 'The server could not be reached.' : e.status === 429 ? 'Too many requests. Please wait a moment.' : 'Request failed.');
        throw new ApiFailure(message, e.status, b?.details);
      }
      throw e;
    }
  }

  // account
  me = () => this.call<Me>('GET', '/me');
  register = (role: 'newcomer' | 'hr' | 'agency', body: unknown) => this.call<Me>('POST', `/register/${role}`, body);

  // public forms
  sendInquiry = (body: unknown) => this.call<void>('POST', '/partner-inquiries', body);
  sendLead = (body: unknown) => this.call<void>('POST', '/leads', body);

  // newcomer
  myProfile = () => this.call<Newcomer>('GET', '/newcomers/me');
  updateProfile = (body: unknown) => this.call<Newcomer>('PUT', '/newcomers/me', body);
  searchProperties = (city: string, maxPrice?: number) =>
    this.call<Property[]>('GET', `/properties?city=${encodeURIComponent(city)}${maxPrice ? `&maxPrice=${maxPrice}` : ''}`);
  favorites = () => this.call<Property[]>('GET', '/favorites');
  addFavorite = (id: string) => this.call<void>('PUT', `/favorites/${id}`);
  removeFavorite = (id: string) => this.call<void>('DELETE', `/favorites/${id}`);

  // hr
  invite = (body: unknown) => this.call<Newcomer>('POST', '/hr/newcomers', body);
  hires = () => this.call<Newcomer[]>('GET', '/hr/newcomers');

  // agency
  agency = () => this.call<Agency>('GET', '/agency/account');
  leads = () => this.call<Lead[]>('GET', '/agency/leads');
  unlock = (id: string) => this.call<Lead>('POST', `/agency/leads/${id}/unlock`);
  notifications = () => this.call<AgencyNotification[]>('GET', '/agency/notifications');
  markRead = (id: string) => this.call<void>('POST', `/agency/notifications/${id}/read`);
  properties = () => this.call<Property[]>('GET', '/agency/properties');
  createProperty = (p: PropertyInput) => this.call<Property>('POST', '/agency/properties', p);
  updateProperty = (id: string, p: PropertyInput) => this.call<Property>('PUT', `/agency/properties/${id}`, p);
  deleteProperty = (id: string) => this.call<void>('DELETE', `/agency/properties/${id}`);

  // admin
  inquiries = () => this.call<PartnerInquiry[]>('GET', '/admin/partner-inquiries');
  verifyAgency = (id: string, verified: boolean) => this.call<void>('PUT', `/admin/agencies/${id}/verify`, { verified });
  setPlan = (id: string, plan: string, activeUntil: string | null) =>
    this.call<void>('PUT', `/admin/agencies/${id}/plan`, { plan, activeUntil });
}
