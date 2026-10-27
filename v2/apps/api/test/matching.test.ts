import { describe, it, expect } from 'vitest';
import { scoreMatch } from '../src/matching.js';

const n = { city: 'Bogotá', arrival_date: '2026-12-01', budget_min: null, budget_max: 2_000_000, bedrooms: 1, furnished: true, districts: ['Chapinero'] };
const p = { city: 'bogota', district: 'chapinero', price: 1_800_000, bedrooms: 1, furnished: true, available_from: '2026-11-25' };

describe('scoreMatch', () => {
  it('perfect match scores 100 (accent/case-insensitive)', () => expect(scoreMatch(n, p)).toBe(100));
  it('different city is a hard reject', () => expect(scoreMatch(n, { ...p, city: 'Medellín' })).toBe(0));
  it('price > 125% of budget is a hard reject', () => expect(scoreMatch(n, { ...p, price: 2_600_000 })).toBe(0));
  it('slightly over budget loses points but still matches', () => {
    const s = scoreMatch(n, { ...p, price: 2_150_000 });
    expect(s).toBeGreaterThan(50); expect(s).toBeLessThan(100);
  });
  it('unavailable until long after arrival loses availability points', () =>
    expect(scoreMatch(n, { ...p, available_from: '2027-03-01' })).toBe(80));
  it('wrong district / unfurnished reduce score', () =>
    expect(scoreMatch(n, { ...p, district: 'Usaquén', furnished: false })).toBe(100 - 15 - 15 + 0));
  it('no preferences set still yields a sensible score', () =>
    expect(scoreMatch({ ...n, budget_max: null, bedrooms: null, furnished: null, districts: [], arrival_date: null }, p)).toBe(25 + 20 + 15 + 15 + 8));
});
