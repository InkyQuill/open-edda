---
name: Edda — quiet author workshop
description: Approved calm writing system integrated into the production frontend.
colors:
  light-shell: "#f1f2ee"
  light-paper: "#fcfcf8"
  light-raised: "#fffefa"
  light-ink: "#293831"
  light-muted: "#667169"
  light-line: "#dce1d9"
  light-hover: "#e8ede5"
  light-selected: "#dce9e0"
  light-accent: "#376452"
  light-accent-ink: "#234d3c"
  light-danger: "#a33332"
  light-overlay: "#1b282955"
  light-selection: "#cfdfd2"
  dark-shell: "#202623"
  dark-paper: "#262d29"
  dark-raised: "#303833"
  dark-ink: "#e0e6dd"
  dark-muted: "#a8b5aa"
  dark-line: "#404b43"
  dark-hover: "#333e36"
  dark-selected: "#364d3f"
  dark-accent: "#aacdb7"
  dark-accent-ink: "#c3e2cd"
  dark-danger: "#ffb2aa"
  dark-overlay: "#0b110dbb"
  dark-selection: "#42604b"
typography:
  interface:
    fontFamily: "Geist Variable, system-ui, sans-serif"
    fontSize: "14px"
  manuscript:
    fontFamily: "Georgia, Noto Serif, serif"
    fontSize: "19px"
    lineHeight: 1.85
  japanese:
    fontFamily: "Georgia, Noto Serif CJK JP, serif"
    fontSize: "19px"
    lineHeight: 2
rounded:
  field: "6px"
  control: "8px"
  dialog: "12px"
spacing:
  compact: "8px"
  control: "12px"
  inset: "16px"
  panel: "20px"
  section: "24px"
  dialog: "32px"
components:
  icon-button:
    textColor: "{colors.light-muted}"
    rounded: "{rounded.control}"
    width: "36px"
    height: "36px"
  quiet-button:
    backgroundColor: "{colors.light-hover}"
    textColor: "{colors.light-accent-ink}"
    rounded: "{rounded.control}"
    padding: "10px 14px"
  quiet-button-dark:
    backgroundColor: "{colors.dark-hover}"
    textColor: "{colors.dark-accent-ink}"
    rounded: "{rounded.control}"
    padding: "10px 14px"
---

# Design System: Edda

## Overview

**Creative North Star: "Тихая авторская мастерская"**

Edda is a calm workspace for sustained writing and literary translation. Warm stone surrounds a pale reading plane; dark mode uses charcoal surfaces and muted ivory text. Restrained green-teal accents mark selection and interaction. Text occupies the main space; files and optional context support it.

Scope: integrated into the production project and file workflows in `frontend/src/features/{projects,files,appearance}` and shared primitives in `frontend/src/shared/ui/{edda.css,edda.tsx}`. Authentication and settings use the same theme bridge; the existing structured workspace retains its capabilities. Backend and synchronization contracts remain in `CONTEXT.md`. The independent `frontend/design.html` prototype remains a design reference. Product constraints live in `PRODUCT.md`; prototype and integration evidence in `docs/design/prototype.md`. Published to `https://edda.inky.su` on 2026-10-07; see `docs/design/production-acceptance.md`.

**Key Characteristics:**
- Text-first composition with optional supporting panes.
- Equal light, dark and system theme support.
- Quiet, named icon actions and explicit file affordances.

