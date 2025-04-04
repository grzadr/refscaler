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

func NewEnlistmentDefault() *Enlistment {
	return &Enlistment{
		records: make(RecordSlice, 0, 32),
	}
}

func NewEnlistment(
	scale string,
	records []string,
	units units.UnitRegistry,
) (enlistment *Enlistment, err error) {
	return NewEnlistmentDefault(), nil
}


