package main

import (
	"fmt"
	//	"io/ioutil"
	"os"
	"rand"
	"time"
)

const FoodScore = 1000000
const VisitScore = 1000
const NeverVisitedScore = 100000
const NeverVisitedScore2 = 50000
const EnemyWithdrawalScore = 4000
const EnemyHillScore = 10000000
const MoveFromMyHillScore = 10000000

const MaxFindNearCount = 30

const GridSize = 8

const XaosP = 0.25

var big = make([]int16, 200*1000*1000)

type MyBot struct {
	p               Params
	t               Torus
	m               *Map
	locsByProv      LocListMap
	locSet          LocSet
	perf            *Timing
	loc             *FairLocator
	pf              PathFinder
	LocatorBudgetMs int
	gridSet         *GridLocatedSet
}

func (b *MyBot) Init(p Params) (err os.Error) {
	rand.Seed(p.PlayerSeed)
	b.p = p
	b.t = Torus{Rows: p.Rows, Cols: p.Cols}
	b.m = NewMap(b.t, p.ViewRadius2)
	b.locsByProv = NewLocListMap(b.t.Size())
	b.locSet = NewLocSet(b.t.Size())
	b.loc = NewFairLocator(b.m, big)
	b.pf = NewPathFinder(b.t, b.m, b.loc)
	b.LocatorBudgetMs = 100
	b.gridSet = NewGridLocatedSet(b.t, b.m, GridSize)
	return nil
}

type GridLocatedSet struct {
	t          Torus
	m          *Map
	k          int
	antsByProv LocListMap
	locSet     LocSet
}

func NewGridLocatedSet(t Torus, m *Map, k int) *GridLocatedSet {
	return &GridLocatedSet{
		t:          t,
		k:          k,
		m:          m,
		antsByProv: NewLocListMap(t.Size()),
		locSet:     NewLocSet(t.Size()),
	}
}

func (s *GridLocatedSet) GridLoc(loc Location) Location {
	row, col := s.t.Row(loc), s.t.Col(loc)
	row = (row / s.k) * s.k
	col = (col / s.k) * s.k
	return s.t.Loc(row, col)
}

func (s *GridLocatedSet) Update() {
	s.antsByProv.Clear()
	for _, ant := range s.m.MyLiveAnts {
		loc := ant.Loc(s.m.Turn())
		s.antsByProv.Add(s.GridLoc(loc), loc)
	}
}

func (s *GridLocatedSet) Conn(loc Location) []Location {
	loc = s.GridLoc(loc)
	return []Location{
		s.t.ShiftLoc(loc, -s.k, 0),
		s.t.ShiftLoc(loc, 0, -s.k),
		s.t.ShiftLoc(loc, s.k, 0),
		s.t.ShiftLoc(loc, 0, s.k),
	}
}

func (s *GridLocatedSet) FindNear(at Location, score int, ok func(Location, int, bool) bool) (Location, bool) {
	s.locSet.Clear()
	start := s.GridLoc(at)

	s.locSet.Add(start)
	q := []Location{start}
	var q2 []Location
	count := 0
	for len(q) > 0 {
		q, q2 = q2[:0], q
		for _, gridLoc := range q2 {
			for _, ant := range s.antsByProv.Get(gridLoc) {
				if ok(ant, score, gridLoc == start) {
					return ant, true
				}
				count++
				if count > MaxFindNearCount {
					return 0, false
				}
			}
			for _, other := range s.Conn(gridLoc) {
				if !s.locSet.Has(other) {
					q = append(q, other)
					s.locSet.Add(other)
				}
			}
		}
	}
	return 0, false
}

func (s *GridLocatedSet) All() (res []Location) {
	res = make([]Location, len(s.m.MyLiveAnts))
	for i, ant := range s.m.MyLiveAnts {
		res[i] = ant.Loc(s.m.Turn())
	}
	return
}

func (b *MyBot) Plan() {
	l := b.loc
	p := NewGreedyPlanner(b.t.Size())
	var workers []Location
	var targets []Location
	var scores []int
	var prev []Assignment

	for _, ant := range b.m.MyLiveAnts {
		workers = append(workers, ant.Loc(b.m.Turn()))
		if ant.Path == nil {
			continue
		}
		loc := ant.Loc(b.m.Turn())
		if loc == ant.Target {
			// The target is reached
			ant.Path = nil
			continue
		}
		prev = append(prev, Assignment{
			Worker: loc,
			Target: ant.Target,
			Score:  ant.Score,
		})
	}

	var addTarget = func(loc Location, score int) {
		targets = append(targets, loc)
		scores = append(scores, score)
	}

	for _, food := range b.m.Food() {
		addTarget(food, FoodScore)
	}

	//	fmt.Fprintf(os.Stderr, "scores: %v\n", scores)
	b.perf.Log("Prepare data for planner")

	plan := p.Plan(l, prev, b.gridSet, targets, scores)
	b.perf.Log("Planner")
	fmt.Fprintf(os.Stderr, "plan = %v\n", plan)
	for _, assign := range plan {
		ant := b.m.MyLiveAntAt(assign.Worker)
		if ant == nil {
			panic("ant == nil")
		}
		if ant.Target == assign.Target {
			continue
		}
		ant.Path = b.pf.Path(ant.Loc(b.m.Turn()), assign.Target)
		fmt.Fprintf(os.Stderr, "path: %v\n", ant.Path)
		//fmt.Fprintf(os.Stderr, "p2  : %v\n", p2)
		ant.Target = assign.Target
		ant.Score = assign.Score
		//		fmt.Fprintf(os.Stderr, "ant: %v\n", *ant)
	}
	b.perf.Log("Finding paths")
	return
}

