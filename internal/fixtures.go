package internal

import (
	"json"
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

func() []UnitEntry {
	var entries []UnitEntry
	err := json.Unmarshal([]byte(fixtureUnitEntriesStr), &entries)

	if err != nil {
		panic(err)
	}

	return entries
}()
