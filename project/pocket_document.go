package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var pocketUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type pocketJSON map[string]json.RawMessage

func pocketInvalid(message string) error {
	return fmt.Errorf("%w: Pocket review: %s", ErrInvalidTree, message)
}
func pocketObject(data []byte) (pocketJSON, error) {
	var object pocketJSON
	if !utf8.Valid(data) || json.Unmarshal(data, &object) != nil || object == nil {
		return nil, pocketInvalid("expected a UTF-8 JSON object")
	}
	return object, nil
}

// Unknown fields stay as raw JSON, including numeric values beyond float64.
func pocketField(object pocketJSON, key string, target any, required bool) error {
	value, ok := object[key]
	if !ok {
		if required {
			return pocketInvalid("missing " + key)
		}
		return nil
	}
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) || json.Unmarshal(value, target) != nil {
		return pocketInvalid("invalid " + key)
	}
	return nil
}
func pocketSet(object pocketJSON, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	object[key] = data
	return nil
}
func pocketChild(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\\x00")
}

type PocketRecord struct {
	ID         string           `json:"id"`
	Type       string           `json:"type,omitempty"`
	Before     string           `json:"before,omitempty"`
	After      string           `json:"after,omitempty"`
	Selected   string           `json:"selected_text,omitempty"`
	Comment    string           `json:"comment,omitempty"`
	Anchor     PocketAnchor     `json:"anchor"`
	Resolution PocketResolution `json:"resolution"`
	raw        pocketJSON
}
type pocketDocument struct {
	raw                         pocketJSON
	ChapterID, SourcePath, Note string
	Edits, Signals              []PocketRecord
}

func decodePocket(data []byte) (pocketDocument, error) {
	d := pocketDocument{Edits: []PocketRecord{}, Signals: []PocketRecord{}}
	var err error
	d.raw, err = pocketObject(data)
	if err != nil {
		return d, err
	}
	version := 1
	for _, f := range []struct {
		key      string
		target   any
		required bool
	}{{"schema_version", &version, false}, {"chapter_id", &d.ChapterID, true}, {"source_path", &d.SourcePath, true}, {"chapter_note", &d.Note, false}} {
		if err = pocketField(d.raw, f.key, f.target, f.required); err != nil {
			return d, err
		}
	}
	if version != 1 || !pocketUUID.MatchString(d.ChapterID) || !pocketChild(d.SourcePath) {
		return d, pocketInvalid("unsupported schema or chapter identity")
	}
	ids := map[string]bool{}
	for _, kind := range []string{"edits", "signals"} {
		var rows []json.RawMessage
		if err = pocketField(d.raw, kind, &rows, false); err != nil {
			return d, err
		}
		if len(rows) > 1000 {
			return d, pocketInvalid("at most 1000 records per kind can be processed in the browser")
		}
		for _, data := range rows {
			raw, err := pocketObject(data)
			if err != nil {
				return d, err
			}
			r := PocketRecord{raw: raw}
			if err = pocketField(raw, "id", &r.ID, true); err != nil {
				return d, err
			}
			if !pocketUUID.MatchString(r.ID) || ids[r.ID] {
				return d, pocketInvalid("invalid or duplicate record ID")
			}
			ids[r.ID] = true
			anchorRaw, err := pocketObject(raw["anchor"])
			if err != nil {
				return d, err
			}
			for _, field := range []struct {
				key    string
				target any
			}{
				{"source_sha256", &r.Anchor.Source}, {"selection_sha256", &r.Anchor.Selection},
				{"start_byte", &r.Anchor.Start}, {"end_byte", &r.Anchor.End},
				{"start_line", &r.Anchor.StartLine}, {"end_line", &r.Anchor.EndLine},
				{"prefix", &r.Anchor.Prefix}, {"suffix", &r.Anchor.Suffix},
			} {
				if err = pocketField(anchorRaw, field.key, field.target, true); err != nil {
					return d, err
				}
			}
			if !validPocketAnchor(r.Anchor) {
				return d, pocketInvalid("invalid anchor")
			}
			if kind == "edits" {
				if err = pocketField(raw, "before", &r.Before, true); err != nil {
					return d, err
				}
				if err = pocketField(raw, "after", &r.After, true); err != nil {
					return d, err
				}
				if r.Before == "" || r.Before == r.After {
					return d, pocketInvalid("invalid proposed replacement")
				}
				d.Edits = append(d.Edits, r)
			} else {
				if err = pocketField(raw, "type", &r.Type, true); err != nil {
					return d, err
				}
				if err = pocketField(raw, "selected_text", &r.Selected, true); err != nil {
					return d, err
				}
				if err = pocketField(raw, "comment", &r.Comment, false); err != nil {
					return d, err
				}
				switch r.Type {
				case "note", "change_required", "warning", "review":
				default:
					return d, pocketInvalid("unsupported signal type")
				}
				d.Signals = append(d.Signals, r)
			}
		}
	}
	return d, nil
}
func (d *pocketDocument) classify(source []byte) {
	hash := pocketHash(source)
	for i := range d.Edits {
		r := &d.Edits[i]
		r.Resolution = resolvePocket(source, hash, r.Anchor, r.Before)
	}
	for i := range d.Signals {
		r := &d.Signals[i]
		r.Resolution = resolvePocket(source, hash, r.Anchor, r.Selected)
	}
	conflicts := map[int]bool{}
	for i, a := range d.Edits {
		if a.Resolution.Range == nil {
			continue
		}
		for j := i + 1; j < len(d.Edits); j++ {
			b := d.Edits[j]
			if b.Resolution.Range != nil && a.Resolution.Range.From < b.Resolution.Range.To && b.Resolution.Range.From < a.Resolution.Range.To {
				conflicts[i] = true
				conflicts[j] = true
			}
		}
	}
	for i := range conflicts {
		d.Edits[i].Resolution = PocketResolution{Kind: "conflicting"}
	}
}
