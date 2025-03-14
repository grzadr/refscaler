package units_entry

import (
	"encoding/json"
	// "fmt"
	"path"
	// "testing"
	"testing/fstest"
)

var fixtureUnitEntriesStr = `
[
	{
		"name": "kilometer",
		"value": 1000.0,
		"aliases": [
			"km",
			"kilometers"
		]
	},
	{
		"name": "meter",
		"value": 1.0,
		"aliases": [
			"m",
			"meters"
		]
	}
]
`
var fixtureUnitEntries = func() []UnitEntry {
	var entries []UnitEntry
	err := json.Unmarshal([]byte(fixtureUnitEntriesStr), &entries)

	if err != nil {
		panic(err)
	}

	return entries
}()

var testFSDirPath = "units"

var testFS = fstest.MapFS{
	path.Join(testFSDirPath, "test_unit.json"): {
		Data: []byte(fixtureUnitEntriesStr),
	},
	path.Join(testFSDirPath, "empty.json"): {
		Data: []byte("{}"),
	},
	path.Join(testFSDirPath, "text.txt"): {
		Data: []byte("FILE"),
	},
}
