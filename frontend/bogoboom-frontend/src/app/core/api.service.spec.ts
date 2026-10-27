import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { describe, expect, it } from 'vitest';

import { ApiFailure, ApiService, describeError } from './api.service';
import { RUNTIME_CONFIG } from './config';

function setup() {
  TestBed.configureTestingModule({
    providers: [
      provideHttpClient(),
      provideHttpClientTesting(),
      { provide: RUNTIME_CONFIG, useValue: { apiUrl: 'http://api.test/api/v1', auth0: { domain: 'd', clientId: 'c', audience: 'a' } } },
    ],
  });
  return { api: TestBed.inject(ApiService), http: TestBed.inject(HttpTestingController) };
}

describe('describeError', () => {
  it('shows the server message and field details', () => {
    const e = new ApiFailure('Invalid data', 400, { email: 'invalid' });
    expect(describeError(e)).toBe('Invalid data (email: invalid)');
  });
  it('never leaks unknown errors', () => {
    expect(describeError(new Error('SELECT * FROM users'))).toBe('Something went wrong. Please try again.');
  });
});

describe('ApiService', () => {
  it('calls the configured API base', async () => {
    const { api, http } = setup();
    const p = api.me();
    const req = http.expectOne('http://api.test/api/v1/me');
    expect(req.request.method).toBe('GET');
    req.flush({ isRegistered: false });
    await expect(p).resolves.toEqual({ isRegistered: false });
  });

  it('turns HTTP errors into ApiFailure with the server message', async () => {
    const { api, http } = setup();
    const p = api.unlock('abc');
    http.expectOne('http://api.test/api/v1/agency/leads/abc/unlock').flush({ message: 'No credits left' }, { status: 402, statusText: 'Payment Required' });
    await expect(p).rejects.toMatchObject({ message: 'No credits left', status: 402 });
  });

  it('gives a friendly message for rate limiting', async () => {
    const { api, http } = setup();
    const p = api.me();
    http.expectOne('http://api.test/api/v1/me').flush(null, { status: 429, statusText: 'Too Many' });
    await expect(p).rejects.toMatchObject({ status: 429 });
  });
});
