---
name: sensitivity-check
description: Review representation choices for stereotype risk, agency imbalance, and avoidable harm while keeping the author in charge of the story.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - story_bible_entry
    - entry_section
    - attached_note
    - project_note
  tags:
    - fiction
    - sensitivity
    - representation
    - optional
  priority: 30
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > application > sensitivity-check > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $sensitivity-check

Advisory representation review for authors who want to catch stereotype risks, agency problems, or avoidable harm without handing over creative authority.

## Use When

- The story includes identities, histories, or harms the author wants to review carefully.
- Representation choices may need another pass for trope risk or agency balance.
- The author wants a constructive memo before deeper revision or publication.

## Do Not Use When

- The author wants automatic rewrites instead of advisory review.
- The request is purely about grammar or pacing.
- The author wants this skill to overrule intent instead of informing decisions.

## Writer Workflow

1. Read the target Story Text, Chapter, Story Bible material, or note set in context.
2. Identify the represented identities, power dynamics, and likely concern areas.
3. Flag concrete passages or patterns, explain why they may land badly, and separate severity levels.
4. Offer revision directions that preserve the author's goals where possible.
5. Leave final canon and wording decisions to the author.

## Writer Output Handling

- Return the advisory review in chat by default.
- Create an Attached Note when the feedback belongs to one Chapter, one passage, or one selection.
- Create or update a Project Note when the author wants a broader representation memo across the project.
- Propose Story Bible updates only when the author wants guidance converted into durable canon or character notes.
- Do not use Structured Writes in this skill.

## Script Compatibility

This rewrite adapts source audit logic into Writer-native review guidance. Source helper scripts are not runnable in Milestone 3.5, so the skill works through analysis, references, and reviewable notes only.
