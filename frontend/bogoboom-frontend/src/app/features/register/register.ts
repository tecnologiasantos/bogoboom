import { Component, computed, inject, input, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { ApiFailure, ApiService, describeError } from '../../core/api.service';
import { HOME, SessionService } from '../../core/session.service';

type RegRole = 'newcomer' | 'hr' | 'agency';
const TITLES: Record<RegRole, string> = {
  newcomer: "I'm a Newcomer",
  hr: "I'm from HR",
  agency: "I'm a Real Estate Agency",
};

@Component({ selector: 'app-register', imports: [ReactiveFormsModule, RouterLink], templateUrl: './register.html' })
export class RegisterComponent {
  private readonly api = inject(ApiService);
  private readonly router = inject(Router);
  private readonly session = inject(SessionService);
  private readonly fb = inject(FormBuilder).nonNullable;

  /** Route parameter :role (bound with withComponentInputBinding). */
  readonly role = input<string>('newcomer');
  protected readonly kind = computed<RegRole | null>(() => {
    const r = this.role();
    return r === 'newcomer' || r === 'hr' || r === 'agency' ? r : null;
  });
  protected readonly title = computed(() => (this.kind() ? TITLES[this.kind()!] : ''));

  protected readonly form = this.fb.group({
    fullName: ['', [Validators.required, Validators.maxLength(120)]],
    phone: ['', Validators.pattern(/^\+?[0-9 ()-]{6,20}$/)],
    // newcomer
    cityOfContract: ['Bogotá', Validators.maxLength(100)],
    arrivalDate: [''],
    budgetMax: [null as number | null, [Validators.min(0), Validators.max(1e12)]],
    bedrooms: [null as number | null, [Validators.min(0), Validators.max(20)]],
    furnished: [false],
    districts: [''],
    shareWithAgencies: [false],
    // hr
    companyName: ['', Validators.maxLength(160)],
    industry: ['', Validators.maxLength(100)],
    // agency
    agencyName: ['', Validators.maxLength(160)],
    description: ['', Validators.maxLength(2000)],
    // hr + agency
    city: ['Bogotá', Validators.maxLength(100)],
    country: ['CO', [Validators.required, Validators.pattern(/^[A-Za-z]{2}$/)]],
    acceptTerms: [false, Validators.requiredTrue],
  });

  protected readonly busy = signal(false);
  protected readonly error = signal('');
  protected readonly fieldErrors = signal<Record<string, string>>({});

  protected fieldError(name: string): string {
    return this.fieldErrors()[name] ?? '';
  }

  /** Checks only the fields of the chosen role. */
  private valid(kind: RegRole): boolean {
    const c = this.form.controls;
    const need: (keyof typeof c)[] = ['fullName', 'phone', 'acceptTerms'];
    if (kind === 'newcomer') need.push('cityOfContract', 'budgetMax', 'bedrooms');
    if (kind === 'hr') need.push('companyName', 'country');
    if (kind === 'agency') need.push('agencyName', 'city', 'country');
    need.forEach((n) => c[n].markAsTouched());
    const missing =
      (kind === 'newcomer' && !c.cityOfContract.value.trim()) ||
      (kind === 'hr' && !c.companyName.value.trim()) ||
      (kind === 'agency' && (!c.agencyName.value.trim() || !c.city.value.trim()));
    return !missing && need.every((n) => c[n].valid);
  }

  async submit() {
    const kind = this.kind();
    if (!kind || !this.valid(kind)) {
      this.error.set('Please fix the highlighted fields.');
      return;
    }
    const v = this.form.getRawValue();
    const base = { acceptTerms: v.acceptTerms, fullName: v.fullName, phone: v.phone };
    const body =
      kind === 'newcomer'
        ? {
            ...base,
            shareWithAgencies: v.shareWithAgencies,
            cityOfContract: v.cityOfContract,
            arrivalDate: v.arrivalDate || null,
            budgetMax: v.budgetMax,
            bedrooms: v.bedrooms,
            furnished: v.furnished,
            preferredDistricts: v.districts.split(',').map((d) => d.trim()).filter(Boolean),
          }
        : kind === 'hr'
          ? { ...base, companyName: v.companyName, industry: v.industry, city: v.city, country: v.country }
          : { ...base, agencyName: v.agencyName, description: v.description, city: v.city, country: v.country };

    this.busy.set(true);
    this.error.set('');
    this.fieldErrors.set({});
    try {
      const me = await this.api.register(kind, body);
      this.session.me.set(me);
      await this.router.navigateByUrl(me.role ? HOME[me.role] : '/app');
    } catch (e) {
      this.error.set(describeError(e));
      if (e instanceof ApiFailure) this.fieldErrors.set(e.details);
    } finally {
      this.busy.set(false);
    }
  }
}
