package refscaler

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"slices"
	"strconv"
	"strings"

	"github.com/grzadr/refscaler/units"
)

type RawMeasure struct {
	value float64
	alias string
}

func newRawMeasure(raw string) (rawMeasure RawMeasure, err error) {
	value, alias, found := strings.Cut(strings.TrimSpace(raw), " ")

	if !found {
		return RawMeasure{}, fmt.Errorf("raw measure '%s' is malformed", raw)
	}

	if len(value) == 0 {
		return RawMeasure{}, fmt.Errorf(
			"raw measure '%s' missing value",
			raw,
		)
	}

	if len(alias) == 0 {
		return RawMeasure{}, fmt.Errorf(
			"raw measure '%s' missing unit alias",
			raw,
		)
	}

	numValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return RawMeasure{}, fmt.Errorf(
			"raw measure '%s' value failed to be parsed: %w",
			raw,
			err,
		)
	}

	rawMeasure.value = numValue
	rawMeasure.alias = alias

	return
}

type RawMeasureSlice []RawMeasure

func newRawMeasureSlice(measures string) (rawSlice RawMeasureSlice, err error) {
	rawMeasures := strings.Split(measures, ",")

	if len(rawMeasures) == 0 {
		return rawSlice, fmt.Errorf("measures are empty")
	}

	rawSlice = make(RawMeasureSlice, 0, len(rawMeasures))

	for _, r := range rawMeasures {
		raw, err := newRawMeasure(r)
		if err != nil {
			return nil, err
		}

		rawSlice = append(rawSlice, raw)
	}

	return
}

func (r *RawMeasureSlice) getFirstUnitLabel() string {
	return (*r)[0].alias
}

type MeasureValue float64

func newMeasureFromSlice(
	measures RawMeasureSlice,
	group *units.UnitGroup,
) (measure MeasureValue, err error) {
	for _, raw := range measures {
		unit, ok := group.Get(raw.alias)

		if !ok {
			return measure, fmt.Errorf(
				"alias '%s' not found",
				raw.alias,
			)
		}

		measure += MeasureValue(raw.value * unit.Multiplier)
	}

	if measure == 0 {
		return 0, fmt.Errorf("value cannot equal 0")
	}

	return
}

func newMeasureValue(
	measures string,
	group *units.UnitGroup,
) (measure MeasureValue, err error) {
	rawMeasures, err := newRawMeasureSlice(measures)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to create measure value from '%s': %w",
			measures,
			err,
		)
	}

	measure, err = newMeasureFromSlice(rawMeasures, group)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to create measure value from '%s': %w",
			measures,
			err,
		)
	}

	return
}

type Entry struct {
	label    string
	measures string
	line     string
}

func splitEntryLine(line string) (label, measures string, err error) {
	label, measures, found := strings.Cut(line, ": ")

	if !found {
		return "", "", fmt.Errorf("line '%s' missing ': ' separator", line)
	}

	if len(label) == 0 {
		return "", "", fmt.Errorf("line '%s' missing label", line)
	}

	if len(measures) == 0 {
		return "", "", fmt.Errorf("line '%s' missing value", line)
	}

	return
}

func newEntry(line string) (entry Entry, err error) {
	label, measures, err := splitEntryLine(line)
	if err != nil {
		return Entry{}, fmt.Errorf("malformed line '%s': %w", line, err)
	}

	entry.label = label
	entry.measures = measures
	entry.line = line

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

	measure_value, err := newMeasureValue(entry.measures, group)
	if err != nil {
		return record, err
	}
	record.absValue = measure_value

	return
}

type RecordSlice []*Record

type Enlistment struct {
	records RecordSlice
	ref *Record
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
			return -1
		} else if a.absValue < b.absValue {
			return 1
		} else {
			return 0
		}
	})
}

func (e *Enlistment) addRecord(entry Entry) error {
	record, err := newRecord(entry, e.group)
	if err != nil {
		return err
	}

	e.records = append(e.records, &record)

	if e.ref == nil || e.ref.absValue < record.absValue {
		e.ref = &record
	}

	return nil
}

const CommentPrefix = "#"

func iterLines(scanner *bufio.Scanner) iter.Seq2[Entry, error] {
	return func(yield func(Entry, error) bool) {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if len(line) == 0 || strings.HasPrefix(line, CommentPrefix) {
				continue
			}
			entry, err := newEntry(line)

			if err != nil && !yield(Entry{}, err) {
				return
			}
			if !yield(entry, nil) {
				return
			}
		}
	}
}

func determineUnitGroup(
	entry Entry,
	registry units.UnitRegistry,
) (group *units.UnitGroup, err error) {
	measures, err := newRawMeasureSlice(entry.measures)
	if err != nil {
		return nil, err
	}

	alias := measures.getFirstUnitLabel()

	group, ok := registry.Find(alias)

	if !ok {
		return nil, fmt.Errorf(
			"failed to determine unit group for alias '%s'",
			alias,
		)
	}

	return
}

func (e *Enlistment) loadFromReader(
	reader io.Reader,
	registry units.UnitRegistry,
) error {
	scanner := bufio.NewScanner(reader)

	var entry Entry
	var err error
	var ok bool

	next, stop := iter.Pull2(iterLines(scanner))
	defer stop()

	entry, err, ok = next()

	if !ok {
		return fmt.Errorf("enlistment is empty")
	}

	if err != nil {
		return err
	}

	group, err := determineUnitGroup(entry, registry)
	if err != nil {
		return err
	}

	e.group = group

	if err := e.addRecord(entry); err != nil {
		return fmt.Errorf("failed to add entry '%s': %w", entry.line, err)
	}

	for {
		entry, err, ok = next()

		if !ok {
			break
		}

		if err != nil {
			return err
		}

		if err := e.addRecord(entry); err != nil {
			return fmt.Errorf("failed to add entry '%s': %w", entry.line, err)
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
	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	return NewEnlistment(file, unit_files)
}
