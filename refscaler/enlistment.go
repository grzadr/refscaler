package refscaler

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/grzadr/refscaler/units"
)

type MeasureValue float64

type EntryMeasure struct {
	Unit  string
	Value MeasureValue
}

func newEntryMeasure(str string) (measure EntryMeasure, err error) {
	value, unit, found := strings.Cut(strings.TrimSpace(str), " ")

	if !found || len(value) == 0 || len(unit) == 0 {
	}
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
				return fmt.Errorf("alias '%s' in %s not found", alias, line)
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
