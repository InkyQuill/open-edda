export type Entry = { id: string; project: string; name: string; parent: string | null; kind: 'folder' | 'text' | 'image' | 'pdf' | 'audio' | 'video' | 'binary'; body?: string; url?: string; size?: number; mime?: string; temporary?: boolean };
export type Project = { id: string; title: string; subtitle: string };
export const prose = `К вечеру море стало цвета олова. Мира остановилась у последнего дома на набережной и посмотрела на окна: во всех горел свет, кроме одного.\n\nНа втором этаже, за узким стеклом, темнела комната отца. Она помнила эту комнату другой — открытой настежь, наполненной запахом бумаги и ветра. На подоконнике всегда стояла чашка, которую никто не решался убрать.\n\n— Ты всё-таки приехала, — сказал смотритель.\n\nОн не удивился. Только отступил в сторону, пропуская её к двери, и провёл ладонью по мокрому рукаву.\n\nМира положила руку на холодную ручку. За дверью было тихо. Даже море здесь казалось далёким, словно дом стоял не у воды, а в глубине чужого воспоминания.\n\nВ кармане лежало письмо. За три дня дороги она так и не открыла его второй раз.`;
export const japanese = `夕方になると、海は錫の色に変わった。ミラは海沿いの通りの最後の家の前で立ち止まり、窓を見上げた。一つを除いて、どの窓にも明かりが灯っていた。\n\n二階の細い窓の向こうで、父の部屋だけが暗かった。彼女の記憶の中では、その部屋の窓はいつも大きく開いていて、紙と風の匂いに満ちていた。窓辺には、誰も片づけようとしないカップが置かれていた。\n\n「やっぱり来たんだね」と管理人が言った。\n\n彼は驚かなかった。ただ戸口を空けるように脇へ寄り、濡れた袖を手のひらで払った。\n\nミラは冷たい取っ手に手をかけた。扉の向こうは静かだった。\n\nポケットには手紙が入っていた。三日間の旅のあいだ、彼女は一度もそれを読み返さなかった。`;
const cover = `<svg xmlns="http://www.w3.org/2000/svg" width="560" height="760" viewBox="0 0 560 760"><rect width="560" height="760" fill="#d8e2db"/><path d="M0 480H560V760H0Z" fill="#45675f"/><path d="M0 510Q140 450 280 510T560 510" fill="none" stroke="#aabeb2" stroke-width="2"/><text x="52" y="114" fill="#263e38" font-family="Georgia,serif" font-size="48">Тихий берег</text><text x="54" y="155" fill="#45675f" font-family="sans-serif" font-size="15" letter-spacing="3">РАБОЧАЯ ОБЛОЖКА</text><text x="54" y="698" fill="#eef1e9" font-family="sans-serif" font-size="17">Edda · демонстрационный проект</text></svg>`;
export const initialProjects: Project[] = [{ id: 'shore', title: 'Тихий берег', subtitle: 'Роман · рукопись и материалы' }, { id: 'letters', title: 'Письма с севера', subtitle: 'Перевод · 日本語 → Русский' }];
export const initialEntries: Entry[] = [
 {id:'manuscript',project:'shore',name:'Рукопись',parent:null,kind:'folder'},
 {id:'chapter',project:'shore',name:'01 — Возвращение.md',parent:'manuscript',kind:'text',body:prose},
 {id:'chapter2',project:'shore',name:'02 — Письмо.md',parent:'manuscript',kind:'text',body:'Письмо оказалось коротким. Всего несколько строк, написанных знакомым почерком.\n\nМира села у окна и начала читать.'},
 {id:'materials',project:'shore',name:'Материалы',parent:null,kind:'folder'},
 {id:'notes',project:'shore',name:'Заметки о героях.md',parent:'materials',kind:'text',body:'Мира возвращается в приморский город после долгого отсутствия.\n\nСмотритель знает историю дома, но не торопится рассказывать её.'},
 {id:'cover',project:'shore',name:'Обложка.svg',parent:'materials',kind:'image',url:`data:image/svg+xml;charset=utf-8,${encodeURIComponent(cover)}`,mime:'image/svg+xml'},
 {id:'archive',project:'shore',name:'Материалы.zip',parent:'materials',kind:'binary',url:'data:application/zip;base64,UEsFBgAAAAAAAAAAAAAAAAAAAAAAAA==',size:22,mime:'application/zip'},
 {id:'translation',project:'letters',name:'01 — Возвращение.md',parent:null,kind:'text',body:prose},
 {id:'source',project:'letters',name:'原文.txt',parent:null,kind:'text',body:japanese},
];
export function moveIssue(entries: Entry[], id: string, parent: string | null): string | null {
 const item = entries.find(e => e.id === id);
 if (!item) return 'Файл больше не существует.';
 const destination = entries.find(e => e.id === parent);
 if (parent && (!destination || destination.kind !== 'folder' || destination.project !== item.project)) return 'Выберите папку этого проекта.';
 let cursor = parent;
 while(cursor) { if(cursor === id) return 'Нельзя переместить папку внутрь самой себя.'; cursor = entries.find(e => e.id === cursor)?.parent ?? null; }
 if(entries.some(e => e.id !== id && e.project === item.project && e.parent === parent && e.name === item.name)) return 'В этой папке уже есть файл с таким именем.';
 return null;
}
export function kindFor(file: File): Entry['kind'] {
 if(file.type.startsWith('image/')) return 'image';
 if(file.type === 'application/pdf') return 'pdf';
 if(file.type.startsWith('audio/')) return 'audio';
 if(file.type.startsWith('video/')) return 'video';
 if(file.type.startsWith('text/') || /\.(md|txt|json|csv|yaml|yml)$/i.test(file.name)) return 'text';
 return 'binary';
}
