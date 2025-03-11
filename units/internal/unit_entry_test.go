package units_entry

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func helpCompareUnitEntry(a, b *UnitEntry) bool {
	if a.Name != b.Name || a.Value != b.Value || len(a.Aliases) != len(b.Aliases) {
		return false
	}

	for i, alias := range a.Aliases {
		if alias != b.Aliases[i] {
			return false
		}
	}

	return true
}

func TestIterUnitEntries_Success(t *testing.T) {
	reader := bytes.NewReader(fixtureUnitEntriesByte)
	// fmt.Printf("%s\n", fixtureUnitEntryRecordsByte)
	for i, next := range IterUnitEntries(reader) {
		if next.Err != nil {
			t.Fatalf("Received unexpected error %v", next.Err)
		}

		if !helpCompareUnitEntry(&next.Entry, &fixtureUnitEntries[i]) {
			t.Errorf(
				"Entry %d: expected %+v, got %+v",
				i,
				fixtureUnitEntries[i],
				next.Entry,
			)
		}
	}
}

func TestIterUnitEntries_EarlyTermination(t *testing.T) {
	count := 0
	reader := bytes.NewReader(fixtureUnitEntriesByte)
	for i, next := range IterUnitEntries(reader) {
		count = i
		if next.Err != nil {
			t.Errorf("Received unexpected error %v", next.Err)
		}
		break
	}

	if count != 0 {
		t.Errorf("Expected to terminate at 0, but terminated at %d", count)
	}
}

func TestIterUnitEntries_InvalidJSON(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "invalid opening delimiter",
			input:   `["not an object"]`,
			wantErr: "expected {, got [",
		},
		{
			name:    "corrupted JSON",
			input:   `{invalid`,
			wantErr: "reading key: invalid character 'i'",
		},
		{
			name: "non-string key",
			// This will trigger the type assertion failure
			input:   `{1: {"value": 1.0, "aliases": ["m"]}}`,
			wantErr: "reading key: invalid character '1'",
		},
		{
			name:    "invalid value structure",
			input:   `{"key": "not an object"}`,
			wantErr: "decoding value for \"key\": json: cannot unmarshal string into Go value of type parsing.UnitEntry",
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: "reading token: EOF",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotErr error
			for _, next := range IterUnitEntries(strings.NewReader(tc.input)) {
				gotErr = next.Err
				break
			}

			if gotErr == nil {
				t.Fatal("expected error, got nil")
			}
			if gotErr.Error() != tc.wantErr {
				t.Errorf("expected error %q, got %q", tc.wantErr, gotErr.Error())
			}
		})
	}
}

func TestParseNextEntry_NonStringKey(t *testing.T) {
	input := []byte(`{"1": {"value": 1.0, "aliases": ["m"]}}`)
	decoder := json.NewDecoder(bytes.NewReader(input))

	for i := 0; i < 2; i++ {
		_, err := decoder.Token()

		if err != nil {
			t.Error(err)
			return
		}
	}
	entry, err := parseNextEntry(decoder)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErr := "expected string key, got json.Delim"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}

	if entry.validate() == nil {
		t.Errorf("expected zero UnitEntry, got %+v", entry)
	}
}

func TestIterUnitEntries_ComplexErrors(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name: "missing value field",
			input: `{
                "meter": {
                    "aliases": ["m"]
                }
            }`,
			wantErr: "invalid entry \"meter\": positive non-zero value field is required",
		},
		{
			name: "invalid value type",
			input: `{
                "meter": {
                    "value": "not a number",
                    "aliases": ["m"]
                }
            }`,
			wantErr: "decoding value for \"meter\": json: cannot unmarshal string into Go struct field UnitEntry.value of type float64",
		},
		{
			name: "invalid aliases type",
			input: `{
                "meter": {
                    "value": 1.0,
                    "aliases": "not an array"
                }
            }`,
			wantErr: "decoding value for \"meter\": json: cannot unmarshal string into Go struct field UnitEntry.aliases of type []string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotErr error
			for _, next := range IterUnitEntries(strings.NewReader(tc.input)) {
				gotErr = next.Err
				break
			}

			if gotErr == nil {
				t.Fatal("expected error, got nil")
			}
			if gotErr.Error() != tc.wantErr {
				t.Errorf("expected error %q, got %q", tc.wantErr, gotErr.Error())
			}
		})
	}
}

func TestUnitEntry_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		entry   UnitEntry
		wantErr string
	}{
		{
			name: "valid entry",
			entry: UnitEntry{
				Name:    "meter",
				Value:   1.0,
				Aliases: []string{"m"},
			},
			wantErr: "",
		},
		{
			name: "missing value",
			entry: UnitEntry{
				Name:    "meter",
				Aliases: []string{"m"},
			},
			wantErr: "positive non-zero value field is required",
		},
		{
			name:    "zero value struct",
			entry:   UnitEntry{},
			wantErr: "positive non-zero value field is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.entry.validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tc.wantErr {
				t.Errorf("expected error %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

// Benchmark to ensure performance
func BenchmarkIterUnitEntries(b *testing.B) {
	for b.Loop() {
		reader := bytes.NewReader(fixtureUnitEntriesByte)
		for _, next := range IterUnitEntries(reader) {
			if next.Err != nil {
				b.Fatal(next.Err)
			}
		}
	}
}
