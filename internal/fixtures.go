package internal

import (
	"encoding/json"
	"path"
	"testing/fstest"
)

func GetFixtureUnitEntriesStr() string {
	return `
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
}

func GetFixtureUnitEntriesByte() []byte {
	return []byte(GetFixtureUnitEntriesStr())
}

type TestUnitEntry struct {
	Name    string   `json:"name"`
	Value   float64  `json:"value"`
	Aliases []string `json:"aliases"`
}

func GetFixtureTestUnitEntries() []TestUnitEntry {
	entries := make([]TestUnitEntry, 2)
	if err := json.Unmarshal(GetFixtureUnitEntriesByte(), &entries); err != nil {
		panic("cannot unmarshal entries")
	}
	return entries
}

func GetFixtureTestFsDirPath() string {
	return "units"
}

func GetFixtureTestFs() fstest.MapFS {
	return fstest.MapFS{
		path.Join(GetFixtureTestFsDirPath(), "test_unit.json"): {
			Data: GetFixtureUnitEntriesByte(),
		},
		path.Join(GetFixtureTestFsDirPath(), "empty.json"): {
			Data: []byte("[]"),
		},
		path.Join(GetFixtureTestFsDirPath(), "text.txt"): {
			Data: []byte("FILE"),
		},
	}
}

func GetFixtureEnlistmentFs() fstest.MapFS {
	return fstest.MapFS{
		"standard": {
			Data: []byte(
				`Item 1: 0.75 hour, 15 minutes
				# Item X: 100 hours
				Item 2: 15 minutes
				Item 3: 60 seconds
				`),
		},
		"unsorted": {
			Data: []byte(
				`Item 2: 15 minutes

				Item 3: 60 seconds
				Item 1: 1 hour
				`),
		},
	}
}

type TestEnlistment struct {
	Label string
	Value float64
}

func GetFixtureEnlistmentExpected() []TestEnlistment {
	return []TestEnlistment{
		{
			Label: "Item 1",
			Value: 3600,
		},
		{
			Label: "Item 2",
			Value: 900,
		},
		{
			Label: "Item 3",
			Value: 60,
		},
	}
}
