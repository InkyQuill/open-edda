# Production UI — 2026-10-07

The approved quiet workshop design is live at https://edda.inky.su. Projects and file writing use the new layout; authentication and existing settings share the theme. The structured workspace retains its existing controls. Tokens and UX rules are in [DESIGN.md](../../DESIGN.md).

## Behavior

- Light, dark and system appearance, reading size and serif preference; real source-file comparison with independently scrolling panes.
- One file tree for text and media. Drag/drop moves files and folders; named management dialogs provide keyboard/touch alternatives. Native directory drops and folder selection import a planned tree before a single publication.
- Image/audio/video and browser-supported PDF previews; unsupported binaries open a normal file page with download. PDF rendering remains browser-dependent; Codex's embedded browser did not render the native viewer, so an explicit download fallback stays visible.
- Per-file dirty drafts survive switching files and reloading the same tab using sessionStorage. A closed tab is not a durable backup. Confirmed saves create server project versions. Retrying a lost response retains the operation ID; conflicts expose mine/theirs/both. History previews a saved tree before an explicit whole-project restore.

## Verification

- 144 unit tests and TypeScript check passed; production build passed. Existing bundle-size warning remains (~1.09 MB uncompressed JS, including the retained structured editor).
- Four API-backed Playwright cases passed on desktop Chrome and Pixel 5 against the final deployment image, using an isolated temporary data filesystem. They cover project/folder/file creation, multiple drafts and reload, concurrent content conflict, lost-response save retry, exact historical text, image/binary import and download, identity-preserving rename/DnD, deletion and whole-tree restore, and dark appearance. No horizontal overflow in the checked mobile/desktop flows.
- CUA browser inspected real manuscript and Japanese source comparison in light/dark at 1280×720; mobile screenshots include writing, conflict and file drawer. Evidence is `.impeccable/review/production-*.png`. Finish reviewer found no material visual/interaction regression in this scope; its documentation-scope finding was corrected and scored resolved. Dashboard/auth were source-reviewed, with published dashboard additionally inspected in the browser.
- Public HTTPS health, login and exact JS/CSS bytes match the final build. All three existing project head IDs are unchanged by deployment. Two historical files from the existing PVC acceptance project were downloaded and matched their recorded sizes and SHA-256 hashes. Browser opened existing text, binary page and history. Live verification did not create projects or modify author content.

## Deployment and recovery

Image: `docker.io/library/open-edda:20261007-quiet-ui-v1`; index `sha256:bcd34587a98fafeacb5beb000524eb58f1b0f9859c3c1c6509dfa9de73f2bf52`.

Built with `deploy/frontend-only.Dockerfile`, `frontend/dist` as context and `open-edda:20261006-files-sync` as base. This preserves the running server and CLI; their SHA-256 hashes match the pre-update deployment. No schema or storage changes were shipped. The local working tree also contains separate backend/CLI work, which is not part of this image.

Before rollout, `edda backup` and `edda verify-backup` succeeded at `/data/backups/pre-quiet-ui-20261007-1342`. A private copy was exported to `~/.local/share/open-edda/backups/pre-quiet-ui-20261007-1342.tar.gz` (884118 bytes, SHA-256 `fc2f9d639f821b952325343ae03e9d274378e7e9db3128e4602ed7e056c97960`). Secrets remain separate. Kubernetes rollout completed and the public route was verified afterward.

The previous imported image remains available. If needed, set deployment `open-edda` container `open-edda` back to `docker.io/library/open-edda:20261006-files-sync`, then wait for rollout and public health. Image rollback must not delete or replace the PVC.
