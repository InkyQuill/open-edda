---
name: cultural-depth
description: Add layered cultural texture through inherited patterns, memetic pressure, norms, taboos, social scripts, and character behavior under culture.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - story_bible_entry
    - entry_section
    - project_note
    - attached_note
    - chapter
    - story_text
  tags:
    - fiction
    - culture
    - character
    - canon-safe
    - optional
  priority: 40
metadata:
  useCases:
    - A culture, family, region, class group, institution, or community feels thin, uniform, stereotyped, or newly invented.
    - Characters need inherited habits, social scripts, taboos, class or regional markers, or internalized contradictions that affect behavior.
    - The author wants customs, artifacts, references, rituals, tensions, or everyday assumptions that imply history without exposition.
    - A scene needs cultural pressure to shape what characters notice, hide, obey, resist, misunderstand, or take for granted.
  doNotUse:
    - The request is mainly about political systems, economics, religion design, or conlang structure rather than lived cultural texture.
    - The author wants only a fast naming pass or surface aesthetic list.
    - The work requires confirmed canon changes and the author has not asked for proposals or Story Bible updates.
  status: optional
  source:
    - docs > skills > suggested > fiction > character > memetic-depth > SKILL.md
  scriptStatus: no-source-helpers
---

# $cultural-depth

Create the perception of lived cultural depth by connecting inherited practices, memetic survival pressures, social scripts, norms, taboos, class and regional variation, and character behavior under those pressures.

## Edda Workflow

1. Read the target Story Text, Attached Note, Project Note, or Story Bible entry. Use `project_map`, `search_content`, `read_content`, `read_chapter`, `read_story_bible_entry`, or `read_entry_section` to identify existing canon about the relevant culture, family, region, class, institution, religion, diaspora, occupation, language community, or historical rupture.
2. Separate confirmed canon from inference. Treat new cultural facts, histories, institutions, rituals, names, taboos, and group assumptions as proposals until the author confirms them.
3. Identify the cultural layer the scene or note needs: inherited family pattern, public norm, taboo, class marker, regional variation, institutional script, survival habit, assimilation pressure, prestige behavior, shame rule, hospitality rule, mourning practice, commercialized tradition, degraded meaning, or cross-cultural synthesis.
4. Apply cognitive triangulation. Build each texture set from roughly 40% recognizable anchors, 40% inferrable transformations, and 20% inscrutable residue. Adjust for context: near-future or opening scenes need more recognizable anchors; fantasy, far-future, first-contact, or emotionally driven climaxes can tolerate more mystery.
5. Track memetic pressure. For each proposed custom, artifact, phrase, or behavior, state why it survived, spread, changed, or became taboo: family enforcement, class aspiration, regional pride, migration, conquest, market demand, religious authority, school discipline, occupational danger, generational rebellion, or political repression.
6. Build a process chain instead of a random list. Show at least one path such as origin -> family adaptation -> regional variant -> commercial version -> misunderstood remnant, or source culture A + source culture B -> contact pressure -> synthesis -> authenticity dispute.
7. Add social scripts. Define what a culturally fluent character is expected to do, what an outsider misses, what a child learns by correction, what a higher-status person can ignore, and what behavior triggers embarrassment, gossip, sanction, suspicion, reverence, or protection.
8. Include norms and taboos as behavior rules, not encyclopedia entries. State what characters avoid saying, who can touch or name an object, when a joke becomes dangerous, what hospitality requires, what grief permits, what public emotion costs, and which violations are forgiven versus unforgivable.
9. Add class, regional, generational, and diaspora variation. Avoid treating the culture as a single voice. Note how elite, rural, urban, borderland, immigrant, youth, elder, professional, religious, secular, or conquered subgroups differ in practice and in what they consider authentic.
10. Surface internalized contradictions. Give characters at least one inherited belief or script that conflicts with their desire, class position, family history, profession, body, language, ambition, faith, politics, or relationships.
11. Filter through point of view. A cultural insider notices misuse, commodification, accent, status, and shame. An outsider notices novelty and may flatten differences. A merchant notices supply, price, authenticity claims, and demand. A rebel notices coercion. A nostalgic character notices loss.
12. Tie texture to action. Convert cultural depth into choices: what a character lies about, performs automatically, refuses to eat, pockets for luck, overpays for, hides from family, corrects in anger, gives as apology, misunderstands, weaponizes, or breaks at a cost.

## Criteria For Good Cultural Depth

- The texture has recognizable anchors, inferrable transformations, and a few unresolved mysteries; it is not all familiar, all strange, or all explained.
- Each detail implies a process: inheritance, degradation, synthesis, commercialization, revival, suppression, migration, class aspiration, regional drift, or generational conflict.
- Culture affects behavior under pressure, not just decor. Characters comply, resist, code-switch, misread, conceal, or exploit cultural expectations.
- Variation exists inside the group. Class, region, generation, institution, family, diaspora status, and proximity to power change the practice.
- Contradictions are internalized. Characters can believe a norm, resent it, depend on it, and violate it in the same story.
- Power dynamics are visible. The output notes who defines authenticity, who profits, whose traditions are mocked or protected, and who pays the social cost.
- Mystery does not block scene comprehension. Inscrutable elements should invite curiosity while the immediate action remains legible.

## Avoiding Stereotype And Monoculture

- Do not reduce a culture to a single trait, food, costume, accent, ritual, moral value, or emotional style.
- Do not make every member obey the same norm with the same intensity. Give exceptions, hypocrites, reformers, rural variants, elite affectations, family-specific rules, and people who only perform the norm in public.
- Do not borrow real-world sacred, traumatic, or marginalized practices as exotic decoration. If a proposal draws from real-world culture, keep it respectful, specific, and transform it through the story world's pressures rather than using it as a costume.
- Do not explain every cultural mystery in narration. Preserve some residue that characters accept, misunderstand, or debate.
- Do not confuse random unfamiliarity with depth. If an element has no process chain or behavioral consequence, cut it or connect it.

## Edda Output Handling

- Return short diagnostics, option lists, and scene-level coaching in chat when the author is deciding.
- Use an Attached Note when the output applies to one chapter, selection, scene, family interaction, or local cultural pressure.
- Use a Project Note for durable option banks, culture audits, ratio checks, process chains, social-script maps, taboo lists, or cross-chapter cultural continuity.
- Use a Story Bible proposal when the work would add or change durable canon: cultural history, institutions, rituals, names, taboos, class structure, regions, family lore, holidays, religious practices, migrations, wars, occupations, language facts, or authenticity disputes.
- Keep canon-safe proposals explicitly labeled as `Confirmed`, `Inferred`, `Proposed`, or `Open Question`. Do not write proposed culture facts into canon unless the author asks for a Story Bible update or confirms them.
- Do not use Structured Writes in this skill unless the author explicitly asks for an applied rewrite. For applied rewrites, preserve the author intent and change only the selected passage or requested scene.

## Script Compatibility

This source skill has no helper scripts. Its useful decision logic is converted into Edda-native guidance: cognitive triangulation, ratio auditing, process chains, social-script mapping, power-dynamics checks, variation checks, anti-stereotype criteria, and canon-safe output handling.
