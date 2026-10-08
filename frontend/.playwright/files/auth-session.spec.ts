import { loginForTest } from './auth';
import { expect, test } from '@playwright/test';

test('rejected token redirects to login and returns to the requested page', async ({page}) => {
  await page.addInitScript(() => localStorage.setItem('open_edda_token','expired-test-token'));
  await page.goto('/projects?resume=1#list');
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.getByRole('status')).toHaveText('Сессия завершилась. Войдите снова, чтобы продолжить работу.');
  expect(await page.evaluate(() => localStorage.getItem('open_edda_token'))).toBeNull();
  await page.getByLabel('Email',{exact:true}).fill('browser@example.invalid');
  await page.getByLabel('Пароль',{exact:true}).fill('browser-test-password');
  await expect.poll(async () => {
    const response = page.waitForResponse(r => r.url().endsWith('/api/auth/login') && r.request().method() === 'POST');
    await page.getByRole('button',{name:'Войти',exact:true}).click();
    const result = await response;
    if (result.status() === 429) return false;
    expect(result.ok()).toBeTruthy();
    return true;
  }, {intervals:[2100], timeout:10000}).toBe(true);
  await expect(page).toHaveURL(/\/projects\?resume=1#list$/);
  await expect(page.getByRole('button',{name:'Настройки вида',exact:true}).locator('svg')).toHaveClass(/lucide-palette/);
  await expect(page.getByRole('link',{name:'Настройки сервиса',exact:true}).locator('svg')).toHaveClass(/lucide-settings/);
});

test('expiry during save preserves the draft across reauthentication', async ({page,request},info) => {
  const login = await loginForTest(request);
  const {token} = await login.json() as {token:string};
  const response=await request.post('/api/projects',{headers:{Authorization:`Bearer ${token}`},data:{title:`Session draft ${info.project.name} ${Date.now()}`,storageMode:'files'}});
  expect(response.ok()).toBeTruthy();
  const {id}=await response.json() as {id:string};
  await page.addInitScript(value=>localStorage.setItem('open_edda_token',value),token);
  await page.goto(`/projects/${id}/files`);
  if(info.project.name==='mobile') await page.getByRole('button',{name:'Файлы проекта',exact:true}).click();
  await page.getByRole('button',{name:'Создать файл',exact:true}).click();
  await page.getByLabel('Имя',{exact:true}).fill('draft.md');
  await page.getByRole('button',{name:'Создать',exact:true}).click();
  const text=page.getByRole('textbox',{name:'Текст файла',exact:true});
  await expect(text).toBeVisible();await text.fill('Черновик переживает повторный вход');
  await page.route(`**/api/projects/${id}/files/objects/*`,route=>route.fulfill({status:401,contentType:'application/json',body:'{"error":"invalid or expired token"}'}),{times:1});
  await page.getByRole('button',{name:'Сохранить файл',exact:true}).click();
  await expect(page).toHaveURL(/\/login$/);
  await page.getByLabel('Email',{exact:true}).fill('browser@example.invalid');
  await page.getByLabel('Пароль',{exact:true}).fill('browser-test-password');
  await expect.poll(async () => {
    const response = page.waitForResponse(r => r.url().endsWith('/api/auth/login') && r.request().method() === 'POST');
    await page.getByRole('button',{name:'Войти',exact:true}).click();
    const result = await response;
    if (result.status() === 429) return false;
    expect(result.ok()).toBeTruthy();
    return true;
  }, {intervals:[2100], timeout:10000}).toBe(true);
  await expect(page).toHaveURL(new RegExp(`/projects/${id}/files$`));
  await expect(text).toHaveValue('Черновик переживает повторный вход');
});

test('refresh cookie silently recovers access and logout revokes it', async ({page,context}) => {
  const login = await loginForTest(page.request,true);
  const body = await login.json() as {token:string;refreshToken?:string;refreshExpiresAt:number};
  expect(body.refreshToken).toBeUndefined();
  const cookie = (await context.cookies()).find(c=>c.name==='edda_refresh');
  expect(cookie?.httpOnly).toBe(true);
  expect(cookie?.sameSite).toBe('Strict');
  await page.goto('/login');
  await page.evaluate(() => {
    localStorage.setItem('open_edda_token','rejected-access-token');
    localStorage.setItem('open_edda_session','test-session');
  });
  const refresh = page.waitForResponse(r=>r.url().endsWith('/api/auth/refresh'));
  await page.goto('/projects?silent=1');
  expect((await refresh).ok()).toBeTruthy();
  await expect(page.getByRole('heading',{name:'Место для ваших историй'})).toBeVisible();
  await expect(page).toHaveURL(/projects\?silent=1$/);
  expect(await page.evaluate(()=>localStorage.getItem('open_edda_token'))).not.toBe('rejected-access-token');
  await page.getByRole('button',{name:'Выйти',exact:true}).click();
  await expect(page).toHaveURL(/login$/);
  const rejected=await page.request.post('/api/auth/refresh',{data:{refreshToken:cookie?.value}});
  expect(rejected.status()).toBe(401);
});
