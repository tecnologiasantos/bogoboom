import { bootstrapApplication } from '@angular/platform-browser';
import { buildAppConfig } from './app/app.config';
import { App } from './app/app';
import { loadRuntimeConfig } from './app/core/config';

// The runtime config (API URL, Auth0 tenant) is read from /config.json before the app starts, so
// the same build runs in every environment (local, AWS, Cloudflare) without rebuilding.
loadRuntimeConfig()
  .then((cfg) => bootstrapApplication(App, buildAppConfig(cfg)))
  .catch((err) => {
    console.error(err);
    document.body.innerHTML =
      '<p style="font-family:Arial;padding:2rem">Bogoboom could not start: the configuration file is missing or invalid.</p>';
  });
