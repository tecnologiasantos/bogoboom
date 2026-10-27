import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { ApiService, ApiFailure, describeError } from '../../core/api.service';

export const COLLABORATION_TYPES = [
  { value: 'real_estate', label: 'Real estate agency / property owner' },
  { value: 'hr', label: 'HR department / employer' },
  { value: 'relocation', label: 'Relocation services' },
  { value: 'franchise', label: 'Franchising and expansion' },
  { value: 'other', label: 'Something else' },
];

@Component({ selector: 'app-partnership', imports: [ReactiveFormsModule, RouterLink], templateUrl: './partnership.html' })
export class PartnershipComponent {
  private readonly api = inject(ApiService);
  private readonly fb = inject(FormBuilder).nonNullable;
  protected readonly types = COLLABORATION_TYPES;

  protected readonly form = this.fb.group({
    fullName: ['', [Validators.required, Validators.maxLength(120)]],
    email: ['', [Validators.required, Validators.email, Validators.maxLength(254)]],
    phone: ['', [Validators.pattern(/^\+?[0-9 ()-]{6,20}$/)]],
    country: ['', Validators.maxLength(100)],
    city: ['', Validators.maxLength(100)],
    collaborationType: ['real_estate', Validators.required],
    acceptPrivacy: [false, Validators.requiredTrue],
    website: [''], // honeypot: real people leave it empty
  });
  protected readonly busy = signal(false);
  protected readonly done = signal(false);
  protected readonly error = signal('');
  protected readonly fieldErrors = signal<Record<string, string>>({});

  async submit() {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.busy.set(true);
    this.error.set('');
    this.fieldErrors.set({});
    try {
      await this.api.sendInquiry(this.form.getRawValue());
      this.done.set(true);
    } catch (e) {
      this.error.set(describeError(e));
      if (e instanceof ApiFailure) this.fieldErrors.set(e.details);
    } finally {
      this.busy.set(false);
    }
  }
}
