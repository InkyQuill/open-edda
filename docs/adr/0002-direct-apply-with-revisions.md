# Direct Apply With Version Safety

Open Edda supports direct agent-applied changes instead of requiring every write to wait for pre-approval. Direct Apply is an explicit per-action mode whose preference can be remembered for an author and project, preserving a visible distinction between previewed changes and immediate application.

This favors the fast creative loop that motivated the product, while version checks, diffs, restore, activity records, and explicit checkpoints provide the safety net for authors who prefer to inspect or reverse changes after the fact. Existing per-item database revisions can remain as implementation detail during migration, but the product-facing long-term history model is project-wide checkpoints over saved files.
