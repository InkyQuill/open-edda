import { useRef, useState } from 'react';
import { Modal } from '../../shared/ui/edda';
import { ApiError } from '../../api';
import { frontmatter, metadataReplacement } from './chapterMetadata';
import { changeReview, type ReviewAction } from './pocketApi';
import type { ProjectVersion } from './fileApi';
export function ChapterMetadata({ source, version, entryId, busy, dirty, error, run, onSaved }: {
 source: string; version: ProjectVersion; entryId: string; busy: boolean; dirty: boolean; error: string;
 run: (work: () => Promise<void>) => Promise<void>; onSaved: () => Promise<void>;
}) {
 const [draft,setDraft]=useState<{yaml:string;source:string;version:string;entryId:string}|null>(null);
 const pending=useRef<{signature:string;action:ReviewAction}|null>(null);
 const fm=frontmatter(source);
 async function save(){
  if(!draft || dirty || frontmatter(draft.source)?.error)return;
  metadataReplacement(draft.source,draft.yaml);
  const signature=JSON.stringify(draft);
  if(pending.current?.signature!==signature)pending.current={signature,action:{action:'metadata',metadata:draft.yaml,entryId:draft.entryId,expectedVersion:draft.version,operationId:crypto.randomUUID()}};
  try { await changeReview(version.projectId,pending.current.action); }
  catch(cause){if(cause instanceof ApiError && cause.status===409)pending.current=null;throw cause;}
  await onSaved();pending.current=null;setDraft(null);
 }
 return <><button className="subtle-action" disabled={busy || dirty} onClick={()=>{pending.current=null;setDraft({yaml:fm?.yaml??'',source,version:version.id,entryId});}}>Метаданные главы</button>
 <Modal open={!!draft} busy={busy} onClose={()=>setDraft(null)} title="Метаданные главы" description="Произвольные свойства текста в YAML. Все поля необязательны; статус и готовность определяете вы." wide>
 {error && <p role="alert" className="form-error">{error}</p>}
 {draft && <form onSubmit={e=>{e.preventDefault();void run(save);}}>
 {frontmatter(draft.source)?.error && <p className="form-error">{frontmatter(draft.source)?.error} Исправьте исходный Markdown перед изменением метаданных.</p>}
 <label className="field-label">Свойства YAML<textarea aria-label="Свойства YAML" rows={10} disabled={busy} value={draft.yaml} onChange={e=>setDraft({...draft,yaml:e.target.value})} placeholder={'title: Название главы\nstatus: В работе'} /></label>
 <div className="dialog-actions"><button className="quiet-button" disabled={busy || dirty || !!frontmatter(draft.source)?.error || draft.yaml===(frontmatter(draft.source)?.yaml??'')}>Сохранить метаданные</button></div>
 </form>}
 </Modal></>;
}
