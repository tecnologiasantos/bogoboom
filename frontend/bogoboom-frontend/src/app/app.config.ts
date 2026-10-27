import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideRouter, withComponentInputBinding } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { authHttpInterceptorFn, provideAuth0 } from '@auth0/auth0-angular';

import { routes } from './app.routes';
import { RUNTIME_CONFIG, RuntimeConfig } from './core/config';

export function buildAppConfig(cfg: RuntimeConfig): ApplicationConfig {
  return {
    providers: [
      provideBrowserGlobalErrorListeners(),
      provideRouter(routes, withComponentInputBinding()),
      provideHttpClient(withInterceptors([authHttpInterceptorFn])),
      { provide: RUNTIME_CONFIG, useValue: cfg },
      provideAuth0({
        domain: cfg.auth0.domain,
        clientId: cfg.auth0.clientId,
        authorizationParams: { redirect_uri: window.location.origin, audience: cfg.auth0.audience },
        // Tokens live in localStorage and renew with refresh tokens (no hidden iframe), so a reload
        // keeps the session. Enable refresh-token rotation on the Auth0 application.
        cacheLocation: 'localstorage',
        useRefreshTokens: true,
        httpInterceptor: {
          allowedList: [
            // Public forms never carry a token.
            { uri: `${cfg.apiUrl}/partner-inquiries`, allowAnonymous: true },
            { uri: `${cfg.apiUrl}/leads`, allowAnonymous: true },
            // Everything else under the API gets the access token for our audience, and nothing
            // outside the API ever does.
            { uri: `${cfg.apiUrl}/*`, tokenOptions: { authorizationParams: { audience: cfg.auth0.audience } } },
          ],
        },
      }),
    ],
  };
}
