import type { CSSProperties } from 'react';
import { GalleyEditor } from '@inkyquill/galley-editor';
import { themeToCssVariables } from '@inkyquill/galley-themes';
import { resolveTheme } from '../appearance/themes';
import { useAppearance } from '../appearance/Appearance';

export function MarkdownEditor({ value, onChange, busy }: { value: string; onChange: (value: string) => void; busy: boolean }) {
  const appearance = useAppearance();
  return <div className={`file-galley ${appearance.serif ? '' : 'sans-prose'}`} style={{ fontSize: appearance.fontSize }}>
    <GalleyEditor value={value} onChange={onChange} editable={!busy} ariaLabel="Текст файла" mode="live" theme={appearance.scheme} toolbar={false} footer={false} tabIndents={false} minRows={12} surface={{ contentPadding: '0', style: { ...themeToCssVariables(resolveTheme(appearance.family, appearance.scheme)), '--ge-content-padding': '0', '--ge-font-body': appearance.serif ? "Georgia, 'Noto Serif', serif" : "'Geist Variable', system-ui, sans-serif", '--ge-font-size': `${appearance.fontSize}px`, '--ge-line-height': '1.85', '--ge-radius-editor': '0px' } as CSSProperties }} placeholder="Начните с первой строки…" />
  </div>;
}
