// Inspect the deployed UI with the test project created by verify-live.py.
import { createRequire } from 'node:module';
import { readFile, mkdir } from 'node:fs/promises';
import { homedir } from 'node:os';
import { resolve } from 'node:path';
const require = createRequire(new URL('../frontend/package.json', import.meta.url));
const { chromium } = require('@playwright/test');
const account = JSON.parse(await readFile(resolve(homedir(), '.config/open-edda/bootstrap-login.json'), 'utf8'));
const report = JSON.parse(await readFile(process.env.EDDA_VERIFY_REPORT, 'utf8'));
const login = await fetch(`${account.server}/api/auth/login`, { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({email:account.email,password:account.password}) });
if (!login.ok) throw new Error(`Login failed: ${login.status}`);
const {token} = await login.json();
const output = resolve('frontend/.playwright/test-results/live');
await mkdir(output, {recursive:true});
const browser = await chromium.launch();
try {
 for (const [name,viewport] of [['desktop',{width:1365,height:900}],['mobile',{width:390,height:844}]]) {
  const context = await browser.newContext({viewport});
  await context.addInitScript(token=>localStorage.setItem('open_edda_token',token),token);
  const page = await context.newPage();const errors=[];page.on('pageerror',error=>errors.push(error.message));
  await page.goto(`${account.server}/projects/${report.project}/files`);
  await page.getByRole('heading',{name:report.title,exact:true}).waitFor();
  if(name==='mobile')await page.getByRole('button',{name:'Files',exact:true}).click();
  await page.getByRole('button',{name:'chapter.md',exact:true}).click();
  await page.getByLabel('File text',{exact:true}).waitFor();
  const text=await page.getByLabel('File text',{exact:true}).inputValue();
  if(!text.includes('Проверочная глава'))throw new Error('Wrong downloaded chapter');
  if(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth))throw new Error('Horizontal overflow');
  if(errors.length)throw new Error(errors.join('\n'));
  await page.screenshot({path:resolve(output,`${name}.png`),fullPage:true});
  await context.close();
 }
 console.log('PASS: production desktop/mobile UI, authenticated chapter read, no browser exceptions or overflow');
}finally{await browser.close();}
