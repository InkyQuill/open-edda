# chapter-drafter

- Source path: `docs/skills/suggested/fiction/orchestrators/chapter-drafter/`
- Final disposition: `archived`
- Why it is not included as a 3.5 built-in skill: the skill assumes autonomous multi-pass orchestration across several other skills, persistent progress tracking, and scoring loops that go beyond the Milestone 3.5 built-in skill model. It also carries script-assisted evaluation behavior instead of a straightforward Writer-native read/check or rewrite flow.
- What would need to change for reconsideration: Writer would need stronger skill orchestration primitives, persistent draft-pass state, and a clear policy for automated quality gates before this can ship as a built-in workflow.
- Scripts/assets involved: yes. The source folder includes `scripts/score-scene.ts` plus templates under `templates/`.
- Future roadmap mapping: partial. Revisit only if Writer grows a future orchestration surface for autonomous multi-pass drafting; audited script runtime would be one prerequisite, not the roadmap reason by itself.
