package main

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

func (it *Items) HasHillAt(loc Location) bool {
	for _, item := range it.At[loc] {
		if item.What == Hill {
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

func (a *MyAnt) HasLoc(turn int) bool {
	return turn-a.BornAt >= 0 &&
		turn-a.BornAt < len(a.Locs)
}

func (a *MyAnt) NewTurn(turn int) {
	if !a.HasLoc(turn) {
		a.Locs = append(a.Locs, a.Locs[len(a.Locs)-1])
	}
}

type Map struct {
	T           Torus
	Terrain     []Terrain
	Items       []*Items
	ViewMaskRow []int
	ViewMaskCol []int
	MyAnts      []*MyAnt
	MyLiveAnts  []*MyAnt
	LastVisited []int
	Next        *Items
}

func NewMap(t Torus, viewRadius2 int) (m *Map) {
	m = &Map{
		T:           t,
		Terrain:     make([]Terrain, t.Size()),
		Items:       []*Items{NewItems(t)},
		LastVisited: make([]int, t.Size()),
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
	for _, ant := range m.MyLiveAnts {
		for i := 0; i < len(m.ViewMaskRow); i++ {
			loc2 := m.T.ShiftLoc(ant.Loc(m.Turn()), m.ViewMaskRow[i], m.ViewMaskCol[i])
			if m.Terrain[loc2] == Unknown {
				m.Terrain[loc2] = Land
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
	return m.Terrain[newLoc] != Water &&
		m.Items[m.Turn()].CanEnter(newLoc) &&
		m.Next.CanEnter(newLoc)
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
	f := func(cell Location) {
		if m.Terrain[cell] == Land {
			res = append(res, cell)
		}
	}
	for _, dir := range Dirs {
		f(m.T.NewLoc(loc, dir))
	}
	return
}

func (m *Map) HasHillAt(loc Location) bool {
	return m.Items[m.Turn()].HasHillAt(loc)
}

func (m *Map) MyLiveAntAt(loc Location) *MyAnt {
	// FIXME: This should have O(1) complexity, instead of O(n)
	for _, ant := range m.MyLiveAnts {
		if ant.Loc(m.Turn()) == loc {
			return ant
		}
	}
	return nil
}
