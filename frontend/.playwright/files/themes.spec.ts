import { expect, test } from '@playwright/test';
import { loginForTest } from './auth';

test('shared themes apply to app, dialogs and editor and follow system changes',async({page,request,isMobile},info)=>{
 const {token}=await(await loginForTest(request)).json() as {token:string};
 await page.addInitScript(value=>localStorage.setItem('open_edda_token',value),token);
 const response=await request.post('/api/projects',{headers:{Authorization:`Bearer ${token}`},data:{title:`Themes ${info.project.name}`,storageMode:'files'}});expect(response.ok()).toBeTruthy();const {id}=await response.json() as {id:string};
 await page.goto(`/projects/${id}/files`);
 await expect(page.getByRole('heading',{name:`Themes ${info.project.name}`,exact:true})).toBeVisible();
 await page.locator('input[type=file]').first().setInputFiles({name:'chapter.md',mimeType:'text/markdown',buffer:Buffer.from('# Тема\n\nТекст **рукописи**.')});
 await page.getByRole('dialog').getByRole('button',{name:'Добавить файлы',exact:true}).click();
 await expect(page.locator('.file-galley .cm-content')).toBeVisible();
 if(isMobile)await page.getByRole('button',{name:'Файлы проекта',exact:true}).click();
 await page.getByRole('button',{name:'Настройки вида',exact:true}).click();
 const family=page.getByRole('combobox',{name:'Тема',exact:true});
 const names=await family.locator('option').allTextContents();
 for(const name of names){
  await family.selectOption({label:name});
  for(const mode of ['Светлое','Тёмное']){
   await page.getByRole('button',{name:mode,exact:true}).click();
   await expect(page.locator('html')).toHaveAttribute('data-theme',mode==='Тёмное'?'dark':'light');
   await expect.poll(()=>page.evaluate(()=>{
    const root=getComputedStyle(document.documentElement),editor=getComputedStyle(document.querySelector('.file-galley .ge-editor-shell')!);
    return root.getPropertyValue('--paper').trim()===editor.getPropertyValue('--ge-color-bg').trim();
   })).toBeTruthy();
  }
 }
 await family.selectOption('Edda');await page.getByRole('button',{name:'Системное',exact:true}).click();
 await page.emulateMedia({colorScheme:'light'});await expect(page.locator('html')).toHaveAttribute('data-theme-id','edda-light');
 await page.emulateMedia({colorScheme:'dark'});await expect(page.locator('html')).toHaveAttribute('data-theme-id','edda-dark');
 await page.getByRole('button',{name:'Закрыть',exact:true}).click();
 if(isMobile)await page.getByRole('button',{name:'Закрыть файлы',exact:true}).click();
 await page.screenshot({path:info.outputPath('edda-dark-editor.png'),fullPage:true});
 await page.reload();await expect(page.locator('html')).toHaveAttribute('data-theme-id','edda-dark');
});
