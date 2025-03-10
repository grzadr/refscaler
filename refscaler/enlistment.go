package refscaler

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
	units     UnitSlice
}

func NewEnlistment(
	scale string,
	records []string,
	units UnitRegistry,
) (*Enlistment, error) {

}
