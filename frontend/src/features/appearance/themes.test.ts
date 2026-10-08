import { expect, it } from 'vitest';
import { resolveTheme, themeFamilies } from './themes';
it('selects shared family variants and safely falls back for missing variants', () => {
  expect(resolveTheme('Edda', 'dark').id).toBe('edda-dark');
  expect(resolveTheme('unknown', 'light').id).toBe('edda-light');
  expect(resolveTheme('Darcula', 'light').id).toBe('edda-light');
  for (const family of themeFamilies) for (const scheme of ['light', 'dark'] as const) expect(resolveTheme(family, scheme).scheme).toBe(scheme);
});
