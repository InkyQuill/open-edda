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
  useCases:
    - A draft is complete or close enough to position honestly.
    - The author needs back-cover copy, store description text, taglines, or query-style pitch language.
    - Existing copy feels generic, spoilery, or mismatched to genre promise.
  doNotUse:
    - The book itself is still too undefined to market clearly.
    - The request is for fiction drafting rather than marketing copy.
    - The author wants guaranteed market comps or sales predictions.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > book-marketing > SKILL.md
  scriptStatus: source-templates-retained
---

# $book-marketing

Publishing-adjacent copy support for authors who need blurbs, descriptions, taglines, or pitch language that sells the reading experience instead of summarizing the plot.

Core principle: marketing copy promises a reading experience. It should create desire for the book, not recap the whole story.

## Edda Workflow

1. Read only the context needed to market honestly: the author's brief, relevant Project Notes, Attached Notes, Story Bible entries for protagonist/genre/setting, and enough Story Text or summary material to understand the first-act setup, core conflict, tone, and stakes.
2. Separate confirmed book facts from positioning guesses. If genre, protagonist, emotional core, stakes, or audience is unclear, ask targeted questions before drafting.
3. Diagnose the marketing state before generating:
   - M1 No Marketing Copy Exists: establish elemental genre, emotional core, one-sentence stakes, and at least one comparable-title direction before drafting.
   - M2 Copy Summarizes Instead of Sells: replace "first X, then Y, then Z" plot recap with hook, emotional context, conflict, and stakes.
   - M3 Copy Signals Wrong Genre: compare the copy's tone, keywords, and comps against the book's actual genre promise; realign copy voice to the book.
   - M4 Copy Is Generic or Forgettable: add one concrete "only in this book" detail and replace abstract adjectives with specific images or pressures.
   - M5 Platform Mismatch: adapt length, formatting, and structure for the destination instead of reusing one block everywhere.
   - M6 Tagline or Hook Does Not Stick: generate short variants, read for rhythm, remove filler, and keep only lines that create curiosity or promise.
4. Define the reader promise in one sentence: genre experience plus emotional payoff plus the pressure that makes the book distinctive. Do not let the promise exceed what the book actually delivers.
5. Build positioning around hook, protagonist, conflict, stakes, and genre fit:
   - Lead with the hook, not backstory or worldbuilding.
   - Ground the protagonist with emotional context, not just biography.
   - State what they want, what blocks them, and what they risk losing.
   - Stop at the moment of choice or pressure; do not reveal climax, solution, twist resolution, or ending.
   - Match the copy's voice to the book's tone and category.
6. Choose one deliverable shape and load only the matching template with `read_skill_file`:
   - `templates/blurb.md` for back-cover or library-style copy, usually 150-200 words.
   - `templates/amazon.md` for store-page description, mobile-readable HTML, comp-title sections, and optional editorial quotes.
   - `templates/taglines.md` for ads, social posts, hooks, and short positioning lines.
   - `templates/query.md` for agent/publisher pitch paragraphs and query-style positioning.
7. Apply genre-aware promise checks:
   - Wonder promises awe, discovery, and new perspective.
   - Horror promises fear, dread, and can't-look-away tension.
   - Mystery promises a puzzle, clues, and an answer withheld from the copy.
   - Thriller promises stakes, countdown pressure, and survival or exposure.
   - Romance or relationship stories promise emotional connection and satisfying relational payoff.
   - Adventure promises excitement and forward momentum.
   - Drama promises character transformation and emotional journey.
   - Humor promises wit, entertainment, and laugh-worthy situations.
   - Idea-driven fiction promises intellectual stimulation or perspective shift.
8. Use comparable titles carefully. Comps can clarify shelf, tone, audience, or differentiator, but they are not proof of sales potential. Prefer recent, recognizable, same-category, accurate comps; avoid mega-bestsellers as primary comparisons unless the author explicitly wants broad shorthand. If current market fit matters, ask the author to supply or verify comps.
9. Test the draft before presenting it:
   - First sentence creates curiosity.
   - Stakes appear early and are personal enough to matter.
   - At least one specific detail could not describe another book.
   - Genre signals match the actual manuscript.
   - Copy ends with tension, not resolution.
   - The deliverable follows the selected platform's length, format, and tone constraints.
10. Name anti-patterns when diagnosing existing copy:
   - Synopsis Trap: plot sequence instead of desire.
   - Vague Intrigue: mystery language with no concrete information.
   - Spoiler Reveal: copy reaches climax, solution, or aftermath.
   - Feature List: elements listed without emotional connection.
   - Throat Clear: backstory or worldbuilding before the hook.
   - Wrong Voice: copy tone mismatches the book's tone or market category.
   - Author Intrusion: author explanation displaces the book's promise.
11. Keep this skill publishing-adjacent. Do not write new fiction, decide the author's genre against their stated intent, promise sales outcomes, invent credentials, create author bios or websites, or build paid-ad campaign strategy. Refer genre diagnosis, story-structure issues, cliche work, naming/sound work, or manuscript revision to more appropriate skills when those are the real blocker.

## Edda Output Handling

- Return diagnosis, questions, and draft copy in chat by default.
- For existing-copy audits, include the marketing state, the main failure mode, the promised reader experience, and the concrete rewrite criteria used.
- For generated copy, label deliverables clearly and include only useful variants, not a large undifferentiated dump.
- Create an Attached Note only when the author asks to keep copy tied to one chapter, excerpt, selection, or focused launch deliverable.
- Create or update a Project Note only when the author explicitly wants a reusable marketing packet, positioning note, or cross-platform copy bank.
- Do not write directly into Story Text.
- Do not create or update Story Bible canon. If marketing work reveals a canon ambiguity, report it as an open question or proposal for the author to confirm elsewhere.
- Do not use Structured Writes in this skill.

## Templates

This skill includes publishing-adjacent templates as Writer-native references:

- `templates/blurb.md` — Book blurb structure and guidance.
- `templates/query.md` — Query letter template.
- `templates/taglines.md` — Tagline and logline templates.
- `templates/amazon.md` — Amazon store page description template.

Load the matching template with `read_skill_file` only after choosing the deliverable shape. Do not load all templates for a short positioning audit. When the author asks for multiple platform versions, load one template at a time, draft that deliverable against its criteria, then proceed to the next.

## Script Compatibility

The source skill has no scripts. This rewrite preserves built-in blurb, store, tagline, and query templates as Edda-native references and converts diagnostic behavior into Edda workflow guidance. The skill works through context reading, `read_skill_file` templates, chat drafts, and optional author-requested notes only.
