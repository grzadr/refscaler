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

var testJsonMap = TestEntryRecords{
	"meter": {
		Value:   1.0,
		Aliases: []string{"m", "meters"},
	},
	"kilometer": {
		Value:   1000.0,
		Aliases: []string{"km", "kilometers"},
	},
}


var testJsonData []byte = func() []byte {
	data, err := json.Marshal(testJsonMap)
	if err != nil {
		panic(err)
	}
	return data
}()

var testFSDirPath = "units"

var testFS = fstest.MapFS{
	path.Join(testFSDirPath, "test_unit.json"): {
		Data: testJsonData,
	},
	path.Join(testFSDirPath, "empty.json"): {
		Data: []byte("{}"),
	},
	path.Join(testFSDirPath, "text.txt"): {
		Data: []byte("FILE"),
	},
}
