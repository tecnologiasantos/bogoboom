import { Component, DestroyRef, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { Router, RouterLink, RouterOutlet } from '@angular/router';
import { AuthService } from '@auth0/auth0-angular';

import { HOME, SessionService } from './core/session.service';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, RouterLink],
  templateUrl: './app.html',
  styleUrl: './app.scss',
})
export class App {
  protected readonly session = inject(SessionService);
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);
  protected readonly year = new Date().getFullYear();
  protected readonly home = HOME;

  constructor() {
    // After the Auth0 redirect, go where the person was heading (a same-site path only).
    this.auth.appState$.pipe(takeUntilDestroyed(inject(DestroyRef))).subscribe((state) => {
      const target = typeof state?.target === 'string' && state.target.startsWith('/') && !state.target.startsWith('//') ? state.target : '/app';
      void this.router.navigateByUrl(target);
    });
    this.auth.isAuthenticated$.pipe(takeUntilDestroyed(inject(DestroyRef))).subscribe((ok) => {
      this.session.isLoggedIn.set(ok);
      if (ok && !this.session.me()) void this.session.refresh().catch(() => undefined);
    });
  }
}
