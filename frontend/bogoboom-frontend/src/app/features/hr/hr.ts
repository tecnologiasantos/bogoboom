import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';

import { ApiFailure, ApiService, describeError } from '../../core/api.service';
import { Newcomer } from '../../core/models';

@Component({ selector: 'app-hr', imports: [ReactiveFormsModule], templateUrl: './hr.html' })
export class HrComponent implements OnInit {
  private readonly api = inject(ApiService);
  private readonly fb = inject(FormBuilder).nonNullable;

  protected readonly hires = signal<Newcomer[]>([]);
  protected readonly message = signal('');
  protected readonly error = signal('');
  protected readonly busy = signal(false);
  protected readonly fieldErrors = signal<Record<string, string>>({});

  protected readonly form = this.fb.group({
    fullName: ['', [Validators.required, Validators.maxLength(120)]],
    email: ['', [Validators.required, Validators.email]],
    phone: ['', Validators.pattern(/^\+?[0-9 ()-]{6,20}$/)],
    cityOfContract: ['Bogotá', Validators.required],
    arrivalDate: [''],
    budgetMax: [null as number | null],
    bedrooms: [null as number | null],
    furnished: [false],
  });

  async ngOnInit() {
    try {
      this.hires.set(await this.api.hires());
    } catch (e) {
      this.error.set(describeError(e));
    }
  }

  async invite() {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const v = this.form.getRawValue();
    this.busy.set(true);
    this.error.set('');
    this.message.set('');
    this.fieldErrors.set({});
    try {
      await this.api.invite({ ...v, arrivalDate: v.arrivalDate || null });
      this.message.set(`${v.fullName} was added. Ask them to sign up with ${v.email} to complete their profile.`);
      this.form.reset({ cityOfContract: v.cityOfContract, furnished: false });
      this.hires.set(await this.api.hires());
    } catch (e) {
      this.error.set(describeError(e));
      if (e instanceof ApiFailure) this.fieldErrors.set(e.details);
    } finally {
      this.busy.set(false);
    }
  }
}
