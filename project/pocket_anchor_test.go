package project

import (
	"bytes"
	"strings"
	"testing"
)

func TestPocketAnchorContextMatchesCodePointWindows(t *testing.T) {
	for _, surroundings := range []string{"", "Я🐦\r\n", strings.Repeat("Я🐦é\r\n", 160), strings.Repeat("a", 1<<19)} {
		source := []byte(surroundings + "выделение" + surroundings)
		from, to := len(surroundings), len(surroundings)+len("выделение")
		anchor := makePocketAnchor(source, PocketRange{From: from, To: to})
		prefix, suffix := []rune(surroundings), []rune(surroundings)
		if len(prefix) > 128 {
			prefix = prefix[len(prefix)-128:]
		}
		if len(suffix) > 128 {
			suffix = suffix[:128]
		}
		if anchor.Prefix != string(prefix) || anchor.Suffix != string(suffix) {
			t.Fatal("context differs from code-point windows")
		}
		if anchor.Source != pocketHash(source) || anchor.Selection != pocketHash(source[from:to]) || anchor.Start != int64(from) || anchor.End != int64(to) || anchor.StartLine != int64(1+bytes.Count(source[:from], []byte{'\n'})) || anchor.EndLine != int64(1+bytes.Count(source[:to-1], []byte{'\n'})) {
			t.Fatal("anchor wire fields changed")
		}
	}
}
func BenchmarkPocketAnchorLargeChapter(b *testing.B) {
	source := bytes.Repeat([]byte("я🐦\n"), 150000)
	r := PocketRange{From: 7 * 75000, To: 7*75000 + 2}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		makePocketAnchor(source, r)
	}
}
