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
  useCases:
    - The story includes identities, histories, or harms the author wants to review carefully.
    - Representation choices may need another pass for trope risk or agency balance.
    - The author wants a constructive memo before deeper revision or publication.
  doNotUse:
    - The author wants automatic rewrites instead of advisory review.
    - The request is purely about grammar or pacing.
    - The author wants this skill to overrule intent instead of informing decisions.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > sensitivity-check > SKILL.md
  scriptStatus: deferred-helper-scripts
---

# $sensitivity-check

Advisory representation review that flags stereotype risk, agency imbalance, and avoidable harm while leaving final creative authority with the author.

## Edda Workflow

1. Read the requested Chapter, Story Text, Story Bible entry, entry section, Attached Note, or Project Note. When reviewing prose, also read enough surrounding context to distinguish depiction from endorsement and to understand genre, point of view, historical setting, narrator reliability, and author intent.
2. Build a representation inventory from the material under review. Track race and ethnicity, culture, nationality, religion, gender, sexuality, disability, neurodivergence, mental health, body size, class, age, and other identities that affect power, danger, or social treatment in the story. Mark identity claims as explicit canon, inference, or unknown.
3. Review representation risk by applying the source states as diagnostic lenses:
   - Cultural context and appropriation: cultural elements used as exotic decoration, sacred or restricted knowledge treated casually, outsider discovery or explanation frames, collapsed distinct cultures, colonial or outdated framing, and language used only for flavor.
   - Gender and misogyny: objectification, unbalanced physical description, gendered violence used as a shortcut, women or non-binary characters existing mainly for men's arcs, "not like other girls" framing, or male gaze treated as neutral.
   - Disability, mental health, neurodivergence, and body size: condition-as-character, cure as required happy ending, inspiration framing, disability as metaphor for evil or deficiency, mental illness linked to violence, fatness moralized or used only for comedy, and accommodations treated as abnormal.
   - Stereotyping patterns: characterization based on assumed group traits, stock roles, dialect caricature, background characters with less individuality than default characters, and any group represented by one person who must stand for everyone.
   - Agency and voice imbalance: who drives plot decisions, who speaks for themselves, whose interiority is centered, who benefits from the resolution, who is explained by outsiders, and whether marginalized characters exist mainly as helpers, lessons, wisdom figures, victims, or proof of another character's goodness.
   - Harmful tropes: death and suffering patterns such as bury-your-gays, fridging, tragic mixed-race identity, dead-disabled-person endings; utility tropes such as magical minority, mystical Indigenous guide, gay best friend, manic pixie dream girl; danger tropes such as depraved bisexual, predatory lesbian, trans deception, or mental illness as violence.
4. Check power, context, and harm before flagging. Ask what affected readers might experience, whether the narrative challenges or endorses the harmful frame, whether the point of view is intentionally biased, whether historical or genre context changes the reading, and whether the scene gives harmed characters dignity, consequence, voice, and aftermath.
5. Review language in context rather than as isolated forbidden words. Flag othering terms, outdated disability phrasing, euphemisms, slurs, dialect rendering, accent jokes, "broken speech", suffering language, gendered body description, and body-size moralizing only with the surrounding narrative function and speaker/narrator stance.
6. Review violence, trauma, and discrimination framing. Distinguish necessary difficult content from gratuitous harm. Check whether violence is used as spectacle, shortcut motivation, punishment for identity, or development for an unaffected character. Look for aftermath, consent-aware framing, proportional detail, and whether survivors keep agency beyond the traumatic event.
7. Assess tokenism and distribution. A single marginalized character is not automatically a failure, and quotas are not the goal. Flag tokenism when one character carries all representation, lacks a personal arc, exists to teach or validate default characters, or has no diversity of experience around them.
8. Rank findings by severity:
   - Critical: likely significant harm or a damaging trope central to the work; strongly recommend addressing before publication.
   - Significant: repeated pattern or structural imbalance; recommend revision.
   - Minor: localized issue, wording problem, or weakly supported pattern; suggest adjustment.
   - Note: context or awareness item that may guide author decisions.
9. For each finding, cite the concrete passage or pattern, name the diagnostic lens, explain the possible reader impact, separate evidence from inference, and suggest alternatives that preserve the author's apparent story goal when possible.
10. Avoid sensitivity-work anti-patterns. Do not act as word police without context, demand representation quotas, require perfect purity, prohibit authors from writing outside their own identities, ignore historical or point-of-view context, or present one review as the final community verdict.

## Edda Output Handling

- Return a concise advisory memo in chat by default. Use headings such as `Representation Inventory`, `Findings`, `Severity`, `Recommendations`, and `Limits`.
- Create an Attached Note when the review belongs to one Chapter, scene, passage, or selected range.
- Create or update a Project Note when the author asks for a durable cross-project representation memo, recurring-risk tracker, or publication-readiness checklist.
- Propose Story Bible changes only when the author explicitly wants recommendations converted into durable character, culture, history, institution, or worldbuilding notes. Mark these as proposals until the author confirms them.
- Do not apply rewrites by default. If the author asks for replacement text, provide options or use Structured Writes only after the target content and author intent are explicit.
- Keep recommendations advisory. This skill informs author decisions; it does not certify safety, grant permission, forbid content, replace research, or replace affected-community sensitivity readers.

## Script Compatibility

The source includes helper scripts for pattern scanning and representation mapping. In Open Edda these scripts remain deferred helpers: do not ask the author to run them, do not treat them as readable reference files, and do not depend on them for the review.

Their methodology is preserved as manual criteria: scan for language patterns, uneven physical description, harmful trope markers, disability and body-size framing, dialogue and dialect risk, identity distribution, agency by identity, survival disparity, trope risks, and tokenism indicators. Pattern absence is not clearance; structural characterization, plot agency, context, and reader impact still require agent review and, when appropriate, human sensitivity readers.
