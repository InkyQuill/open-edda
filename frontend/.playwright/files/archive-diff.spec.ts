import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { expect, test } from '@playwright/test';
import { loginForTest } from './auth';
import type { ProjectVersion } from '../../src/features/files/fileApi';

const archive = execFileSync('python3', ['-c', `import io,zipfile,sys
b=io.BytesIO()
with zipfile.ZipFile(b,'w') as z:
 z.writestr('book/章.md','# 原稿\\n')
 z.writestr('book/empty/','')
 z.writestr('cover.bin',bytes([0,255,13,10]))
 z.writestr('.env','private')
sys.stdout.buffer.write(b.getvalue())`]);
test('ZIP import is reviewed, atomic, retryable, and comparable by stable identity', async ({page, request, isMobile}, info) => {
  const {token} = await (await loginForTest(request)).json() as {token:string};
  const headers = {Authorization:`Bearer ${token}`};
  await page.addInitScript(value => localStorage.setItem('open_edda_token',value), token);
  const project = await (await request.post('/api/projects',{headers,data:{title:`Archive ${info.project.name}`,storageMode:'files'}})).json() as {id:string};
  const root = `/api/projects/${project.id}/files`;
  const head = async () => await (await request.get(`${root}/versions/current`,{headers})).json() as ProjectVersion;
  await page.goto(`/projects/${project.id}/files`);
  await expect(page.getByRole('heading',{name:`Archive ${info.project.name}`,exact:true})).toBeVisible();
  const original = await head();
  await page.getByLabel('ZIP-архив',{exact:true}).setInputFiles({name:'book.zip',mimeType:'application/zip',buffer:archive});
  await expect(page.getByRole('dialog')).toContainText('book/empty/');
  await expect(page.getByRole('dialog')).toContainText('Исключено локальных настроек и кэшей: 1');
  expect((await head()).id).toBe(original.id);
  await page.route(`**${root}/versions`, async route => { await route.fetch(); await route.abort(); }, {times:1});
  await page.getByRole('dialog').getByRole('button',{name:'Добавить файлы',exact:true}).click();
  await expect(page.getByRole('alert')).toBeVisible();
  const lost = await head();
  await page.getByRole('dialog').getByRole('button',{name:'Добавить файлы',exact:true}).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  expect((await head()).id).toBe(lost.id);
  expect(lost.entries.map(e=>e.path)).toEqual(['book','book/empty','book/章.md','cover.bin']);
  const file = lost.entries.find(e=>e.path==='book/章.md')!;
  const binary = lost.entries.find(e=>e.path==='cover.bin')!;
  expect(await (await request.get(`${root}/versions/${lost.id}/entries/${binary.id}`,{headers})).body()).toEqual(Buffer.from([0,255,13,10]));
  const body = '# 新しい原稿\n', sha256 = createHash('sha256').update(body).digest('hex');
  expect((await request.put(`${root}/objects/${sha256}`,{headers,data:body})).status()).toBe(204);
  expect((await request.post(`${root}/versions`,{headers,data:{expectedVersion:lost.id,operationId:'rename-and-edit',entries:lost.entries.map(e=>e.id===file.id?{...e,path:'book/renamed.md',sha256,bytes:Buffer.byteLength(body)}:e)}})).status()).toBe(201);
  await page.getByRole('button',{name:'Обновить файлы',exact:true}).click();
  await expect(page.getByRole('button',{name:'Обновить файлы',exact:true})).toBeEnabled();
  await page.getByRole('button',{name:'История проекта',exact:true}).click();
  await page.locator('.revision-row').filter({hasText:lost.id.slice(-8)}).click();
  await page.getByRole('button',{name:'Сравнить с текущей',exact:true}).click();
  await expect(page.getByRole('dialog')).toContainText('Перемещено · Изменено');
  await expect(page.getByLabel('Изменения текста')).toContainText('− # 原稿');
  await expect(page.getByLabel('Изменения текста')).toContainText('+ # 新しい原稿');
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBeTruthy();
  await page.screenshot({path:info.outputPath(`comparison-${isMobile?'mobile':'desktop'}.png`),fullPage:true});
});
