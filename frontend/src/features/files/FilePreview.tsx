import { useState } from 'react';
import { ArrowDownToLine, File } from 'lucide-react';
import { downloadFile } from '../../shared/ui/edda';
import { basename, type FileView } from './fileModel';
export function FilePreview({ view, compact = false }: { view: FileView; compact?: boolean }) {
  const [failed, setFailed] = useState(false);
  const download = () => downloadFile(view.blob, basename(view.entry.path));
  return <div className={compact ? 'history-preview' : 'file-page'}>{!compact && <div className="file-page-title"><div><h1>{basename(view.entry.path)}</h1><p>{new Intl.NumberFormat('ru').format(view.entry.bytes)} байт</p></div></div>}
    {view.kind === 'pdf' && <p className="preview-note">Просмотр PDF зависит от браузера. Если документ не появился, <button className="inline-download" onClick={download}>скачайте файл</button>.</p>}
    {!failed && view.kind === 'image' ? <div className="image-preview"><img src={view.url} alt={basename(view.entry.path)} onError={() => setFailed(true)} /></div> : !failed && view.kind === 'audio' ? <div className="media-preview"><audio controls src={view.url} aria-label={basename(view.entry.path)} onError={() => setFailed(true)} /></div> : !failed && view.kind === 'video' ? <div className="media-preview"><video controls src={view.url} aria-label={basename(view.entry.path)} onError={() => setFailed(true)} /></div> : !failed && view.kind === 'pdf' && navigator.pdfViewerEnabled ? <object className="pdf-preview" data={view.url} type="application/pdf" aria-label={`Просмотр ${basename(view.entry.path)}`}><button className="quiet-button" onClick={download}>Скачать PDF</button></object> : view.kind === 'text' ? <pre className="history-preview">{view.body}</pre> : <div className="binary-preview"><File size={54} strokeWidth={1} /><h2>Файл готов к скачиванию</h2><p>{failed ? 'Браузер не смог показать этот файл.' : 'У этого файла нет доступного встроенного просмотра.'}<br />Откройте его в подходящем приложении.</p><button className="quiet-button" onClick={download}><ArrowDownToLine size={17} />Скачать файл</button></div>}
  </div>;
}
