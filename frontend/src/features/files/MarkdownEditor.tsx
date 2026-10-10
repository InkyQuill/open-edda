import { flushSync } from 'react-dom';
import { editorText, sourceText } from './editorText';
import type { Extension } from '@codemirror/state';
import type { CSSProperties, Ref } from 'react';
import { GalleyEditor, type GalleyHandle } from '@inkyquill/galley-editor';
import { themeToCssVariables } from '@inkyquill/galley-themes';
import { resolveTheme } from '../appearance/themes';
import { useAppearance } from '../appearance/Appearance';

export function MarkdownEditor({ value, onChange, busy, readOnly = false, editorRef, extensions }: { value: string; onChange: (value: string) => void; busy: boolean; readOnly?: boolean; editorRef?: Ref<GalleyHandle>; extensions?: Extension[] }) {
  const appearance = useAppearance();
  function change(next: string) {
    const updated = sourceText(value, next);
    // Galley's controlled-value effect must see this keystroke before the next
    // native input event, otherwise rapid mobile typing can replay an older value.
    if (updated !== value) flushSync(() => onChange(updated));
  }
  return <div className={`file-galley ${appearance.serif ? '' : 'sans-prose'}`} style={{ fontSize: appearance.fontSize }}>
    <GalleyEditor ref={editorRef} extensions={extensions} value={editorText(value)} onChange={change} editable={!busy && !readOnly} ariaLabel="Текст файла" mode="live" theme={appearance.scheme} toolbar={false} footer={false} tabIndents={false} minRows={12} surface={{ contentPadding: '0', style: { ...themeToCssVariables(resolveTheme(appearance.family, appearance.scheme)), '--ge-content-padding': '0', '--ge-font-body': appearance.serif ? "Georgia, 'Noto Serif', serif" : "'Geist Variable', system-ui, sans-serif", '--ge-font-size': `${appearance.fontSize}px`, '--ge-line-height': '1.85', '--ge-radius-editor': '0px' } as CSSProperties }} placeholder="Начните с первой строки…" />
  </div>;
}
