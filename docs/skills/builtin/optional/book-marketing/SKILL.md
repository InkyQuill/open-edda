---
name: book-marketing
description: Turn a finished or near-finished story into blurbs, platform copy, taglines, and query-ready positioning without overselling the wrong book.
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
    - publishing
    - marketing
    - optional
  priority: 24
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > application > book-marketing > SKILL.md
  scriptStatus: source-templates-retained
---

# $book-marketing

Publishing-adjacent copy support for authors who need blurbs, descriptions, taglines, or pitch language that sells the reading experience instead of summarizing the plot.

## Use When

- A draft is complete or close enough to position honestly.
- The author needs back-cover copy, store description text, taglines, or query-style pitch language.
- Existing copy feels generic, spoilery, or mismatched to genre promise.

## Do Not Use When

- The book itself is still too undefined to market clearly.
- The request is for fiction drafting rather than marketing copy.
- The author wants guaranteed market comps or sales predictions.

## Edda Workflow

1. Read enough Story Text, notes, or summary material to understand genre promise and stakes.
2. Identify the core reader promise, not just the plot events.
3. Choose the deliverable shape: blurb, store description, taglines, or query pitch.
4. Match the copy to the actual emotional experience of the book.
5. Keep marketing language separate from canon and Story Text unless the author wants it stored as notes.

## Edda Output Handling

- Return marketing copy in chat by default.
- Create an Attached Note when the copy belongs to one excerpt, one launch conversation, or one focused deliverable.
- Create or update a Project Note when the author wants a reusable marketing packet.
- Do not propose Story Bible changes in this skill.
- Do not use Structured Writes in this skill.

## Bundled Templates

This skill includes publishing-adjacent templates as Writer-native references:

- `templates/blurb.md` — Book blurb structure and guidance.
- `templates/query.md` — Query letter template.
- `templates/taglines.md` — Tagline and logline templates.
- `templates/amazon.md` — Amazon store page description template.

The agent should reference these templates in Agent Session responses when the author asks for marketing copy.

## Script Compatibility

This rewrite preserves built-in blurb, store, tagline, and query templates as Edda-native references. Script execution is unavailable in Milestone 3.5, and this skill works through guidance, templates, and reviewable marketing drafts only.
