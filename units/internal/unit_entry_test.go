package units_entry

import (
	// "bytes"
	// "encoding/json"
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
	reader := strings.NewReader(fixtureUnitEntriesStr)
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
	reader := strings.NewReader(fixtureUnitEntriesStr)
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
			name:    "unexpected token",
			input:   `{"not an object"}`,
			wantErr: "error reading token: expected [, got {",
		},
		{
			name:    "invalid opening delimiter",
			input:   `["not an object"]`,
			wantErr: "cannot unmarshal string `\"not an object\"` into units_entry.UnitEntry",
		},
		{
			name:    "corrupted JSON",
			input:   `[ invalid`,
			wantErr: "reading JSON: invalid character 'i' looking for beginning of value",
		},
		{
			name: "non-string name",
			// This will trigger the type assertion failure
			input:   `[{"name": 1, "value": 1.0}]`,
			wantErr: "cannot unmarshal number `{\"name\": 1, \"value\": 1.0}` into string field `name`",
		},
		{
			name:    "missing name key",
			input:   `[{"key": "not an object"}]`,
			wantErr: "error validating entry {\"key\": \"not an object\"}: unit name cannot be empty",
		},
		{
			name:    "empty name value",
			input:   `[{"name": ""}]`,
			wantErr: "error validating entry {\"name\": \"\"}: unit name cannot be empty",
		},
		{
			name:    "zero value",
			input:   `[{"name": "a","value": 0}]`,
			wantErr: "error validating entry {\"name\": \"a\",\"value\": 0}: unit value must be positive non-zero",
		},
		{
			name:    "negative value",
			input:   `[{"name": "a","value": -0.001}]`,
			wantErr: "error validating entry {\"name\": \"a\",\"value\": -0.001}: unit value must be positive non-zero",
		},
		{
			name:    "empty input",
			input:   ``,
			wantErr: "error reading token: EOF",
		},
		{
			name:    "syntax error - unclosed object",
			input:   `[{"name": "test", "value": 1.0`,
			wantErr: "syntax error at offset 28 (content: [{\"name\": \"test\", \"value\": 1.0): unexpected EOF",
		},
		{
			name:    "syntax error - invalid escape sequence",
			input:   `[{"name": "test with \invalid escape", "value": 1.0}]`,
			wantErr: "syntax error at offset 21 (content: [{\"name\": \"test with \\invalid escape\", \"value\": 1.0}]): invalid character 'i' in string escape code",
		},
		{
			name:    "other unmarshaling error - large number",
			input:   `[{"name": "test", "value": 1e1000}]`,
			wantErr: "error decoding entry (content: [{\"name\": \"test\", \"value\": 1e1000}]): json: cannot unmarshal number 1e1000 into Go struct field UnitEntry.value of type float64",
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

// Benchmark to ensure performance
func BenchmarkIterUnitEntries(b *testing.B) {
	for b.Loop() {
		reader := strings.NewReader(fixtureUnitEntriesStr)
		for _, next := range IterUnitEntries(reader) {
			if next.Err != nil {
				b.Fatal(next.Err)
			}
		}
	}
}
