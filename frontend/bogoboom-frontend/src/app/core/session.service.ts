import { Injectable, computed, inject, signal } from '@angular/core';
import { AuthService } from '@auth0/auth0-angular';
import { firstValueFrom } from 'rxjs';

import { ApiService } from './api.service';
import { Me, Role } from './models';

/** Where each role lands after signing in. */
export const HOME: Record<Role, string> = {
  newcomer: '/newcomer',
  hr: '/hr',
  agency: '/agency',
  admin: '/admin',
};

/** Who is signed in and which account they have. The API stays the source of truth. */
@Injectable({ providedIn: 'root' })
export class SessionService {
  private readonly auth = inject(AuthService);
  private readonly api = inject(ApiService);

  readonly me = signal<Me | null>(null);
  readonly role = computed(() => this.me()?.role ?? null);
  readonly isLoggedIn = signal(false);

  /** Asks Auth0 and the API who the caller is. Safe to call again. */
  async refresh(): Promise<Me | null> {
    const authenticated = await firstValueFrom(this.auth.isAuthenticated$);
    this.isLoggedIn.set(authenticated);
    if (!authenticated) {
      this.me.set(null);
      return null;
    }
    const me = await this.api.me();
    this.me.set(me);
    return me;
  }

  login(returnTo = '/app') {
    this.auth.loginWithRedirect({ appState: { target: returnTo } });
  }

  signup(returnTo: string) {
    this.auth.loginWithRedirect({ appState: { target: returnTo }, authorizationParams: { screen_hint: 'signup' } });
  }

  logout() {
    this.me.set(null);
    this.isLoggedIn.set(false);
    this.auth.logout({ logoutParams: { returnTo: window.location.origin } });
  }
}
