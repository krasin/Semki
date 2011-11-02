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
const EnemyWithdrawalScore = 2000

const MaxFindNearCount = 10

type MyBot struct {
	p          Params
	t          Torus
	m          *Map
	cn         *Country
	gov        *Goverment
	locsByProv LocListMap
	locSet     LocSet
	perf       *Timing
}

func (b *MyBot) Init(p Params) (err os.Error) {
	rand.Seed(p.PlayerSeed)
	b.p = p
	b.t = Torus{Rows: p.Rows, Cols: p.Cols}
	b.m = NewMap(b.t, p.ViewRadius2)
	b.locsByProv = NewLocListMap(b.t.Size())
	b.locSet = NewLocSet(b.t.Size())
	return nil
}

type myLocator struct {
	cn *Country
}

func (l *myLocator) Dist(from, to Location) int {
	panic("myLocator.Dist is not implemented")
}

type MyLocatedSet struct {
	m          *Map
	cn         *Country
	locs       []Location
	locSet     LocSet
	locsByProv LocListMap
}

func NewMyLocatedSet(m *Map, cn *Country, locs []Location, locSet LocSet, locsByProv LocListMap) (s *MyLocatedSet) {
	s = &MyLocatedSet{m: m, cn: cn, locs: locs, locSet: locSet, locsByProv: locsByProv}
	s.locsByProv.Clear()
	for _, loc := range locs {
		s.locsByProv.Add(cn.Prov(loc).Center, loc)
	}
	return
}

func (s *MyLocatedSet) All() (res []Location) {
	res = make([]Location, len(s.locs))
	copy(res, s.locs)
	return
}

func (s *MyLocatedSet) FindNear(at Location, ok func(Location) bool) (Location, bool) {
	if ant := s.m.MyLiveAntAt(at); ant != nil && ok(at) {
		return at, true
	}
	prov := s.cn.Prov(at)
	if prov == nil {
		return 0, false
	}
	s.locSet.Clear()
	s.locSet.Add(prov.Center)
	q := []*Province{prov}
	var q2 []*Province
	count := 0
	for len(q) > 0 {
		q, q2 = q2[:0], q
		for _, prov := range q2 {
			for _, w := range s.locsByProv.Get(prov.Center) {
				if ok(w) {
					return w, true
				}
				count++
				if count > MaxFindNearCount {
					return 0, false
				}
			}
			for _, other := range prov.Conn {
				otherProv := s.cn.ProvByIndex(other)
				if !s.locSet.Has(otherProv.Center) {
					q = append(q, s.cn.Prov(otherProv.Center))
					s.locSet.Add(otherProv.Center)
				}
			}
		}
	}
	return 0, false
}

func (b *MyBot) Plan() {
	l := &myLocator{cn: b.cn}
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
	for i := 0; i < b.cn.ProvCount(); i++ {
		prov := b.cn.ProvByIndex(i)
		if !prov.Live() || b.m.HasHillAt(prov.Center) {
			continue
		}
		score := NeverVisitedScore
		age := b.m.Turn() - b.m.LastVisited[prov.Center]
		if b.m.LastVisited[prov.Center] > 0 {
			score = age * VisitScore
		}
		if score > 0 {
			addTarget(prov.Center, score)
		}
	}

	fmt.Fprintf(os.Stderr, "scores: %v\n", scores)
	b.perf.Log("Prepare data for planner")

	plan := p.Plan(l, prev, NewMyLocatedSet(b.m, b.cn, workers, b.locSet, b.locsByProv), targets, scores)
	b.perf.Log("Planner")
	fmt.Fprintf(os.Stderr, "plan = %v\n", plan)
	for _, assign := range plan {
		ant := b.m.MyLiveAntAt(assign.Worker)
		if ant == nil {
			panic("ant == nil")
		}
		ant.Path = b.cn.Path(ant.Loc(b.m.Turn()), assign.Target)
		ant.Target = assign.Target
		ant.Score = assign.Score
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
	b.cn.Dump("/tmp/country.txt")
	b.perf.Log("Map dump")

	b.Plan()

	turn := b.m.Turn()
	for provInd, rep := range b.gov.TurnRep {
		prov := b.cn.ProvByIndex(provInd)
		if len(rep.MyLiveAnts) == 0 {
			continue
		}
		// Enemies
		if len(rep.Enemy) > 0 {
			closerProv := b.cn.CloserProv(prov)
			for _, ant := range rep.MyLiveAnts {
				ant.Score = EnemyWithdrawalScore
				ant.Target = closerProv.Center
				ant.Path = b.cn.Path(ant.Loc(turn), closerProv.Center)

				if ant.Path == nil {
					panic("Unable to find a path between an ant and the center of a closer prov")
				}
			}
			continue
		}
	}
	b.perf.Log("Withdrawal from enemies")

	b.m.MoveAnts()
	b.perf.Log("MoveAnts")

	for _, ant := range b.m.MyLiveAnts {
		if ant.HasLoc(turn + 1) {
			// This ant has been moved
			dir := b.t.GuessDir(ant.Loc(turn), ant.Loc(turn+1))
			fmt.Fprintf(os.Stderr, "guess dir: %c\n", dir)
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
