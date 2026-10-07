# Product

<!-- impeccable:product-schema 1 -->

## Platform
web

## Users
Authors and literary translators working for long periods with prose, sources and reference material. Desktop is the primary writing surface; phones support reading and small edits.

## Product Purpose
Open Edda is a self-hosted workspace for portable, versioned writing projects. Text and translation are primary; files, history and optional AI support that work.

## Capabilities and Constraints
Preserve arbitrary project folders and immutable server versions. React, Tailwind, Radix and Galley Editor are existing foundations. Backend authority and synchronization semantics remain as described in CONTEXT.md and ADR 0014/0015.

The approved calm design is integrated into the production frontend: projects, file writing, source/translation comparison, light/dark/system themes and responsive layouts. Existing structured workspace and settings capabilities remain available; authentication and settings share the palette. The independent development prototype remains at `frontend/design.html`. Files and folders require drag and drop with an accessible alternative. Preview browser-supported media in the main workspace; other binaries have a normal file page with download. No separate sidebar download picker.

## Brand Commitments
In-app name: Edda. Modern, soft, calm and comfortable. Compact icon actions with accessible names and tooltips. The user authorized rethinking the entire layout on 2026-10-07.

## Product Principles
- Preserve the author's work and show saving states honestly.
- Let the text occupy the primary space; supporting panes are optional.
- Preserve folder organization and explicit version/conflict choices.
- Both themes receive equal attention.

## Evidence on Hand
Current source and CONTEXT.md. Per-file drafts persist in sessionStorage; explicit server saves create immutable project versions. Conflict resolution offers mine, theirs and both; history restoration explicitly restores the whole project. Source comparison reads a real project file. API-backed desktop/mobile verification passed against the deployed server image; this frontend integration was deployed to `https://edda.inky.su` on 2026-10-07. Prototype content remains fictional demonstration material.

## Accessibility & Inclusion
Keyboard equivalents for drag and drop, visible focus, named icon buttons, readable contrast, reduced motion, Cyrillic and Japanese content.
