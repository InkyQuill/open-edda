import { useState, useRef } from 'react';
import { ArrowRight, LogOut, Plus, RefreshCw, Settings2, Trash2 } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import { createProject, deleteProject } from '../../api';
import { logout } from '../../authApi';
import { IconButton, Modal } from '../../shared/ui/edda';
import { AppearanceButton } from '../appearance/Appearance';
import { useProjects } from './projectHooks';

export function ProjectsPage() {
  const navigate = useNavigate();
  const { projects, loading, error, reload } = useProjects();
  const [open, setOpen] = useState(false), [title, setTitle] = useState(''), [language, setLanguage] = useState('ru');
  const [creating, setCreating] = useState(false), [createError, setCreateError] = useState('');
  const [deletingProject, setDeletingProject] = useState<typeof projects[number] | null>(null);
  const [confirmation, setConfirmation] = useState(''), [deleting, setDeleting] = useState(false), [deleteError, setDeleteError] = useState('');
  const deletingRef = useRef(false);
  async function remove(event: React.FormEvent) {
    event.preventDefault();
    if (!deletingProject || confirmation !== deletingProject.title || deletingRef.current) return;
    deletingRef.current = true; setDeleting(true); setDeleteError('');
    try {
      await deleteProject(deletingProject.id, confirmation);
      try { if (localStorage.getItem('edda.last-project') === deletingProject.id) localStorage.removeItem('edda.last-project'); } catch { /* Optional resume state. */ }
      setDeletingProject(null); reload();
    } catch (cause) { setDeleteError(cause instanceof Error ? cause.message : 'Не удалось удалить проект. Повторите попытку.'); }
    finally { deletingRef.current = false; setDeleting(false); }
  }
  const [logoutError, setLogoutError] = useState('');
  const [loggingOut, setLoggingOut] = useState(false);
  async function signOut() {
    setLoggingOut(true); setLogoutError('');
    try { await logout(); navigate('/login', {replace:true}); }
    catch (cause) { setLogoutError(cause instanceof Error ? cause.message : 'Не удалось выйти.'); }
    finally { setLoggingOut(false); }
  }
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
  return <div className="edda"><div className="dashboard"><header className="dashboard-header"><Link to="/projects" className="wordmark">edda<span className="brand-dot" /></Link><div className="header-actions"><AppearanceButton /><Link className="icon-button" title="Настройки сервиса" aria-label="Настройки сервиса" to="/settings"><Settings2 size={18} /></Link><IconButton icon={LogOut} label="Выйти" disabled={loggingOut} onClick={() => void signOut()} /></div></header><main className="project-main"><div className="project-heading"><div><h1>Место для ваших историй</h1><p>Рукописи, переводы и всё, что помогает им появиться.</p></div><IconButton icon={Plus} label="Создать проект" onClick={() => setOpen(true)} /></div>
    {logoutError && <div className="error-banner" role="alert">{logoutError}</div>}
    {error && <div className="error-banner" role="alert">{error}<IconButton icon={RefreshCw} label="Повторить загрузку" onClick={reload} /></div>}
    {loading ? <p className="muted" role="status">Загружаем проекты…</p> : resume ? <><div className="continue-project"><div className="book-spine" aria-hidden="true"><span>{resume.title}</span></div><div><span className="muted">Продолжить работу</span><h2>{resume.title}</h2><p>Текст, материалы и сохранённые версии</p><Link className="text-action" to={route(resume)}>Открыть проект <ArrowRight size={17} /></Link></div></div><div className="list-heading"><h2>Ваши проекты</h2><span>{projects.length}</span></div><nav className="project-list" aria-label="Ваши проекты">{projects.map(project => <div key={project.id} className="project-list-item"><Link className="project-row" to={route(project)}><span className="project-monogram" aria-hidden="true">{project.title.slice(0, 1)}</span><span><strong>{project.title}</strong><small>{project.language || 'Язык не указан'}</small></span><ArrowRight size={18} aria-hidden="true" /></Link><IconButton icon={Trash2} label={`Удалить проект «${project.title}»`} onClick={() => { setDeletingProject(project); setConfirmation(''); setDeleteError(''); }} /></div>)}</nav></> : !error && <div className="empty-workspace"><h2>Ваша первая история</h2><p>Создайте проект, затем добавьте текст или перенесите файлы с компьютера.</p><button className="quiet-button" onClick={() => setOpen(true)}><Plus size={17} />Создать проект</button></div>}
  </main></div><Modal open={open} onClose={() => setOpen(false)} title="Новая история" description="Дайте проекту имя. Структуру можно выбрать позже." busy={creating}><form onSubmit={event => void create(event)}><label className="field-label">Название проекта<input autoFocus value={title} onChange={event => setTitle(event.target.value)} required disabled={creating} /></label><label className="field-label">Язык текста<select value={language} onChange={event => setLanguage(event.target.value)} disabled={creating}><option value="ru">Русский</option><option value="en">English</option><option value="ja">日本語</option><option value="">Другой</option></select></label>{createError && <p className="form-error" role="alert">{createError}</p>}<div className="dialog-actions"><button className="quiet-button" disabled={creating} type="submit">{creating ? 'Создаём…' : 'Создать проект'}<ArrowRight size={16} /></button></div></form></Modal><Modal open={deletingProject !== null} onClose={() => { if (!deletingRef.current) setDeletingProject(null); }} title="Удалить проект?" description={`Проект «${deletingProject?.title ?? ''}», все его файлы и история версий будут удалены с сервера без возможности отмены. Копии на компьютере останутся, но синхронизация с этим проектом прекратится.`} busy={deleting}><form onSubmit={event => void remove(event)}><label className="field-label">Введите название проекта для подтверждения<input autoFocus autoComplete="off" value={confirmation} onChange={event => setConfirmation(event.target.value)} disabled={deleting} /></label>{deleteError && <p className="form-error" role="alert">{deleteError}</p>}<div className="dialog-actions"><button type="button" className="quiet-button" disabled={deleting} onClick={() => setDeletingProject(null)}>Отмена</button><button type="submit" className="quiet-button danger-action" disabled={deleting || !deletingProject || confirmation !== deletingProject.title}>{deleting ? 'Удаляем…' : 'Удалить проект'}</button></div></form></Modal></div>;
}
