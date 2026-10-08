import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import { applyTheme, themeFamilies } from './themes';
import { Monitor, Moon, Palette, Sun } from 'lucide-react';
import { IconButton, Modal } from '../../shared/ui/edda';

type Appearance = { family: string; theme: 'light' | 'dark' | 'system'; fontSize: number; serif: boolean };
const defaults: Appearance = { family: 'Edda', theme: 'system', fontSize: 19, serif: true };
function read(): Appearance {
  try { const value = JSON.parse(localStorage.getItem('edda.appearance') ?? 'null') as Partial<Appearance> | null;
    return { family: typeof value?.family === 'string' && themeFamilies.includes(value.family) ? value.family : 'Edda', theme: value?.theme === 'light' || value?.theme === 'dark' ? value.theme : 'system', fontSize: typeof value?.fontSize === 'number' ? Math.min(25, Math.max(16, value.fontSize)) : 19, serif: value?.serif !== false };
  } catch { return defaults; }
}
const Context = createContext({ ...defaults, scheme: 'light' as 'light' | 'dark', update: (_change: Partial<Appearance>) => {} });
export function AppearanceProvider({ children }: { children: ReactNode }) {
  const [settings, setSettings] = useState(read);
  const [scheme, setScheme] = useState<'light' | 'dark'>('light');
  useEffect(() => {
    const media = matchMedia('(prefers-color-scheme: dark)');
    const apply = () => { const dark = settings.theme === 'dark' || (settings.theme === 'system' && media.matches); setScheme(dark ? 'dark' : 'light'); applyTheme(settings.family, dark ? 'dark' : 'light', document.documentElement); };
    apply(); media.addEventListener('change', apply);
    try { localStorage.setItem('edda.appearance', JSON.stringify(settings)); } catch { /* Preferences still work for this session. */ }
    return () => media.removeEventListener('change', apply);
  }, [settings]);
  return <Context.Provider value={{ ...settings, scheme, update: change => setSettings(previous => ({ ...previous, ...change })) }}>{children}</Context.Provider>;
}
export const useAppearance = () => useContext(Context);
export function AppearanceButton() {
  const [open, setOpen] = useState(false);
  const appearance = useAppearance();
  return <><IconButton icon={Palette} label="Настройки вида" onClick={() => setOpen(true)} /><Modal open={open} onClose={() => setOpen(false)} title="Удобно именно вам" description="Настройте пространство для спокойной работы."><label className="field-label">Тема<select value={appearance.family} onChange={event => appearance.update({ family: event.target.value })}>{themeFamilies.map(family => <option key={family} value={family}>{family}</option>)}</select></label><fieldset><legend>Оформление</legend><div className="theme-options">{([{ value: 'light', label: 'Светлое', icon: Sun }, { value: 'dark', label: 'Тёмное', icon: Moon }, { value: 'system', label: 'Системное', icon: Monitor }] as const).map(option => <button key={option.value} className={appearance.theme === option.value ? 'chosen' : ''} aria-pressed={appearance.theme === option.value} onClick={() => appearance.update({ theme: option.value })}><option.icon size={20} />{option.label}</button>)}</div></fieldset><label className="range-label">Размер текста <span>{appearance.fontSize} px</span><input type="range" min={16} max={25} value={appearance.fontSize} onChange={event => appearance.update({ fontSize: Number(event.target.value) })} /></label><label className="check-label"><input type="checkbox" checked={appearance.serif} onChange={event => appearance.update({ serif: event.target.checked })} />Шрифт с засечками для рукописи</label></Modal></>;
}
