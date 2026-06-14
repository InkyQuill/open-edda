# Open Edda — Roadmap & ADR Gap Report

**Date:** 2026-06-14  
**Scope:** Roadmap milestones 1–4, все 12 ADR, планы в `docs/superpowers/plans/`

---

## Итоговая сводка

| Milestone | Roadmap | Факт | Вердикт |
|---|---|---|---|
| M1: Project Core | Implemented | Реализован | ✅ PASS — 1 minor gap |
| M2: Agent Core | Implemented | Реализован | ✅ PASS — 4 gaps |
| M3: Skill Core | Planned | Реализован | ✅ PASS — roadmap устарел |
| M3.5: Skill Library | Planned | Реализован | ⚠️ PASS — 3 gaps |
| M3.6: Script Runtime | Planned | Не начат | 🔴 GAP |
| M4: Daily Writing Polish | Planned | Не начат | 🔴 GAP |
| Later: Local Sync | Deferred | — | — |
| Later: Story Dashboard | Deferred | — | — |

---

## M1: Project Core — ✅ PASS (1 gap)

### Реализовано полностью

- SQLite schema: 7 таблиц + FTS5 + триггеры (миграция 00001)
- 13 sqlc-запросов, все методы `project.Service`
- Structured writes: `AppendToContent`, `InsertIntoContent`, `ReplaceContentRange`, `UpdateEntrySectionBody`
- UTF-8 boundary validation в `InsertIntoContent` и `ReplaceContentRange`
- Optimistic concurrency: двойная проверка (Go-level + SQL WHERE + affected rows)
- `SearchContent` с post-фильтрацией по `metadataFilters` и `tags`
- `ProjectMap` с sections и relations для story_bible_entry
- HTTP API: все 9 endpoints из плана
- Elysium Import/Export через `archive/zip`
- Error mapping: 400/404/409/500
- React: Project Dashboard + Writing Workspace shell

### Gaps

| # | Описание | Серьёзность |
|---|---|---|
| M1.1 | **Нет HTTP-рута `GET /projects/{id}/map`.** `ProjectMap` есть в сервисе, но не экспонирован через HTTP — фронтенд не может его использовать. | Medium |
| M1.2 | **Auth placeholder.** `author-1` захардкожен, API-ключи хранятся plaintext. Нет отдельного плана на auth. | High |

---

## M2: Agent Core — ✅ PASS (4 gaps)

### Реализовано полностью

- OpenAI-compatible provider client с нормализацией usage (включая DeepSeek `prompt_cache_hit_tokens`)
- Расчёт стоимости: `tokens * price_per_million / 1_000_000`
- 4 pricing поля: input/output/cache-read/cache-write на model variant
- Prompt assembly с 7 context source snapshots + layered writing briefs
- System prompt verbatim совпадает с планом
- 13 tools: все context tools + все write tools + skill tool
- Tool result artifacts с `full_result_json` + `model_visible_markdown`
- Truncation logic: 2000 строк / 50 KiB с флагом `truncated`
- Chat turns с tool loop (max 4 раунда)
- Quick actions: Continuation, Rewrite, Read and Check
- Preview accept/reject с re-check `expectedRevision`
- `PrunePromptRecords` с retention_days из prompt_profile
- Prompt records с context snapshots
- Activity events на каждый tool call
- Все HTTP endpoints из плана (17+)
- Frontend: AgentSettings, AgentPanel, activity trail, spend display
- Apply mode toggle (preview/direct_apply)
- Continuation controls: target type (words/sentences), count, guidance
- Preview Accept/Reject в UI
- Provider disclosure (model label в quick actions header)

### Gaps

| # | Описание | Серьёзность |
|---|---|---|
| M2.1 | **`update_entry_section` tool без `expectedRevision`.** В отличие от всех остальных write tools, нет optimistic concurrency. Два конкурентных обновления секции могут silently overwrite друг друга. `tools.go:57-62`, `service.go:453`. | **Critical** |
| M2.2 | **Нет явной redaction API-ключей в prompt records.** Ключ не попадает в request/response JSON при текущем коде, но нет sanitization-функции как safety net. План требует: «Prompt records redact API keys.» | Medium |
| M2.3 | **Нет insert toggle в Continuation UI.** `insert: false` захардкожен в `App.tsx:1442`. Бэкенд поддерживает выбор позиции вставки, но UI не даёт авторам этого контролировать. | Medium |
| M2.4 | **Нет selection-based Rewrite/Read&Check в UI.** Быстрые действия всегда используют `selectionStart: 0, selectionEnd: конец_главы`. Бэкенд поддерживает byte-offset ranges, но UI не даёт выделить фрагмент текста. | Medium |

