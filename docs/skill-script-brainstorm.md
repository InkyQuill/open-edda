# Open Edda Skill Script & Prompt Orchestration Brainstorm

This document outlines the architectural brainstorm for exposing database context, executing helper scripts safely, and orchestrating prompt-based alternative workflows within Open Edda.

---

## 1. Exposing Database Info (Input Isolation)

To prevent scripts from directly querying the database (which could lead to SQL injection, privilege escalation, or cross-project data exposure), we enforce the **JSON Input Envelope** pattern.

1. **Pre-Fetched Envelopes:** The Go backend (via [Service.RunScript](skill/script_runtime.go#L122)) queries the SQLite store for the precise records requested (e.g., specific chapters, notes, or world-bible entries).
2. **Stdin Stream:** This data is marshaled into a standard [Envelope](skill/runtime/types.go#L521) JSON schema and passed to the script's `stdin`.
3. **No Direct Connections:** Scripts do not receive a database DSN, connection pool, or raw file path to the database. They operate purely on memory representations passed via `stdin`.

### Gap: Content body text is not included in the envelope

The current `EnvelopeInputs` struct carries `ContentIDs` and `EntrySections` as **references only** — just IDs and headings, no body text. This means a script like `rhythm.ts` has no way to read the chapter prose it needs to analyze. Similarly, `fate-pressure.ts` cannot read the current fate-tracking state stored in a story bible entry's body.

The `AssetInput` type does carry `BodyText`, but it is only populated for skill-internal data files (`references/`, `data/`), not for project content items.

**This is the single most important gap to close before any analytical script can run inside Edda.** Without body text in the envelope, every script that processes actual story content is dead on arrival.

#### Proposed solution: add `ContentItems` to `EnvelopeInputs`

```go
type EnvelopeInputs struct {
    ContentIDs    []string          `json:"contentIds,omitempty"`    // keep for backward compat
    ContentItems  []ContentInput    `json:"contentItems,omitempty"`  // NEW: full content
    EntrySections []EntrySectionRef `json:"entrySections,omitempty"`
    Assets        []AssetInput      `json:"assets,omitempty"`
    Arguments     map[string]any    `json:"arguments,omitempty"`
}

type ContentInput struct {
    ID              string `json:"id"`
    Kind            string `json:"kind"`
    Title           string `json:"title"`
    BodyMarkdown    string `json:"bodyMarkdown"`
    CurrentRevision int64  `json:"currentRevision"`
}
```

The Go backend would fetch each content item by ID (reusing the existing project-service access checks) and include its full body in the envelope. The `skill_script` tool arguments from the agent would specify `contentIds` (which content to load), and the backend resolves them into `ContentItems` before calling the runner.

**Considerations:**
- Body text can exceed 100k words. The envelope should enforce a configurable per-item byte limit and a total envelope size limit (e.g., 2 MB). If a content item exceeds the limit, include a truncation marker and the item's metadata so the script knows it got a partial view.
- `ContentIDs` should remain for backward compatibility with scripts that only need the ID list (e.g., to count items). New scripts should use `ContentItems`.
- The `buildScriptEnvelope` method in `script_runtime.go` currently validates that content IDs exist but doesn't fetch body text. This is where the change would land.

#### Alternative: separate `contentBody` flag in `skill_script` arguments

Instead of always including body text, the agent could pass `includeBody: true` in the `skill_script` arguments, and the backend only populates `ContentInput.BodyMarkdown` when that flag is set. This avoids bloating the envelope for scripts that don't need body text.

**My recommendation:** Always include body text for requested content items. The envelope is already ephemeral (stdin, not persisted), and the scripts that need it are the ones we're designing for. The extra bandwidth is negligible compared to the LLM API calls that triggered the script execution.

---

## 2. Execution and Output Control

To ensure scripts cannot damage the story or project repository:

1. **Ephemeral Working Directory:** Scripts run inside a temporary folder (e.g., `/tmp/open-edda-skill-script-*`) which is recursively deleted immediately upon execution (see [Runner.Run](skill/runtime/runner.go#L45-L50)).
2. **Proposal-Only Outputs:** Scripts return a JSON envelope on `stdout`. The backend validates the payload against a strict schema (e.g., `ScriptOutput` in [runner.go](skill/runtime/runner.go#L117-L135)).
3. **Review Queues:** Scripts are strictly forbidden from directly updating database records. All modifications are parsed as **Proposals** (e.g., `proposals` array) and sent to a review queue where the author must explicitly accept or reject them.

### Gap: Proposal format is too narrow for generated-data scripts

The current `ScriptOutput` requires:
- `kind`: one of `report`, `proposal`, `draft`, `generated_data`
- `title`: required non-empty string
- At least one of: `markdown` (prose), `proposals` (array of `Proposal` structs), or `generatedData` (freeform `map[string]any`)

For `rhythm.ts`, the output is a structured statistical report — sentence lengths, standard deviations, variety scores. This maps cleanly to `kind: "report"` with the analysis in `markdown`. No problem here.

For `words.ts`, the output is 50 generated words with their syllable structure. This maps to `kind: "generated_data"` with the word list in `generatedData`. This also works, though the agent needs guidance on what to do with the output (create a project note? add to a story bible entry?).

For `fate-roll.ts`, the output is a roll result (numeric values, severity, death eligibility). This is a `report` with structured data. The agent then decides how to narrate it.

**The current output contract is adequate but not prescriptive enough.** Scripts that generate content (words, names, simulation results) should include a `suggestedDestination` field telling the agent where the output belongs — a project note, a story bible entry, a chat message — so the agent doesn't have to guess.

### Gap: No way for scripts to signal "I need more context"

A script receiving a content ID might determine that it also needs related content (e.g., `rhythm.ts` analyzing one chapter might want to compare against the project's average). The current design gives the script exactly what the agent asked for, nothing more. If the agent didn't request the right content items, the script produces an incomplete result.

**This is acceptable for v1.** The agent is responsible for figuring out what the script needs and requesting it. Future versions could add a "script hints" mechanism where scripts declare their required inputs in metadata, and the backend validates completeness before execution.

---

## 3. Sandboxing & Isolation Strategies

To protect the host system from untrusted script execution:

* **Deno Permission Sandbox (Recommended):** Since **96% of the original scripts** are written in Deno TypeScript, we can execute Deno with strict permission flags:

  ```bash
  deno run --no-net --allow-read=/tmp/workdir --allow-write=/tmp/workdir script.ts
  ```

  This isolates filesystem access and blocks network socket creation natively.
* **WebAssembly (WASM) Isolation:** Compile script runtimes into WebAssembly modules and run them via a WASI-compliant engine (like Wasmtime), preventing host system access entirely.
* **Linux Namespaces:** Spawn commands using `unshare` to restrict PID, mount, and network namespaces.
* **Prompt-Based Translation:** Eliminate code execution by translating script heuristics directly into LLM prompts.

### Consideration: Deno sandboxing requires Deno on the host

The current `Runner` uses `sh -c` to execute whatever `RuntimeCommand` is stored in the approval. If we want Deno-specific sandboxing, the `RuntimeCommand` for Deno scripts must be the full `deno run` invocation with permission flags, not just a bare script path.

This means the approval system needs to store the **complete command** including Deno flags. The admin approving the script would see something like:

```bash
deno run --no-net --allow-read --allow-write=/tmp script.ts
```

And the audit system would verify that the flags are restrictive enough. This is already compatible with the current `RuntimeCommand` field — we just need the approval UX to guide admins toward safe Deno invocations.

**Problem:** What if Deno is not installed? The script runner doesn't check for runtime availability before execution. A failed `deno: not found` would produce a `StatusFailed` result, which is correct but unhelpful. Consider adding a runtime-availability check during approval that verifies the command's base interpreter exists on the host.

### Consideration: WASM is a v2+ goal

WASM gives the strongest isolation but requires compiling every script to WASM, which adds build complexity and limits what scripts can do (no filesystem access at all, no subprocess spawning). For v1, Deno sandboxing or plain `sh -c` with stripped environment is sufficient. WASM becomes interesting when we want to run untrusted community-contributed scripts without admin pre-approval.

### Consideration: Linux namespaces add host-specific complexity

`unshare` is Linux-only. The project already has platform-specific runner files (`runner_unix.go`, `runner_windows.go`). Adding namespace isolation would require another platform split and significantly more complex process management. Not worth it for a single-author self-hosted tool.

---

## 4. Scripts that Cannot be Converted to LLM Prompts

While many diagnostic scripts can be converted to prompt instructions, several are better run as code:

| Script / Pipeline | Original Path | Why it Cannot be Prompt-Based |
| --- | --- | --- |
| **Graphviz Rendering** | [render-graphs.js](docs/skills/important/writing-skills/render-graphs.js) | Requires executing the `dot` CLI compiler to output physical SVG files. |
| **Phonology & Combinatorial Lexicons** | `conlang/scripts/words.ts` | Generating thousands of unique words conforming to strict mathematical syllable constraints (e.g., `(C)V(C)`) is slow, expensive, and error-prone for LLMs. |
| **Statistical Prose Metrics** | `craft/prose-style/scripts/rhythm.ts` | Exact calculations of sentence length variance and standard deviation are highly inaccurate in LLMs due to tokenization. Heuristics are better computed in code. |
| **Batch Document Pipelines** | `structure/reverse-outliner/scripts/reverse-outline.ts` | Sequencing multi-step analysis across an entire 100k-word manuscript exceeds LLM context windows. It requires scripts to batch, save intermediate JSON, and coordinate API calls. |
| **Fates Simulation Engine** | `worldbuilding/world-fates/scripts/fate-roll.ts` | Stateful tracking of numeric variables and rolling true random numbers cannot be maintained reliably by LLMs across long sessions. |

### Updated classification: three categories, not two

The brainstorm currently treats scripts as either "convert to prompts" or "keep as code." But the user's requirement is more nuanced. There are **three** categories:

#### Category A: Remove — scripts with no place in Edda

These scripts serve development or authoring workflows that don't map to the Edda author experience. They should not be converted, retained, or deferred — they should simply be dropped during skill staging.

| Script | Reason for removal |
| --- | --- |
| `render-graphs.js` | Requires Graphviz CLI, outputs SVG files to disk. Edda has no image rendering surface in v1. This is an authoring-toolchain helper, not a writing assistant. The dot blocks in SKILL.md files can be preserved as readable flow diagrams in `references/` (text, not rendered images). |

**Staging skill instruction:** When the staging skill or `skill-writer` encounters a script that shells out to external CLI tools to produce non-text artifacts (images, PDFs, binaries), the script should be classified as `removed`. The source skill should either drop the script entirely or convert the input data (e.g., dot source) into a `references/` file the agent can read as context.

#### Category B: Convert to prompt instructions — scripts whose logic is better as LLM guidance

Many "audit" scripts are essentially checklist runners: they parse text, apply heuristics, and produce a report. The heuristics are often judgment calls that an LLM can apply more flexibly than rigid code. The script's value is in its **rubric**, not its execution.

| Script | Prompt conversion approach |
| --- | --- |
| `prose-check.ts` | Convert the audit checklist into a `## Prose Audit Checklist` section in the skill body. The agent follows the checklist using `read_chapter` and produces a report. |
| `dialogue-audit.ts` | Convert the function/anti-pattern list into a `references/dialogue-audit-rubric.md`. The agent loads it via `read_skill_file` and applies it. |
| `voice-check.ts` | Convert the voice differentiation criteria into structured instructions. The agent compares characters' dialogue. |
| `genre-check.ts`, `genre-blend.ts` | Convert into guided analysis instructions backed by `data/genre-elements.json`. The agent uses the data file for genre-specific checks. |
| `ending-check.ts`, `setup-payoff.ts` | Convert into evaluation rubrics the agent applies to selected chapters. |
| `scene-sequencing/analyze-scene.ts` | Convert into a scene review guide. |
| `revision-audit.ts` | Convert into a revision-pass checklist. |
| `cliche-transcendence/orthogonality-check.ts` | Convert into guided questioning instructions. |
| `story-sense/entropy.ts` | Convert the entropy diagnostic criteria into an LLM-analyzable checklist. |
| `story-sense/functions.ts` | Convert into a lookup guide backed by `data/functions-forms.json`. |

**Key principle:** If the script's logic is a sequence of pattern matches, counts, or rules that an LLM can apply at least as well as code (and often better, because LLMs understand context), convert it to instructions. If the script performs exact numerical computation that LLMs get wrong, keep it as code.

#### Category C: Isolate as Edda skill scripts — scripts that must run as code

These scripts perform computation that LLMs cannot reliably replicate: exact math, random number generation, combinatorial generation, or multi-step orchestration over large inputs. They should become proper Edda skill scripts that:

1. Receive project content through the `Envelope` stdin (once the content-body gap from Section 1 is closed).
2. Return results through the `ScriptOutput` contract.
3. Never directly mutate the database.
4. Run in the existing sandboxed `Runner` with admin-approved commands.

| Script | What it computes | Edda integration |
| --- | --- | --- |
| `rhythm.ts` | Sentence length distribution, standard deviation, variety scores, opening-word repetition | Agent calls `skill_script` with `contentIds` referencing the target chapter. Script receives chapter body in `ContentItems`. Returns `kind: "report"` with rhythm metrics and verdict. Agent uses the report to coach the author. |
| `conlang/words.ts` | Combinatorial word generation from phoneme inventories and syllable templates | Agent calls `skill_script` with `arguments` specifying phoneme inventory, count, seed. Script reads `data/phoneme-frequencies.json` and `data/syllable-templates.json` as assets. Returns `kind: "generated_data"` with word list. Agent proposes adding results to a story bible entry or project note. |
| `conlang/phonology.ts` | Phonology generation from frequency distributions | Similar to `words.ts`. Returns generated phoneme inventory as `generated_data`. |
| `fate-roll.ts` | Weighted random fate outcomes based on pressure/danger/fortune | Agent calls `skill_script` with `arguments` for pressure, danger, fortune. Script rolls random numbers. Returns `kind: "report"` with roll results. Agent narrates the outcome and proposes story bible updates. |
| `fate-pressure.ts` | Pressure scoring based on accumulated exposure events | Agent calls `skill_script` with `contentIds` referencing fate-tracking entries. Script reads entries from `ContentItems`, computes pressure. Returns report. |
| `fate-choice.ts` | Choice generation from shift-type data | Agent calls `skill_script` with `arguments` for severity and context. Returns generated choices as `generated_data`. |
| `exposure-log.ts` | Exposure report compilation | Agent calls `skill_script` with `contentIds` referencing relevant story bible entries. Returns structured exposure report. |
| `propose-shift.ts` | Fate-shift proposal generation | Agent calls `skill_script` with roll results and context in `arguments`. Returns `kind: "proposal"` with suggested story bible changes. |
| `reverse-outliner/*.ts` | Multi-step batch analysis pipeline over entire manuscripts | **Deferred.** This requires a pipeline orchestration model that the current single-invocation script runner doesn't support. The script needs to process content in batches, accumulate intermediate results, and coordinate across multiple invocations. This needs a "job" abstraction that doesn't exist yet. |
| `story-zoom/*.ts` | Long-running watcher/dashboard | **Deferred.** Requires event subscription and persistent state, not a request-response script execution. |

### Problem: Reverse-outliner needs multi-step orchestration

The `reverse-outliner` pipeline runs 5+ scripts in sequence, with each consuming the previous one's output. The current `skill_script` tool is a single invocation. To support this, we would need either:

1. **A pipeline definition format** in the skill that tells Edda "run these scripts in order, passing outputs as inputs to the next." This is a non-trivial runtime addition.
2. **A single orchestration script** that calls the other scripts internally. But our sandbox prevents subprocess spawning, so this won't work with `--no-net` and restricted filesystem.
3. **Agent-driven orchestration** where the agent calls `skill_script` multiple times across tool rounds, passing intermediate results in `arguments`. This works with the current system but requires the agent to understand the pipeline, which is fragile.

**My recommendation:** Defer `reverse-outliner` and `story-zoom` until we have a job/pipeline abstraction. For now, the agent can approximate a reverse outline by reading chapters and producing outline notes through prompt-based analysis — it just won't be as thorough as the scripted pipeline over a full manuscript.

---

## 5. Orchestrating Prompt-Based Subagents (No-Code Alternatives)

For scripts converted into LLM workflows, the agent system must handle coordination, parallel execution, and resilience.

### A. Scoped Parallel Subagents
1. **Scoped Contexts:** The parent agent spans multiple concurrent subagents (using `invoke_subagent` or a similar multi-agent orchestrator). Each subagent is restricted to a narrow scope (e.g., reviewing dialogue in one chapter using a specific rubric file).
2. **Tool Sandboxing:** Subagents are provisioned with read-only tools (`read_chapter`, `read_story_bible_entry`) to prevent race conditions or conflicting writes.
3. **Execution Wait Loop:** Like CLI subagents, the main process halts active tool calls and yields control, waking up reactively once all child subagents post completion logs or message tokens.
4. **Hang and Timeout Management:** A watchdog timer monitors active subagents. If a subagent hangs, enters an infinite loop, or exceeds time/token limits, the parent issues a cancel signal (similar to `manage_subagents` with action `kill`) to prune the process.

### Consideration: Subagents don't exist yet in Edda

The current agent system has a single agent per session with tool calls. There is no `invoke_subagent` tool, no parallel execution, no subagent scoping. This entire section is forward-looking design for a future capability.

**For v1, the simpler model is sufficient:** the agent processes one task at a time, calling tools sequentially (or in limited tool-call rounds). Prompt-converted skills work within this model — the agent follows the rubric, calls `read_chapter` for each chapter it needs to check, and produces a report. No parallelism needed.

**Subagents become important when:** an author asks "check the dialogue consistency across all 30 chapters" and the agent needs to process them in parallel to finish in reasonable time. This is a v2+ concern.

### B. Skill-to-Prompt Translation Workflow
We must ensure our staging and built-in skill authoring tools translate procedural scripts into declarative LLM instructions:

* **Staging Skill Authoring (`open-edda-skill-staging`):**
  * When auditing staging folders (using `/home/inky/Development/writer/.agents/skills/open-edda-skill-staging`), any reference to `skill_script` tools must be scrutinized. 
  * The staging tool enforces that procedural code (e.g., a dialogue passive-voice audit) is translated into a detailed markdown guide in `references/` or `SKILL.md` (e.g., an exact voice checklist).
* **Native Skill Writer (`skill-writer`):**
  * The built-in authoring skill `skill-writer` (located in `/home/inky/Development/writer/docs/skills/builtin/default/skill-writer`) must coach authors on translating coding logic to prompt logic.
  * It provides templates for **Evaluation Rubrics** and **Structured Response Formats** instead of script templates, helping creators specify *how* the model should analyze text rather than writing custom JS/Python scripts to parse it.

### Updated skill-authoring instructions: three-category decision tree

Both `open-edda-skill-staging` and `skill-writer` need explicit rules for handling each script category:

#### Rule 1: Remove external-tool-dependent scripts

When encountering a script that:
- Invokes an external CLI tool not bundled with Edda (e.g., `dot`, `ffmpeg`, `inkscape`)
- Produces non-text output formats (SVG, PNG, PDF)
- Requires a graphical environment or display server

**Action:** Classify as `removed`. Do not convert, retain, or defer. If the source data fed to the external tool is valuable (e.g., Graphviz dot source), preserve it in `references/` as a readable text artifact. Add a note in `## Script Compatibility` explaining what was removed and why.

**Example:** `render-graphs.js` → remove the script, keep `.dot` blocks in `references/flow-diagrams.dot` as context the agent can read.

#### Rule 2: Convert heuristic/audit scripts to prompt instructions

When encountering a script that:
- Applies pattern-matching rules, checklists, or scoring heuristics to text
- Produces a text report or structured findings
- Does not perform exact numerical computation where precision matters
- Does not generate combinatorial output beyond what an LLM can produce

**Action:** Classify as `converted-to-reference` or `converted-to-data-template`. Extract the script's rules into the skill body or a reference file. The agent applies the rules using Edda context tools instead of running code.

**Conversion checklist:**
1. List every rule, pattern, threshold, and category the script checks.
2. Preserve exact thresholds where they matter (e.g., "sentence > 35 words = very long").
3. Preserve the scoring/weighting formula if applicable (e.g., "sentence variety = 40%, paragraph variety = 30%, opening variety = 30%").
4. Turn each check into an explicit instruction: "Count the number of sentences starting with each word. Flag any word that starts more than 15% of sentences or appears 4+ times."
5. Add the converted rubric to `SKILL.md` or `references/`.
6. In `## Script Compatibility`, note: "This skill was converted from a Deno script. The agent now applies the same rubric through Edda context tools."

#### Rule 3: Isolate computation scripts as approved helpers

When encountering a script that:
- Performs exact numerical computation (statistics, probability, random generation)
- Generates combinatorial output (word lists, name pools, simulation results)
- Processes large inputs that exceed practical LLM context windows
- Requires deterministic, reproducible results with seeded randomness

**Action:** Classify as `retained-for-runtime`. Adapt the script to consume the Edda `Envelope` on stdin and produce `ScriptOutput` on stdout. The script must:
1. Read all input from stdin (JSON `Envelope`), not from CLI args or local files.
2. Output a valid `ScriptOutput` JSON to stdout.
3. Not write to the filesystem, open network connections, or spawn subprocesses.
4. Include all required data (phoneme tables, syllable templates) as skill assets loaded through `assets` in the `skill_script` arguments, or embed them directly in the script.

**Adaptation pattern for existing Deno scripts:**

Most existing scripts read from CLI args (`--pressure 0.65`) or local files (`Deno.readTextFile`). They need to be adapted to:

```typescript
// Before (CLI-oriented):
const args = Deno.args;
const pressure = parseFloat(args[args.indexOf("--pressure") + 1]);

// After (Envelope-oriented):
const input = JSON.parse(await new Response(Deno.stdin.readable).text());
const pressure = input.arguments?.pressure ?? 0.5;
const chapterText = input.contentItems?.[0]?.bodyMarkdown ?? "";
```

**This is the main conversion work for Category C scripts.** Each script needs a thin adapter layer that maps `Envelope` fields to the script's internal parameters. The core computation logic can remain unchanged.

---

## 6. Implementation Roadmap

### Phase 1: Close the content-body gap (prerequisite for everything) — DONE

Without body text in the envelope, no analytical script can run. This phase is now complete.

1. ~~Add `ContentInput` struct to `skill/runtime/types.go`.~~ ✅
2. ~~Extend `EnvelopeInputs` with `ContentItems []ContentInput`.~~ ✅
3. ~~Update `buildScriptEnvelope` in `skill/script_runtime.go` to fetch content item bodies when `contentIds` are specified in the `skill_script` arguments.~~ ✅
4. ~~Add envelope size limits (per-item and total) to prevent oversized stdin payloads.~~ ✅ (`maxContentBodyBytes = 512 KB`, `maxEnvelopeBodyBytes = 2 MB`)
5. ~~Update the `skill_script` tool definition in `agent/tools.go` to pass `contentIds` through to the envelope builder.~~ Already worked; `contentIds` was already passed through.
6. ~~Add `BodyMarkdown` to `EntrySectionRef` and populate it in `buildScriptEnvelope`.~~ ✅

**Implementation notes:**
- `ContentInput` includes `ID`, `Kind`, `Title`, `BodyMarkdown`, `MetadataJSON`, `CurrentRevision`, and `Truncated` (set when body exceeds per-item limit).
- `EntrySectionRef` now includes `BodyMarkdown` populated from the entry section record.
- Per-item truncation appends `contentTruncationMarker` and sets `Truncated: true` so scripts know they got a partial view.
- Total envelope body limit is enforced across all content items; exceeding it returns `ErrInvalidInput`.
- `validateEntrySection` now returns `(string, error)` with the section body instead of just `error`.

### Phase 2: Adapt Category C scripts to the envelope contract

For each retained script:

1. Add an envelope-reading preamble that parses `Envelope` from stdin.
2. Map envelope fields to the script's internal parameters.
3. Replace `console.log(formatReport(...))` output with `console.log(JSON.stringify(scriptOutput))` producing valid `ScriptOutput`.
4. Bundle required data files as skill assets (referenced in `skill_script` arguments via `assetPaths`).
5. Remove CLI arg parsing, file reads, and any local filesystem access.
6. Test with a mock envelope JSON piped to stdin.

**Scripts to adapt (priority order):**
1. `rhythm.ts` — simplest, single chapter input, report output
2. `fate-roll.ts` — no content input needed, just arguments
3. `words.ts` — needs bundled data assets, generated_data output
4. `phonology.ts` — similar to words.ts
5. Remaining `world-fates` scripts — depend on story bible entries as content input

### Phase 3: Update skill-authoring skills

1. Update `skill-writer/SKILL.md` with the three-category decision tree (Remove / Convert / Isolate) as explicit instructions.
2. Update `skill-writer/references/external-skill-conversion.md` with Category C adaptation patterns.
3. Update `open-edda-skill-staging` skill to enforce the three-category classification during audit.
4. Add a `## Script Classification Guide` section to both skills that maps common script patterns to their category.

### Phase 4: Approval and runtime UX

1. Extend `SkillScriptAudit` to include a `scriptCategory` field (`removed`, `converted`, `isolated`).
2. For `isolated` scripts, the approval flow should verify:
   - The script reads only from stdin.
   - The script outputs valid `ScriptOutput` JSON.
   - The `RuntimeCommand` includes appropriate sandboxing flags (Deno `--no-net`, etc.).
   - The script does not spawn subprocesses or access the network.
3. For the admin approval UI, show the script's category, its envelope expectations, and a sample input/output pair for manual verification.

---

## 7. Open Questions and Risks

### Q1: Should `ContentItems` in the envelope include metadata JSON?

Some scripts (e.g., `fate-pressure.ts`) might need to read structured metadata from story bible entries (e.g., `{"fatePressure": 0.65}`). If `ContentInput` includes `MetadataJSON`, scripts can parse it. If not, they have to parse it out of `BodyMarkdown`, which is fragile.

**My inclination:** Include `MetadataJSON` in `ContentInput`. It's already stored as a separate field in the database and is typically small.

### Q2: How do scripts access their own bundled data files?

Currently, `AssetInput` in the envelope carries skill data file contents. The `buildScriptEnvelope` method loads them when `assetPaths` are specified. This works for small JSON data files (phoneme tables, syllable templates). But some scripts have large data sets (character naming cultures: 30+ JSON files).

**Options:**
1. Load all requested assets into the envelope (current approach). Large data sets bloat the envelope.
2. Write requested assets to the script's temp working directory and let the script read them from disk. This gives the script `--allow-read` access to its own temp dir.
3. Merge related small data files into a single JSON asset during skill staging.

**My recommendation:** Option 2 for v1. The runner already creates a temp working directory. Writing asset files there before execution is simple and keeps the envelope lean. The `RuntimeCommand` for Deno scripts would include `--allow-read=<workdir>`.

### Q3: What about the `reverse-outliner` pipeline?

This is the hardest case. It requires multi-step orchestration with intermediate state. The current single-invocation model cannot handle it.

**Options:**
1. **Agent-driven orchestration:** The agent calls each step as a separate `skill_script` invocation, passing intermediate results in `arguments`. Fragile but works with the current system.
2. **Pipeline definition:** Add a `pipelines/` section to skills that defines multi-step execution plans. The runtime executes the pipeline, managing intermediate state. Significant new feature.
3. **Single monolithic script:** Combine all pipeline steps into one large script. This works but defeats the purpose of modular pipeline steps and makes debugging harder.
4. **Keep deferred:** Wait for a future job/pipeline abstraction.

**My recommendation:** Keep deferred for now. The agent can approximate reverse outlining through prompt-based analysis of individual chapters. When we build a proper pipeline/job system (which we'll need for other reasons too), the reverse-outliner scripts become the first consumer.

### Q4: Seeded randomness in scripts vs. reproducibility

`fate-roll.ts` and `words.ts` both use seeded PRNGs for reproducibility. When running through Edda, the seed needs to come from somewhere. Options:
1. Agent passes `seed` in `arguments`. Author can re-run with the same seed for reproducible results.
2. Backend generates a seed automatically if none is provided. Store the seed in the `SkillScriptRun` record for audit/replay.
3. Both: default to auto-generated, allow override.

**My recommendation:** Option 3. The `ScriptRun` record already persists `InputJSON`, which would include the seed. This gives full reproducibility.

### Q5: Error handling when scripts produce invalid output

The current runner handles this (`StatusRejected` for invalid JSON or failed validation). But what should the agent do when a script it called returns `rejected`? Currently it gets an error message. Should it retry? Fall back to prompt-based analysis?

**My recommendation:** The agent should treat `rejected` as a final failure for that tool call, report the error to the author, and suggest alternatives (e.g., "The prose rhythm script returned invalid output. I can analyze the rhythm manually instead."). Do not auto-retry — if the script failed once, it will likely fail again with the same input.

### Risk: Script execution latency

Each `skill_script` call spawns a new process (Deno or Node startup time), reads stdin, computes, writes stdout. For simple scripts like `fate-roll.ts`, this is fast (<1s). For `rhythm.ts` on a long chapter, it could take several seconds. For `reverse-outliner` on a full manuscript, minutes.

The current agent tool-call loop has `maxToolRounds = 4`. If a script takes 30 seconds, that's 30 seconds of blocked agent session. For v1 this is acceptable. For v2, consider async script execution with a polling pattern.

### Risk: Script versioning and compatibility

When the `Envelope` schema changes (e.g., adding `ContentItems`), existing scripts that only read `ContentIDs` must continue to work. The envelope should always include both fields during a transition period. Scripts should ignore unknown fields (standard JSON decoding behavior).

### Risk: Deno remote imports in retained scripts

Several scripts import from `https://deno.land/std@0.208.0/path/mod.ts`. In a sandboxed environment with `--no-net`, these imports will fail at runtime. Scripts must either:
1. Vendor their dependencies (include them as skill assets).
2. Use only Deno's built-in APIs.
3. Run without `--no-net` but with `--allow-net=deno.land` (less secure).

**My recommendation:** Scripts must be self-contained. Vendor any Deno std imports or rewrite to use only built-in APIs. The `--no-net` flag is a critical security boundary and should not be relaxed for convenience.
