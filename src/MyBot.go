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

var big = make([]int16, 200*1000*1000)

type MyBot struct {
	p               Params
	t               Torus
	m               *Map
	cn              *Country
	gov             *Goverment
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
		if b.cn.IsOwn(food) {
			addTarget(food, FoodScore)
		}
	}
	for _, hill := range b.m.EnemyHills() {
		if b.cn.IsOwn(hill.Loc) {
			addTarget(hill.Loc, EnemyHillScore)
		}
	}

	for i := 0; i < b.cn.ProvCount(); i++ {
		prov := b.cn.ProvByIndex(i)
		if !prov.Live() || b.m.HasMyHillAt(prov.Center) {
			continue
		}
		if b.m.LastVisited[prov.Center] > 0 {
			age := b.m.Turn() - b.m.LastVisited[prov.Center]
			if age > 0 {
				addTarget(prov.Center, age*VisitScore)
			}
		} else {
			addTarget(prov.Center, NeverVisitedScore)
			for _, dir := range Dirs {
				newLoc := b.t.NewLoc(prov.Center, dir)
				if b.m.Terrain[newLoc] == Land {
					addTarget(newLoc, NeverVisitedScore2)
					break
				}
			}
		}
	}

	//	fmt.Fprintf(os.Stderr, "scores: %v\n", scores)
	b.perf.Log("Prepare data for planner")

	//	mls := NewMyLocatedSet(b.m, b.cn, workers, b.locSet, b.locsByProv)
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
		//		p2 := b.cn.Path(ant.Loc(b.m.Turn()), assign.Target)
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

	if b.cn == nil {
		b.cn = NewCountry(b.m)
	} else {
		b.cn.Update()
	}
	b.perf.Log("Country update")
	if b.gov == nil {
		b.gov = NewGoverment(b.cn, b.m)
	} else {
		b.gov.Update()
	}
	b.perf.Log("Goverment update")
	//	b.cn.Dump("/tmp/country.txt")
	//	b.perf.Log("Map dump")

	b.Plan()

	turn := b.m.Turn()
	for provInd, rep := range b.gov.TurnRep {
		prov := b.cn.ProvByIndex(provInd)
		if len(rep.MyLiveAnts) == 0 {
			continue
		}
		if b.m.HasMyHillAt(prov.Center) && len(prov.Conn) > 0 {
			ant := b.m.MyLiveAntAt(prov.Center)
			// Don't stay in the hill
			if ant != nil {
				ant.Target = b.cn.ProvByIndex(prov.Conn[rand.Intn(len(prov.Conn))]).Center
				ant.Score = MoveFromMyHillScore
				ant.Path = b.pf.Path(prov.Center, ant.Target)
			}
		}
	}
	b.perf.Log("Withdrawal from enemies")

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
