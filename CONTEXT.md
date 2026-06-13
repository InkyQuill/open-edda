# Writer

Writer is a private writing workspace for hobby novelists who want AI assistance over an entire story project, not only the currently selected text.

## Language

**AI Writing Studio**:
A private workspace where an author develops story text, story-world material, and AI-assisted revisions as one connected project.
_Avoid_: Generic writing app, AI text editor

**Story Project**:
The complete body of material for one fiction work or series, including draft prose, story bible material, notes, and collaboration history.
_Avoid_: File folder, document, workspace

**Project Dashboard**:
The author-facing entry point for choosing, creating, importing, exporting, and configuring story projects.
_Avoid_: Home page, admin panel

**Writing Workspace**:
The project-level surface where the author edits chapters, manages the story bible, reviews notes and reports, works with skills, and runs agent sessions.
_Avoid_: Editor page, project page

**Story Bible**:
The durable reference material that defines a story project's characters, worldbuilding, continuity, style constraints, and other canon-relevant context.
_Avoid_: Lore dump, notes database, worldbuilding folder

**Story Text**:
Ordered prose intended to become part of the manuscript.
_Avoid_: Draft file, chapter file

**Chapter**:
The primary story text unit in the first version, edited as a continuous document with cursor and selection-aware agent actions.
_Avoid_: Scene, manuscript file

**Story Bible Entry**:
A focused piece of durable reference material inside the story bible, such as a character, location, faction, world system, continuity fact, or style constraint.
_Avoid_: Lore page, wiki article

**Entry Relation**:
A connection from one story bible entry to another that helps authors and agents navigate related canon, optionally labeled with the nature of the relationship.
_Avoid_: Backlink, wiki link, unstructured reference

**Entry Section**:
A freeform named section within a story bible entry that can be retrieved independently when an agent needs focused context.
_Avoid_: Field, property, subpage

**Project Note**:
Exploratory or temporary material that can inform the author and agent but is not treated as canon.
_Avoid_: Brain dump, scratch file

**Attached Note**:
A note or agent report linked to a chapter or text selection without becoming story text, story bible material, or a project note by default.
_Avoid_: Inline comment, canon note, review file

**Writing Brief**:
The durable instructions that describe what the agent is writing and how it should write it, including genre, tense, point of view, style, and project-level prose constraints.
_Avoid_: Prompt prefix, generation settings, author's note

**Layered Writing Brief**:
A writing brief model where story project defaults can be refined or overridden for a specific story text unit.
_Avoid_: Prompt stack, local settings

**Project Source of Truth**:
The authoritative version of a story project's text, reference material, versions, and collaboration history.
_Avoid_: File source, sync folder

**Markdown Export**:
A portable representation of a story project for local editing, backup, and use with external text-based agent tools.
_Avoid_: Canonical files, filesystem database

**Elysium Layout**:
The initial Markdown folder convention for importing and exporting story projects, with folders such as `story/`, `characters/`, `worldbuilding/`, and `braindump/`.
_Avoid_: Arbitrary layout, sync profile

**Markdown Import**:
A conservative process that creates a story project from an Elysium Layout folder while preserving Markdown content, metadata, entry sections, relations, and chapters.
_Avoid_: Live sync, merge import

**Local Sync Tool**:
A future command-line workflow that detects local Markdown changes and replays them into the service database with merge handling.
_Avoid_: V1 import, filesystem source of truth

**Markdown-Based Content**:
Story project content stored in the database as Markdown-compatible text so it can be edited in a Markdown editor and exported cleanly to Markdown files.
_Avoid_: Rich text document, proprietary document model

**Galley Editor**:
The Markdown-native editor foundation intended for editing chapters and other Markdown-based content in Writer.
_Avoid_: Custom editor, generic textarea

**React Frontend**:
The browser application layer used for Writer because Galley Editor is React-based.
_Avoid_: Framework-neutral frontend, Svelte frontend

