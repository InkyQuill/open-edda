# Edda Design System

Status: approved and implemented, 2026-09-30.

## Product and intent

Edda is a private writing workspace for long-form fiction: manuscript, story bible, planning, notes, assistant sessions and review. The manuscript is the primary surface. The interface should support long writing sessions without competing for attention.

The author explicitly approved replacing the previous visual design. Existing colors, card layouts and framing are not constraints. Preserve product capabilities and authored data while improving presentation.

## Authoritative palette

Use `@inkyquill/galley-themes` as the source of color values for both the application and editor. Do not duplicate or invent a separate Edda palette.

- Initial default: `thoth-light`.
- Default dark palette: `thoth-dark`.
- All built-in catalog palettes may be selected in Settings → Appearance.
- Persist the selected catalog ID in the authenticated author's server-side profile preferences. Browser storage alone is insufficient.
- Validate IDs against the catalog. Missing or unsupported preferences fall back to `thoth-light`.
- The chosen palette applies to every route, overlay and editor surface.

Thoth Light uses a warm off-white background, dark brown text and a copper accent. Thoth Dark uses the package's matching dark palette. The earlier Sand + Sage proposal was superseded by the author's Galley/Thoth requirement; do not introduce a green accent independently.

Use `themeToCssVariables()` to apply `--app-*` and `--ge-*` roles. Map application component tokens to palette roles:

| UI role | Galley role |
| --- | --- |
| Background / text | `app.bg` / `app.text` |
| Cards and popovers | `app.panel` |
| Quiet panels | `app.panelMuted` |
| Hover and selected backgrounds | `app.hover` |
| Primary actions and focus | `app.focus` |
| Secondary text | `app.textMuted` |
| Borders | `app.border` |
| Errors | `app.errorBg`, `app.errorBorder`, `app.errorText` |
| Editor text, selection, caret, syntax | Package `editor`, `markdown` and `syntax` roles |

Color indicates interaction and state. Avoid decorative colored panels. Flat does not mean faint: keep text readable and keyboard focus visible. Verify contrast in both Thoth schemes and any additional selected palette.

## Flat surfaces and components

- Opaque, flat surfaces; no gradients, glass, paper texture or decorative shadows.
- Separate major regions with spacing, subtle background differences or a single thin divider.
- Use compact navigation rows instead of framing each item as a card.
- Small controls use 4–6 px corner radii. Avoid pill shapes for ordinary controls.
- On coarse-pointer devices, shared buttons, tabs and inputs have at least 44 px height; editor toolbar buttons have at least 40 px dimensions on narrow screens.
- Narrow settings navigation stacks vertically; tab-list height must grow with its content.
- Default controls are quiet; use solid accent fills for the principal action only.
- Popovers and dialogs may use a restrained shadow to communicate layering.
- Preserve visible focus outlines, disabled states, accessible labels and error messages.
- Use existing Radix-based primitives and Lucide icons. Do not add a second component library for decoration.

## Typography, spacing and layout

- Interface: readable sans serif with Cyrillic support; the current implementation uses Geist Variable with system fallbacks. The Impeccable detector flags Geist as an overused font; retaining the approved interface font is an intentional exception for refinement work.
- Typical interface text: 13–14 px; secondary labels may use 12 px.
- Manuscript: 16 px on desktop and 14 px on phones, approximately 1.55–1.6 line height, target line length 65–75 characters. Prefer comfortable reading over maximizing visible text.
- Author-selectable manuscript fonts may be added as a separate feature; do not confuse font choice with palette selection.
- Spacing uses a 4 px base, commonly 8, 12, 16, 24 and 32 px.
- Desktop: project navigation on the left, manuscript in the center, contextual tools on the right when requested.
- Start fresh workspaces in writing mode with the assistant hidden. Preserve saved workspace preferences.
- Provide a way to hide navigation for focused writing and reopen it.
- Avoid duplicating the same navigation actions in multiple prominent toolbars.
- On phones, the manuscript dominates. Use a single compact workspace header with a collapsed menu; collapse chapter metadata and formatting by default. Keep Save visible. Hide bottom navigation while the editor has focus, and size the workspace to the visual viewport so the software keyboard does not obscure its scroll area. Secondary panels remain sheets. Settings and navigation must remain reachable through the menu.

## Galley Editor

Use published `@inkyquill/galley-editor`; the author identified `~/dev/galley-editor` as its source repository for API inspection. Do not modify that separate repository as part of Edda integration unless requested.

- Use `theme="inherit"` so the editor consumes the same Galley Theme variables as Edda.
- Remove the outer manuscript card frame and editor shell borders, radii and shadows.
- Make toolbar and footer backgrounds blend into the writing surface; retain necessary internal structure in tables, code blocks and menus.
- Supply Lucide icons through the public `toolbar.icons` API. Galley retains ownership of command behavior, accessible button names and keyboard operation.
- Keep Markdown controlled state, UTF-8 selection adapters, undo, save and review behavior intact.
- Theme switching must not remount the editor or discard unsaved text, cursor or selection.
- Keep word count quiet and close to the manuscript. Do not add decorative branding inside the editor.

## Assistant and review

Assistant and review tools are available on demand. They must not visually dominate the manuscript. Distinguish generated suggestions, applied changes and review reports with functional states and clear actions. Keep technical provider/runtime administration in Settings.

## Motion

Use brief transitions only to explain a state change or panel opening. No decorative motion. Respect `prefers-reduced-motion`; keyboard use must never depend on animation.

## Reference provenance

- [Radix Themes: color](https://www.radix-ui.com/themes/docs/theme/color) and [buttons](https://www.radix-ui.com/themes/docs/components/button): semantic roles, consistent interaction states and quiet controls. Adapted to Galley palette tokens and flat styling; Radix Themes itself is not a required dependency.
- [Ulysses](https://ulysses.app/): library beside a focused writing surface. Adapted to Edda's manuscript, story bible, notes and assistant; not a pixel clone.
- [iA Writer](https://ia.net/writer): text priority and focused writing. Adapted to a project workspace with optional tools; no copying of proprietary fonts or assets.
- [Galley Themes](https://www.npmjs.com/package/@inkyquill/galley-themes): authoritative catalog and token mapping, including Thoth Light/Dark. Package is MIT licensed.
- [Galley Editor](https://www.npmjs.com/package/@inkyquill/galley-editor): editor integration and public customization APIs. Package is MIT licensed.

## Verification for interface changes

Read this file before designing or editing UI. Check the running interface, not only source or compilation:

- Desktop, laptop and mobile widths; long names and empty/error/loading states.
- Thoth Light and Thoth Dark; application/editor/overlay consistency.
- Theme selection, server save, reload and re-login; another author's preference isolation.
- Save failure: clear error and restoration of the previously stored palette.
- Editor text, selection, formatting, undo, save and review after theme changes.
- Keyboard navigation, visible focus and reduced motion.

Run relevant frontend tests/build and backend tests for profile persistence. Report any unavailable visual verification rather than declaring it complete.