---

## M3: Skill Core — ✅ PASS

**Roadmap устарел** — статус «Planned», но реализация полностью завершена (20 коммитов, от `3028297` до `2380050`).

- Схема: skills, skill_files, skill_routing_hints, agent_session_skills (миграция 00003)
- Парсер `ParseSkillArchive`: zip, path traversal protection, YAML-like frontmatter, file classification
- `skill.Service`: Install, List, Get, GetByName, ListRoutable, SelectSessionSkills, ListSessionSkills, RenderForModel
- `RenderForModel`: XML-wrapped вывод с 40KB budget и truncation
- HTTP API: 5 endpoints
- Agent integration: `SkillProvider` interface, `skill` tool, prompt context sources, skill_ids propagation
- Frontend: SkillBrowser с import zip, `$skill` autocomplete, skill selectors, script-disabled badges

### Gaps

| # | Описание | Серьёзность |
|---|---|---|
| M3.1 | **Нет поддержки `source_type: local_directory` для импорта.** План (Skill Core, строка 5) говорит: «Skill import/install from uploaded zip archives and server-local directories.» HTTP API реализует только upload. | Low |
| M3.2 | **Roadmap нуждается в обновлении** — M3 всё ещё «Planned». | Low |

---

## M3.5: Elysium Skill Library Rewrite — ⚠️ PASS (3 gaps)

Библиотека **реализована**: 44 SKILL.md файла (18 default + 26 optional), манифест, script audit, README, archive notes.

### Что сделано хорошо

- **18 Default Skills** — все переписаны под Edda-native инструкции:
  - Полный frontmatter: `name`, `description`, `route` (actionKinds, contentKinds, tags, priority), `metadata` (status, source, scriptStatus)
  - Секции «Use when» / «Do not use when»
  - «Edda Workflow» и «Edda Output Handling»
  - «Script Compatibility» для skills со скриптами в исходниках
  - Ни один default skill не требует скриптов
- **26 Optional Skills** — аналогичное качество
- **`$children-stories`** — новый Edda-native skill с age bands
- **Манифест** (`manifest.md`): 93 строки, все 56 source SKILL.md учтены
- **Script audit** (`script-audit.md`): 145 строк, каждый скрипт классифицирован
- **`builtin/README.md`**: документирует политику default/optional/archive
- **8 archive notes** в `archive-notes/`

### Gaps

| # | Описание | Серьёзность |
|---|---|---|
| M3.5.1 | **Нет `reverse-outliner.md` в `archive-notes/`.** План Task 5 требует: «Defer `reverse-outliner` with a pointer to a future Edda-native analysis pipeline.» Манифест документирует его как deferred, но archive note не создан. | **High** |
| M3.5.2 | **`story-sense` и `genre-check` не скопировали свои `data/`.** Script audit (`script-audit.md:75-85`) пометил `story-sense/data/` (functions-forms.json, genre-elements.json) и `genre-conventions/data/genre-elements.json` как «likely to be copied». Фактически — только `SKILL.md` в каждой папке, без data-файлов. | Medium |
| M3.5.3 | **Ни один SKILL.md не ссылается на свои `data/` или `templates/`.** Агенты, запускающие skill, не получают structured pointer к bundled data. Инструкции говорят «built-in data» в Script Compatibility, но не указывают конкретные пути/файлы. | Medium |
| M3.5.4 | **Нет built-in-specific import fixtures.** План Task 7 требует фикстуры для representative Default/Optional skills, включая merged skill и script-bearing skill. Существующий `skill/http_test.go` тестирует только синтетический `style-pass`. | **High** |
| M3.5.5 | **`$skill-writer` использует `actionKind: skill_authoring`.** Такого значения нет в Agent Core (`chat`, `continuation`, `rewrite`, `read_check`). Может потребоваться runtime mapping. | Low |

