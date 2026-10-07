import {describe,expect,it} from 'vitest';
import {diffLines} from './lineDiff';
describe('bounded line diff',()=>{
 it('retains a small precise diff',()=>{expect(diffLines('a\nb','a\nc')).toEqual([{kind:'same',text:'a'},{kind:'removed',text:'b'},{kind:'added',text:'c'}]);});
 it('preserves both large texts with common edges',()=>{
  const a=['start',...Array.from({length:2000},(_,i)=>`old ${i}`),'end'].join('\n');
  const b=['start',...Array.from({length:2000},(_,i)=>`new ${i}`),'end'].join('\n');
  const diff=diffLines(a,b);
  expect(diff.filter(x=>x.kind!=='added').map(x=>x.text).join('\n')).toBe(a);
  expect(diff.filter(x=>x.kind!=='removed').map(x=>x.text).join('\n')).toBe(b);
  expect(diff[0].kind).toBe('same');expect(diff.at(-1)?.kind).toBe('same');
 });
});
