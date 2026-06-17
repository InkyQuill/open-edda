---
name: identity-denial
description: Structure character arcs where a protagonist denies the identity their actions, evidence, and external labels increasingly confirm.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
    - story_bible_entry
  tags:
    - fiction
    - character
    - arc
    - optional
  priority: 46
metadata:
  useCases:
    - A character insists they are not becoming something while their behavior, choices, or consequences suggest otherwise.
    - A fall, corruption, addiction, trauma, inheritance, class, role, professional, or monster-transformation arc needs denial structure.
    - The author asks for mirror moments, rationalizations, recognition scenes, external labels, or costs of accepting a changed identity.
    - A scene or chapter needs tests for whether self-concept, evidence, and social recognition are escalating together.
  doNotUse:
    - The story needs a straightforward growth arc without self-deception or resistance to identity change.
    - The request is mainly about plot logistics, lore generation, or world rules rather than character self-concept.
    - The character already understands and accepts the identity shift, and the task is only about consequences.
    - The author wants broad feedback without a denial-focused character-arc lens.
  status: optional
  source:
    - docs > skills > suggested > fiction > structure > identity-denial > SKILL.md
  scriptStatus: no-source-helpers
---

# $identity-denial

Character-arc support for diagnosing, designing, and revising stories where denial of a changing identity acts as both internal conflict and plot engine.

## Edda Workflow

1. Read the target Chapter, Story Text selection, Attached Note, Project Note, or Story Bible entry named by the author. If the denial arc depends on established character history, relationships, institutions, supernatural rules, or prior choices, use `project_map`, `search_content`, `read_chapter`, `read_content`, `read_story_bible_entry`, or `read_entry_section` to separate confirmed canon from inference.
2. Name the denied identity in concrete terms: moral identity, social identity, psychological identity, relational identity, professional identity, or genre-specific transformation. State the character's self-concept as an "I am not X" or "I am still Y" claim, then list the evidence that contradicts it.
3. Locate the denial level:
   - Surface denial: the character says they are not like a group while behaving like it.
   - Rationalized denial: the character creates rules that exempt their actions from the identity label.
   - Projected denial: the character condemns in others what they refuse to see in themselves.
   - Desperate denial: the character stages increasingly elaborate proof that the label is false while the proof deepens the truth.
4. Map the staged denial arc. Identify the inception point, justification phase, escalation markers, mirror moments, crisis point, and likely resolution. If any stage is missing, explain the functional gap it creates on the page.
5. Test the pressure system:
   - Self-concept: what the character says they are, what vocabulary they avoid, and what old rituals they maintain as proof.
   - Evidence: actions, props, bodies, records, consequences, repeated choices, and irreversible thresholds.
   - External labels: what allies, enemies, institutions, communities, victims, or witnesses call the character.
   - Plot pressure: how each attempt to disprove the identity creates new evidence for it.
6. Track the justification ladder from "just this once" through "just until," "only when necessary," "they deserved it," and possible acceptance. Keep the logic internally consistent even when it is morally or factually false.
7. Assign truth mirrors by function, not decoration:
   - The Namer speaks the denied identity aloud.
   - The Corrupted Sage is farther along the same path.
   - The Innocent sees clearly without theory.
   - The Abandoned carries the cost of the denial.
   - The Dark Twin accepts what the protagonist denies.
   Also identify enablers, challengers, witnesses, and parallels when they create pressure or contrast.
8. Design or evaluate reveal and recognition beats. A strong beat changes what can be denied, what others can safely ignore, or what the character must sacrifice to keep denial alive. A weak beat only repeats information the reader and character already understand.
9. Choose an ending pattern that pays the cost of acceptance or refusal: tragic collapse, dark acceptance, redemptive recognition, delusional victory, integration, recontextualization, partial recognition, perpetual tension, or cyclical return. State what the character loses if they accept the identity and what they lose if they refuse it.
10. For scene-level work, run scene tests:
    - What identity label is being resisted in this scene?
    - What concrete evidence appears on the page?
    - Who notices, names, enables, or challenges it?
    - What rationalization does the character use now?
    - How is this denial different from the previous denial beat?
    - What pressure rises by the end of the scene?
11. Diagnose anti-patterns directly:
    - Obvious from start: make the first denial understandable and evidence gradual.
    - Inconsistent rationalization: define the character's false-but-coherent rules and exceptions.
    - Missing point of no return: add a specific irreversible action or recognition threshold.
    - Consequence-free acceptance: attach real loss to truthful recognition.
    - Unsympathetic entry: show why denial protects something the character plausibly values.
12. Propose revisions at the smallest useful scope: a sharper label, a better mirror character, a stronger pressure beat, a more coherent rationalization, a clearer point of no return, or a resolution with costs. Do not invent durable canon as fact; mark new history, identity labels, relationships, supernatural transformations, or institutional recognition as proposals for author confirmation.

## Edda Output Handling

- Use chat for short diagnosis, coaching, option sets, scene tests, and author decisions.
- Create an Attached Note when the analysis belongs to one Chapter, selection, beat, or scene cluster.
- Create or update a Project Note when the author wants a durable denial-arc map, mirror-character plan, justification ladder, or cross-chapter pressure tracker.
- Propose Story Bible changes only for durable canon such as changed character identity, backstory, addiction history, faction membership, supernatural status, public reputation, relationships, institutions, names, rules, or timeline facts. Label these as proposals until the author confirms them.
- Use Structured Writes or direct write tools only when the author explicitly asks to revise selected Story Text. Keep rewrites canon-safe by preserving established facts unless the author approved a proposed change.

## Script Compatibility

This source skill has no helper scripts. Its useful decision logic is converted into Edda-native diagnosis, planning, and rewrite guidance.
