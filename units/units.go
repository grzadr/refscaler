package units

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"maps"
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
	// baseUnit *Unit
}

func (g *UnitGroup) add(entry unit_entry.UnitEntry) error {
	unit := &Unit{
		Name:       entry.Name,
		Multiplier: entry.Value,
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

func (g *UnitGroup) IterBackward() iter.Seq[*Unit] {
	return func(yield func(*Unit) bool) {
		for _, u := range slices.Backward(g.units) {
			if !yield(u) {
				return
			}
		}
	}
}

func (g *UnitGroup) Length() int {
	return len(g.units)
}

type UnitJSON struct {
	Name    string   `json:"name"`
	Value   float64  `json:"value"`
	Aliases []string `json:"aliases"`
}

func (u *UnitJSON) AddAlias(alias string) {
	u.Aliases = append(u.Aliases, alias)
}

type UnitGroupJSON []UnitJSON

func (g *UnitGroup) Serialize() UnitGroupJSON {
	json_units := make(UnitGroupJSON, 0, g.Length())
	visited_units := make(map[string]*UnitJSON, g.Length())

	for _, unit := range g.units {
		temp := UnitJSON{
			Name:    unit.Name,
			Value:   unit.Multiplier,
			Aliases: make([]string, 0, 4),
		}
		json_units = append(json_units, temp)

		visited_units[unit.Name] = &json_units[len(json_units)-1]
	}

	for alias, unit := range g.aliases {
		name := unit.Name

		if alias == unit.Name {
			continue
		}

		visited_units[name].AddAlias(alias)
	}

	for i := range json_units {
		slices.Sort(json_units[i].Aliases)
	}

	return json_units
}

func newUnitGroupDefault() *UnitGroup {
	return &UnitGroup{
		units:   make(UnitsSlice, 0, 32),
		aliases: make(UnitAliases, 128),
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

type UnitRegistryJSON map[string]UnitGroupJSON

type UnitRegistry interface {
	Find(alias string) (group *UnitGroup, ok bool)
	Add(key string, group *UnitGroup)
	Serialize() UnitRegistryJSON
	ToJSON() (string, error)
}

type UnitRegistryFiles map[string]*UnitGroup

func NewUnitRegistryFilesDefault() UnitRegistryFiles {
	return make(UnitRegistryFiles, 16)
}

func (r *UnitRegistryFiles) Length() int {
	return len(*r)
}

func (r *UnitRegistryFiles) Add(key string, group *UnitGroup) {
	(*r)[key] = group
}

func (r *UnitRegistryFiles) Find(alias string) (group *UnitGroup, ok bool) {
	for group = range maps.Values(*r) {
		_, ok = group.Get(alias)

		if ok {
			return
		}
	}

	return nil, false
}

func (r *UnitRegistryFiles) Serialize() UnitRegistryJSON {
	serialized := make(map[string]UnitGroupJSON, len(*r))

	for name, group := range *r {
		serialized[name] = group.Serialize()
	}

	return serialized
}

func (r *UnitRegistryFiles) ToJSON() (string, error) {
	str, err := json.MarshalIndent(r.Serialize(), "", "  ")
	if err != nil {
		return "", err
	}

	return string(str), nil
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

var (
	EmbeddedUnitRegistry     *UnitRegistryFiles = nil
	EmbeddedUnitRegistryJSON                    = ""
)

func newEmbeddedUnitRegistry() (registry UnitRegistryFiles, err error) {
	return NewUnitRegistryFiles(unitsFS, UNITS_PATH)
}

func init() {
	registry, err := newEmbeddedUnitRegistry()
	if err != nil {
		panic(err)
	}

	EmbeddedUnitRegistry = &registry

	EmbeddedUnitRegistryJSON, err = registry.ToJSON()
	if err != nil {
		panic(err)
	}
}