type Timing struct {
	start int64
	last  int64
}

func NewTiming() *Timing {
	now := time.Nanoseconds()
	return &Timing{start: now, last: now}
}

func (t *Timing) CurMs() int {
	now := time.Nanoseconds()
	return int((now - t.last) / (1000 * 1000))
}

func (t *Timing) Log(name string) {
	now := time.Nanoseconds()
	fmt.Fprintf(os.Stderr, "%s: %d ms\n", name, (now-t.last)/(1000*1000))
	t.last = now
}

func (t *Timing) Total() {
	now := time.Nanoseconds()
	fmt.Fprintf(os.Stderr, "total: %d ms\n", (now-t.start)/(1000*1000))
}

func (b *MyBot) FindClosestHill(at Location) (res Location) {
	res = at
	bestDist := -1
	for _, hill := range b.m.MyHills() {
		dist := b.loc.Dist(at, hill.Loc)
		if dist != NoPath && (bestDist == -1 || dist < bestDist) {
			bestDist = dist
			res = hill.Loc
		}
	}
	return
}

func GetRandomDirection(dirs []Direction, scores []int) Direction {
	fmt.Fprintf(os.Stderr, "GetRandomDirection, dirs: %v, scores: %v\n", dirs, scores)
	min := (1 << 31) - 1
	for _, s := range scores {
		if min > s {
			min = s
		}
	}
	fmt.Fprintf(os.Stderr, "min: %v\n", min)

	var sum int
	for i := range scores {
		scores[i] -= min
		sum += scores[i]
	}
	sumf := float64(sum)
	fmt.Fprintf(os.Stderr, "scores: %v\n", scores)
	h := XaosP / float64(len(scores))
	p := make([]float64, len(scores))
	var x float64
	if sum > 0 {
		x = h * sumf / (1 - float64(len(scores))*h)
	} else {
		x = 1
	}
	fmt.Fprintf(os.Stderr, "h: %v, x: %v\n", h, x)

	for i := range scores {
		p[i] = (float64(scores[i]) + x) / (sumf + float64(len(scores))*x)
	}
	fmt.Fprintf(os.Stderr, "p: %v\n", p)
	v := rand.Float64()
	for i, cur := range p {
		if v <= cur {
			return dirs[i]
		}
		v -= cur
	}
	return dirs[0]
}

func (b *MyBot) CheckMax(hill, loc Location, max int) bool {
	panic("CheckMax not implemented")
}

func (b *MyBot) DoTurn(input []Input) (orders []Order, err os.Error) {
	b.perf = NewTiming()
	b.m.Update(input)
	b.perf.Log("Map update")

	fmt.Fprintf(os.Stderr, "len(NewCells): %d\n", len(b.m.NewCells))
	b.loc.Add(b.m.NewCells...)
	b.loc.Update(func() bool {
		return b.perf.CurMs() < b.LocatorBudgetMs
	})
	b.perf.Log("Fair locator update")

	b.gridSet.Update()
	b.perf.Log("GridLocatedSet update")

	//	b.Plan()

	turn := b.m.Turn()
	for _, ant := range b.m.MyLiveAnts {
		loc := ant.Loc(b.m.Turn())
		var a []Direction
		var s []int
		hill := b.FindClosestHill(loc)
		dist := b.loc.Dist(hill, loc)
		var newLoc Location

		try := func(dir Direction) {
			newLoc = b.t.NewLoc(loc, dir)
			if b.m.Terrain[newLoc] == Land {
				score := 0
				newDist := b.loc.Dist(hill, newLoc)
				switch {
				case dist < newDist:
					if dist < 10 && !b.CheckMax(hill, newLoc, 10) {
						return
					}
					score++
				case dist > newDist:
					score--
				}

				a = append(a, dir)
				s = append(s, score)
			}
		}
		try(North)
		try(East)
		try(South)
		try(West)
		if len(a) == 0 {
			continue
		}
		dir := GetRandomDirection(a, s)
		ant.Target = b.t.NewLoc(loc, dir)
		path := NewPath(b.t, loc)
		path.Append(dir)
		ant.Path = path
		ant.Score = 1
	}

	b.m.MoveAnts()
	b.perf.Log("MoveAnts")

	for _, ant := range b.m.MyLiveAnts {
		if ant.HasLoc(turn + 1) {
			// This ant has been moved
			dir := b.t.GuessDir(ant.Loc(turn), ant.Loc(turn+1))
			//			fmt.Fprintf(os.Stderr, "guess dir: %c\n", dir)
			orders = append(orders,
				Order{
					Row: b.t.Row(ant.Loc(turn)),
					Col: b.t.Col(ant.Loc(turn)),
					Dir: dir,
				})
		}
	}
	b.perf.Log("Generate output")
	b.perf.Total()
	return
}
