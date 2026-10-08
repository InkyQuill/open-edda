import { describe, expect, it } from 'vitest';
import { compareTrees } from './versionDiff';
import type { TreeEntry } from './fileApi';
const file: TreeEntry = { id: 'a', path: 'old.md', kind: 'file', bytes: 1, sha256: 'a' };
describe('version tree comparison', () => {
  it('tracks rename and content change together by identity', () => {
    expect(compareTrees([file], [{ ...file, path: 'new.md', sha256: 'b' }])[0]?.label).toBe('Перемещено · Изменено');
  });
  it('distinguishes replacement at the same path from editing', () => {
    expect(compareTrees([file], [{ ...file, id: 'b' }]).map(c => c.label).sort()).toEqual(['Добавлено', 'Удалено']);
  });
  it('includes empty folders and additions and omits unchanged files', () => {
    expect(compareTrees([file], [file, { id: 'd', path: 'empty', kind: 'directory', bytes: 0 }])).toHaveLength(1);
    expect(compareTrees([file], [file])).toEqual([]);
  });
});
