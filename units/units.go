package units

import (
	"fmt"
	"io"

	"github.com/grzadr/refscaler/units/unit_entry"
)

type Unit struct {
	name       string
	multiplier float64
}

type (
	UnitsSlice  []*Unit
	UnitAliases map[string]*Unit
)

type UnitGroup struct {
	units   UnitsSlice
	aliases UnitAliases
	// allowPrefix bool // TODO
	baseUnit Unit
}

func (g *UnitGroup) add(entry unit_entry.UnitEntry) error {
	unit := &Unit{
		name: entry.Name,
		multiplier: entry.Value,
	}

	g.units = append(g.units, unit)
	g.aliases[unit.name] = unit

	for _, a := range entry.Aliases {
		g.aliases[a] = unit
	}

	return nil
}

func newUnitGroupDefault() *UnitGroup {
	return &UnitGroup{
		units:    make(UnitsSlice, 0, 32),
		baseUnit: Unit{},
	}
}

func NewUnitGroup(unitsData io.Reader) (group *UnitGroup, err error) {
	group = newUnitGroupDefault()
	for entry, err := range unit_entry.IterUnitEntries(unitsData) {
		if err != nil {
			return group, fmt.Errorf("Error reading unit entry: %w", err)
		}
		if err := group.add(entry); err != nil {
			return group, fmt.Errorf("Error adding unit entry %v: %w", entry, err)
		}
	}
	return group, nil
}

type UnitRegistry interface {
	Find(unit string) (group *UnitGroup, ok bool)
}

type UnitRegistryFiles struct {
	groups map[string]UnitGroup
}

func NewUnitRegistryFiles() (registry *UnitRegistryFiles, err error) {
	// TODO
	return registry, err
}