References and adaptations:
- [Linear’s UI redesign](https://linear.app/now/how-we-redesigned-the-linear-ui): hierarchy of navigation, header and work area, adapted to lower density and long prose.
- [Radix color scale](https://www.radix-ui.com/colors/docs/palette-composition/understanding-the-scale): semantic roles for surfaces and interaction states; Edda uses its own palette.
- [Radix Dialog](https://www.radix-ui.com/primitives/docs/components/dialog): focus and keyboard behavior through the existing Radix dependency.

## Colors

The frontmatter preserves the approved palette, now implemented in `frontend/src/shared/ui/edda.css`. The `light-` and `dark-` prefixes map to the same CSS custom properties under the corresponding theme. Production uses `--edda-muted` and `--edda-accent` for secondary text and accent color to avoid collisions with existing semantic surface tokens; the bridge maps this palette to existing app and shadcn roles.

- `shell`, `paper`, `raised`: surrounding chrome, reading plane, dialogs.
- `ink`, `muted`: primary content and secondary labels.
- `line`, `hover`, `selected`: boundaries, hover, selected navigation.
- `accent`, `accent-ink`: focus, icons and interactive text.
- `danger`: actionable errors; `overlay`: modal backdrop; `selection`: text selection.

Choose light, dark or system in view settings; the production appearance provider persists the choice and observes system changes. Use semantic roles consistently in both themes. Selection also uses shape or state semantics rather than color alone.

## Typography

Interface: bundled Geist Variable with system sans-serif fallback, 14px base. Manuscript: 19px initially, adjustable from 16–25px, line height 1.85; a sans-serif reading option is available. Japanese text uses a CJK serif fallback and line height 2. Font coverage must be checked in the actual browser, not inferred from the stack.

The implementation currently uses Georgia for document and dashboard headings (38px and 40px; 33px on mobile), with 29px headings in translation. This system-display fallback is an implementation limitation, not a prescribed display-face identity. Production increases tree labels to 13px and several peripheral labels to 11–12px; remaining compact labels are observed details, not a recommended type scale for new surfaces.

## Layout

Desktop: 248px file sidebar, flexible main area, optional 300px context pane. The actual workspace header is at least 66px high. A manuscript column is capped at 730px; translation expands to 1200px with independently scrolling columns and an adjustable 30–70% split. Focus mode hides surrounding navigation.

At 1150px and below, the sidebar is 224px and context overlays the work area. At 800px and below, files move into a 280px drawer, translation becomes source/translation tabs, and writing receives 26px side padding. At 1600px and above, manuscript spacing grows. Dashboard content is capped at 1056px including its horizontal padding.

Use the recorded spacing values for controls and groups; the reading layout intentionally has broader, content-led margins rather than forcing every dimension onto one grid.

## Elevation & Depth

Tonal surfaces and fine borders carry the resting hierarchy. Diffuse shadows belong to dialogs, transient confirmations and overlay panes. The shared shadow is `0 16px 56px #17221d26` in light mode and `0 16px 56px #0006` in dark mode. No decorative lifting animation is required.

Buttons transition color and background over 130ms ease-out. Toast entry lasts 160ms ease-out. Reduced-motion preference disables both animation and transitions.

## Shapes

Controls use 8px corners, fields and tree rows 6px, dialogs and the continue-project surface 12px. Use these modest radii without rounding every container into a card. SVG line icons use consistent proportions; ordinary icon actions render at 18px with a 1.6 stroke.

## Components

- **Actions:** transparent icon buttons use muted ink, then accent ink and hover surface. Accessible name and native title describe the action; toggle actions expose state. Quiet text buttons use hover surface, accent ink, 10px × 14px padding and at least 40px height. Visible text remains appropriate for explicit choices such as downloading an unsupported file.
- **Focus and dialogs:** controls use a 2px accent outline with a 3px offset. Radix handles dialog focus, Escape and restoration. The editor currently substitutes a subtle gutter indicator; this is not a precedent for suppressing clear keyboard focus elsewhere.
- **Fields:** paper surface, 1px line border, 6px corners, 11px × 12px padding, at least 42px height. Put errors next to the relevant form or operation.
- **Files:** one tree includes text, folders and binaries. Selected rows use selected surface and accent ink. Drag targets have an accent outline; a named move action offers the same destination choice without dragging. Reject cycles, cross-project destinations and duplicate names. Mobile tree rows are at least 44px high and move actions remain visible.
- **Preview:** images, PDF, audio and video occupy the main file page using browser-supported viewers. Unsupported or failed previews retain file identity and a download action. Do not create a separate binary-download section in the sidebar.
- **Context:** project history opens one optional pane with an explicit close action. Preview real version files before explicitly restoring the whole project. Conflicts offer mine, theirs and both; source comparison selects a real project text and scrolls independently. Prototype-only notes and assistant demonstrations are not production capabilities.
- **Feedback:** use transient status announcements for completed actions; retain errors and saving information by the affected workspace. Per-file browser drafts use sessionStorage; only a confirmed explicit server save is reported as saved. Lost-response retries retain the operation identity. Media previews use real blobs; PDF rendering depends on browser support and always retains a download fallback.

The sidecar contains representative renderable primitives. Its synthesized tonal ramps are panel visualization metadata, not additional approved UI colors.

## Do's and Don'ts

- Do keep every file in the project tree and open its content or download page in the main workspace.
- Do provide keyboard and touch alternatives to drag and drop.
- Do distinguish editing, browser drafts, confirmed server saves and failures truthfully.
- Do preserve visible focus, Cyrillic and Japanese readability, and reduced-motion preferences.
- Don’t use saturated call-to-action blocks for routine file operations.
- Don’t treat a binary file or an unavailable browser preview as a project error.
- Don’t hide essential actions behind hover alone on touch screens.
- Don’t adopt prototype-only persistence or simulated conflicts as production semantics.
