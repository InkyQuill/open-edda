---
name: governance-systems
description: Build states, councils, empires, guild powers, and internal factions with believable authority limits and competing interests.
route:
  actionKinds:
    - chat
    - read_check
    - story_bible
  contentKinds:
    - story_bible_entry
    - entry_section
    - project_note
    - attached_note
    - chapter
    - story_text
  tags:
    - fiction
    - worldbuilding
    - governance
    - optional
  priority: 42
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > governance-systems > SKILL.md
  scriptStatus: no-source-helpers
---

# $governance-systems

Political worldbuilding for authors who need power structures, legitimacy, factional conflict, and administrative limits to feel concrete.

## Use When

- A story depends on institutions, succession, border tensions, bureaucracy, or elite conflict.
- A kingdom, republic, corporate state, fleet, or council needs internal logic.
- The author wants political stakes that feel bigger than one villain.

## Do Not Use When

- The request is mainly about prose, pacing, or local scene revision.
- The world only needs a decorative title hierarchy.
- The author wants to freeze canon before exploring alternatives.

## Edda Workflow

1. Read the relevant Story Bible context and any Chapters already using the polity.
2. Define how authority is claimed, distributed, resisted, and enforced.
3. Map internal factions, practical limits, and external pressures.
4. Check whether the current structure supports the story's conflicts and scale.
5. Keep proposed institutional facts reviewable until the author confirms them.

## Edda Output Handling

- Return the governance model in chat by default.
- Create an Attached Note when the work serves one Chapter, one polity, or one conflict.
- Create or update a Project Note for broader political exploration.
- Propose Story Bible updates only after author confirmation of offices, laws, factions, or legitimacy claims.
- Do not use Structured Writes in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native political analysis and canon-safe proposal output.
