import { describe, expect, it } from 'vitest';
import { inspectFile, moveEntries, validatePath } from './fileModel';
import type { TreeEntry } from './fileApi';
const entries: TreeEntry[] = [{id:'a',path:'story',kind:'directory',bytes:0},{id:'b',path:'story/one.md',kind:'file',bytes:1,sha256:'old'},{id:'c',path:'materials',kind:'directory',bytes:0},{id:'d',path:'story/nested',kind:'directory',bytes:0}];
describe('production file operations', () => {
  it('moves a folder recursively without changing identities or content', () => {
    const result = moveEntries(entries,'a','materials/story');
    expect(result.find(e => e.id === 'b')).toEqual({...entries[1],path:'materials/story/one.md'});
    expect(result.find(e => e.id === 'd')?.path).toBe('materials/story/nested');
    expect(entries[1].path).toBe('story/one.md');
  });
  it('rejects cycles, collisions and missing parents', () => {
    expect(() => moveEntries(entries,'a','story/nested/story')).toThrow();
    expect(() => moveEntries(entries,'a','materials')).toThrow();
    expect(() => moveEntries(entries,'b','missing/one.md')).toThrow();
  });
  it('rejects unsafe and reserved paths', () => { for(const path of ['../chapter','/chapter','a//b','a/../b','.edda/state','.env','a\\b']) expect(() => validatePath(path)).toThrow(); });
  it('accepts Unicode paths', () => expect(() => validatePath('原文/Глава 1.md')).not.toThrow());
  it('shows unknown binary as ordinary downloadable content', async () => {const view=await inspectFile({...entries[1],path:'cover.bin'},new Blob([new Uint8Array([0,255])]));expect(view.kind).toBe('binary');URL.revokeObjectURL(view.url);});
  it('recognizes images despite generic object MIME',async () => {const view=await inspectFile({...entries[1],path:'cover.svg'},new Blob(['<svg/>'],{type:'application/octet-stream'}));expect(view.kind).toBe('image');expect(view.blob.type).toBe('image/svg+xml');URL.revokeObjectURL(view.url);});
  it('retains UTF-8 and BOM exactly when editing',async () => {const body='\ufeff日本語\nРусский\r\n';const view=await inspectFile(entries[1],new Blob([body]));expect(view.body).toBe(body);URL.revokeObjectURL(view.url);});
  it('offers large text for download without rejecting the file',async () => {const view=await inspectFile(entries[1],new Blob(['x'.repeat(1024*1024+1)]));expect(view.kind).toBe('binary');URL.revokeObjectURL(view.url);});
});
