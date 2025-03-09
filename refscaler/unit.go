package refscaler

type Unit struct {
	name       string
	multiplier MeasureValue
}

type UnitGroup struct {
	units UnitsMap
	// allowPrefix bool #TODO
	baseUnit Unit
}

type UnitsMap map[string]Unit

func NewUnitGroupDefault() UnitGroup {
	return UnitGroup{
		units:    make(UnitsMap),
		baseUnit: Unit{},
	}
}

type UnitSlice []Unit

type UnitRegistry interface {
	find(unitQuery string) (group UnitGroup, ok bool)
}

type UnitRegistryFiles struct {
	groups map[string]UnitGroup
}
