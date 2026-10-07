---
target: Projects and workspace direction (source review; current runtime blocked)
total_score: 16
max_score: 40
na_heuristics:
p0_count: 0
p1_count: 4
target_identity: "file:/home/inky/Development/open-edda/frontend/src/features/projects/ProjectsPage.tsx"
target_fingerprint: "sha256:b5ca0637d589fa686b04b14f3a9c1165e16aabddf9d88d18cb58193bef50e63f"
target_path: /home/inky/Development/open-edda/frontend/src/features/projects/ProjectsPage.tsx
timestamp: 2026-10-06T08-59-02Z
slug: frontend-src-features-projects-projectspage-tsx
---
Method: dual-agent (A: `/root/ui_design` · B: `/root/ui_evidence`). Assessments were isolated; A completed before detector results entered synthesis. Mode: Operate. Target: `frontend/src/features/projects/ProjectsPage.tsx` with the workspace it opens. Scores are **provisional source judgments**, not measured usability or live visual acceptance.

Specificity: the present entry screen is a generic project CRUD launcher. The strongest product identity would be the author's recognizable project structure, dependable version history and clear local synchronization. Existing editor-centered routing is a useful base; AI configuration should not dominate initial project use.

| Heuristic | Score / 4 | Main source observation |
| --- | --- | --- |
| System status | 2 | Basic load/save states; no acknowledged project sync state |
| Real-world match | 1 | Language codes/content kinds/revision terminology |
| Control and freedom | 2 | Modes exist; no direct Projects return in workspace chrome |
| Consistency | 2 | Shared primitives; duplicate mode/context navigation |
| Error prevention | 1 | Draft replacement risk on navigation; runtime reproduction pending |
| Recognition | 2 | Labels exist; custom project tree absent |
| Efficiency | 1 | Large project cards, no project search; fixed panels |
| Minimalism | 2 | Quiet palette but nested permanent frames and duplicate controls |
| Error recovery | 2 | Error states/retry exist; limited task-specific recovery |
| Help | 1 | No clear existing-folder/sync model |
| **Total** | **16 / 40** | **Poor under the skill rubric; provisional, not a release score** |

Priority changes for the projects/sync delivery:

1. **P1 — Entry workflow:** `ProjectsPage.tsx:139–195` offers title/language and Elysium ZIP, not existing CWS project attachment. Create/open/connect should lead to a preserved file inventory and first-sync outcome. Suggested Impeccable workflow: `onboard`.
2. **P1 — Project navigation:** `WorkspacePage.tsx` and `ContextDrawer.tsx` impose four content kinds and additional World/Notes placeholders. Use the actual directory tree with optional semantic views. Suggested workflow: `clarify`.
3. **P1 — Draft safety:** `EditorFrame.tsx:128–144` hydrates selected content and `editorSlice.ts:65–83` resets dirty state. Preserve drafts across navigation/refresh and distinguish saved from synchronized. This is a source-backed risk, not reproduced data loss. Suggested workflow: `harden`.
4. **P1 — Prose width:** `workspaceSlice.ts:34–41` defaults to Assistant with 304/380px panels; `WorkspaceShell.tsx` shows them at `md`. Their nominal width plus rail leaves almost no center at 768px. Validate actual rendering after build repair; collapse tools before prose. Suggested workflow: `layout`.
5. **P2 — Reference design transfer:** shared semantic themes, independent UI/book typography and flat prose surfaces should replace nested permanent frames and competing typography rules. Use Galley Desk's actual transfer guide; avoid copying exact Electron dimensions. Suggested workflow: `distill`.

Strengths: short creation form with pending/error states; explicit loading/empty/retry handling; shared UI primitives and a distinct central editor. These support incremental correction, not a framework replacement.

Cognitive load: creation is moderate; workspace has competing mode controls, context tabs, filters and assistant setup. The concern is simultaneous decisions, not claiming each small list exceeds four choices. Emotional gap: creating a project should end with recognizable work safely available, rather than empty categories and assistant setup.

Persona checks: first-time authors lack an explanation of language/setup; experienced translators cannot recognize their existing folder structure; selected mode/filter state needs accessible semantics and keyboard verification. Contrast, focus restoration, mobile rendering and draft outcomes remain unverified.

The detector ran once over `frontend/src`: zero errors and one `overused-font` warning at `styles.css:10` (`Inter`). Later rules use Geist; the warning does not prove rendered typography or a usability defect. It does not detect the storage/onboarding problems. No live overlay is claimed.

Reference evidence: Galley Desk `apps/desktop/DESIGN.md`, `docs/design-system-transfer.md`; Timeline Helper `DESIGN.md`, `PRODUCT.md`. Transfer their quiet, content-first surfaces and progressive disclosure. Do not copy Timeline's product-specific exclusion of sync status: synchronization is central to Open Edda.

Questions skipped: the user already selected the product direction and requested documentation planning, not an additional design interview. The implementation brief should answer how the author sees that their folder is preserved, identifies originals versus translations, and knows that synchronization has been acknowledged.