---

## M3.6: Skill Script Runtime — 🔴 GAP (не начат)

Всё отсутствует:
- `skill/runtime/` package
- Миграция 00004
- Agent tool `skill_script`
- HTTP API для audits/approvals/runs
- Frontend admin controls
- `scriptRuntimeTypes.ts` / `scriptRuntimeApi.ts`

---

## M4: Daily Writing Polish — 🔴 GAP (не начат)

- Нет dedicated plan (roadmap: «Needs dedicated plan»)
- Galley Editor не подключен (контент — read-only textarea)
- Нет streaming (SSE)
- Нет mobile-friendly layout
- Нет diff/restore UI
- Нет model-switching UX polish

---

## ADR Cross-Reference

| ADR | Решение | Статус | Комментарий |
|---|---|---|---|
| 0001 | Database Source of Truth | ✅ | SQLite + Markdown-based content |
| 0002 | Direct Apply with Revisions | ✅ | preview/direct_apply + conflict detection |
| 0003 | Tool-Accessible Project Context | ✅ | 7 context tools |
| 0004 | Layered Writing Briefs | ✅ | project-wide → per-chapter order |
| 0005 | Skill Scripts Require Admin Approval | ⚠️ | Scripts disabled. Runtime — M3.6 |
| 0006 | OpenAI-Compatible Providers | ✅ | `/v1/chat/completions` + DeepSeek cache |
| 0007 | Galley Editor | ⚠️ | ADR принят, редактор не подключён (M4) |
| 0008 | Elysium Layout | ✅ | Import/export реализован |
| 0009 | Prompt Records for Debugging | ⚠️ | Реализовано, но нет redaction-функции (M2.2) |
| 0010 | React Frontend for Galley | ⚠️ | React есть, Galley Editor нет (M4) |
| 0011 | Go Backend | ✅ | Go 1.26 + chi |
| 0012 | Go Backend Baseline | ✅ | chi + sqlc + goose + SQLite |

---

## Все gaps одним списком

| ID | Milestone | Описание | Severity |
|---|---|---|---|
| M2.1 | Agent Core | `update_entry_section` без `expectedRevision` | **Critical** |
| M1.2 | Project Core | Auth placeholder (нет auth вообще) | High |
| M3.5.1 | Skill Library | Нет `reverse-outliner.md` archive note | High |
| M3.5.4 | Skill Library | Нет built-in-specific import fixtures | High |
| M1.1 | Project Core | Нет HTTP-рута `GET /projects/{id}/map` | Medium |
| M2.2 | Agent Core | Нет явной redaction API-ключей в prompt records | Medium |
| M2.3 | Agent Core | Нет insert toggle в Continuation UI | Medium |
| M2.4 | Agent Core | Нет selection-based Rewrite/Read&Check в UI | Medium |
| M3.5.2 | Skill Library | `story-sense` и `genre-check` — нет data/ | Medium |
| M3.5.3 | Skill Library | SKILL.md не ссылаются на свои data/templates | Medium |
| M3.1 | Skill Core | Нет `local_directory` импорта (только upload) | Low |
| M3.2 | Skill Core | Roadmap устарел (M3 всё ещё Planned) | Low |
| M3.5.5 | Skill Library | `$skill-writer` использует `skill_authoring` actionKind | Low |
| M3.6 | Script Runtime | Не начат | Blocking |
| M4 | Writing Polish | Не начат, нет dedicated plan | Blocking |

---

## Приоритетный порядок действий

1. **M2.1 (Critical fix):** Добавить `expectedRevision` в `update_entry_section` tool
2. **M3.2:** Обновить `roadmap.md` — M3 → Implemented
3. **M3.5.1:** Создать `archive-notes/reverse-outliner.md`
4. **M3.5.4:** Написать import fixtures для built-in skills
5. **M1.2:** Создать dedicated plan для Auth/Security
6. **M3.5.2:** Скопировать `data/` для `story-sense` и `genre-check`
7. **M2.2:** Добавить `redactAPIKey` в prompt record creation
8. **M4:** Создать dedicated plan для Daily Writing Polish
9. **M3.6:** Реализовать Skill Script Runtime
