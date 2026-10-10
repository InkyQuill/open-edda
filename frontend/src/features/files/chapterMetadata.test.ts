import { expect, it } from 'vitest';
import { frontmatter, metadataReplacement } from './chapterMetadata';
it('keeps YAML extensions and comments editable without serializing the body', () => {
 const source = '\ufeff---\r\n# comment\r\ntitle: Текст\r\ncustom: [a, b]\r\n...\r\nТело 🐦\n';
 const fm = frontmatter(source)!;
 expect(fm.error).toBeNull(); expect(fm.title).toBe('Текст');
 expect(source.slice(fm.end)).toBe('Тело 🐦\n');
 expect(metadataReplacement(source, fm.yaml.replace('Текст', 'Новое'))).toContain('# comment\r\ntitle: Новое');
});
it.each(['a: 1\na: 2', '[a, b]', 'title: x\n---\nbody'])('rejects invalid or escaping YAML', yaml => expect(() => metadataReplacement('Body', yaml)).toThrow());
it('detects unterminated metadata', () => expect(frontmatter('---\ntitle: x\nBody')?.error).toContain('закрывающей'));
