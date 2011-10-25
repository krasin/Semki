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

func (it *Items) HasAntAt(loc Location, owner int) bool {
	for _, item := range it.At[loc] {
		if item.What == Ant && item.Owner == owner {
			return true
		}
	}
	return false
}

type MyAnt struct {
	Locs   []Location
	BornAt int
	DiedAt int
	Alive  bool
}

func (a *MyAnt) Loc(turn int) Location {
	return a.Locs[turn-a.BornAt]
}

func (a *MyAnt) NewTurn(turn int) {
	if a.BornAt+len(a.Locs) < turn+1 {
		a.Locs = append(a.Locs, a.Locs[len(a.Locs)-1])
	}
}

type Map struct {
	Terrain     []Terrain
	Items       []*Items
	Rows        int
	Cols        int
	ViewMaskRow []int
	ViewMaskCol []int
	MyAnts      []*MyAnt
	MyLiveAnts  []*MyAnt
	Next        *Items
}

func NewMap(rows, cols int, viewRadius2 int) (m *Map) {
	m = &Map{
		Rows:    rows,
		Cols:    cols,
		Terrain: make([]Terrain, cols*rows),
		Items:   []*Items{NewItems(rows, cols)},
	}
	m.GenerateViewMask(viewRadius2)
	return m
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

func (m *Map) MyHills() (hills []*Item) {
	for _, item := range m.Items[m.Turn()].All {
		if item.What == Hill && item.Owner == Me {
			hills = append(hills, item)
		}
	}
	return
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

func ShiftLoc(loc Location, rowShift, colShift, rows, cols int) Location {
	row := Row(loc, cols)
	col := Col(loc, cols)
	return Loc((row+rowShift+rows)%rows, (col+colShift+cols)%cols, cols)
}

func (m *Map) UpdateLiveAnts() {
	// Find dead ants and remove them from MyLiveAnts
	var dead []int
	items := m.Items[m.Turn()]
	for i, ant := range m.MyLiveAnts {
		ant.NewTurn(m.Turn())
		if !items.HasAntAt(ant.Loc(m.Turn()), Me) {
			dead = append(dead, i)
		}
	}
	for i := len(dead) - 1; i >= 0; i-- {
		m.MyLiveAnts[dead[i]].Alive = false
		m.MyLiveAnts[dead[i]].DiedAt = m.Turn()
		m.MyLiveAnts[dead[i]] = m.MyLiveAnts[len(m.MyLiveAnts)-1]
		m.MyLiveAnts = m.MyLiveAnts[:len(m.MyLiveAnts)-1]
	}

	// Find newly born ants. They born in hills
	for _, hill := range m.MyHills() {
		if items.HasAntAt(hill.Loc, Me) {
			ant := &MyAnt{
				BornAt: m.Turn(),
				Alive:  true,
				Locs:   []Location{hill.Loc},
			}
			m.MyAnts = append(m.MyAnts, ant)
			m.MyLiveAnts = append(m.MyLiveAnts, ant)
		}
	}
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
		case Food:
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
	m.UpdateLiveAnts()
	m.UpdateVisibility()
}

func (m *Map) GenerateViewMask(viewRadius2 int) {
	for i := 0; i*i <= viewRadius2; i++ {
		for j := 0; j*j+i*i <= viewRadius2; j++ {
			m.ViewMaskRow = append(m.ViewMaskRow, i)
			m.ViewMaskCol = append(m.ViewMaskCol, j)
			if i > 0 {
				m.ViewMaskRow = append(m.ViewMaskRow, -i)
				m.ViewMaskCol = append(m.ViewMaskCol, j)
			}
			if j > 0 {
				m.ViewMaskRow = append(m.ViewMaskRow, i)
				m.ViewMaskCol = append(m.ViewMaskCol, -j)
			}
			if i > 0 && j > 0 {
				m.ViewMaskRow = append(m.ViewMaskRow, -i)
				m.ViewMaskCol = append(m.ViewMaskCol, -j)
			}
		}
	}
}

func (m *Map) UpdateVisibility() {
	for _, ant := range m.MyLiveAnts {
		for i := 0; i < len(m.ViewMaskRow); i++ {
			loc2 := ShiftLoc(ant.Loc(m.Turn()), m.ViewMaskRow[i], m.ViewMaskCol[i], m.Rows, m.Cols)
			if m.Terrain[loc2] == Unknown {
				m.Terrain[loc2] = Land
			}
		}
	}
}

func (m *Map) CanMove(loc Location, d Direction) bool {
	newLoc := m.NewLoc(loc, d)
	return m.Terrain[newLoc] != Water &&
		m.Items[m.Turn()].CanEnter(newLoc) &&
		m.Next.CanEnter(newLoc)
}

func (m *Map) Move(ant *MyAnt, d Direction) {
	newLoc := m.NewLoc(ant.Loc(m.Turn()), d)
	m.Next.Add(newLoc, &Item{What: Ant, Owner: Me, Loc: newLoc})
	ant.Locs = append(ant.Locs, newLoc)
}
