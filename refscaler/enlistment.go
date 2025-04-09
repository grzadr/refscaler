package refscaler

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"io/fs"
	"github.com/grzadr/refscaler/units"
)

type MeasureValue float64

type EntryMeasure struct {
	Unit  string
	Value MeasureValue
}

func newEntryMeasure(str string) (measure EntryMeasure, err error) {
	raw_value, unit, found := strings.Cut(strings.TrimSpace(str), " ")
	raw_value = strings.TrimSpace(raw_value)
	unit = strings.TrimSpace(unit)

	if !found || len(raw_value) == 0 || len(unit) == 0 {
		return measure, fmt.Errorf("measure '%s' has wrong format", str)
	}

	value, err := strconv.ParseFloat(raw_value, 64)
	if err != nil {
		return measure, fmt.Errorf(
			"value '%s' cannot be parsed to float",
			raw_value,
		)
	}

	measure.Unit = unit
	measure.Value = MeasureValue(value)

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
	record.absValue = 1

	for _, measure := range entry.measures {
		unit, ok := group.Get(measure.Unit)

		if !ok {
			return record, fmt.Errorf(
				"unit alias '%s' not found in group",
				measure.Unit,
			)
		}

		record.absValue *= measure.Value * MeasureValue(unit.Multiplier)
	}

	return
}

type RecordSlice []Record

type Enlistment struct {
	records RecordSlice
	// scale     MeasureValue
	// unitRef   string
	ref *Record
	// units     units.UnitsSlice
}

func NewEnlistmentDefault() *Enlistment {
	return &Enlistment{
		records: make(RecordSlice, 0, 32),
	}
}

func (e *Enlistment) addRecord(record Record) error {
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
