# Galley Editor for Markdown Editing

Open Edda uses Galley Editor as the intended Markdown-native editing surface for chapters and other Markdown-based content. It is already usable enough to anchor the first version, and the service can improve its polish and integration while building cursor-aware continuation, selection-aware rewrite, read-and-check, diffs, saves, and checkpoint workflows.

Galley Editor is at https://github.com/InkyQuill/galley-editor and is published to npm as `@inkyquill/galley-editor`.

If the package stops exposing the editor hooks Open Edda needs for cursor, selection, and context actions, Open Edda may vendor or fork the editor code instead of treating the external dependency as mandatory.
