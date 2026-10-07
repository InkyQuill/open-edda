import { useState, useRef } from 'react';
import { ArrowRight, LogOut, Plus, RefreshCw, Settings2 } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import { createProject } from '../../api';
import { clearToken } from '../../authApi';
import { IconButton, Modal } from '../../shared/ui/edda';
import { AppearanceButton } from '../appearance/Appearance';
import { useProjects } from './projectHooks';

export function ProjectsPage() {
  const navigate = useNavigate();
  const { projects, loading, error, reload } = useProjects();
  const [open, setOpen] = useState(false), [title, setTitle] = useState(''), [language, setLanguage] = useState('ru');
  const [creating, setCreating] = useState(false), [createError, setCreateError] = useState('');
  const inFlight = useRef(false);
  let lastProject = '';
  try { lastProject = localStorage.getItem('edda.last-project') ?? ''; } catch { /* Empty resume state. */ }
  const resume = projects.find(project => project.id === lastProject) ?? projects[0];
  const route = (project: typeof projects[number]) => `/projects/${encodeURIComponent(project.id)}${project.storageMode === 'files' ? '/files' : ''}`;
  async function create(event: React.FormEvent) {
    event.preventDefault(); if (!title.trim() || inFlight.current) return;
    inFlight.current = true; setCreating(true); setCreateError('');
    try { const project = await createProject({ title: title.trim(), language: language.trim(), storageMode: 'files' }); navigate(route(project)); }
    catch (cause) { setCreateError(cause instanceof Error ? cause.message : 'Не удалось создать проект. Повторите попытку.'); }
    finally { inFlight.current = false; setCreating(false); }
  }
  return <div className="edda"><div className="dashboard"><header className="dashboard-header"><Link to="/projects" className="wordmark">edda<span className="brand-dot" /></Link><div className="header-actions"><AppearanceButton /><Link className="icon-button" title="Настройки сервиса" aria-label="Настройки сервиса" to="/settings"><Settings2 size={18} /></Link><IconButton icon={LogOut} label="Выйти" onClick={() => { clearToken(); navigate('/login', { replace: true }); }} /></div></header><main className="project-main"><div className="project-heading"><div><h1>Место для ваших историй</h1><p>Рукописи, переводы и всё, что помогает им появиться.</p></div><IconButton icon={Plus} label="Создать проект" onClick={() => setOpen(true)} /></div>
    {error && <div className="error-banner" role="alert">{error}<IconButton icon={RefreshCw} label="Повторить загрузку" onClick={reload} /></div>}
    {loading ? <p className="muted" role="status">Загружаем проекты…</p> : resume ? <><div className="continue-project"><div className="book-spine" aria-hidden="true"><span>{resume.title}</span></div><div><span className="muted">Продолжить работу</span><h2>{resume.title}</h2><p>Текст, материалы и сохранённые версии</p><Link className="text-action" to={route(resume)}>Открыть проект <ArrowRight size={17} /></Link></div></div><div className="list-heading"><h2>Ваши проекты</h2><span>{projects.length}</span></div><nav className="project-list" aria-label="Ваши проекты">{projects.map(project => <Link key={project.id} className="project-row" to={route(project)}><span className="project-monogram" aria-hidden="true">{project.title.slice(0, 1)}</span><span><strong>{project.title}</strong><small>{project.language || 'Язык не указан'}</small></span><ArrowRight size={18} aria-hidden="true" /></Link>)}</nav></> : !error && <div className="empty-workspace"><h2>Ваша первая история</h2><p>Создайте проект, затем добавьте текст или перенесите файлы с компьютера.</p><button className="quiet-button" onClick={() => setOpen(true)}><Plus size={17} />Создать проект</button></div>}
  </main></div><Modal open={open} onClose={() => setOpen(false)} title="Новая история" description="Дайте проекту имя. Структуру можно выбрать позже." busy={creating}><form onSubmit={event => void create(event)}><label className="field-label">Название проекта<input autoFocus value={title} onChange={event => setTitle(event.target.value)} required disabled={creating} /></label><label className="field-label">Язык текста<select value={language} onChange={event => setLanguage(event.target.value)} disabled={creating}><option value="ru">Русский</option><option value="en">English</option><option value="ja">日本語</option><option value="">Другой</option></select></label>{createError && <p className="form-error" role="alert">{createError}</p>}<div className="dialog-actions"><button className="quiet-button" disabled={creating} type="submit">{creating ? 'Создаём…' : 'Создать проект'}<ArrowRight size={16} /></button></div></form></Modal></div>;
}
