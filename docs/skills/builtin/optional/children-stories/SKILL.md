---
name: children-stories
description: Draft age-appropriate stories with clear emotional safety, read-aloud rhythm, and themes that land without preaching.
route:
  actionKinds:
    - chat
    - read_check
    - continuation
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - children
    - read-aloud
    - optional
  priority: 60
metadata:
  useCases:
    - The author is drafting or revising fiction for children.
    - Age band, vocabulary level, sentence complexity, and emotional safety all matter.
    - The story needs a moral or theme handled with warmth instead of lectures.
  doNotUse:
    - The target audience is older teen or adult.
    - The author wants high-intensity fear, violence, or ambiguity beyond the intended age band.
    - The request is mainly for marketing copy rather than the story itself.
  status: optional
  source:
    - writer-native
  scriptStatus: writer-native-original
---

# $children-stories

Children's-story drafting and revision support that fits the reader's developmental stage, keeps tension safe, and preserves warmth without condescending or preaching.

## Edda Workflow

1. Establish the audience contract before giving story text. If the author has not named an age, infer only when the project context makes it clear; otherwise ask for the age band, target format, and read-alone versus read-aloud use. Use these working bands:
   - Picture book or preschool read-aloud, ages 3-5: one clear situation, concrete objects, repeated phrases, immediate emotions, short sentences, and an adult voice that can carry the rhythm.
   - Early reader or young read-alone, ages 6-8: simple chapters or scenes, decodable wording where possible, one main problem, explicit cause and effect, and enough repetition to support confidence without flattening the story.
   - Middle grade, ages 9-12: fuller interiority, stronger stakes, subplots or ensemble dynamics when appropriate, richer vocabulary with context support, and emotional complexity that still resolves with orientation and care.
   - Cross-age family read-aloud: layered humor or feeling, clean surface action for younger listeners, and enough musical prose for an adult to enjoy reading it repeatedly.
2. Read available project context with `project_map`, `read_content`, `read_chapter`, `read_story_bible_entry`, or `read_entry_section` when the story belongs to an existing project. Separate confirmed facts from new suggestions before changing names, relationships, setting rules, or recurring motifs.
3. Choose the form before judging prose:
   - Picture books need page-turn logic, imageable beats, low exposition, and room for illustrations rather than dense description.
   - Early readers need lexical control, clear scene geography, short paragraphs, and reward loops such as pattern, joke, discovery, or mastery.
   - Middle grade needs plot causality, character agency, social and emotional consequences, and prose that respects the reader's intelligence.
4. Calibrate reading level against the chosen band. Check vocabulary load, sentence length, abstraction, idiom density, implied background knowledge, paragraph length, and whether a child could track who wants what in each beat. Prefer vivid simple words over babyish substitutions. Define or contextualize rare words through action, not glossary-like interruption.
5. Build gentle conflict with a visible safety frame. Give the child reader a worry, want, mystery, mistake, or social friction that matters, but keep threat intensity proportional to the age band. For younger children, fear should be brief, named, and held by a reassuring narrative presence. For older children, uncertainty can last longer, but avoid hopelessness, graphic harm, humiliation-as-comedy, or irreversible loss unless the author explicitly wants a grief-focused book and the age band can hold it.
6. Use repetition, rhythm, and pattern as structure, not filler. Repeat a phrase, action, sound, counting pattern, contrast, or return image when it helps anticipation, memory, participation, or page turns. Vary repeated material enough that each return escalates, reveals, comforts, or pays off.
7. Test adult read-aloud flow. Read the prose mentally for breath length, mouth feel, stress pattern, dialogue handoffs, and places where the adult reader would stumble. Break long sentences, remove tongue-twisting clutter unless intentional, and make character voices distinct without relying on stereotypes or dialect caricature.
8. Handle morals and themes through choice, consequence, care, and change. Do not state the lesson as a lecture unless the genre convention or author brief calls for it. Let the child protagonist participate in solving the problem; avoid making an adult deliver the whole answer. A good theme can be named softly near the end, but it should already be proven by the story.
9. Respect parent, caregiver, educator, and classroom constraints. Flag content that may need author approval: unsafe imitable behavior, food or body shame, exclusion, religious or political messaging, stereotypes, bullying methods, medical claims, nightmares at bedtime, or school-policy-sensitive material. Offer a lower-intensity alternative instead of silently removing the author's premise.
10. When revising, diagnose before rewriting. Identify the current intended band, the mismatched elements, and the smallest changes that would align form, reading level, fear level, rhythm, and theme. Preserve the author's premise and voice unless they ask for a new version.

## Quality Criteria

Good children's-story output should:

- Name the target age band or format assumptions when they affect the answer.
- Give the child character meaningful agency at the story's scale.
- Use concrete sensory details and actions before abstract explanation.
- Keep conflict legible, emotionally safe, and worth resolving.
- Make repetition, rhythm, or page-turn shape serve comprehension and delight.
- Sound natural when read aloud by an adult.
- Let the theme arise from the story's events rather than from scolding.
- Maintain canon boundaries when drafting inside an existing project.

Weak output includes:

- Generic sweetness with no problem, desire, surprise, or consequence.
- Vocabulary that is either babyish for the audience or needlessly adult.
- Fear, danger, shame, or moral punishment that exceeds the requested age band.
- A lesson delivered by authorial lecture after the plot has stopped.
- Dense narration that leaves no room for illustrations in picture-book work.
- "For kids" simplification that removes emotional truth, humor, or agency.

## Edda Output Handling

- Return age-band guidance, diagnosis, revision notes, or draft story text in chat by default.
- Create an Attached Note when the result belongs to one story draft, one chapter, one selected passage, or one read-aloud test.
- Create or update a Project Note when the author wants a reusable children's-story brief, series age ladder, classroom constraints list, parent-facing constraints list, or revision checklist.
- For Story Text changes, use Structured Writes only when the author explicitly asks to draft, continue, or rewrite a known chapter or selection and the target content is available.
- Treat new recurring character facts, setting rules, family relationships, invented places, school details, magical rules, and series continuity as proposals until the author confirms them for the Story Bible.
- Label canon-sensitive suggestions as `Existing canon`, `Inference`, or `Proposal` when project context is involved.

## Script Compatibility

This is an Edda-native optional skill with no source folder and no scripts. It works through Edda project-reading tools, chat guidance, notes, and explicit Structured Writes.
