---
name: story-analysis
description: Structured analysis for completed chapters or stories, focused on strengths, weaknesses, and the highest-value revision opportunities.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - analysis
    - chapter-review
    - revision
  priority: 84
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > core > story-analysis > SKILL.md
  scriptStatus: no-source-helpers
---

# $story-analysis

Structured post-draft analysis for authors who want a thorough read on what a chapter or complete story is accomplishing and where revision effort will matter most.

## Use When

- A chapter or complete story draft is ready for serious evaluation.
- The author wants more than a quick check and needs a structured report.
- The author wants strengths, weaknesses, and prioritized revision targets.

## Do Not Use When

- The text is still being actively discovered and the author mainly needs drafting momentum.
- The request is narrowly about dialogue, prose polish, or pacing.
- The author wants direct rewriting instead of analysis.

## Writer Workflow

1. Read the full target chapter or story in context.
2. Evaluate narrative role, character work, pacing, clarity, and payoff.
3. Separate high-impact problems from smaller craft notes.
4. Summarize what is already working so revision does not damage strengths.
5. End with a prioritized set of revision targets.

## Writer Output Handling

- Return the core analysis in chat for quick discussion.
- Create an Attached Note when the report belongs to one chapter.
- Create or update a Project Note when the report should guide later revision passes across the project.
- Do not propose Story Bible changes unless the analysis uncovers a canon conflict worth tracking.
- Do not use Structured Writes in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Writer-native analysis and note output.
