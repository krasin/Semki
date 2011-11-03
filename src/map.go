package main

import (
	"fmt"
	"os"
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

var Dirs = []Direction{North, East, South, West}

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

func NewItems(t Torus) *Items {
	return &Items{
		At: make([][]*Item, t.Size()),
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

func (it *Items) HasMyHillAt(loc Location) bool {
	for _, item := range it.At[loc] {
		if item.What == Hill && item.Owner == Me {
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

	Path   Path
	Target Location
	Score  int
}

func (a *MyAnt) Loc(turn int) Location {
	return a.Locs[turn-a.BornAt]
}

func (a *MyAnt) HasLoc(turn int) bool {
	return turn-a.BornAt >= 0 &&
		turn-a.BornAt < len(a.Locs)
}

func (a *MyAnt) NewTurn(turn int) {
	if !a.HasLoc(turn) {
		a.Locs = append(a.Locs, a.Locs[len(a.Locs)-1])
	}
}

func (a *MyAnt) String() string {
	return fmt.Sprintf("(%v)", a.Locs[len(a.Locs)-1])
}

type MyAntIndex struct {
	myAnts []*MyAnt
	s      LocSet
}

func NewMyAntIndex(size int) *MyAntIndex {
	return &MyAntIndex{
		myAnts: make([]*MyAnt, size),
		s:      NewLocSet(size),
	}
}

func (in *MyAntIndex) Add(loc Location, ant *MyAnt) {
	in.s.Add(loc)
	in.myAnts[loc] = ant
}

func (in *MyAntIndex) At(loc Location) *MyAnt {
	if in.s.Has(loc) {
		return in.myAnts[loc]
	}
	return nil
}

func (in *MyAntIndex) Clear() {
	in.s.Clear()
}

type Map struct {
	T               Torus
	Terrain         []Terrain
	Items           []*Items
	ViewMaskRow     []int
	ViewMaskCol     []int
	MyAnts          []*MyAnt
	MyLiveAnts      []*MyAnt
	MyLiveAntsIndex *MyAntIndex
	LastVisited     []int
	Next            *Items
	NewCells        []Location
}

func NewMap(t Torus, viewRadius2 int) (m *Map) {
	m = &Map{
		T:               t,
		Terrain:         make([]Terrain, t.Size()),
		Items:           []*Items{NewItems(t)},
		LastVisited:     make([]int, t.Size()),
		MyLiveAntsIndex: NewMyAntIndex(t.Size()),
	}
	m.GenerateViewMask(viewRadius2)
	return m
}

func (m *Map) Turn() int {
	return len(m.Items) - 1
}

func (m *Map) Food() (res []Location) {
	for _, item := range m.Items[m.Turn()].All {
		if item.What == Food {
			res = append(res, item.Loc)
		}
	}
	return
}

func (m *Map) Enemy() (res []Location) {
	for _, item := range m.Items[m.Turn()].All {
		if item.What == Ant && item.Owner != Me {
			res = append(res, item.Loc)
		}
	}
	return
}

func (m *Map) MyHills() (hills []*Item) {
	for _, item := range m.Items[m.Turn()].All {
		if item.What == Hill && item.Owner == Me {
			hills = append(hills, item)
		}
	}
	return
}

func (m *Map) EnemyHills() (hills []*Item) {
	for _, item := range m.Items[m.Turn()].All {
		if item.What == Hill && item.Owner != Me {
			hills = append(hills, item)
		}
	}
	return
}

func (m *Map) UpdateLiveAnts() {
	m.MyLiveAntsIndex.Clear()
	// Find dead ants and remove them from MyLiveAnts
	var dead []int
	items := m.Items[m.Turn()]
	for i, ant := range m.MyLiveAnts {
		ant.NewTurn(m.Turn())
		if !items.HasAntAt(ant.Loc(m.Turn()), Me) {
			dead = append(dead, i)
		}
		m.MyLiveAntsIndex.Add(ant.Loc(m.Turn()), ant)
	}
	for i := len(dead) - 1; i >= 0; i-- {
		m.MyLiveAnts[dead[i]].Alive = false
		m.MyLiveAnts[dead[i]].DiedAt = m.Turn()
		m.MyLiveAnts[dead[i]] = m.MyLiveAnts[len(m.MyLiveAnts)-1]
		m.MyLiveAnts = m.MyLiveAnts[:len(m.MyLiveAnts)-1]
	}

	// Find newly born ants. They born in hills
	for _, hill := range m.MyHills() {
		if items.HasAntAt(hill.Loc, Me) &&
			m.MyLiveAntAt(hill.Loc) == nil {
			ant := &MyAnt{
				BornAt: m.Turn(),
				Alive:  true,
				Locs:   []Location{hill.Loc},
			}
			m.MyAnts = append(m.MyAnts, ant)
			m.MyLiveAnts = append(m.MyLiveAnts, ant)
			m.MyLiveAntsIndex.Add(ant.Loc(m.Turn()), ant)
		}
	}

}

func (m *Map) Update(input []Input) {
	m.Items = append(m.Items, NewItems(m.T))
	for _, in := range input {
		loc := m.T.Loc(in.Row, in.Col)
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
	m.Next = NewItems(m.T)
	m.UpdateLiveAnts()
	m.UpdateVisibility()
	m.UpdateLastVisited()
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
	m.NewCells = m.NewCells[:0]
	for _, ant := range m.MyLiveAnts {
		for i := 0; i < len(m.ViewMaskRow); i++ {
			loc2 := m.T.ShiftLoc(ant.Loc(m.Turn()), m.ViewMaskRow[i], m.ViewMaskCol[i])
			if m.Terrain[loc2] == Unknown {
				m.Terrain[loc2] = Land
				m.NewCells = append(m.NewCells, loc2)
			}
		}
	}
}

func (m *Map) UpdateLastVisited() {
	for _, ant := range m.MyLiveAnts {
		m.LastVisited[ant.Loc(m.Turn())] = m.Turn()
	}
}

func (m *Map) CanMove(loc Location, d Direction) bool {
	newLoc := m.T.NewLoc(loc, d)
	if m.Terrain[newLoc] == Water {
		return false
	}
	if !m.Next.CanEnter(newLoc) {
		return false
	}
	ant := m.MyLiveAntAt(newLoc)
	if ant != nil {
		fmt.Fprintf(os.Stderr, "CanMove(%d, %c) is false because of the ant at: %d\n", loc, d, newLoc)
		return false
	}
	return m.Items[m.Turn()].CanEnter(newLoc)
}

func (m *Map) Move(ant *MyAnt, d Direction) {
	newLoc := m.T.NewLoc(ant.Loc(m.Turn()), d)
	m.Next.Add(newLoc, &Item{What: Ant, Owner: Me, Loc: newLoc})
	ant.Locs = append(ant.Locs, newLoc)
}

func (m *Map) Discovered(loc Location) bool {
	return m.Terrain[loc] != Unknown
}

func (m *Map) LandNeighbours(loc Location) (res []Location) {
	for _, dir := range Dirs {
		cell := m.T.NewLoc(loc, dir)
		if m.Terrain[cell] == Land {
			res = append(res, cell)
		}
	}
	return
}

func (m *Map) MoveAnts() {
	m.ResolveConflicts()
	//	fmt.Fprintf(os.Stderr, "MyLiveAnts: %v\n", m.MyLiveAnts)
	for _, ant := range m.MyLiveAnts {
		if ant.Path == nil {
			continue
		}
		dir := ant.Path.Dir(0)
		if m.CanMove(ant.Loc(m.Turn()), dir) {
			m.Move(ant, dir)
			ant.Path.Advance(1)
		} else {
			//			fmt.Fprintf(os.Stderr, "Can't move, dir: %c, ant loc: %d\n", dir, ant.Loc(m.Turn()))
		}
	}
}

func (m *Map) Conn(loc Location) (res []Location) {
	return m.LandNeighbours(loc)
}

func (m *Map) ResolveConflicts() {
	q := make([]*MyAnt, len(m.MyLiveAnts))
	copy(q, m.MyLiveAnts)
	var q2 []*MyAnt
	for len(q) > 0 {
		tmpQ := q2
		q2 = q
		q = tmpQ[:0]

		for _, ant := range q2 {
			//			fmt.Fprintf(os.Stderr, "ResolveConflicts for ant at %v, ", ant.Loc(m.Turn()))
			if ant.Path == nil {
				//				fmt.Fprintf(os.Stderr, "Path == nil\n")
				continue
			}
			if ant.Path.Len() == 0 {
				//				fmt.Fprintf(os.Stderr, "Path is empty\n")
				ant.Path = nil
				continue
			}
			dir := ant.Path.Dir(0)
			//			fmt.Fprintf(os.Stderr, "dir = %c ", dir)
			to := m.T.NewLoc(ant.Loc(m.Turn()), dir)
			ant2 := m.MyLiveAntAt(to)
			if ant == ant2 {
				panic(fmt.Sprintf("ant == ant2! loc: %d, newLoc: %d, dir: %c", ant.Loc(m.Turn()), to, dir))
			}
			if ant2 == nil {
				//				fmt.Fprintf(os.Stderr, "ant2 == nil\n")
				// FIXME: do something with the case when two ants want to enter one cell
				continue
			}
			if ant2.Path != nil && ant2.Path.Len() == 0 {
				ant2.Path = nil
			}
			if ant2.Path == nil {
				//				fmt.Fprintf(os.Stderr, "ant2.Path == nil. Success\n")
				tmpPath := ant.Path
				ant.Path = ant2.Path
				ant2.Path = tmpPath
				tmpTarget := ant.Target
				ant.Target = ant2.Target
				ant2.Target = tmpTarget
				tmpScore := ant.Score
				ant.Score = ant2.Score
				ant2.Score = tmpScore

				ant2.Path.Advance(1)
				if ant2.Path.Len() == 0 {
					ant2.Path = nil
				}
				q = append(q, ant)
				q = append(q, ant2)
				continue
			}
			dir2 := ant2.Path.Dir(0)
			if !ant2.HasLoc(m.Turn()) {
				panic("ant2 does not have current loc!")
			}
			to2 := m.T.NewLoc(ant2.Loc(m.Turn()), dir2)
			if to2 == ant.Loc(m.Turn()) {
				//				fmt.Fprintf(os.Stderr, "*-><-* case. Success!\n")
				if ant.Path == nil {
					panic("ant.Path == nil!")
				}
				tmpPath := ant.Path
				ant.Path = ant2.Path
				ant2.Path = tmpPath
				if ant2.Path == nil {
					panic(fmt.Sprintf("1. ant2.Path == nil! tmpPath: %v", tmpPath))
				}
				tmpTarget := ant.Target
				ant.Target = ant2.Target
				ant2.Target = tmpTarget
				tmpScore := ant.Score
				ant.Score = ant2.Score
				ant2.Score = tmpScore

				ant.Path.Advance(1)
				if ant.Path.Len() == 0 {
					ant.Path = nil
				}
				if ant2.Path == nil {
					panic(fmt.Sprintf("2. ant2.Path == nil! tmpPath: %v", tmpPath))
				}
				ant2.Path.Advance(1)
				if ant2.Path.Len() == 0 {
					ant2.Path = nil
				}
				q = append(q, ant)
				q = append(q, ant2)
				continue
			}
			//			fmt.Fprintf(os.Stderr, "Unknown case, ant2 at %v, dir2: %c \n", ant2.Loc(m.Turn()), dir2)
		}
	}
}

func (m *Map) HasMyHillAt(loc Location) bool {
	return m.Items[m.Turn()].HasMyHillAt(loc)
}

func (m *Map) MyLiveAntAt(loc Location) *MyAnt {
	res := m.MyLiveAntsIndex.At(loc)
	if res == nil {
		return nil
	}
	if res.Loc(m.Turn()) != loc {
		panic(fmt.Sprintf("MyLiveAntAt lies: loc: %d, res: %v", loc, *res))
	}
	return res
}
