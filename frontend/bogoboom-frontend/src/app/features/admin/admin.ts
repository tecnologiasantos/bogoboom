import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';

import { ApiService, describeError } from '../../core/api.service';
import { PartnerInquiry } from '../../core/models';

const UUID = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;

@Component({ selector: 'app-admin', imports: [ReactiveFormsModule], templateUrl: './admin.html' })
export class AdminComponent implements OnInit {
  private readonly api = inject(ApiService);
  private readonly fb = inject(FormBuilder).nonNullable;

  protected readonly inquiries = signal<PartnerInquiry[]>([]);
  protected readonly message = signal('');
  protected readonly error = signal('');
  protected readonly form = this.fb.group({
    agencyId: ['', [Validators.required, Validators.pattern(UUID)]],
    plan: ['pro'],
    activeUntil: [''],
  });

  async ngOnInit() {
    try {
      this.inquiries.set(await this.api.inquiries());
    } catch (e) {
      this.error.set(describeError(e));
    }
  }

  private async act(fn: () => Promise<void>, ok: string) {
    if (this.form.controls.agencyId.invalid) {
      this.form.controls.agencyId.markAsTouched();
      return;
    }
    this.error.set('');
    this.message.set('');
    try {
      await fn();
      this.message.set(ok);
    } catch (e) {
      this.error.set(describeError(e));
    }
  }

  verify(verified: boolean) {
    return this.act(() => this.api.verifyAgency(this.form.getRawValue().agencyId, verified), verified ? 'Agency verified.' : 'Verification removed.');
  }

  setPlan() {
    const v = this.form.getRawValue();
    const until = v.activeUntil ? new Date(v.activeUntil + 'T23:59:59Z').toISOString() : null;
    return this.act(() => this.api.setPlan(v.agencyId, v.plan, until), 'Plan updated.');
  }
}
