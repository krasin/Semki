package main

type Location uint64

type ItemType int
type Terrain int

const (
	Unknown = 0
	Land    = '.'
	Water   = 'w'
	Food    = 'f'
	Hill    = 'h'
	Ant     = 'a'
	DeadAnt = 'd'
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

type Map struct {
	Terrain []Terrain
	Items   []Items
	Rows    int
	Cols    int
	Turn    int
	Next    Items
}

func NewMap(rows, cols int) *Map {
	return &Map{
		Rows:    rows,
		Cols:    cols,
		Terrain: make([]Terrain, cols*rows),
		Items:   make([]Items, 1),
	}
}