**Go Backend**:
The service layer used for Writer's API, persistence, agent tooling, provider calls, import/export, and self-hosted deployment.
_Avoid_: Python backend, Next.js backend

**SQLite Store**:
The first-version relational database for story projects, revisions, story bible metadata, agent sessions, and settings in a single-author self-hosted deployment.
_Avoid_: Postgres-first database, file storage

**Mobile-Friendly Web**:
A responsive browser experience that supports reading, chat, small edits, and agent actions on phones without requiring a native mobile app.
_Avoid_: Android app, mobile-first editor

**Agent Session**:
A bounded conversation or task run where an AI assistant uses story project context and approved tools to help the author think, revise, analyze, or generate.
_Avoid_: Chat, prompt, run

**Chat Context Reset**:
An author action that starts a fresh model context for an agent session while preserving the prior conversation and outputs as history.
_Avoid_: Delete chat, clear history, forget project

**Activity Trail**:
A readable record of agent actions, tool use, model variants, context reads, and content changes that is compact by default and expandable for inspection.
_Avoid_: Raw prompt log, audit dump

**Prompt Record**:
An advanced debugging record of assembled model input and model output for an agent session, subject to retention controls.
_Avoid_: Activity trail, public transcript

**Skill**:
A project-installable agent procedure that can include instructions, routing metadata, templates, data files, and optional executable helpers.
_Avoid_: Prompt snippet, preset, macro

**Admin-Approved Script**:
A skill helper script that a server administrator has explicitly allowed to run in the web service environment.
_Avoid_: Bundled script, trusted script

**OpenAI-Compatible Provider**:
Any model service or gateway that exposes an OpenAI-style API surface usable by Writer for agent sessions and generation actions.
_Avoid_: Vendor integration, model backend

**Provider Disclosure**:
The story project content and instructions sent to a configured model provider for a specific agent session or generation action.
_Avoid_: Data leak, sync, upload

**Model Variant**:
A configured model option within the same provider setup that the author can switch between for speed, quality, cost, or task fit.
_Avoid_: Backend, preset

**Whole-Project Context**:
An agent's ability to discover, inspect, and use the relevant parts of a story project through project maps, retrieval, summaries, and read tools.
_Avoid_: Full prompt dump, whole context window

**Versioned Change**:
A proposed or applied modification to story text or story bible material that preserves the previous state and can be reviewed through a diff.
_Avoid_: Edit, save, overwrite

**Per-Item Revision**:
A saved prior state of one chapter, story bible entry, writing brief, or project note that can be compared, labeled, and restored independently.
_Avoid_: Project snapshot, branch, checkpoint

**Direct Apply**:
An agent write mode where changes are applied immediately while preserving revisions so the author can inspect, revert, or continue editing afterward.
_Avoid_: YOLO mode, unsafe edit

**Continuation**:
An agent action that generates new story text at the current insertion point or at the end of a story text unit, optionally guided by the author's instruction and bounded by a requested length.
_Avoid_: Autocomplete, continue button

**Generation Buffer**:
A transient UI buffer that can display streamed generated text before it is committed to Markdown-based content as a complete change.
_Avoid_: Partial document write, streaming save

**Structured Write**:
An agent tool operation that changes Markdown-based content through a bounded action such as insert, append, replace range, or update section.
_Avoid_: Raw patch, arbitrary diff, token write

**Rewrite**:
An agent action that transforms selected story text according to the author's guidance while preserving the original through revisions.
_Avoid_: Regenerate, paraphrase

**Read and Check**:
An agent action that evaluates selected story text or a larger story text unit and returns a report with observations and suggestions rather than directly changing prose.
_Avoid_: Review, critique, analysis

**Author**:
The human owner and creative authority for a story project.
_Avoid_: User, customer, writer

**Single-Author Instance**:
A deployment where one author owns and controls the story projects, even if the product leaves room for future collaboration.
_Avoid_: Solo mode, personal account
