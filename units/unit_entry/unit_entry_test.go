package unit_entry

import (
	"strings"
	"testing"

	// "bytes"
	// "encoding/json"
	"github.com/grzadr/refscaler/internal"
)

func TestUnitEntry_IsBase(t *testing.T) {
	if entry := (UnitEntry{Value: 1.0}); !entry.IsBase() {
		t.Fatal("object UnitEntry with value 1.0 should be base")
	}
	if entry := (UnitEntry{Value: 0.0}); entry.IsBase() {
		t.Fatal("object UnitEntry with value 0.0 should not be base")
	}
}

func helpCompareUnitEntry(a *UnitEntry, b *internal.TestUnitEntry) bool {
	if a.Name != b.Name || a.Value != b.Value ||
		len(a.Aliases) != len(b.Aliases) {
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
	reader := strings.NewReader(internal.GetFixtureUnitEntriesStr())
	i := 0
	entries := internal.GetFixtureTestUnitEntries()
	for next, err := range IterUnitEntries(reader) {
		if err != nil {
			t.Fatalf("Received unexpected error %v", err)
		}

		if !helpCompareUnitEntry(&next, &entries[i]) {
			t.Errorf(
				"Entry %d: expected %+v, got %+v",
				i,
				entries[i],
				next,
			)
		}
		i++
	}
}

func TestIterUnitEntries_EarlyTermination(t *testing.T) {
	reader := strings.NewReader(internal.GetFixtureUnitEntriesStr())
	i := 0
	for _, err := range IterUnitEntries(reader) {
		if err != nil {
			t.Errorf("Received unexpected error %v", err)
		}
		i++
		break
	}

	if i != 1 {
		t.Errorf("Expected to terminate at 1, but terminated at %d", i)
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
			wantErr: "unexpected token: expected [, got {",
		},
		{
			name:    "invalid opening delimiter",
			input:   `["not an object"]`,
			wantErr: "cannot unmarshal string `\"not an object\"` into unit_entry.UnitEntry",
		},
		{
			name:    "corrupted JSON",
			input:   `[ invalid`,
			wantErr: "reading JSON: invalid character 'i' looking for beginning of value",
		},
		{
			name:    "non-string name",
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
			wantErr: "unexpected token: EOF",
		},
		{
			name:    "syntax error - unclosed object",
			input:   `[{"name": "test", "value": 1.0}}`,
			wantErr: "unexpected token: invalid character '}' after array element",
		},
		{
			name:    "syntax error - invalid escape sequence",
			input:   `[{"name": "test with \invalid escape", "value": 1.0}]`,
			wantErr: "reading JSON: invalid character 'i' in string escape code",
		},
		{
			name:    "other unmarshaling error - large number",
			input:   `[{"name": "test", "value": 1e1000}]`,
			wantErr: "cannot unmarshal number 1e1000 `{\"name\": \"test\", \"value\": 1e1000}` into float64 field `value`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotErr error
			for _, err := range IterUnitEntries(strings.NewReader(tc.input)) {
				gotErr = err
				if gotErr != nil {
					break
				}
			}

			if gotErr == nil {
				t.Fatal("expected error, got nil")
			}
			if gotErr.Error() != tc.wantErr {
				t.Errorf(
					"expected error %q, got %q",
					tc.wantErr,
					gotErr.Error(),
				)
			}
		})
	}
}

// Benchmark to ensure performance
func BenchmarkIterUnitEntries(b *testing.B) {
	for b.Loop() {
		reader := strings.NewReader(internal.GetFixtureUnitEntriesStr())
		for _, err := range IterUnitEntries(reader) {
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
