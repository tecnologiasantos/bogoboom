import { CurrencyPipe } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';

import { ApiService, describeError } from '../../core/api.service';
import { Agency, AgencyNotification, Lead, Property, PropertyType } from '../../core/models';

export const PROPERTY_TYPES: PropertyType[] = ['apartment', 'room', 'house', 'studio'];

@Component({ selector: 'app-agency', imports: [ReactiveFormsModule, CurrencyPipe], templateUrl: './agency.html', styleUrl: './agency.scss' })
export class AgencyComponent implements OnInit {
  private readonly api = inject(ApiService);
  private readonly fb = inject(FormBuilder).nonNullable;
  protected readonly types = PROPERTY_TYPES;

  protected readonly tab = signal<'leads' | 'listings' | 'notifications'>('leads');
  protected readonly account = signal<Agency | null>(null);
  protected readonly leads = signal<Lead[]>([]);
  protected readonly listings = signal<Property[]>([]);
  protected readonly notes = signal<AgencyNotification[]>([]);
  protected readonly editing = signal<string | null>(null);
  protected readonly error = signal('');
  protected readonly message = signal('');
  protected readonly busy = signal(false);

  protected readonly form = this.fb.group({
    title: ['', [Validators.required, Validators.maxLength(160)]],
    description: [''],
    price: [0, [Validators.required, Validators.min(0)]],
    currency: ['COP', [Validators.required, Validators.pattern(/^[A-Z]{3}$/)]],
    type: ['apartment' as PropertyType, Validators.required],
    bedrooms: [null as number | null],
    furnished: [false],
    city: ['Bogotá', Validators.required],
    district: [''],
    address: [''],
    availableFrom: [''],
    isActive: [true],
  });

  async ngOnInit() {
    await this.reload();
  }

  private async run(fn: () => Promise<void>) {
    this.busy.set(true);
    this.error.set('');
    try {
      await fn();
    } catch (e) {
      this.error.set(describeError(e));
    } finally {
      this.busy.set(false);
    }
  }

  async reload() {
    await this.run(async () => {
      const [a, l, p, n] = await Promise.all([this.api.agency(), this.api.leads(), this.api.properties(), this.api.notifications()]);
      this.account.set(a);
      this.leads.set(l);
      this.listings.set(p);
      this.notes.set(n);
    });
  }

  protected get unreadCount(): number {
    return this.notes().filter((n) => !n.readAt).length;
  }

  protected get hasPaidPlan(): boolean {
    const a = this.account();
    if (!a || a.plan === 'free') return false;
    return !a.planActiveUntil || new Date(a.planActiveUntil) > new Date();
  }

  async unlock(lead: Lead) {
    await this.run(async () => {
      const unlocked = await this.api.unlock(lead.id);
      this.leads.update((list) => list.map((l) => (l.id === lead.id ? { ...l, ...unlocked } : l)));
      this.account.set(await this.api.agency());
    });
  }

  async markRead(n: AgencyNotification) {
    await this.run(async () => {
      await this.api.markRead(n.id);
      this.notes.set(await this.api.notifications());
    });
  }

  edit(p: Property) {
    this.editing.set(p.id);
    this.form.patchValue({
      title: p.title, description: p.description, price: p.price, currency: p.currency, type: p.type,
      bedrooms: p.bedrooms ?? null, furnished: p.furnished, city: p.city, district: p.district,
      address: p.address, availableFrom: p.availableFrom ?? '', isActive: p.isActive,
    });
    this.tab.set('listings');
  }

  cancelEdit() {
    this.editing.set(null);
    this.form.reset({ currency: 'COP', type: 'apartment', city: 'Bogotá', price: 0, isActive: true, furnished: false });
  }

  async saveListing() {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const v = this.form.getRawValue();
    const body = { ...v, availableFrom: v.availableFrom || null };
    await this.run(async () => {
      const id = this.editing();
      if (id) await this.api.updateProperty(id, body);
      else await this.api.createProperty(body);
      this.message.set(id ? 'Listing updated.' : 'Listing created. Matching newcomers were notified.');
      this.cancelEdit();
      this.listings.set(await this.api.properties());
      this.notes.set(await this.api.notifications());
    });
  }

  async remove(p: Property) {
    if (!window.confirm(`Delete "${p.title}"?`)) return;
    await this.run(async () => {
      await this.api.deleteProperty(p.id);
      this.listings.set(await this.api.properties());
    });
  }
}
