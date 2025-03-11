package units_entry

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
)

// UnitEntry represents a single unit definition.
// Fields are exported to work with json.Decoder
type UnitEntry struct {
	Name    string   `json:"name"`
	Value   float64  `json:"value"`
	Aliases []string `json:"aliases"`
}

func (u UnitEntry) validate() error {
	if u.Value == 0 {
		return fmt.Errorf("positive non-zero value field is required")
	}
	return nil
}

// NextUnitEntry wraps a UnitEntry with potential error.
// This is cleaner than having a separate error field.
type NextUnitEntry struct {
	Entry UnitEntry
	Err   error
}

// IterUnitEntries returns an iterator over unit entries in JSON data.
// The iterator yields an index and Result for each entry.
func IterUnitEntries(jsonData io.Reader) iter.Seq2[int, NextUnitEntry] {
	return func(yield func(int, NextUnitEntry) bool) {
		decoder := json.NewDecoder(jsonData)

		// Check for opening delimiter
		if err := expectToken(decoder, json.Delim('[')); err != nil {
			yield(0, NextUnitEntry{Err: err})
			return
		}

		// Iterate through entries
		for i := 0; decoder.More(); i++ {
			entry, err := parseNextEntry(decoder)
			if err != nil {
				yield(i, NextUnitEntry{Err: err})
				return
			}

			if !yield(i, NextUnitEntry{Entry: entry}) {
				return
			}
		}
	}
}

// expectToken checks for an expected JSON token.
func expectToken(decoder *json.Decoder, expected json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}

	delim, ok := token.(json.Delim)
	if !ok || delim != expected {
		return fmt.Errorf("expected %v, got %v", expected, token)
	}
	return nil
}

// parseNextEntry reads the next unit entry from the decoder.
func parseNextEntry(decoder *json.Decoder) (entry UnitEntry, err error) {
	// Read key (unit name)
	// key, err := decoder.Token()
	// if err != nil {
	// 	return UnitEntry{}, fmt.Errorf("reading key: %w", err)
	// }

	// name, ok := key.(string)
	// if !ok {
	// 	return UnitEntry{}, fmt.Errorf("expected string key, got %T", key)
	// }

	if err := decoder.Decode(&entry); err != nil {
		return UnitEntry{}, fmt.Errorf("error decoding entry: %w", err)
	}

	// entry.Name = name

	if err := entry.validate(); err != nil {
		return UnitEntry{}, fmt.Errorf("invalid entry: %w", err)
	}

	return entry, nil
}
