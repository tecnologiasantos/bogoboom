import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-not-found',
  imports: [RouterLink],
  template: `<section class="page-narrow center">
    <div class="card"><h2>Page not found</h2><p>The page you are looking for does not exist.</p>
    <a class="btn" routerLink="/">Back to the home page</a></div></section>`,
})
export class NotFoundComponent {}
