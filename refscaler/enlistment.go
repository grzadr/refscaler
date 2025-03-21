package refscaler

import (
	"github.com/grzadr/refscaler/units"
)

type MeasureValue float64

type Record struct {
	label    string
	absValue MeasureValue
}

type RecordSlice []Record

type Enlistment struct {
	records   RecordSlice
	scale     MeasureValue
	unitRef   string
	recordRef *Record
	units     units.UnitsSlice
}

func NewEnlistment(
	scale string,
	records []string,
	units units.UnitRegistry,
) (*Enlistment, error) {
	return nil, nil
}
