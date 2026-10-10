import { useState } from 'react';
import type { ReviewChange, ReviewRecord } from './pocketApi';

export type ReviewComposition = { mode: 'add'; selection: { from: number; to: number; selected: string } } | { mode: 'update'; record: ReviewRecord } | { mode: 'note'; note: string };
export function ReviewComposer({ initial, busy, onSubmit }: { initial: ReviewComposition; busy: boolean; onSubmit: (change: ReviewChange) => void }) {
  const [id] = useState(() => initial.mode === 'update' ? initial.record.id : crypto.randomUUID());
  const initialType = initial.mode === 'update' ? initial.record.type ?? 'edit' : 'note';
  const initialText = initial.mode === 'note' ? initial.note : initial.mode === 'update' ? initial.record.after ?? initial.record.comment ?? '' : '';
  const [type, setType] = useState(initialType);
  const [text, setText] = useState(initialText);
  const selected = initial.mode === 'add' ? initial.selection.selected : initial.mode === 'update' ? initial.record.before ?? initial.record.selected_text : '';
  return <form onSubmit={event => { event.preventDefault(); onSubmit(initial.mode === 'note' ? { action: 'note', note: text } : { action: initial.mode, recordId: id, signalType: type, after: text, comment: text, ...(initial.mode === 'add' ? initial.selection : {}) }); }}>
    {selected && <blockquote className="review-quote">{selected}</blockquote>}
    {initial.mode !== 'note' && <label className="field-label">Тип<select aria-label="Тип рецензии" value={type} disabled={busy || (initial.mode === 'update' && !initial.record.type)} onChange={event => setType(event.target.value)}><option value="note">Заметка</option><option value="review">Замечание</option><option value="warning">Предупреждение</option><option value="change_required">Нужна правка</option><option value="edit" disabled={initial.mode === 'update' && !!initial.record.type}>Предложить замену</option></select></label>}
    <label className="field-label">{initial.mode === 'note' ? 'Заметка о главе' : type === 'edit' ? 'Предлагаемый текст' : 'Комментарий'}<textarea autoFocus className="note-editor" value={text} disabled={busy} onChange={event => setText(event.target.value)} /></label>
    {type === 'edit' && initial.mode !== 'note' && <p className="fine-print">Пустая замена предлагает удалить выделенный фрагмент. Текст изменится только после принятия предложения.</p>}
    <div className="dialog-actions"><button className="quiet-button" disabled={busy || (type === 'edit' && text === selected) || (initial.mode !== 'add' && text === initialText && type === initialType)}>Сохранить рецензию</button></div>
  </form>;
}
