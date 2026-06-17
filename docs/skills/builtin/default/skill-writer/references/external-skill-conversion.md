# External Skill Conversion Reference

Use this reference only when converting a source skill from another agent system into an Open Edda runtime skill.

## Conversion Targets

Convert source material into the Open Edda skill folder contract:

```txt
skill-folder/
  SKILL.md
  references/
  templates/
  data/
  scripts/
```

Only `SKILL.md` is required. Add other folders only when they reduce context load or preserve useful reusable material.

## Platform Mapping

| Source pattern | Edda conversion |
| --- | --- |
| `Use When`, `When to Use`, `Don't Use`, trigger lists | Move to `metadata.useCases` and `metadata.doNotUse`. |
| Long setup explanation | Delete unless it changes future agent behavior. |
| Local file paths, repository reads, `read_file` | Replace with `read_content`, `read_chapter`, `read_story_bible_entry`, `read_entry_section`, `search_content`, or `project_map`. |
| Shell commands, grep, file watchers | Replace with Edda search/read tools, approved `skill_script`, or deferred status. |
| Subagents or terminal agents | Replace with explicit Edda workflow steps for the current agent. |
| Large examples | Move to `references/` and tell the agent when to call `read_skill_file`. |
| Output examples or reusable document shapes | Move to `templates/`. |
| Tables, taxonomies, lexicons, checklists | Move to `data/` or `references/`. |
| Scripts | Treat as disabled by default; convert to guidance/data or mark as deferred/approved helper. |
| Canon or lore edits | Convert to Story Bible proposals unless author explicitly confirms durable changes. |

## Source-Specific Notes

### OpenClaw

OpenClaw-style skills often mix routing, procedures, and tool assumptions. Keep the procedure only when it can run through Edda tools. Move all routing into metadata. Replace local resource reads with `read_skill_file` if the resource ships inside the skill.

### skills.sh

Shell-oriented skills often assume executable scripts, pipes, and filesystem inspection. Convert the deterministic part into a checklist or data file. Keep script execution deferred unless Skill Script Runtime has an approved helper.

### Claude Skills

Claude skills often include excellent discipline rules and rationalization blockers. Preserve those if they affect agent behavior. Remove platform-specific mentions of Claude Code, local worktrees, filesystem reads, and subagents unless the same behavior has an Edda tool equivalent.

### Codex Skills

Codex skills often separate `SKILL.md`, `references/`, `scripts/`, and `assets/`. Preserve this progressive disclosure pattern, but translate Codex filesystem access into Edda `read_skill_file` and project context tools. Do not copy Codex-only validation commands into runtime instructions.

## Rewrite Checklist

1. Name the target skill with lowercase letters, digits, and hyphens.
2. Write a trigger-focused `description`.
3. Add route fields for the expected action and content kinds.
4. Move positive selection cases to `metadata.useCases`.
5. Move routing-away cases to `metadata.doNotUse`.
6. Write a concrete `## Edda Workflow`.
7. Write explicit `## Edda Output Handling`.
8. Add `## Script Compatibility`.
9. Add reference/template/data routing only if files exist.
10. Verify the skill body tells the future agent what to read before acting.

## Red Flags

Rewrite again if the draft:

- Requires the agent to know full worldbuilding or all chapters without choosing what to read.
- Tells the agent to inspect local files, paths, folders, or repository state.
- Says to run shell commands or grep project files.
- Requires a script but does not state whether it is approved, deferred, or converted.
- Contains vague instructions such as "improve the writing" without context, action, and output rules.
- Lets the agent silently change canon.
- Forces the agent to load long examples before it knows they are relevant.
