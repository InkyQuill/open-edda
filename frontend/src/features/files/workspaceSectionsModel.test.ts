import { expect, test } from 'vitest';
import { decodeSections } from './workspaceSectionsModel';
test('author can configure only finished text with custom labels; no draft/KB prerequisites', () => {
  const config = { schemaVersion: 1, sections: [{ id: 'text', label: 'Готово по моему решению', folderId: 'final', custom: 'preserve' }], future: { option: true } };
  expect(decodeSections(JSON.stringify(config))).toEqual(config);
  expect(decodeSections('{"schemaVersion":1,"sections":[]}').sections).toEqual([]);
});
test('unsupported settings are rejected rather than silently replaced', () => {
  for (const input of ['null', '{"schemaVersion":2,"sections":[]}', '{"schemaVersion":1,"sections":[{"id":"a","label":"","folderId":"kb"}]}']) expect(() => decodeSections(input)).toThrow();
});
