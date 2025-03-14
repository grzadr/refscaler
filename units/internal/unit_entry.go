package units_entry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
)

var (
	ErrEmptyName = errors.New("unit name cannot be empty")
	ErrZeroValue = errors.New("unit value must be positive non-zero")
)

// UnitEntry represents a single unit definition.
// Fields are exported to work with json.Decoder
type UnitEntry struct {
	Name    string   `json:"name"`
	Value   float64  `json:"value"`
	Aliases []string `json:"aliases"`
}

func (u UnitEntry) validate() error {
	if len(u.Name) == 0 {
		return ErrEmptyName
	}
	if u.Value <= 0 {
		return ErrZeroValue
	}
	return nil
}

// NextUnitEntry wraps a UnitEntry with potential error.
// This is cleaner than having a separate error field.
type NextUnitEntry struct {
	Entry UnitEntry
	Err   error
}

// expectToken checks for an expected JSON token.
func expectToken(decoder *json.Decoder, expected json.Delim) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	delim, ok := token.(json.Delim)
	if !ok || delim != expected {
		return fmt.Errorf("expected %v, got %v", expected, token)
	}
	return nil
}

// IterUnitEntries returns an iterator over unit entries in JSON data.
// The iterator yields an index and Result for each entry.
func IterUnitEntries(jsonData io.Reader) iter.Seq2[int, NextUnitEntry] {
	return func(yield func(int, NextUnitEntry) bool) {
		decoder := json.NewDecoder(jsonData)

		// Check for opening delimiter
		if err := expectToken(decoder, json.Delim('[')); err != nil {
			yield(0, NextUnitEntry{Err: fmt.Errorf("error reading token: %w", err)})
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

// parseNextEntry reads the next unit entry from the decoder.
func parseNextEntry(decoder *json.Decoder) (entry UnitEntry, err error) {
	var rawJSON json.RawMessage
	if err := decoder.Decode(&rawJSON); err != nil {
		return UnitEntry{}, fmt.Errorf("reading JSON: %w", err)
	}

	entry = UnitEntry{Aliases: make([]string, 0, 4)}

	// Try to unmarshal the raw JSON into the entry struct
	if err := json.Unmarshal(rawJSON, &entry); err != nil {
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			return UnitEntry{}, fmt.Errorf(
				"syntax error at offset %d (content: %s): %w",
				syntaxErr.Offset,
				string(rawJSON),
				err,
			)

		case errors.As(err, &typeErr):
			if field := typeErr.Field; field != "" {
				return UnitEntry{}, fmt.Errorf(
					"cannot unmarshal %s `%s` into %s field `%s`",
					typeErr.Value,
					string(rawJSON),
					typeErr.Type.String(),
					typeErr.Field,
				)
			}
			return UnitEntry{}, fmt.Errorf(
				"cannot unmarshal %s `%s` into %s",
				typeErr.Value,
				string(rawJSON),
				typeErr.Type.String(),
			)

		default:
			return UnitEntry{}, fmt.Errorf(
				"error decoding entry (content: %s): %w",
				string(rawJSON),
				err,
			)
		}
	}

	if err := entry.validate(); err != nil {
		return UnitEntry{}, fmt.Errorf(
			"error validating entry %s: %w",
			string(rawJSON),
			err,
		)
	}

	return entry, nil
}
