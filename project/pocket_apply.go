package project

import "encoding/json"

func changePocket(review PocketReview, input PocketAction) ([]byte, []byte, error) {
	source := []byte(review.Source)
	d := review.document
	switch input.Action {
	case "metadata":
		next, err := replacePocketMetadata(source, &d, input.Metadata)
		if err != nil {
			return nil, nil, err
		}
		source = next
	case "add", "update":
		if err := composePocket(&d, source, input); err != nil {
			return nil, nil, err
		}
	case "note":
		if err := pocketSet(d.raw, "chapter_note", input.Note); err != nil {
			return nil, nil, err
		}
	case "remove", "apply":
		var selected *PocketRecord
		for i := range d.Edits {
			if d.Edits[i].ID == input.RecordID {
				selected = &d.Edits[i]
			}
		}
		if input.Action == "remove" {
			for i := range d.Signals {
				if d.Signals[i].ID == input.RecordID {
					selected = &d.Signals[i]
				}
			}
		}
		if selected == nil {
			return nil, nil, pocketInvalid("record not found")
		}
		if input.Action == "apply" {
			if selected.Resolution.Kind != "resolved" || selected.Resolution.Range == nil {
				return nil, nil, pocketInvalid("edit is stale, ambiguous or overlapping; resolve it explicitly")
			}
			r := *selected.Resolution.Range
			replacement := []byte(selected.After)
			next := make([]byte, 0, len(source)-r.To+r.From+len(replacement))
			next = append(next, source[:r.From]...)
			next = append(next, replacement...)
			next = append(next, source[r.To:]...)
			shift := len(replacement) - (r.To - r.From)
			for i := range d.Edits {
				other := &d.Edits[i]
				old := other.Resolution.Range
				if other.ID == selected.ID || other.Resolution.Kind != "resolved" || old == nil {
					continue
				}
				moved := *old
				if old.From >= r.To {
					moved.From += shift
					moved.To += shift
				}
				if err := movePocketRecord(other, next, moved, false); err != nil {
					return nil, nil, err
				}
			}
			for i := range d.Signals {
				signal := &d.Signals[i]
				old := signal.Resolution.Range
				if signal.Resolution.Kind != "resolved" || old == nil {
					continue
				}
				moved := *old
				switch {
				case old.To <= r.From:
				case old.From >= r.To:
					moved.From += shift
					moved.To += shift
				case old.From >= r.From && old.To <= r.To:
					continue
				case old.From <= r.From && old.To >= r.To:
					moved.To += shift
				default:
					continue
				}
				if err := movePocketRecord(signal, next, moved, true); err != nil {
					return nil, nil, err
				}
			}
			source = next
		}
		for _, group := range []struct {
			key     string
			records []PocketRecord
		}{{"edits", d.Edits}, {"signals", d.Signals}} {
			rows := []pocketJSON{}
			for _, r := range group.records {
				if r.ID != input.RecordID {
					rows = append(rows, r.raw)
				}
			}
			// Avoid adding an absent, untouched default array during a decision.
			if _, exists := d.raw[group.key]; exists {
				if err := pocketSet(d.raw, group.key, rows); err != nil {
					return nil, nil, err
				}
			}
		}
	default:
		return nil, nil, pocketInvalid("unsupported review action")
	}
	data, err := json.MarshalIndent(d.raw, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	if _, err = decodePocket(data); err != nil {
		return nil, nil, err
	}
	return source, append(data, '\n'), nil
}
func movePocketRecord(record *PocketRecord, source []byte, r PocketRange, signal bool) error {
	anchor := makePocketAnchor(source, r)
	// Preserve future/extension fields of an anchor as well as of its record.
	raw, err := pocketObject(record.raw["anchor"])
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(anchor)
	if err != nil {
		return err
	}
	known, err := pocketObject(encoded)
	if err != nil {
		return err
	}
	for key, value := range known {
		raw[key] = value
	}
	if err = pocketSet(record.raw, "anchor", raw); err != nil {
		return err
	}
	if signal {
		return pocketSet(record.raw, "selected_text", string(source[r.From:r.To]))
	}
	return nil
}
