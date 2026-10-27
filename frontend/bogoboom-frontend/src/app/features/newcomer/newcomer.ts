import { CurrencyPipe } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';

import { ApiService, describeError } from '../../core/api.service';
import { Newcomer, Property, Stage } from '../../core/models';

export const STAGES: { value: Stage; label: string }[] = [
  { value: 'profile_done', label: 'Profile done' },
  { value: 'searching', label: 'Searching for a home' },
  { value: 'visiting', label: 'Visiting places' },
  { value: 'contract_signed', label: 'Contract signed' },
  { value: 'moved_in', label: 'Moved in' },
  { value: 'cancelled', label: 'Cancelled' },
];

@Component({ selector: 'app-newcomer', imports: [ReactiveFormsModule, CurrencyPipe], templateUrl: './newcomer.html' })
export class NewcomerComponent implements OnInit {
  private readonly api = inject(ApiService);
  private readonly fb = inject(FormBuilder).nonNullable;
  protected readonly stages = STAGES;

  protected readonly tab = signal<'profile' | 'search' | 'favorites'>('profile');
  protected readonly profile = signal<Newcomer | null>(null);
  protected readonly results = signal<Property[]>([]);
  protected readonly favorites = signal<Property[]>([]);
  protected readonly message = signal('');
  protected readonly error = signal('');
  protected readonly busy = signal(false);
  protected readonly searched = signal(false);

  protected readonly form = this.fb.group({
    cityOfContract: [''],
    arrivalDate: [''],
    budgetMax: [null as number | null],
    bedrooms: [null as number | null],
    furnished: [false],
    districts: [''],
    stage: ['searching'],
    allowContact: [false],
  });
  protected readonly searchForm = this.fb.group({ city: [''], maxPrice: [null as number | null] });

  async ngOnInit() {
    await this.guard(async () => {
      const p = await this.api.myProfile();
      this.profile.set(p);
      this.form.patchValue({
        cityOfContract: p.cityOfContract,
        arrivalDate: p.arrivalDate ?? '',
        budgetMax: p.budgetMax ?? null,
        bedrooms: p.bedrooms ?? null,
        furnished: p.furnished ?? false,
        districts: p.preferredDistricts.join(', '),
        stage: p.stage === 'invited' ? 'profile_done' : p.stage,
        allowContact: p.allowContact,
      });
      this.searchForm.patchValue({ city: p.cityOfContract, maxPrice: p.budgetMax ?? null });
      this.favorites.set(await this.api.favorites());
    });
  }

  private async guard(fn: () => Promise<void>) {
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

  async save() {
    const v = this.form.getRawValue();
    await this.guard(async () => {
      const p = await this.api.updateProfile({
        cityOfContract: v.cityOfContract,
        arrivalDate: v.arrivalDate,
        budgetMax: v.budgetMax,
        bedrooms: v.bedrooms,
        furnished: v.furnished,
        preferredDistricts: v.districts.split(',').map((d) => d.trim()).filter(Boolean),
        stage: v.stage,
        allowContact: v.allowContact,
      });
      this.profile.set(p);
      this.message.set('Profile saved.');
    });
  }

  async search() {
    const v = this.searchForm.getRawValue();
    await this.guard(async () => {
      this.results.set(await this.api.searchProperties(v.city, v.maxPrice ?? undefined));
      this.searched.set(true);
    });
  }

  isFavorite(id: string): boolean {
    return this.favorites().some((f) => f.id === id);
  }

  async toggleFavorite(p: Property) {
    await this.guard(async () => {
      if (this.isFavorite(p.id)) await this.api.removeFavorite(p.id);
      else await this.api.addFavorite(p.id);
      this.favorites.set(await this.api.favorites());
    });
  }
}
