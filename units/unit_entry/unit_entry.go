package unit_entry

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
func IterUnitEntries(jsonData io.Reader) iter.Seq2[UnitEntry, error] {
	return func(yield func(UnitEntry, error) bool) {
		decoder := json.NewDecoder(jsonData)
		decoder.DisallowUnknownFields()

		// Check for opening delimiter
		if err := expectToken(decoder, json.Delim('[')); err != nil {
			yield(UnitEntry{}, fmt.Errorf("unexpected token: %w", err))
			return
		}

		// Iterate through entries
		for decoder.More(){
			entry, err := parseNextEntry(decoder)
			if err != nil {
				yield(UnitEntry{}, err)
				return
			}

			if !yield(entry, nil) {
				return
			}
		}

		if err := expectToken(decoder, json.Delim(']')); err != nil {
			yield(UnitEntry{}, fmt.Errorf("unexpected token: %w", err))
			return
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

	err = json.Unmarshal(rawJSON, &entry)

	var typeErr *json.UnmarshalTypeError

	if errors.As(err, &typeErr) {
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
	} else if err != nil {
		return UnitEntry{}, fmt.Errorf("unsupported JSON error: %w", err)
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
