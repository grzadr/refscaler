package units

import (
	"fmt"
	"strings"
	"testing"

	"github.com/grzadr/refscaler/internal"
	"github.com/grzadr/refscaler/units/unit_entry"
)

func mapKeysToString[T any](m map[string]T) string {
	keys := make([]string, len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}

	return strings.Join(keys, ", ")
}

func verifyGroupWithEntryRef(ref unit_entry.UnitEntry, group *UnitGroup) error {
	unit, ok := group.Get(ref.Name)

	if !ok {
		return fmt.Errorf(
			"%s not present in UnitGroup aliases %s",
			ref.Name,
			mapKeysToString(group.aliases),
		)
	}

	if unit.multiplier != ref.Value {
		return fmt.Errorf("%s ref value %f differs from %f", unit.name, ref.Value, unit.multiplier)
	}

	for _, alias := range ref.Aliases {
		aliased_unit, ok := group.Get(alias)
		if !ok {
			return fmt.Errorf(
				"Alias %s for %s not present in UnitGroup aliases %s",
				alias,
				unit.name,
				mapKeysToString(group.aliases),
			)
		}

		if unit != aliased_unit {
			return fmt.Errorf(
				"%s unit aliased as %s points to different instance",
				unit.name, alias,
			)
		}
	}

	return nil
}

func TestNewUnitsGroup(t *testing.T) {
	entriesStr := internal.GetFixtureUnitEntriesStr()
	reader := strings.NewReader(entriesStr)

	group, err := NewUnitGroup(reader)
	if err != nil {
		t.Fatalf("Failed to initialize UnitGroup: %s", err)
	}
	for entry, err := range unit_entry.IterUnitEntries(strings.NewReader(entriesStr)) {
		if err != nil {
			t.Fatalf("Failed to prepare ref entries: %s", err)
		}

		if err := verifyGroupWithEntryRef(entry, group); err != nil {
			t.Fatalf("Failed to verify ref entry %v: %s", entry, err)
		}
	}
}
