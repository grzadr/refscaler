package refscaler

type MeasureValue float64

type Record struct {
	label string
	absValue MeasureValue
}

type Unit struct {
	name string
	multiplier MeasureValue
}

type UnitGroup struct {
	units map[string]Unit
	hasPrefix bool
	baseUnit *Unit
}

type UnitSlice []Unit

type UnitRegistry string {
	groups map[string]UnitGroup
}

type RecordSlice []Record

type Enlistment struct {
	records RecordSlice
	scale MeasureValue
	unitRef string
	recordRef *Record
	units UnitSlice
}

func NewEnlistment (
	scale string,
	records []string,
	units UnitRegistry,
) (*Enlistment, error) {
	
}
