package refscaler

import (
	"fmt"
	"testing"

	"github.com/grzadr/refscaler/internal"
	"github.com/grzadr/refscaler/units"
)

func helperCompareEnlistments(
	expected []internal.TestEnlistment,
	enlistment *Enlistment,
) error {
	if len(enlistment.records) != len(expected) {
		return fmt.Errorf(
			"expected %d records, got %d",
			len(expected),
			len(enlistment.records),
		)
	}

	i := 0

	for exp, rec := range internal.IterZip(expected, enlistment.records) {
		if rec.absValue != MeasureValue(exp.Value) {
		}

		if rec.label != exp.Label {
			return fmt.Errorf(
				"expected label '%s', got '%s' for record %d",
				exp.Label,
				rec.label,
				i,
			)
		}

		i++
	}

	return nil
}

func TestNewEnlistmentFromFile(t *testing.T) {
	expected := internal.GetFixtureEnlistmentExpected()

	for _, filename := range []string{"standard", "unsorted"} {
		enlistment, err := NewEnlistmentFromFile(
			internal.GetFixtureEnlistmentFs(),
			filename,
			units.EmbeddedUnitRegistry,
		)
		if err != nil {
			t.Fatal(err)
		}

		if err := helperCompareEnlistments(expected, enlistment); err != nil {
			t.Fatal(err)
		}
	}
}
