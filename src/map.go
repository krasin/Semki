package main

import (
	"fmt"
)

type Location int
type ItemType int
type Terrain int
type Direction int

const (
	Unknown = 0
	Land    = '.'
	Water   = 'w'
	Food    = 'f'
	Hill    = 'h'
	Ant     = 'a'
	DeadAnt = 'd'

	North = 'N'
	East  = 'E'
	South = 'S'
	West  = 'W'

	Me = 0
)

type Item struct {
	What  ItemType
	Owner int
	Loc   Location
}

type ItemFeed func() *Item

type Items struct {
	At  [][]*Item
	All []*Item
}

func NewItems(rows, cols int) *Items {
	return &Items{
		At: make([][]*Item, rows*cols),
	}
}

func (it *Items) Add(loc Location, item *Item) {
	it.At[loc] = append(it.At[loc], item)
	it.All = append(it.All, item)
}

func (it *Items) CanEnter(loc Location) bool {
	for _, item := range it.At[loc] {
		if item.What == Ant || item.What == Food {
			return false
		}
	}
	return true
}

type Map struct {
	Terrain []Terrain
	Items   []*Items
	Rows    int
	Cols    int
	Next    *Items
}

func NewMap(rows, cols int) *Map {
	return &Map{
		Rows:    rows,
		Cols:    cols,
		Terrain: make([]Terrain, cols*rows),
		Items:   []*Items{NewItems(rows, cols)},
	}
}

func (m *Map) Turn() int {
	return len(m.Items) - 1
}

func (m *Map) Loc(row, col int) Location {
	return Loc(row, col, m.Cols)
}

func (m *Map) Row(loc Location) int {
	return Row(loc, m.Cols)
}

func (m *Map) Col(loc Location) int {
	return Col(loc, m.Cols)
}

func (m *Map) NewLoc(loc Location, d Direction) Location {
	return NewLoc(loc, d, m.Rows, m.Cols)
}

func Loc(row, col, cols int) Location {
	return Location(row*cols + col)
}

func Row(loc Location, cols int) int {
	return int(loc) / cols
}

func Col(loc Location, cols int) int {
	return int(loc) % cols
}

func NewLoc(loc Location, d Direction, rows, cols int) Location {
	row := Row(loc, cols)
	col := Col(loc, cols)
	switch d {
	case North:
		return Loc((row+rows-1)%rows, col, cols)
	case South:
		return Loc((row+1)%rows, col, cols)
	case West:
		return Loc(row, (col+cols-1)%cols, cols)
	case East:
		return Loc(row, (col+1)%cols, cols)

	}
	panic(fmt.Sprintf("Unknown direction: %d", d))
}

func (m *Map) Update(input []Input) {
	m.Items = append(m.Items, NewItems(m.Rows, m.Cols))
	for _, in := range input {
		loc := m.Loc(in.Row, in.Col)
		switch in.What {
		case Water:
			m.Terrain[loc] = Water
		case Hill:
			fallthrough
		case Ant:
			fallthrough
		case DeadAnt:
			item := &Item{
				What:  ItemType(in.What),
				Loc:   loc,
				Owner: in.Owner,
			}
			m.Items[m.Turn()].Add(loc, item)
		}
	}
	m.Next = NewItems(m.Rows, m.Cols)
}

func (m *Map) MyAnts() (res []*Item) {
	items := m.Items[m.Turn()]
	for _, item := range items.All {
		if item.What == Ant && item.Owner == Me {
			res = append(res, item)
		}

	}
	return
}

func (m *Map) CanMove(loc Location, d Direction) bool {
	newLoc := m.NewLoc(loc, d)
	return m.Terrain[newLoc] != Water &&
		m.Items[m.Turn()].CanEnter(newLoc) &&
		m.Next.CanEnter(newLoc)
}

func (m *Map) Move(loc Location, d Direction) {
	m.Next.At[loc] = append(m.Next.At[loc], &Item{What: Ant, Owner: Me, Loc: loc})
}
