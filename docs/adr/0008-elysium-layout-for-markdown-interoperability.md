# Edda Project Layout for Markdown Interoperability

Open Edda defines one comfortable Markdown project layout instead of trying to support arbitrary folder shapes. The target layout is based on the stronger `alchemist` project structure:

```text
AGENTS.md
BOOTSTRAP.md
.agents/skills/
story/
story/_index.md
story/chapter-01.md
storyline/
storyline/_index.md
storyline/chapter-01-plan.md
storyline/vol-01-plan.md
characters/
characters/_index.md
characters/Protagonist.md
worldbuilding/
worldbuilding/_index.md
worldbuilding/culture/
worldbuilding/magic/
worldbuilding/monsters/
worldbuilding/places/
drafts/
drafts/chapter-01-draft.md
```

The central model is a file-first Edda project folder described by `.edda/project.json`. Edda should be strict enough to stay understandable: users can import, move, rename, or refine their existing projects into the Edda layout, but Open Edda does not need to support every personal organization style.

Elysium remains useful prior art and may be a conversion source, but it is not the target layout. Import tools can help convert from Elysium or other common source layouts into the Edda layout. Once a project is under Edda management, indexing, checkpoints, and sync operate against that defined structure.
