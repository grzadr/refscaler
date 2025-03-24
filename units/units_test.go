package units

import (
	"strings"
	"testing"

	"github.com/grzadr/refscaler/internal"
)

func TestNewUnitsGroup(t *testing.T) {
	reader := strings.NewReader(internal.GetFixtureUnitEntriesStr())

	group, err := NewUnitGroup(reader)
	if err != nil {
		t.Fatal(err)
	}
}
