import { describe, expect, it } from 'vitest';
import { initialEntries, kindFor, moveIssue, type Entry } from './model';
describe('prototype file movement', () => {
 it('moves a file between folders', () => expect(moveIssue(initialEntries,'chapter','materials')).toBeNull());
 it('preserves the project boundary', () => expect(moveIssue(initialEntries,'translation','materials')).not.toBeNull());
 it('rejects a folder inside itself or any descendant', () => {
  const entries: Entry[] = [...initialEntries,{id:'nested',project:'shore',name:'Nested',kind:'folder',parent:'manuscript'}];
  expect(moveIssue(entries,'manuscript','manuscript')).not.toBeNull();
  expect(moveIssue(entries,'manuscript','nested')).not.toBeNull();
 });
 it('rejects name collisions without rejecting a no-op', () => {
  const entries: Entry[] = [...initialEntries,{...initialEntries[1],id:'duplicate',parent:'materials'}];
  expect(moveIssue(entries,'chapter','materials')).not.toBeNull();
  expect(moveIssue(entries,'chapter','manuscript')).toBeNull();
 });
 it('accepts moving a folder to the root', () => expect(moveIssue(initialEntries,'materials',null)).toBeNull());
 it('uses an ordinary binary page for unknown formats', () => expect(kindFor(new File(['binary'],'archive.zip',{type:'application/zip'}))).toBe('binary'));
 it('detects media and markdown', () => {
  expect(kindFor(new File(['image'],'photo.png',{type:'image/png'}))).toBe('image');
  expect(kindFor(new File(['pdf'],'source.pdf',{type:'application/pdf'}))).toBe('pdf');
  expect(kindFor(new File(['# Text'],'chapter.md'))).toBe('text');
 });
});
