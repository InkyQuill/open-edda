import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { Dialog } from 'radix-ui';
import { X, type LucideIcon } from 'lucide-react';

export function IconButton({ icon: Icon, label, active, className = '', ...props }: { icon: LucideIcon; label: string; active?: boolean } & ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button type="button" title={label} aria-label={label} {...props} className={`icon-button ${active ? 'active' : ''} ${className}`}><Icon size={18} strokeWidth={1.6} aria-hidden="true" /></button>;
}
export function Modal({ open, onClose, title, description, children, wide = false, busy = false }: { open: boolean; onClose: () => void; title: string; description: string; children: ReactNode; wide?: boolean; busy?: boolean }) {
  return <Dialog.Root open={open} onOpenChange={next => { if (!next && !busy) onClose(); }}><Dialog.Portal><Dialog.Overlay className="dialog-overlay" /><Dialog.Content className={`dialog edda-modal ${wide ? 'conflict-dialog' : ''}`} aria-busy={busy} onEscapeKeyDown={event => { if (busy) event.preventDefault(); }} onPointerDownOutside={event => { if (busy) event.preventDefault(); }}><Dialog.Title>{title}</Dialog.Title><Dialog.Description>{description}</Dialog.Description><Dialog.Close disabled={busy} className="dialog-close icon-button" aria-label="Закрыть"><X size={18} /></Dialog.Close>{children}</Dialog.Content></Dialog.Portal></Dialog.Root>;
}
export function downloadFile(blob: Blob, name: string) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a'); anchor.href = url; anchor.download = name; anchor.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 30_000);
}
