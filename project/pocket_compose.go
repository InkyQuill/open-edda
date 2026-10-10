package project

import (
	"context"
	"unicode/utf8"
)

func composePocket(d *pocketDocument, source []byte, input PocketAction) error {
	if !pocketUUID.MatchString(input.RecordID) {
		return pocketInvalid("invalid record ID")
	}
	kind := "signals"
	if input.SignalType == "edit" {
		kind = "edits"
	}
	var rows []pocketJSON
	if err := pocketField(d.raw, kind, &rows, false); err != nil {
		return err
	}
	index := -1
	for i, r := range rows {
		var id string
		if err := pocketField(r, "id", &id, true); err != nil {
			return err
		}
		if id == input.RecordID {
			index = i
		}
	}
	if input.Action == "update" {
		if index < 0 {
			return pocketInvalid("record not found or type changed")
		}
		// Updating a comment/proposal never guesses a new anchor, even when stale.
		if kind == "edits" {
			if err := pocketSet(rows[index], "after", input.After); err != nil {
				return err
			}
		} else {
			if err := pocketSet(rows[index], "type", input.SignalType); err != nil {
				return err
			}
			if err := pocketSet(rows[index], "comment", input.Comment); err != nil {
				return err
			}
		}
	} else {
		if input.From < 0 || input.To <= input.From || input.To > len(source) || !utf8.Valid(source[:input.From]) || !utf8.Valid(source[:input.To]) || string(source[input.From:input.To]) != input.Selected {
			return pocketInvalid("selection does not match source bytes")
		}
		row := pocketJSON{}
		for key, value := range map[string]any{"id": input.RecordID, "anchor": makePocketAnchor(source, PocketRange{From: input.From, To: input.To})} {
			if err := pocketSet(row, key, value); err != nil {
				return err
			}
		}
		if kind == "edits" {
			_ = pocketSet(row, "before", input.Selected)
			_ = pocketSet(row, "after", input.After)
		} else {
			_ = pocketSet(row, "type", input.SignalType)
			_ = pocketSet(row, "comment", input.Comment)
			_ = pocketSet(row, "selected_text", input.Selected)
		}
		rows = append(rows, row)
	}
	return pocketSet(d.raw, kind, rows)
}

// Undo only the chapter/sidecar pair, preserving unrelated work. Comparing with
// the immutable decision result refuses to overwrite later edits to either file.
// Undoing this resulting version provides redo, with the same conflict checks.
func (s *VersionStore) undoPocket(ctx context.Context, author, project string, input PocketAction) (ProjectVersion, error) {
	if input.ActionVersion == "" {
		return ProjectVersion{}, pocketInvalid("decision version is required")
	}
	current, err := s.Version(ctx, author, project, input.ExpectedVersion)
	if err != nil {
		return ProjectVersion{}, err
	}
	after, err := s.ReadPocketReview(ctx, author, project, input.ActionVersion, input.EntryID)
	if err != nil {
		return ProjectVersion{}, err
	}
	if after.version.ParentID == "" {
		return ProjectVersion{}, pocketInvalid("no review decision to undo")
	}
	before, err := s.Version(ctx, author, project, after.version.ParentID)
	if err != nil {
		return ProjectVersion{}, err
	}
	lookup := func(v ProjectVersion, id string) (TreeEntry, bool) {
		for _, e := range v.Entries {
			if e.ID == id {
				return e, true
			}
		}
		return TreeEntry{}, false
	}
	ids := map[string]bool{after.Chapter.ID: true, after.Sidecar.ID: true}
	entries := append([]TreeEntry(nil), current.Entries...)
	changed := false
	for id := range ids {
		was, existed := lookup(before, id)
		now, present := lookup(after.version, id)
		if existed == present && was == now {
			continue
		}
		actual, found := lookup(current, id)
		if found != present || actual != now {
			return ProjectVersion{}, ErrVersionConflict
		}
		changed = true
		filtered := entries[:0]
		for _, e := range entries {
			if e.ID != id {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
		if existed {
			entries = append(entries, was)
		}
	}
	if !changed {
		return ProjectVersion{}, pocketInvalid("version has no review changes")
	}
	return s.Publish(ctx, PublishVersionInput{AuthorID: author, ProjectID: project, ExpectedVersion: input.ExpectedVersion, OperationID: input.OperationID, Entries: entries})
}
