package units

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"slices"

	"github.com/grzadr/refscaler/units/unit_entry"
	"github.com/grzadr/refscaler/walkentry"
)

type Unit struct {
	Name       string
	Multiplier float64
}

type (
	UnitsSlice  []*Unit
	UnitAliases map[string]*Unit
)

type UnitGroup struct {
	units   UnitsSlice
	aliases UnitAliases
	// allowPrefix bool // TODO
	baseUnit *Unit
}

func (g *UnitGroup) add(entry unit_entry.UnitEntry) error {
	unit := &Unit{
		Name:       entry.Name,
		Multiplier: entry.Value,
	}

	if entry.IsBase() {
		g.baseUnit = unit
	}

	g.units = append(g.units, unit)
	g.aliases[unit.Name] = unit

	for _, a := range entry.Aliases {
		g.aliases[a] = unit
	}

	slices.SortFunc(g.units, func(a *Unit, b *Unit) int {
		if a.Multiplier < b.Multiplier {
			return -1
		} else if a.Multiplier > b.Multiplier {
			return 1
		} else {
			return 0
		}
	})

	return nil
}

func (g *UnitGroup) Get(alias string) (unit *Unit, ok bool) {
	unit, ok = g.aliases[alias]
	return
}

func (g *UnitGroup) Length() int {
	return len(g.units)
}

func newUnitGroupDefault() *UnitGroup {
	return &UnitGroup{
		units:    make(UnitsSlice, 0, 32),
		aliases:  make(UnitAliases, 128),
		baseUnit: &Unit{},
	}
}

func NewUnitGroup(unitsData io.Reader) (group *UnitGroup, err error) {
	group = newUnitGroupDefault()
	for entry, err := range unit_entry.IterUnitEntries(unitsData) {
		if err != nil {
			return group, fmt.Errorf("error reading unit entry: %s", err)
		}
		if err := group.add(entry); err != nil {
			return group, fmt.Errorf(
				"error adding unit entry %v: %s",
				entry,
				err,
			)
		}
	}
	return group, nil
}

type UnitRegistry interface {
	Find(alias string) (group *UnitGroup, ok bool)
	Add(key string, group *UnitGroup)
}

type UnitRegistryFiles map[string]*UnitGroup

func NewUnitRegistryFilesDefault() UnitRegistryFiles {
	// temp := make(UnitRegistryFiles, 16)
	// return &temp
	return make(UnitRegistryFiles, 16)
}

func (r *UnitRegistryFiles) Length() int {
	return len(*r)
}

func (r *UnitRegistryFiles) Add(key string, group *UnitGroup) {
	(*r)[key] = group
}

func (r *UnitRegistryFiles) Find(alias string) (group *UnitGroup, ok bool) {
	return
}

func loadUnitGroupFromJsonFile(
	fsys fs.FS,
	json_path string,
) (group *UnitGroup, err error) {
	file, err := fsys.Open(json_path)
	if err != nil {
		return group, err
	}

	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	group, err = NewUnitGroup(file)
	return
}

func NewUnitRegistryFiles(
	fsys fs.FS,
	dir_path string,
) (registry UnitRegistryFiles, err error) {
	registry = NewUnitRegistryFilesDefault()

	for walk_entry, err := range walkentry.WalkFS(fsys, dir_path) {
		if err != nil {
			return registry, err
		}

		if !walk_entry.IsJSONFile() {
			continue
		}

		unit_group, err := loadUnitGroupFromJsonFile(fsys, walk_entry.Path)
		if err != nil {
			return registry, err
		}

		if unit_group == nil {
			continue
		}

		registry.Add(walk_entry.Name, unit_group)
	}

	return registry, err
}

//go:embed units_db/*.json
var unitsFS embed.FS

const UNITS_PATH = "units_db"

var EmbeddedUnitRegistry *UnitRegistryFiles = nil

func newEmbeddedUnitRegistry() (registry UnitRegistryFiles, err error) {
	return NewUnitRegistryFiles(unitsFS, UNITS_PATH)
}

func init() {
	registry, err := newEmbeddedUnitRegistry()
	if err != nil {
		panic(err)
	}

	EmbeddedUnitRegistry = &registry
}
