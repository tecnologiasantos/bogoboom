import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

/** Role picker for people who signed in but have no account yet. */
@Component({
  selector: 'app-start',
  imports: [RouterLink],
  template: `
    <section class="page center">
      <h2>Welcome! Tell us who you are</h2>
      <p>Choose the account that fits you. You can always contact us if you picked the wrong one.</p>
      <div class="card-row">
        <div class="card"><h3>I'm a Newcomer</h3><p>Find housing before you land.</p><a class="btn" routerLink="/register/newcomer">Continue</a></div>
        <div class="card"><h3>I'm from HR</h3><p>Support the people you hire.</p><a class="btn" routerLink="/register/hr">Continue</a></div>
        <div class="card"><h3>I'm a Real Estate Agency</h3><p>Reach relocating professionals.</p><a class="btn" routerLink="/register/agency">Continue</a></div>
      </div>
    </section>`,
})
export class StartComponent {}
