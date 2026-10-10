package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"unicode/utf8"
)

// PocketAnchor follows Pocket Editor/Galley byte offsets, hashes and code-point
// context windows. It is never interpreted as a JavaScript string offset.
type PocketAnchor struct {
	Source    string `json:"source_sha256"`
	Selection string `json:"selection_sha256"`
	Start     int64  `json:"start_byte"`
	End       int64  `json:"end_byte"`
	StartLine int64  `json:"start_line"`
	EndLine   int64  `json:"end_line"`
	Prefix    string `json:"prefix"`
	Suffix    string `json:"suffix"`
}
type PocketRange struct {
	From int `json:"from"`
	To   int `json:"to"`
}
type PocketResolution struct {
	Kind  string       `json:"kind"`
	Range *PocketRange `json:"range,omitempty"`
}

func pocketHash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func resolvePocket(source []byte, sourceHash string, a PocketAnchor, selected string) PocketResolution {
	stale := PocketResolution{Kind: "stale"}
	if selected == "" || pocketHash([]byte(selected)) != a.Selection {
		return stale
	}
	resolved := func(from, to int) PocketResolution {
		return PocketResolution{Kind: "resolved", Range: &PocketRange{from, to}}
	}
	if sourceHash == a.Source {
		if a.Start < 0 || a.End <= a.Start || a.End > int64(len(source)) {
			return stale
		}
		if string(source[a.Start:a.End]) != selected {
			return stale
		}
		return resolved(int(a.Start), int(a.End))
	}
	// Count all matches but retain only the first. Ambiguity must not allocate an
	// unbounded candidate list for repeated short selections in a large chapter.
	count, contextual := 0, 0
	var first, contextMatch PocketRange
	for offset := 0; offset < len(source); {
		i := bytes.Index(source[offset:], []byte(selected))
		if i < 0 {
			break
		}
		i += offset
		r := PocketRange{i, i + len(selected)}
		count++
		first = r
		if bytes.HasSuffix(source[:i], []byte(a.Prefix)) && bytes.HasPrefix(source[r.To:], []byte(a.Suffix)) {
			contextual++
			contextMatch = r
		}
		offset = i + 1
	}
	if count == 1 {
		return resolved(first.From, first.To)
	}
	if count == 0 {
		return stale
	}
	if contextual == 1 {
		return resolved(contextMatch.From, contextMatch.To)
	}
	return PocketResolution{Kind: "ambiguous"}
}
func makePocketAnchor(source []byte, r PocketRange) PocketAnchor {
	prefix := []rune(string(source[:r.From]))
	suffix := []rune(string(source[r.To:]))
	if len(prefix) > 128 {
		prefix = prefix[len(prefix)-128:]
	}
	if len(suffix) > 128 {
		suffix = suffix[:128]
	}
	return PocketAnchor{Source: pocketHash(source), Selection: pocketHash(source[r.From:r.To]), Start: int64(r.From), End: int64(r.To), StartLine: int64(1 + bytes.Count(source[:r.From], []byte{'\n'})), EndLine: int64(1 + bytes.Count(source[:r.To-1], []byte{'\n'})), Prefix: string(prefix), Suffix: string(suffix)}
}
func validPocketAnchor(a PocketAnchor) bool {
	return validObjectHash(a.Source) && validObjectHash(a.Selection) && a.Start >= 0 && a.End > a.Start && a.StartLine >= 1 && a.EndLine >= a.StartLine && a.EndLine <= 2147483647 && utf8.RuneCountInString(a.Prefix) <= 128 && utf8.RuneCountInString(a.Suffix) <= 128
}
