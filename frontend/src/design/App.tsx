import React, { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Dialog } from 'radix-ui';
import { ArrowDownToLine, ArrowLeft, ArrowRight, Check, ChevronDown, ChevronRight, File, FileText, Folder, FolderInput, FolderPlus, History, Image, Languages, Maximize2, Menu, MessageSquare, Monitor, Moon, PanelLeftClose, PanelRight, Pencil, Plus, Search, Settings2, Sun, Upload, X, type LucideIcon } from 'lucide-react';
import { initialEntries, initialProjects, japanese, kindFor, moveIssue, type Entry, type Project } from './model';

type Theme = 'light' | 'dark' | 'system';
type Revision = { body: string; time: string };
type Saved = { entries: Entry[]; projects: Project[]; theme: Theme; active: string; resume?: string };
const storageKey = 'edda-design-v1';
function readSaved(): Saved | null { try { const value = JSON.parse(localStorage.getItem(storageKey) ?? 'null') as Saved | null; return value && Array.isArray(value.entries) && Array.isArray(value.projects) ? value : null; } catch { return null; } }
const saved = readSaved();
function IconButton({ icon: Icon, label, active, ...props }: { icon: LucideIcon; label: string; active?: boolean } & React.ButtonHTMLAttributes<HTMLButtonElement>) {
 return <button type="button" className={`icon-button${active ? ' active' : ''}`} title={label} aria-label={label} {...props}><Icon size={18} strokeWidth={1.6}/></button>;
}
export function App() {
 const [projects,setProjects] = useState<Project[]>(saved?.projects ?? initialProjects);
 const [entries,setEntries] = useState<Entry[]>(saved?.entries ?? initialEntries);
 const [project,setProject] = useState<string | null>(null);
 const [active,setActive] = useState(saved?.active ?? 'chapter');
 const [resume,setResume] = useState(saved?.resume ?? 'chapter');
 const [split,setSplit] = useState(50);
 const sidebarRef = useRef<HTMLElement>(null);
 const [theme,setTheme] = useState<Theme>(saved?.theme ?? 'system');
 const [systemDark,setSystemDark] = useState(matchMedia('(prefers-color-scheme: dark)').matches);
 const [panel,setPanel] = useState<'notes'|'history'|'assistant'|null>(null);
 const [mobileTree,setMobileTree] = useState(false);
 const [focus,setFocus] = useState(false);
 const [translate,setTranslate] = useState(false);
 const [sourceTab,setSourceTab] = useState(false);
 const [query,setQuery] = useState('');
 const [collapsed,setCollapsed] = useState<Set<string>>(new Set());
 const [dropTarget,setDropTarget] = useState<string | null | undefined>(undefined);
 const [notice,setNotice] = useState('');
 const [error,setError] = useState('');
 const [saveState,setSaveState] = useState('Демо · изменения хранятся в браузере');
 const [modal,setModal] = useState<'settings'|'move'|'new'|'project'|'conflict'|null>(null);
 const [moveId,setMoveId] = useState('');
 const [destination,setDestination] = useState('');
 const [newName,setNewName] = useState('');
 const [newKind,setNewKind] = useState<'text'|'folder'>('text');
 const [fontSize,setFontSize] = useState(19);
 const [serif,setSerif] = useState(true);
 const [revisions,setRevisions] = useState<Record<string,Revision[]>>({});
 const [previewRevision,setPreviewRevision] = useState<Revision | null>(null);
 const [note,setNote] = useState('Проверить ритм первого абзаца. Сохранить ощущение тихого возвращения.');
 const [uploading,setUploading] = useState(false);
 const [failedPreview,setFailedPreview] = useState<string | null>(null);
 const fileInput = useRef<HTMLInputElement>(null);
 const searchInput = useRef<HTMLInputElement>(null);
 const editorRef = useRef<HTMLTextAreaElement>(null);
 const uploadParent = useRef<string | null>(null);
 const urls = useRef<string[]>([]);
 const current = entries.find(e => e.id === active && e.project === project);
 const resumeEntry = entries.find(e => e.id === resume) ?? entries.find(e => e.kind === 'text');
 const resumeProject = projects.find(p => p.id === resumeEntry?.project) ?? projects[0];
 const currentProject = projects.find(p => p.id === project);
 const isDark = theme === 'dark' || (theme === 'system' && systemDark);
 const projectEntries = entries.filter(e => e.project === project);
 const folders = projectEntries.filter(e => e.kind === 'folder');
 useLayoutEffect(() => { const editor = editorRef.current; if(!editor) return; const resize = () => {editor.style.height = 'auto';editor.style.height = `${editor.scrollHeight}px`;}; resize(); const observer = new ResizeObserver(resize); observer.observe(editor.parentElement!); return () => observer.disconnect(); },[current?.id,current?.body,fontSize,serif,translate,sourceTab]);
 const words = current?.body?.trim().split(/\s+/).filter(Boolean).length ?? 0;
 useEffect(() => { const media = matchMedia('(prefers-color-scheme: dark)'); const change = () => setSystemDark(media.matches); media.addEventListener('change',change); return () => media.removeEventListener('change',change); },[]);
 useEffect(() => { document.documentElement.dataset.theme = isDark ? 'dark' : 'light'; document.documentElement.style.colorScheme = isDark ? 'dark' : 'light'; },[isDark]);
 useEffect(() => { setSaveState('Сохраняем черновик…'); const timer = setTimeout(() => { try { localStorage.setItem(storageKey,JSON.stringify({entries:entries.filter(e => !e.temporary),projects,theme,active,resume})); setSaveState('Черновик сохранён в браузере'); } catch { setSaveState('Не удалось сохранить в браузере'); setError('Хранилище браузера недоступно или заполнено. Скачайте текст перед закрытием.'); } },450); return () => clearTimeout(timer); },[entries,projects,theme,active,resume]);
 useEffect(() => { if(!notice) return; const timer = setTimeout(() => setNotice(''),3500); return () => clearTimeout(timer); },[notice]);
 useEffect(() => () => urls.current.forEach(url => URL.revokeObjectURL(url)),[]);
 useEffect(() => { const shortcut = (e: KeyboardEvent) => { if((e.ctrlKey || e.metaKey) && e.key === 'k') { e.preventDefault(); if(matchMedia('(max-width: 800px)').matches) setMobileTree(true); setTimeout(() => searchInput.current?.focus(),0); } if(e.key === 'Escape') { setFocus(false); setMobileTree(false); } }; window.addEventListener('keydown',shortcut); return () => window.removeEventListener('keydown',shortcut); },[]);
 useEffect(() => {
  if(!mobileTree) return;
  const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  const sidebar = sidebarRef.current; sidebar?.querySelector<HTMLButtonElement>('button')?.focus();
  const trap = (event: KeyboardEvent) => {if(event.key !== 'Tab' || !sidebar) return; const elements = Array.from(sidebar.querySelectorAll<HTMLElement>('button, input, a[href], select, [tabindex="0"]')).filter(el => el.offsetParent !== null); const first = elements[0], last = elements.at(-1); if(event.shiftKey && document.activeElement === first) {event.preventDefault();last?.focus();} else if(!event.shiftKey && document.activeElement === last) {event.preventDefault();first?.focus();}};
  const media = matchMedia('(max-width: 800px)'); const closeDesktop = () => {if(!media.matches) setMobileTree(false);};
  document.addEventListener('keydown',trap);media.addEventListener('change',closeDesktop);
  return () => {document.removeEventListener('keydown',trap);media.removeEventListener('change',closeDesktop);if(previous?.isConnected) previous.focus();};
 },[mobileTree]);
 function openProject(id: string) { setProject(id); const last = entries.find(e => e.project === id && e.id === active && e.kind !== 'folder'); setActive(last?.id ?? entries.find(e => e.project === id && e.kind !== 'folder')?.id ?? ''); const firstText = last?.kind === 'text' ? last : entries.find(e => e.project === id && e.kind === 'text'); if(firstText) setResume(firstText.id); setTranslate(id === 'letters'); setSourceTab(false); setPanel(null); setQuery(''); }
 function openEntry(entry: Entry) { setActive(entry.id); if(entry.kind === 'text') setResume(entry.id); setMobileTree(false); setPreviewRevision(null); setFailedPreview(null); }
 function rememberRevision(entry: Entry) { setRevisions(old => ({...old,[entry.id]:[{body:entry.body ?? '',time:new Date().toLocaleTimeString('ru',{hour:'2-digit',minute:'2-digit'})},...(old[entry.id] ?? [])].slice(0,12)})); }
 function edit(body: string) { if(!current) return; setEntries(old => old.map(e => e.id === current.id ? {...e,body} : e)); }
 function move(id: string,parent: string | null) { const issue = moveIssue(entries,id,parent); if(issue) { setError(issue); return false; } setEntries(old => old.map(e => e.id === id ? {...e,parent} : e)); if(parent) setCollapsed(old => new Set([...old].filter(x => x !== parent))); setNotice('Перемещено'); setError(''); return true; }
 async function upload(files: FileList | File[], parent: string | null) {
  if(!project) return;
  setUploading(true); setError('');
  const additions: Entry[] = [];
  try {
   for(const file of Array.from(files)) {
    if(file.size > 30 * 1024 * 1024) throw new Error('Для прототипа выберите файлы до 30 МБ.');
    if([...entries,...additions].some(e => e.project === project && e.parent === parent && e.name === file.name)) throw new Error(`«${file.name}» уже есть в папке. Переименуйте файл и повторите.`);
    const kind = kindFor(file); const url = URL.createObjectURL(file); urls.current.push(url);
    additions.push({id:crypto.randomUUID(),project,name:file.name,parent,kind,url,body:kind === 'text' ? await file.text() : undefined,size:file.size,mime:file.type,temporary:true});
   }
   setEntries(old => [...old,...additions]); if(parent) setCollapsed(old => new Set([...old].filter(x => x !== parent))); if(additions[0]) openEntry(additions[0]); setNotice('Открыто локально · доступно до перезагрузки');
  } catch(cause) { setError(cause instanceof Error ? cause.message : 'Не удалось открыть файлы.'); } finally { setUploading(false); }
 }
 function drop(e: React.DragEvent, parent: string | null) { e.preventDefault(); e.stopPropagation(); setDropTarget(undefined); if(e.dataTransfer.files.length) { void upload(e.dataTransfer.files,parent); return; } const id = e.dataTransfer.getData('application/x-edda-entry'); if(id) move(id,parent); }
 function dragOver(e: React.DragEvent,parent: string | null) { if(!e.dataTransfer.types.some(t => t === 'application/x-edda-entry' || t === 'Files')) return; e.preventDefault(); e.stopPropagation(); e.dataTransfer.dropEffect = e.dataTransfer.types.includes('Files') ? 'copy' : 'move'; setDropTarget(parent); }
 function download() { if(!current) return; const url = current.kind === 'text' ? URL.createObjectURL(new Blob([current.body ?? ''],{type:'text/plain;charset=utf-8'})) : current.url; if(!url) return; const a = document.createElement('a'); a.href = url; a.download = current.name; a.click(); if(current.kind === 'text') setTimeout(() => URL.revokeObjectURL(url),1000); }
 function create() {
  const name = newName.trim(); if(!name) return;
  if(name.includes('/') || name === '.' || name === '..') { setError('Введите имя без /; папку назначения выберите отдельно.'); return; }
  if(modal === 'project') { const id = crypto.randomUUID(); setProjects(old => [...old,{id,title:name,subtitle:'Новый проект'}]); setProject(id); setActive(''); setTranslate(false); setPanel(null); }
  else { if(!project) return; const parent = destination || null; if(entries.some(e => e.project === project && e.parent === parent && e.name === name)) { setError('Это имя уже занято в выбранной папке.'); return; } const entry: Entry = {id:crypto.randomUUID(),project,name,parent,kind:newKind,body:newKind === 'text' ? '' : undefined}; setEntries(old => [...old,entry]); if(newKind === 'text') openEntry(entry); if(parent) setCollapsed(old => new Set([...old].filter(x => x !== parent))); }
  setModal(null); setNewName(''); setError('');
 }
 function showNew(kind: 'text'|'folder') { setNewKind(kind); setNewName(''); setDestination(''); setModal('new'); setError(''); }
 function path(id: string): string { const item = entries.find(e => e.id === id); return item ? (item.parent ? `${path(item.parent)} / ` : '') + item.name : ''; }
 function tree(parent: string | null,depth = 0): React.ReactNode {
  return projectEntries.filter(e => e.parent === parent).sort((a,b) => Number(b.kind === 'folder')-Number(a.kind === 'folder')).map(entry => {
   const isFolder = entry.kind === 'folder';
   if(query && !isFolder && !entry.name.toLocaleLowerCase().includes(query.toLocaleLowerCase())) return null;
   const Icon = isFolder ? Folder : entry.kind === 'image' ? Image : entry.kind === 'text' ? FileText : File;
   const folded = collapsed.has(entry.id) && !query;
   return <div key={entry.id} className="tree-branch">
    <div className={`tree-row ${active === entry.id ? 'selected' : ''} ${dropTarget === entry.id ? 'drop-target' : ''}`} style={{paddingLeft:12+depth*14}} draggable onDragStart={e => { e.dataTransfer.setData('application/x-edda-entry',entry.id); e.dataTransfer.effectAllowed='move'; }} onDragEnd={() => setDropTarget(undefined)} onDragOver={isFolder ? e => dragOver(e,entry.id) : undefined} onDrop={isFolder ? e => drop(e,entry.id) : undefined}>
     <button type="button" className="tree-label" aria-expanded={isFolder ? !folded : undefined} aria-current={active === entry.id ? 'page' : undefined} onClick={() => isFolder ? setCollapsed(old => {const next = new Set(old); if(next.has(entry.id)) next.delete(entry.id); else next.add(entry.id); return next;}) : openEntry(entry)} title={entry.name}>
      {isFolder ? folded ? <ChevronRight size={13}/> : <ChevronDown size={13}/> : <span className="tree-indent"/>}<Icon size={16}/><span>{entry.name}</span>
     </button>
     <button type="button" className="tree-move" aria-label={`Переместить ${entry.name}`} title="Переместить" onClick={() => {setMoveId(entry.id);setDestination(entry.parent ?? '');setModal('move');setError('');}}><FolderInput size={14}/></button>
    </div>
    {isFolder && !folded && tree(entry.id,depth+1)}
   </div>;
  });
 }
 const settingsButton = <IconButton icon={Settings2} label="Настройки вида" onClick={() => setModal('settings')}/>;
 return <div className={`edda ${focus ? 'focus-mode' : ''}`}>
  <input ref={fileInput} type="file" multiple hidden onChange={e => {if(e.target.files) void upload(e.target.files,uploadParent.current); e.target.value='';}}/>
  {!project ? <div className="dashboard">
   <header className="dashboard-header"><a className="wordmark" href="/design.html">edda<span className="brand-dot"/></a><div className="header-actions"><span className="demo-label">Дизайн-прототип</span>{settingsButton}</div></header>
   <main className="project-main"><div className="project-heading"><div><h1>Место для ваших историй</h1><p>Рукописи, переводы и всё, что помогает им появиться.</p></div><IconButton icon={Plus} label="Создать проект" onClick={() => {setNewName('');setModal('project');setError('');}}/></div>
    <div className="continue-project"><div className="book-spine" aria-hidden="true"><span>{resumeProject?.title}</span></div><div><span className="muted">Продолжить работу</span><h2>{resumeProject?.title ?? 'Ваш первый проект'}</h2><p>{resumeEntry?.name ?? 'Начните новый текст'}</p><button className="text-action" onClick={() => {if(resumeProject) openProject(resumeProject.id); if(resumeEntry) {setActive(resumeEntry.id);setResume(resumeEntry.id);}}}>Открыть рукопись <ArrowRight size={17}/></button></div><p className="continue-excerpt">{resumeEntry?.body ? `«${resumeEntry.body.slice(0,55)}…»` : 'История начинается с первой строки.'}</p></div>
    <div className="list-heading"><h2>Ваши проекты</h2><span>{projects.length}</span></div>
    <div className="project-list">{projects.map(p => <button key={p.id} className="project-row" onClick={() => openProject(p.id)}><span className="project-monogram">{p.title.slice(0,1)}</span><span><strong>{p.title}</strong><small>{p.subtitle}</small></span><ArrowRight size={18}/></button>)}</div>
    <footer className="project-footer">Демонстрационные проекты · изменения остаются в этом браузере</footer>
   </main>
  </div> : <>
   {mobileTree && <button className="tree-backdrop" aria-label="Закрыть список файлов" onClick={() => setMobileTree(false)}/>}
   <aside ref={sidebarRef} className={`sidebar ${mobileTree ? 'mobile-open' : ''}`} aria-label="Навигация проекта">
    <div className="sidebar-brand"><button className="wordmark" onClick={() => {setProject(null);setFocus(false);}}>edda<span className="brand-dot"/></button><IconButton icon={PanelLeftClose} label="Закрыть файлы" className="icon-button mobile-only" onClick={() => setMobileTree(false)}/></div>
    <button className="project-switch" onClick={() => {setProject(null);setFocus(false);}}><span>{currentProject?.title}</span><ChevronDown size={16}/></button>
    <label className="search-box"><Search size={15}/><input ref={searchInput} aria-label="Найти файл" value={query} onChange={e => setQuery(e.target.value)} placeholder="Найти файл"/><kbd>Ctrl K</kbd></label>
    <div className="tree-heading"><span>Файлы проекта</span><div><IconButton icon={Plus} label="Создать файл" onClick={() => showNew('text')}/><IconButton icon={FolderPlus} label="Создать папку" onClick={() => showNew('folder')}/><IconButton icon={Upload} label="Открыть локальные файлы" onClick={() => {uploadParent.current=null;fileInput.current?.click();}}/></div></div>
    <nav className="file-tree" aria-label="Файлы и папки">{tree(null)}{query && !projectEntries.some(e => e.kind !== 'folder' && e.name.toLowerCase().includes(query.toLowerCase())) && <p className="tree-empty">Файлы не найдены</p>}{!projectEntries.length && <p className="tree-empty">Добавьте первый текст или перетащите файлы сюда.</p>}</nav>
    <div className={`root-drop ${dropTarget === null ? 'drop-target' : ''}`} onDragOver={e => dragOver(e,null)} onDragLeave={e => {if(!e.currentTarget.contains(e.relatedTarget as Node)) setDropTarget(undefined);}} onDrop={e => drop(e,null)}><Folder size={15}/><span>{uploading ? 'Открываем файлы…' : 'Перетащите сюда → в корень'}</span></div>
    <div className="sidebar-footer"><span className="demo-label">Демо · локальная копия</span>{settingsButton}</div>
   </aside>
   <div className="work-area" inert={mobileTree}>
    <header className="workspace-header"><IconButton icon={Menu} label="Файлы проекта" className="icon-button mobile-only" onClick={() => setMobileTree(true)}/><div className="breadcrumbs"><button onClick={() => setProject(null)}>{currentProject?.title}</button><ChevronRight size={13}/><span>{current?.name ?? 'Проект'}</span></div><div className="header-actions">{current && <><IconButton icon={ArrowDownToLine} label="Скачать файл" onClick={download}/>{current.kind === 'text' && <><IconButton icon={Languages} label="Оригинал и перевод" active={translate} aria-pressed={translate} onClick={() => {setTranslate(!translate);setSourceTab(false);}}/><IconButton icon={Maximize2} label={focus ? 'Выйти из режима сосредоточения' : 'Сосредоточиться на тексте'} aria-pressed={focus} onClick={() => {setFocus(!focus);setPanel(null);}}/></>}<IconButton icon={History} label="История документа" active={panel === 'history'} onClick={() => {setPanel(panel === 'history' ? null : 'history');setPreviewRevision(null);}}/></>}<IconButton icon={PanelRight} label="Заметки к тексту" active={panel === 'notes'} onClick={() => setPanel(panel === 'notes' ? null : 'notes')}/></div></header>
    {error && !modal && <div className="error-banner" role="alert"><span>{error}</span><IconButton icon={X} label="Закрыть сообщение" onClick={() => setError('')}/></div>}
    <div className="work-body">
     <main className={`document-stage ${translate && current?.kind === 'text' ? 'translation-mode' : ''}`}>
      {current ? current.kind === 'text' ? <>
       <div className="document-toolbar"><span><Pencil size={14}/> {translate ? 'Перевод' : 'Рукопись'}</span>{translate && <label className="split-control">Ширина оригинала<input aria-label="Ширина оригинала" type="range" min={30} max={70} value={split} onChange={e => setSplit(Number(e.target.value))}/></label>}<button className="subtle-action" onClick={() => {rememberRevision(current);setPanel('history');setNotice('Версия сохранена для этой сессии');}}>Сохранить версию</button></div>
       {translate && <div className="translation-tabs" role="tablist" aria-label="Сторона перевода"><button role="tab" aria-selected={!sourceTab} onClick={() => setSourceTab(false)}>Перевод</button><button role="tab" aria-selected={sourceTab} onClick={() => setSourceTab(true)}>Оригинал</button></div>}
       <div className={`writing-columns ${sourceTab ? 'show-source' : ''}`} style={translate ? {gridTemplateColumns:`${split}fr ${100-split}fr`} : undefined}>
        {translate && <section className="source-column" aria-label="Оригинал"><div className="column-label">日本語 <span>Оригинал · демо</span></div><h1 lang="ja">帰郷</h1><div className="source-prose" lang="ja" style={{fontSize}}>{japanese.split('\n\n').map((p,i) => <p key={i}>{p}</p>)}</div></section>}
        <section className="prose-column" aria-label="Текст документа">{translate && <div className="column-label">Русский <span>Перевод</span></div>}<h1>{current.name.replace(/\.(md|txt)$/i,'').replace(/^\d+\s*[—-]\s*/,'')}</h1><textarea ref={editorRef} key={current.id} className={`prose-editor ${serif ? '' : 'sans-prose'}`} aria-label="Текст рукописи" value={current.body ?? ''} onFocus={() => {if(!revisions[current.id]?.length) rememberRevision(current);}} onChange={e => edit(e.target.value)} placeholder="Начните с первой строки…" spellCheck={false} style={{fontSize}}/></section>
       </div>
      </> : <section className="file-page"><div className="file-page-title"><div><h1>{current.name}</h1><p>{current.mime || 'Файл'}{current.size !== undefined ? ` · ${new Intl.NumberFormat('ru').format(current.size)} байт` : ''}</p></div></div>
       {current.temporary && <p className="file-session-note">Открыт с вашего устройства. Доступен до перезагрузки страницы.</p>}
       {current.kind === 'pdf' && <p className="file-session-note">Просмотр PDF зависит от браузера. Если документ не появился, <button className="inline-download" onClick={download}>скачайте файл</button>.</p>}
       {failedPreview !== current.id && current.kind === 'image' ? <div className="image-preview"><img src={current.url} alt={`Предпросмотр ${current.name}`} onError={() => setFailedPreview(current.id)}/></div> : failedPreview !== current.id && current.kind === 'pdf' && navigator.pdfViewerEnabled ? <object className="pdf-preview" type="application/pdf" data={current.url} aria-label={`Просмотр ${current.name}`}><p>Браузер не поддерживает просмотр PDF. <button className="text-action" onClick={download}>Скачать файл</button></p></object> : failedPreview !== current.id && current.kind === 'audio' ? <div className="media-preview"><audio controls src={current.url} onError={() => setFailedPreview(current.id)} aria-label={current.name}/></div> : failedPreview !== current.id && current.kind === 'video' ? <div className="media-preview"><video controls src={current.url} onError={() => setFailedPreview(current.id)} aria-label={current.name}/></div> : <div className="binary-preview"><File size={54} strokeWidth={1}/><h2>Файл готов к скачиванию</h2><p>{current.kind === 'pdf' ? 'Этот браузер не поддерживает встроенный просмотр PDF.' : failedPreview === current.id ? 'Этот формат не удалось показать в браузере.' : 'У этого формата нет встроенного просмотра.'}<br/>Скачайте его и откройте в подходящем приложении.</p><button className="quiet-button" onClick={download}><ArrowDownToLine size={17}/>Скачать файл</button></div>}
      </section> : <div className="empty-workspace"><FileText size={32} strokeWidth={1.2}/><h1>Здесь начинается история</h1><p>Откройте файл слева или создайте первый текст.</p><button className="quiet-button" onClick={() => showNew('text')}><Plus size={17}/>Новый текст</button></div>}
     </main>
     {panel && <aside className="context-panel" aria-label={panel === 'notes' ? 'Заметки' : panel === 'history' ? 'История' : 'Помощник'}><header><h2>{panel === 'notes' ? 'На полях' : panel === 'history' ? 'История документа' : 'Помощник'}</h2><IconButton icon={X} label="Закрыть панель" onClick={() => {setPanel(null);setPreviewRevision(null);}}/></header>
      {panel === 'notes' ? <><p className="muted">Мысли, к которым хочется вернуться.</p><textarea className="note-editor" aria-label="Заметка" value={note} onChange={e => setNote(e.target.value)}/><p className="fine-print">Заметка хранится только до перезагрузки.</p><div className="context-divider"/><button className="context-link" onClick={() => setPanel('assistant')}><MessageSquare size={17}/>Обсудить с помощником<ArrowRight size={15}/></button></> : panel === 'assistant' ? <><p className="muted">Помощник доступен по запросу и работает рядом с текстом.</p><p className="fine-print">В этом прототипе ИИ не подключён. Текст никуда не отправляется.</p><button className="quiet-button" onClick={() => setPanel('notes')}><ArrowLeft size={16}/>К заметкам</button></> : <><p className="fine-print">Локальные версии этой сессии. Серверная история не подключена.</p>{!current || current.kind !== 'text' ? <p className="muted">Выберите текстовый документ.</p> : <><button className="quiet-button" onClick={() => {rememberRevision(current);setNotice('Версия сохранена');}}><Plus size={16}/>Сохранить версию</button>{!(revisions[current.id]?.length) && <p className="muted">Сохраните версию, чтобы вернуться к ней позже.</p>}{revisions[current.id]?.map((revision,i) => <button className={`revision-row ${previewRevision === revision ? 'chosen' : ''}`} key={i} onClick={() => setPreviewRevision(revision)}><History size={16}/><span>Версия {revisions[current.id].length-i}<small>Сегодня, {revision.time}</small></span><ChevronRight size={15}/></button>)}{previewRevision && <div className="revision-preview"><h3>Предпросмотр</h3><p>{previewRevision.body || 'Пустой документ'}</p><button className="quiet-button" onClick={() => {rememberRevision(current);edit(previewRevision.body);setPreviewRevision(null);setNotice('Версия восстановлена. Предыдущий текст сохранён в истории.');}}>Восстановить эту версию</button></div>}<button className="subtle-action conflict-demo" onClick={() => setModal('conflict')}>Посмотреть пример конфликта</button></> }</>}
     </aside>}
    </div>
    <footer className="status-bar"><span role="status"><span className="status-dot"/>{current?.temporary ? 'Локальный файл · до перезагрузки' : saveState}</span><span>{current?.kind === 'text' ? `${words} ${new Intl.PluralRules('ru').select(words) === 'one' ? 'слово' : new Intl.PluralRules('ru').select(words) === 'few' ? 'слова' : 'слов'}` : 'Дизайн-прототип'}<span className="desktop-status"> · {translate ? 'JA → RU' : 'Русский'}</span></span></footer>
   </div>
  </>}
  {notice && <div className="toast" role="status"><Check size={16}/>{notice}</div>}
  <Dialog.Root open={modal !== null} onOpenChange={open => {if(!open) {setModal(null);setError('');}}}><Dialog.Portal><Dialog.Overlay className="dialog-overlay"/><Dialog.Content className={`dialog ${modal === 'conflict' ? 'conflict-dialog' : ''}`}><Dialog.Title>{modal === 'settings' ? 'Удобно именно вам' : modal === 'move' ? 'Переместить' : modal === 'project' ? 'Новая история' : modal === 'conflict' ? 'Две версии одного текста' : newKind === 'folder' ? 'Новая папка' : 'Новый текст'}</Dialog.Title><Dialog.Description>{modal === 'settings' ? 'Настройте пространство для спокойной работы.' : modal === 'move' ? entries.find(e => e.id === moveId)?.name : modal === 'project' ? 'Дайте проекту имя. Структуру можно выбрать позже.' : modal === 'conflict' ? 'Пример: файл изменился на другом устройстве. Ваш текст сохранён.' : 'Файлы и папки можно перемещать в любое время.'}</Dialog.Description><Dialog.Close className="dialog-close icon-button" aria-label="Закрыть"><X size={18}/></Dialog.Close>
   {error && <p className="form-error" role="alert">{error}</p>}
   {modal === 'settings' ? <><fieldset><legend>Оформление</legend><div className="theme-options">{([{value:'light',label:'Светлое',icon:Sun},{value:'dark',label:'Тёмное',icon:Moon},{value:'system',label:'Системное',icon:Monitor}] as const).map(option => <button key={option.value} className={theme === option.value ? 'chosen' : ''} aria-pressed={theme === option.value} onClick={() => setTheme(option.value)}><option.icon size={20}/>{option.label}</button>)}</div></fieldset><label className="range-label">Размер текста <span>{fontSize} px</span><input type="range" min={16} max={25} value={fontSize} onChange={e => setFontSize(Number(e.target.value))}/></label><label className="check-label"><input type="checkbox" checked={serif} onChange={e => setSerif(e.target.checked)}/>Шрифт с засечками для рукописи</label><p className="fine-print">Тема сохраняется в браузере. Настройки чтения действуют в этой сессии.</p></> : modal === 'conflict' ? <><div className="conflict-columns"><section><h3>Ваш черновик</h3><p>{current?.body?.slice(0,210)}</p></section><section><h3>Другая версия · пример</h3><p>К вечеру море потемнело. Мира остановилась у дома и подняла глаза к единственному тёмному окну.</p></section></div><p className="fine-print">Демонстрация интерфейса. Серверной версии здесь нет; выбор не меняет текст.</p><div className="dialog-actions"><button className="quiet-button" onClick={() => {setModal(null);setNotice('Демо: выбран ваш черновик');}}>Оставить мой текст</button><button className="quiet-button" onClick={() => {setModal(null);setNotice('Демо: выбрано сохранение обеих версий');}}>Сохранить обе версии</button></div></> : <form onSubmit={e => {e.preventDefault(); if(modal === 'move') {if(move(moveId,destination || null)) setModal(null);} else create();}}>{modal !== 'move' && <label className="field-label">{modal === 'project' ? 'Название проекта' : 'Имя'}<input autoFocus value={newName} onChange={e => setNewName(e.target.value)} placeholder={modal === 'project' ? 'Название вашей истории' : newKind === 'folder' ? 'Материалы' : 'Новая глава.md'} required/></label>}{modal !== 'project' && <label className="field-label">Папка<select value={destination} onChange={e => setDestination(e.target.value)}><option value="">Корень проекта</option>{folders.filter(folder => modal !== 'move' || !moveIssue(entries,moveId,folder.id)).map(folder => <option key={folder.id} value={folder.id}>{path(folder.id)}</option>)}</select></label>}<div className="dialog-actions"><button className="quiet-button" type="submit">{modal === 'move' ? 'Переместить' : 'Создать'}<ArrowRight size={16}/></button></div></form>}
  </Dialog.Content></Dialog.Portal></Dialog.Root>
 </div>;
}
