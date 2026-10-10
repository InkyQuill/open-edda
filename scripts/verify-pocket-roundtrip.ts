/** Optional cross-check against a local Galley Desk checkout; no production dependency.
 * bun scripts/verify-pocket-roundtrip.ts GALLEY_DESK_DIR CHAPTER.md REVIEW.json [MANIFEST.json]
 */
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';
import { strict as assert } from 'node:assert';

const [galley, chapterPath, reviewPath, manifestPath] = process.argv.slice(2);
if (!galley || !chapterPath || !reviewPath) throw new Error('Expected GALLEY_DESK_DIR CHAPTER.md REVIEW.json');
const { decodeReview, encodeReview, classifyReview, decodeManifest } = await import(pathToFileURL(resolve(galley, 'packages/pocket-format/src/index.ts')).href);
const { openChapter } = await import(pathToFileURL(resolve(galley, 'packages/review-engine/src/index.ts')).href);
const wire = readFileSync(reviewPath, 'utf8');
const identity = JSON.parse(wire) as { chapter_id: string; source_path: string };
const review = decodeReview(wire, identity.chapter_id, identity.source_path);
assert.deepEqual(decodeReview(encodeReview(review), identity.chapter_id, identity.source_path), review);
const source = readFileSync(chapterPath);
const model = openChapter(source, review);
assert.deepEqual(Buffer.from(model.source), source);
if (manifestPath) { const book = decodeManifest(readFileSync(manifestPath, 'utf8')); assert.ok(book.chapters.some((chapter: { id: string; path: string }) => chapter.id === identity.chapter_id && chapter.path === identity.source_path)); }
const state = classifyReview(source, review);
console.log(JSON.stringify({ chapter: identity.source_path, edits: review.edits.length, signals: review.signals.length, activeEdits: state.edits.size, activeSignals: state.signals.size, codecRoundTrip: 'passed' }));
