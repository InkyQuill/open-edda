import { X } from 'lucide-react';
import { IconButton } from '../../shared/ui/edda';
import type { PocketReview, ReviewRecord } from './pocketApi';
import './review.css';

const states = { stale: 'Привязка устарела', ambiguous: 'Несколько совпадений', conflicting: 'Пересекающиеся предложения', resolved: 'В тексте' };
const labels = { note: 'Заметка', change_required: 'Нужна правка', warning: 'Предупреждение', review: 'Замечание' };
export function ReviewPanel({ review, loading, error, busy, dirty, canSave, selected, onSelect, onDecide, onSave, onClose, onRetry, onCompose, onEdit, onNote, onUndo, canUndo, canRedo, onStart }: {
  onStart: () => void; onCompose: () => void; onEdit: (record: ReviewRecord) => void; onNote: () => void; onUndo: (direction: 'undo' | 'redo') => void; canUndo: boolean; canRedo: boolean;
  review: PocketReview | null; loading: boolean; error: string; busy: boolean; dirty: boolean; canSave: boolean; selected: string;
  onSelect: (record: ReviewRecord) => void; onDecide: (action: 'apply' | 'remove', id: string) => void; onSave: () => void; onClose: () => void; onRetry: () => void;
}) {
  const records = review ? [...review.edits, ...review.signals].sort((a, b) => (a.resolution.range?.from ?? Infinity) - (b.resolution.range?.from ?? Infinity)) : [];
  return <aside className="review-panel" aria-label="Рецензия главы" aria-busy={loading || busy}>
    <header><h2>Рецензия главы</h2><IconButton icon={X} label="Скрыть рецензию" onClick={onClose} /></header>
    {loading && <p role="status">Проверяем привязки к тексту…</p>}
    {error && <div role="alert"><p>{error}</p><button className="quiet-button" disabled={busy} onClick={onRetry}>Повторить загрузку</button></div>}
    {!loading && !error && !review && <div><p>Для этого текста ещё нет рецензии.</p><button className="quiet-button" disabled={busy || dirty} onClick={onStart}>Начать рецензию главы</button>{dirty && <p className="fine-print">Сначала сохраните черновики.</p>}</div>}
    {review && <>
      <div className="review-actions"><button className="quiet-button" disabled={busy || dirty} onClick={onCompose}>Добавить к выделению</button><button className="subtle-action" disabled={busy || dirty} onClick={onNote}>Изменить заметку о главе</button><button className="subtle-action" disabled={busy || dirty || !canUndo} onClick={() => onUndo('undo')}>Отменить решение</button><button className="subtle-action" disabled={busy || dirty || !canRedo} onClick={() => onUndo('redo')}>Повторить решение</button></div>
      {dirty && <div className="review-dirty" role="status"><p>Черновики не сохранены. Сохраните их, чтобы проверить привязки и принять или закрыть замечания.</p>{canSave && <button className="quiet-button" disabled={busy} onClick={onSave}>Сохранить и проверить привязки</button>}</div>}
      {review.note && <details className="review-note"><summary>Заметка о главе</summary><p>{review.note}</p></details>}
      <p className="review-count">Предложений: {review.edits.length} · Замечаний: {review.signals.length}</p>
      {!records.length && <p>Все замечания разобраны.</p>}
      <ol className="review-records">{records.map(record => <li key={record.id} data-record-id={record.id} className={selected === record.id ? 'selected' : ''}>
        <button className="review-reveal" disabled={busy || dirty || !record.resolution.range} onClick={() => onSelect(record)} aria-label={`Показать в тексте: ${record.before ?? record.selected_text}`}><span>{record.type ? labels[record.type] : 'Предложение'}</span><small>{states[record.resolution.kind]}</small></button>
        {record.before !== undefined ? <div className="review-replacement"><del>{record.before}</del><ins>{record.after || 'Удалить фрагмент'}</ins></div> : <blockquote>{record.selected_text}</blockquote>}
        {record.comment && <p>{record.comment}</p>}
        {record.resolution.kind !== 'resolved' && <p className="fine-print">Проверьте текст вручную. Автоматическое применение недоступно.</p>}
        <div className="review-actions"><button className="subtle-action" disabled={busy || dirty} onClick={() => onEdit(record)}>Изменить рецензию</button>{record.before !== undefined && <button className="quiet-button" disabled={busy || dirty || record.resolution.kind !== 'resolved'} onClick={() => onDecide('apply', record.id)}>Принять правку</button>}<button className="subtle-action" disabled={busy || dirty} onClick={() => onDecide('remove', record.id)}>{record.before !== undefined ? 'Отклонить' : 'Закрыть замечание'}</button></div>
      </li>)}</ol>
    </>}
  </aside>;
}
