---
name: dialogue-check
description: Dialogue diagnosis for flat conversations, identical voices, weak subtext, exposition dumps, and dramatically inert exchanges.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
    - story_bible_entry
  tags:
    - fiction
    - dialogue
    - voice
    - subtext
  priority: 82
metadata:
  useCases:
    - Characters sound too similar, interchangeable, wooden, or identifiable only by dialogue tags.
    - Conversations explain information but do not create tension, subtext, conflict, relationship movement, or scene change.
    - The author asks why a chapter exchange feels flat, on-the-nose, over-expository, too balanced, or dramatically inert.
    - The author wants chapter-level or selection-level dialogue diagnosis before deciding whether to revise.
  doNotUse:
    - The author wants the skill to draft replacement dialogue by default.
    - The main problem is scene structure, chapter pacing, plot logic, or prose style rather than dialogue.
    - The author is still in early brainstorming and has no text to inspect.
  status: default
  source:
    - docs > skills > suggested > fiction > character > dialogue > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $dialogue-check

Focused dialogue review for authors who need a concrete diagnosis of why an exchange feels flat, indistinct, over-explanatory, or too literal.

## Edda Workflow

1. Read the target exchange in its surrounding chapter context with `read_chapter` or `read_content`. If the author references a selection, inspect enough before and after the selection to understand the scene goal, speaker relationship, and immediate pressure on the conversation.
2. Read relevant Story Bible entries or entry sections for the speakers when voice, status, relationship history, rank, dialect, profession, secrets, or canon facts affect the exchange. Treat missing character context as an uncertainty, not as a license to invent canon.
3. Identify the failing layer:
   - Text: the spoken words, rhythm, diction, tags, beats, interruptions, and naturalness.
   - Subtext: the gap between what characters say and what they mean, hide, want, or refuse to admit.
   - Context: power, status, history, stakes, agenda, and the scene function shaping the exchange.
4. Diagnose voice differentiation. Check whether each speaker can be recognized without tags; whether vocabulary, sentence length, formality, directness, question rate, contractions, fragments, evasions, and verbal avoidances differ; and whether emotional state changes each character's speech in character-specific ways.
5. Diagnose subtext. Ask what each speaker wants, what each cannot ask for directly, what is being avoided, and where words contradict behavior or situation. Flag dialogue that states feelings, motives, themes, or intentions too directly.
6. Diagnose agenda and conflict. For every major participant, name the surface goal, hidden agenda, source of resistance, and whether the exchange creates pressure through disagreement, bargaining, deflection, lying, testing, threat, seduction, apology, or refusal.
7. Diagnose exposition. Flag lines where characters tell each other facts they already know, one speaker exists mainly to ask prompting questions, backstory arrives as lecture, or information has no disagreement, discovery, cost, error, or changed relationship attached to it.
8. Diagnose rhythm, pacing, and interruption. Check whether tense moments use shorter exchanges, overlap, silence, fragments, unfinished sentences, or compression; whether slower moments earn longer speeches; whether the back-and-forth is too evenly alternating; and whether pauses or action beats control emphasis.
9. Diagnose status and power. Track who dominates, interrupts, evades, yields, redirects, withholds, wins, loses, or remains silent. Flag exchanges where all speakers have equal weight despite unequal rank, leverage, knowledge, intimacy, or fear.
10. Diagnose scene function. Apply the double-duty test: a dialogue scene should do at least two and preferably three of these jobs: advance plot, reveal character, create or exploit subtext, shift a relationship, intensify tension, clarify stakes, or change the scene state. Flag single-function exchanges that merely transfer information.
11. Diagnose dialogue tags and action beats. Prefer invisible tags such as `said` when attribution is needed; flag decorative said-bookisms, emotional adverbs that tell the reader how to feel, beats that only choreograph empty movement, and missing beats where silence, gesture, or contradiction would carry subtext.
12. Classify the main issue using concrete labels: identical voices, wooden dialogue, exposition dump, no subtext, single-function dialogue, pacing mismatch, weak agenda/conflict, flat status dynamics, tag/beat overhandling, or unclear scene function. Note secondary issues separately.
13. Provide evidence from the text by paraphrase or very short excerpts. Do not quote long passages back to the author.
14. Give revision guidance, not replacement dialogue by default. Recommend interventions such as assigning each speaker a verbal DNA profile, adding hidden agendas, converting exposition into conflict or discovery, varying rhythm, adding interruption or silence, making status visible, changing the scene endpoint, or replacing explanatory tags with revealing action beats.
15. Ask targeted questions only when needed to resolve diagnosis, such as what each speaker wants, what they are hiding, who has power, what information must reach the reader, or how the relationship should change by the end.
16. If the author explicitly asks for an applied rewrite after diagnosis, state that the task has shifted to rewriting and use the appropriate rewrite workflow or `$story-collaborator`. Preserve the diagnosis as constraints for that rewrite.

## Edda Output Handling

- Return the diagnosis in chat for quick review.
- Create an Attached Note when the report belongs to a chapter or selected exchange.
- Create or update a Project Note when repeated dialogue patterns should guide later revision.
- Structure output as: main diagnosis, evidence, affected criteria, likely cause, revision moves, and open questions.
- Keep the distinction between diagnosis and rewrite explicit. This skill can suggest techniques and constraints, but it does not supply polished replacement dialogue unless the author explicitly changes the task.
- Do not propose Story Bible changes unless the dialogue problem comes from canon inconsistency or missing durable character voice guidance. If durable character facts, relationships, status, secrets, or speech rules need to change, present them as Story Bible proposals for author confirmation.
- Do not use Structured Writes or direct text replacement in this skill unless the author explicitly switches to an applied rewrite workflow.

## Script Compatibility

The source `voice-check.ts` and `dialogue-audit.ts` helpers are converted to Edda-native guidance in this skill. Their methodology may inform manual diagnosis: compare vocabulary overlap, sentence length, contractions, questions, fragments, function coverage, subtext signals, tag usage, action beats, and anti-patterns.

Do not ask the runtime agent to read, run, shell out to, or import source scripts as reference files. If equivalent helpers are later approved as non-mutating Edda `skill_script` tools, use them only when `skill` reports that the script is enabled and approved; treat their output as advisory evidence that must be checked against the actual scene context. Until then, script execution is deferred.
