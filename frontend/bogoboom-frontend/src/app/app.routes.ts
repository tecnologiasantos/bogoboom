import { Routes } from '@angular/router';

import { roleGuard, signedInGuard } from './core/guards';

export const routes: Routes = [
  // public
  { path: '', pathMatch: 'full', loadComponent: () => import('./features/landing/landing').then((m) => m.LandingComponent) },
  { path: 'partnership', loadComponent: () => import('./features/partnership/partnership').then((m) => m.PartnershipComponent) },
  { path: 'privacy', loadComponent: () => import('./features/privacy/privacy').then((m) => m.PrivacyComponent) },
  { path: 'start', loadComponent: () => import('./features/register/start').then((m) => m.StartComponent) },

  // sign up (needs an Auth0 login, not yet an account)
  {
    path: 'register/:role',
    canActivate: [signedInGuard],
    loadComponent: () => import('./features/register/register').then((m) => m.RegisterComponent),
  },

  // after login: send each role to its dashboard
  {
    path: 'app',
    canActivate: [roleGuard('newcomer', 'hr', 'agency', 'admin')],
    loadComponent: () => import('./features/common/redirect').then((m) => m.RedirectComponent),
  },
  {
    path: 'newcomer',
    canActivate: [roleGuard('newcomer')],
    loadComponent: () => import('./features/newcomer/newcomer').then((m) => m.NewcomerComponent),
  },
  {
    path: 'hr',
    canActivate: [roleGuard('hr')],
    loadComponent: () => import('./features/hr/hr').then((m) => m.HrComponent),
  },
  {
    path: 'agency',
    canActivate: [roleGuard('agency')],
    loadComponent: () => import('./features/agency/agency').then((m) => m.AgencyComponent),
  },
  {
    path: 'admin',
    canActivate: [roleGuard('admin')],
    loadComponent: () => import('./features/admin/admin').then((m) => m.AdminComponent),
  },
  { path: '**', loadComponent: () => import('./features/common/not-found').then((m) => m.NotFoundComponent) },
];
