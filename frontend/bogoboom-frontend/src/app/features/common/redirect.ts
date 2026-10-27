import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';

import { HOME, SessionService } from '../../core/session.service';

/** Sends the signed-in person to the dashboard of their role. */
@Component({ selector: 'app-redirect', template: '<p class="page center">Loading…</p>' })
export class RedirectComponent {
  constructor() {
    const role = inject(SessionService).role();
    void inject(Router).navigateByUrl(role ? HOME[role] : '/start', { replaceUrl: true });
  }
}
