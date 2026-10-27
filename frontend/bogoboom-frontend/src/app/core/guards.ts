import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { Role } from './models';
import { HOME, SessionService } from './session.service';

/** Needs a signed-in, registered account whose role is one of `roles`. */
export function roleGuard(...roles: Role[]): CanActivateFn {
  return async (_route, state) => {
    const session = inject(SessionService);
    const router = inject(Router);
    let me = null;
    try {
      me = await session.refresh();
    } catch {
      return router.parseUrl('/');
    }
    if (!session.isLoggedIn()) {
      session.login(state.url);
      return false;
    }
    if (!me?.isRegistered || !me.role) return router.parseUrl('/start');
    if (!roles.includes(me.role)) return router.parseUrl(HOME[me.role]);
    return true;
  };
}

/** Needs a signed-in account (registered or not): used by the registration forms. */
export const signedInGuard: CanActivateFn = async (_route, state) => {
  const session = inject(SessionService);
  const router = inject(Router);
  try {
    const me = await session.refresh();
    if (!session.isLoggedIn()) {
      session.signup(state.url);
      return false;
    }
    if (me?.isRegistered && me.role) return router.parseUrl(HOME[me.role]);
    return true;
  } catch {
    return router.parseUrl('/');
  }
};
