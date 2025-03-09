package refscaler

import (
	"encoding/json"
	// "fmt"
	"path"
	// "testing"
	"testing/fstest"
)

type TestUnitEntry struct {
	Value   float64  `json:"value"` // Changed from Value to value to match JSON
	Aliases []string `json:"aliases"`
}

type TestEntryRecords map[string]TestUnitEntry

var fixtureUnitEntryRecords = TestEntryRecords{
	"meter": {
		Value:   1.0,
		Aliases: []string{"m", "meters"},
	},
	"kilometer": {
		Value:   1000.0,
		Aliases: []string{"km", "kilometers"},
	},
}

var fixtureUnitEntryRecordsStr = func() []byte {
	jsonData, err := json.Marshal(fixtureUnitEntryRecords)

	if err != nil {
		panic(err)
	}

	return jsonData
}()

var fixtureUnitEntries = func() []UnitEntry {
	entries := make([]UnitEntry, 0, len(fixtureUnitEntryRecords))

	for name, entry := range fixtureUnitEntryRecords {
		entries = append(entries, UnitEntry{
			Name:    name,
			Value:   entry.Value,
			Aliases: entry.Aliases,
		})
	}

	return entries
}

var testFSDirPath = "units"

var testFS = fstest.MapFS{
	path.Join(testFSDirPath, "test_unit.json"): {
		Data: fixtureUnitEntryRecordsStr,
	},
	path.Join(testFSDirPath, "empty.json"): {
		Data: []byte("{}"),
	},
	path.Join(testFSDirPath, "text.txt"): {
		Data: []byte("FILE"),
	},
}
