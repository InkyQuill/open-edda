# Edda prototype

Approved 2026-10-07: quiet writing and translation workspace, full layout redesign; native previews and drag/drop, quiet icon actions. Isolated /design.html, demo data only. Operate/Read. No production API calls.

## Direction contract
THESIS: The author returns directly to a text. File operations happen around that text without displacing it.
OWN-WORLD: Warm stone shell, ivory reading plane, graphite typography, restrained teal selection; charcoal and muted ivory in dark mode. Small consistent Lucide icons, soft 8px control corners, restrained borders.
STORY: Choose a project, open or continue a document, write or compare source and translation, consult one optional contextual pane. Drop files onto folders; previews belong in the main workspace.
FIRST VIEWPORT: 248px project tree beside a broad quiet reading plane. A 60px header holds breadcrumbs and icon actions. The manuscript begins with a document title and readable prose. The optional 300px context pane has an explicit close action. On mobile the tree is a drawer, translation uses source/translation tabs.
FORM: Code-led interactive prototype; user-approved quiet author workshop overrides the random assignment (seed fc75ea99). No additional concept selection is needed.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance

## Coverage
Projects; editor; translation; light/dark/system; local draft feedback; simulated conflict; local history preview/restore; keyboard move alternative; native image/PDF/audio/video preview and ordinary download fallback. Uploads are temporary session data, explicitly labeled. Drag/drop rearranges the prototype tree; it does not synchronize with Edda.

## References
- https://linear.app/now/how-we-redesigned-the-linear-ui — navigation hierarchy adapted to lower density and long-form reading.
- https://www.radix-ui.com/colors/docs/palette-composition/understanding-the-scale — semantic state roles, custom palette.
- https://www.radix-ui.com/primitives/docs/components/dialog — accessible focus and keyboard behavior using existing Radix dependency.

## Run and build
From `frontend/`, `bun run dev --host 127.0.0.1 --port 4173`; open `/design.html`. Standalone build: `bun node_modules/vite/bin/vite.js build --config vite.design.config.ts` → `frontend/dist/design-prototype/`. Production entry/routes are independent.

## Verification — 2026-10-07
- Typecheck, 136 tests (including 7 tree/file-kind cases), production build and standalone prototype build passed.
- Browser inspected at 1440×960, 768×1024 and 390×844. Light/dark desktop and phone captures retained in `.impeccable/review/`. No horizontal document overflow at checked sizes.
- Real drag: chapter moved from Manuscript to Materials; DOM nesting confirmed. Model tests reject self/descendant moves, cross-project moves and collisions. Folder/file movement has keyboard/touch dialog alternative. OS folder import is not implemented; external drop accepts files.
- Adjusted translation split with keyboard; right pane scrolled 158px while original remained at 0. Mobile tree focus entry/wrap/Escape/return and background inertness verified.
- Continue action retained last translation project and file after reload. Edited text, previewed saved version and restored it; displaced text retained as a new local session version.
- SVG rendered. Uploaded WAV and WebM reached readyState 4, correct duration and video dimensions. Native PDF object remains blank in Codex's embedded browser despite its PDF capability flag; PDF rendering is therefore unverified. Persistent download affordance remains available, including an explanatory PDF fallback link.
- Calculated muted text contrast: 4.52:1 on light shell, 4.94:1 on light paper, 6.62:1 on dark paper. Primary prose exceeds 11:1.
- Finish reviewer identified continuation, translation scrolling/split, mobile drawer focus issues. All three corrected and scored resolved; verdict ship covers these fixes.

## Prototype boundaries
No server integration or AI calls. Text/tree/theme/resume persist in browser storage; revisions, notes, reading settings and uploaded media are session-only. Conflict view explicitly demonstrates choices without pretending to modify a server version. The prototype remains independently runnable for design review. Its simulated storage and conflicts do not describe production behavior.

## Production integration — 2026-10-07
The approved design now serves production project/file workflows, with shared appearance and UI primitives; authentication and existing settings receive the theme bridge. The structured workspace retains its capabilities. See `DESIGN.md` for current scope and tokens. Published to `https://edda.inky.su` on 2026-10-07; [deployment and acceptance evidence](production-acceptance.md).

Real project files, per-file sessionStorage drafts, explicit immutable server saves, stable operation IDs on retry, mine/theirs/both conflicts, and explicit whole-project restoration replace the demonstrations. Translation selects a real source file with independent scrolling. Media previews read real blobs; PDF remains browser-dependent with download fallback. No simulated assistant actions are exposed.

Integration verification: typecheck, 144 tests and four API-backed desktop/mobile end-to-end cases passed against the deployed server image. Five production captures are retained as `.impeccable/review/production-*.png`. Finish review found no material visual or interaction issues; its stale scope-documentation finding is corrected here and in `PRODUCT.md`, `DESIGN.md` and the sidecar.
