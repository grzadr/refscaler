package units

import (
	"io"
	"fmt"
	"github.com/grzadr/refscaler/units/unit_entry"
)

type Unit struct {
	name       string
	multiplier float64
}

type UnitsSlice []Unit
type UnitAliases map[string]*Unit

type UnitGroup struct {
	units UnitsSlice
	aliases UnitAliases
	// allowPrefix bool // TODO
	baseUnit Unit
}

func (g *UnitGroup) add(entry unit_entry.UnitEntry) error {
	return nil
}

func newUnitGroupDefault() UnitGroup {
	return UnitGroup{
		units:    make(UnitsSlice, 0, 32),
		baseUnit: Unit{},
	}
}

func NewUnitGroup(unitsData io.Reader) (group *UnitGroup, err error) {
	// TODO
	group := newUnitGroupDefault()
	for next, err := range unit_entry.IterUnitEntries(unitsData) {
		if err != nil {
			return group, fmt.Errorf("Error reading unit entry: %w", err)
		}


	}
	return nil, nil
}



type UnitRegistry interface {
	Find(unit string) (group *UnitGroup, ok bool)
}

type UnitRegistryFiles struct {
	groups map[string]UnitGroup
}

func NewUnitRegistryFiles () (reg *UnitRegistryFiles, err error) {
	// TODO
	return nil, nil
}
