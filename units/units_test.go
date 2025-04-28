package units

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/grzadr/refscaler/internal"
)

func mapKeysToString[T any](m map[string]T) string {
	keys := make([]string, len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}

	slices.Sort(keys)
	return strings.Join(keys, ", ")
}

func verifyGroupWithEntryRef(
	ref internal.TestUnitEntry,
	group *UnitGroup,
) error {
	unit, ok := group.Get(ref.Name)

	if !ok {
		return fmt.Errorf(
			"%s not present in UnitGroup aliases %s",
			ref.Name,
			mapKeysToString(group.aliases),
		)
	}

	if unit.Multiplier != ref.Value {
		return fmt.Errorf(
			"%s ref value %f differs from %f",
			unit.Name,
			ref.Value,
			unit.Multiplier,
		)
	}

	for _, alias := range ref.Aliases {
		aliased_unit, ok := group.Get(alias)
		if !ok {
			return fmt.Errorf(
				"alias %s for %s not present in UnitGroup aliases %s",
				alias,
				unit.Name,
				mapKeysToString(group.aliases),
			)
		}

		if unit != aliased_unit {
			return fmt.Errorf(
				"%s unit aliased as %s points to different instance",
				unit.Name, alias,
			)
		}
	}

	return nil
}

func verifyUnitGroup(
	entries *[]internal.TestUnitEntry,
	group *UnitGroup,
) error {
	for _, entry := range internal.GetFixtureTestUnitEntries() {
		if err := verifyGroupWithEntryRef(entry, group); err != nil {
			return fmt.Errorf("failed to verify ref entry %v: %w", entry, err)
		}
	}
	return nil
}

func TestNewUnitsGroup(t *testing.T) {
	entriesStr := internal.GetFixtureUnitEntriesStr()
	reader := strings.NewReader(entriesStr)

	group, err := NewUnitGroup(reader)
	if err != nil {
		t.Fatalf("fFailed to initialize UnitGroup: %s", err)
	}
	entries := internal.GetFixtureTestUnitEntries()
	if err := verifyUnitGroup(&entries, group); err != nil {
		t.Fatal(err)
	}
}

func TestNewUnitRegistryFiles(t *testing.T) {
	registry, err := NewUnitRegistryFiles(
		internal.GetFixtureTestFs(),
		internal.GetFixtureTestFsDirPath(),
	)
	if err != nil {
		t.Fatalf("failed to initialize UnitRegistryFiles: %s", err)
	}

	expected_length := 2

	if length := registry.Length(); length != expected_length {
		t.Fatalf(
			"expected length of registry to be %d, not %d",
			expected_length,
			length,
		)
	}

	keys := mapKeysToString(registry)
	expected_keys := "empty, test_unit"

	if keys != expected_keys {
		t.Fatalf("expected keys to equal `%s` not `%s`", expected_keys, keys)
	}

	empty_group, ok := (registry)["empty"]

	if !ok {
		t.Fatalf(
			"expected key 'empty' in registry: %s",
			mapKeysToString(registry),
		)
	}

	if length := empty_group.Length(); length != 0 {
		t.Fatalf(
			"expected length of empty group to be 0, not %d",
			length,
		)
	}

	test_unit_group, ok := (registry)["test_unit"]

	if !ok {
		t.Fatalf(
			"expected key 'test_unit' in registry: %s",
			mapKeysToString(registry),
		)
	}

	entries := internal.GetFixtureTestUnitEntries()
	if err := verifyUnitGroup(&entries, test_unit_group); err != nil {
		t.Fatal(err)
	}

	json_str, err := registry.ToJSON()

	if err != nil {
		t.Fatal(err)
	}

	t.Error(json_str)
}
