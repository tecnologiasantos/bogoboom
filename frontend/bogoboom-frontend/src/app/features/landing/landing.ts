import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

const S3 = 'https://tecnologiasantos-mira.s3.sa-east-1.amazonaws.com/bogoboom';

@Component({
  selector: 'app-landing',
  imports: [RouterLink],
  templateUrl: './landing.html',
  styleUrl: './landing.scss',
})
export class LandingComponent {
  protected readonly front = `${S3}/websiteimages/bogoboom-front.png`;
  protected readonly innovation = [1, 2, 3, 4, 5, 6, 7].map((n) => `${S3}/inovation/inovation${n}.png`);
  protected readonly audio = `${S3}/audio+description/bogoboom.com.wav`;

  protected readonly partner = [
    'List real estate properties as an official partner',
    'Promote jobs and receive housing support for relocated hires',
    "Use Bogoboom's partner badge/logo on your website and media",
    "Participate in Bogoboom's franchising and expansion programs",
    'Be listed as a featured partner on the Bogoboom portal',
    'Earn rewards by referring businesses or newcomers',
  ];
  protected readonly hr = [
    'Save money on temporary hotel stays and last-minute relocation support',
    'Delegate accommodation logistics to a dedicated, efficient platform',
    'Offer a better employee experience from day one',
    "Strengthen your company's reputation as relocation-friendly",
    'Access a pre-screened list of real estate agencies offering fair, quality housing',
    'Focus more on hiring and retention—and less on real estate headaches',
  ];
  protected readonly newcomers = [
    'Register for free and get matched with HR teams and local agencies',
    'Explore housing options before you even land in Bogotá',
    'Enjoy a smooth relocation with trusted partners',
    'Choose whether your housing preferences are shared with real estate agents',
    'Start your life in Bogotá with less stress and more support',
  ];
  protected readonly agencies = [
    'Access a growing network of international professionals relocating to Bogotá',
    'Get visibility among HR departments of top employers',
    "List properties that match newcomers' expectations and budgets",
    'Use our portal to receive qualified, verified housing leads',
    'Convert registrations into real contracts—quickly and reliably',
  ];
}
