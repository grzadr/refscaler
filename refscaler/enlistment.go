package refscaler

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strconv"
	"strings"

	"github.com/grzadr/refscaler/units"
)

type MeasureValue float64

func newMeasureValue(
	value string,
	unit_alias string,
	group *units.UnitGroup,
) (measure MeasureValue, err error) {
	measure_value, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return measure, fmt.Errorf(
			"value '%s' cannot be parsed to float",
			value,
		)
	}

	unit, ok := group.Get(unit_alias)

	if !ok {
		return measure, fmt.Errorf(
			"unit alias '%s' not found in group",
			unit_alias,
		)
	}

	measure = MeasureValue(measure_value * unit.Multiplier)
	return
}

func newMeasureValueFromEntry(
	entries []EntryMeasure,
	group *units.UnitGroup,
) (measure MeasureValue, err error) {
	for _, entry := range entries {
		measure_value, err := newMeasureValue(entry.Value, entry.Unit, group)
		if err != nil {
			return measure, err
		}

		measure += measure_value
	}
	return
}

type EntryMeasure struct {
	Unit  string
	Value string
}

func newEntryMeasure(str string) (measure EntryMeasure, err error) {
	value, unit, found := strings.Cut(strings.TrimSpace(str), " ")
	value = strings.TrimSpace(value)
	unit = strings.TrimSpace(unit)

	if !found || len(value) == 0 || len(unit) == 0 {
		return measure, fmt.Errorf("measure '%s' has wrong format", str)
	}

	measure.Unit = unit
	measure.Value = value

	return
}

type Entry struct {
	label    string
	measures []EntryMeasure
}

func (e *Entry) getFirstUnit() string {
	return e.measures[0].Unit
}

func newDefaultEntry() *Entry {
	return &Entry{
		measures: make([]EntryMeasure, 4),
	}
}

func (e *Entry) loadMeasures(str string) error {
	measures_query := strings.Split(str, ",")

	if len(measures_query) == 0 {
		return fmt.Errorf("no measures detected")
	}

	for _, query := range measures_query {
		measure, err := newEntryMeasure(query)
		if err != nil {
			return err
		}

		e.measures = append(e.measures, measure)
	}

	return nil
}

func newEntry(line string) (entry *Entry, err error) {
	entry = newDefaultEntry()

	label, measures, found := strings.Cut(line, ": ")

	if !found {
		return entry, fmt.Errorf("line `%s` missing `: ` separator", line)
	}

	label = strings.TrimSpace(label)

	if len(label) == 0 {
		return entry, fmt.Errorf("line `%s` contains empty label", line)
	}

	entry.label = label
	if err := entry.loadMeasures(measures); err != nil {
		return entry, err
	}

	return
}

type Record struct {
	label    string
	absValue MeasureValue
}

func newRecord(
	entry Entry,
	group *units.UnitGroup,
) (record Record, err error) {
	record.label = entry.label

	measure_value, err := newMeasureValueFromEntry(entry.measures, group)
	if err != nil {
		return record, err
	}
	record.absValue = measure_value

	return
}

type RecordSlice []*Record

type Enlistment struct {
	records RecordSlice
	// scale     MeasureValue
	// unitRef   string
	ref *Record
	// units     units.UnitsSlice
	group *units.UnitGroup
}

func NewEnlistmentDefault() *Enlistment {
	return &Enlistment{
		records: make(RecordSlice, 0, 32),
	}
}

func (e *Enlistment) sort() {
	slices.SortFunc(e.records, func(a, b *Record) int {
		if a.absValue > b.absValue {
			return 1
		} else if a.absValue < b.absValue {
			return -1
		} else {
			return 0
		}
	})
}

func (e *Enlistment) addRecord(record Record) error {
	e.records = append(e.records, &record)

	if e.ref == nil || e.ref.absValue < record.absValue {
		e.ref = &record
	}

	return nil
}

const CommentPrefix = "#"

func (e *Enlistment) loadFromReader(
	reader io.Reader,
	registry units.UnitRegistry,
) error {
	scanner := bufio.NewScanner(reader)

	var group *units.UnitGroup = nil
	var ok bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || strings.HasPrefix(line, CommentPrefix) {
			continue
		}

		entry, err := newEntry(line)
		if err != nil {
			return err
		}

		if group == nil {
			alias := entry.getFirstUnit()
			group, ok = registry.Find(alias)

			if !ok {
				return fmt.Errorf("alias '%s' from %s not found", alias, line)
			}
		}

		record, err := newRecord(*entry, group)
		if err != nil {
			return err
		}

		if err := e.addRecord(record); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	e.sort()

	return nil
}

func NewEnlistment(
	reader io.Reader,
	units units.UnitRegistry,
) (enlistment *Enlistment, err error) {
	enlistment = NewEnlistmentDefault()
	err = enlistment.loadFromReader(reader, units)
	return enlistment, err
}

func NewEnlistmentFromFile(
	fsys fs.FS,
	filename string,
	unit_files units.UnitRegistry,
) (enlistment *Enlistment, err error) {
	file, err := fsys.Open(filename)
	if err != nil {
		return enlistment, err
	}
	defer file.Close()

	return NewEnlistment(file, unit_files)
}
