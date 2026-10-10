# Pocket/Galley interoperability fixtures

`pocket-editor.review.json` is the real Android review fixture from
`InkyQuill/pocket-editor`, commit `2ac9a6ba809647fbe2578c9c9142f611106bbc2f`,
`app/src/test/resources/fixtures/review-v1.json`. Its synthetic hashes deliberately
resolve as stale; importing it must not discard its records.

`chapter.md`, `chapter.review.json`, `manifest.json` and `applied.*` were generated
with the shared codecs and review engine in `InkyQuill/galley-desk`, commit
`5de5ae75cd6bac8ab3a6a853fc8fa8b3a5d59e31`:

- `@galley-desk/pocket-format`: `makeAnchor`, `encodeReview`, `decodeReview`;
- `@galley-desk/review-engine`: `openChapter`, `applyEdit`.

The source includes Cyrillic, emoji and exact CRLF bytes. The first edit grows a
word; the second edit follows it. One signal spans the applied edit and another
is wholly inside it. `applied.*` is the independent Galley result for accepting
`a65eef9e-7318-4777-b2a2-c58a169bfcf6`.

Go integration tests compare Edda's actual saved chapter bytes and decoded JSON
against those results. They also import two folders with the same book/chapter
UUIDs, preserve unknown nested metadata, and replay a lost response after a newer
version. Browser tests use these same input files through real authenticated HTTP.
These checks are codec/server/browser checks; they do not claim Android-device
or packaged Galley UI qualification.
