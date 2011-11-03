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

const bigN = 20000

const NoPath = (1 << 31) - 1

var big = make([]int16, 200*1000*1000)

type locPair struct {
	a Location
	b Location
}

type fairPathLocator struct {
	conn    Connector
	big     []int16
	loc2ind []int
	ind2loc []Location

	// indices to update
	toUpdate []locPair
	buf      []locPair
}

func NewFairPathLocator(conn Connector, big []int16) *fairPathLocator {
	return &fairPathLocator{
		conn:    conn,
		big:     big,
		loc2ind: make([]int, 40000),
	}
}

func (l *fairPathLocator) hasLoc(loc Location) bool {
	return l.loc2ind[int(loc)] > 0
}

func (l *fairPathLocator) bigIndex(from, to Location) int {
	fromInd := l.loc2ind[int(from)] - 1
	toInd := l.loc2ind[int(to)] - 1
	if fromInd == -1 || toInd == -1 {
		return -1
	}
	if fromInd == toInd {
		return 0
	}
	if fromInd > toInd {
		fromInd, toInd = toInd, fromInd
	}
	s := bigN*fromInd - (fromInd*(fromInd+1))/2
	return s + (toInd - fromInd - 1)
}

func (l *fairPathLocator) set(from, to Location, dist int) {
	//	fmt.Printf("set(%d, %d, dist=%d)\n", from, to, dist)
	ind := l.bigIndex(from, to)
	//	fmt.Printf("ind: %d\n", ind)
	l.big[ind] = int16(dist)
}

// Update distances for loc looking at from
func (l *fairPathLocator) updatePair(loc, from Location) {
	was := false
	if loc == from {
		return
	}
	for i := 0; i < len(l.ind2loc); i++ {
		to := l.ind2loc[i]
		if to == loc || to == from || !l.hasLoc(to) {
			continue
		}
		curDist := l.Dist(loc, to)
		newDist := l.Dist(from, to)
		if newDist != NoPath && newDist+1 < curDist {
			l.set(loc, to, newDist+1)
			was = true
		}
	}
	if was {
		for _, conn := range l.conn.Conn(loc) {
			if conn != from && l.hasLoc(conn) {
				l.toUpdate = append(l.toUpdate, locPair{conn, loc})
			}
		}
	}
}

func (l *fairPathLocator) NeedUpdate() bool {
	return len(l.toUpdate) > 0
}

func (l *fairPathLocator) Add(locs ...Location) {
	//	fmt.Printf("Add(%d)\n", loc)
	for _, loc := range locs {
		if l.hasLoc(loc) {
			continue
		}
		ind := len(l.ind2loc)
		l.ind2loc = append(l.ind2loc, loc)
		l.loc2ind[int(loc)] = ind + 1

		for _, conn := range l.conn.Conn(loc) {
			if !l.hasLoc(conn) {
				continue
			}
			l.set(loc, conn, 1)
			l.toUpdate = append(l.toUpdate, locPair{loc, conn})
			l.toUpdate = append(l.toUpdate, locPair{conn, loc})
		}
	}
}

func (l *fairPathLocator) UpdateStep() {
	l.toUpdate, l.buf = l.buf[:0], l.toUpdate
	for _, pair := range l.buf {
		l.updatePair(pair.a, pair.b)
	}
}

func (l *fairPathLocator) Update() {
	for l.NeedUpdate() {
		l.UpdateStep()
	}
}

func (l *fairPathLocator) Dist(from, to Location) int {
	if from == to {
		return 0
	}
	ind := l.bigIndex(from, to)
	if ind == -1 {
		return NoPath
	}
	val := int(l.big[ind])
	if val == 0 {
		return NoPath
	}
	return val
}

type MyBot struct {
	p          Params
	t          Torus
	m          *Map
	cn         *Country
	gov        *Goverment
	locsByProv LocListMap
	locSet     LocSet
	perf       *Timing
	loc        *fairPathLocator
}

func (b *MyBot) Init(p Params) (err os.Error) {
	rand.Seed(p.PlayerSeed)
	b.p = p
	b.t = Torus{Rows: p.Rows, Cols: p.Cols}
	b.m = NewMap(b.t, p.ViewRadius2)
	b.locsByProv = NewLocListMap(b.t.Size())
	b.locSet = NewLocSet(b.t.Size())
	b.loc = NewFairPathLocator(b.m, big)
	return nil
}

type myLocator struct {
	cn      *Country
	fairLoc *fairPathLocator
}

func (l *myLocator) Dist(from, to Location) int {
	fromProv := l.cn.Prov(from)
	toProv := l.cn.Prov(to)
	if fromProv != toProv {
		return -1
	}
	p := l.cn.PathSlow(from, to)
	if p == nil {
		return -1
	}
	p2 := l.fairLoc.Dist(from, to)
	if p.Len() != p2 {
		fmt.Fprintf(os.Stderr, "DIFF on (from=%d, to=%d) ", from, to)
	}
	fmt.Fprintf(os.Stderr, "p.Len(): %d, p2: %d\n", p.Len(), p2)
	return p.Len()
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

func (s *MyLocatedSet) FindNear(at Location, score int, ok func(Location, int, bool) bool) (Location, bool) {
	if ant := s.m.MyLiveAntAt(at); ant != nil && ok(at, score, true) {
		return at, true
	}
	start := s.cn.Prov(at)
	if start == nil {
		return 0, false
	}
	s.locSet.Clear()
	s.locSet.Add(start.Center)
	q := []*Province{start}
	var q2 []*Province
	count := 0
	for len(q) > 0 {
		q, q2 = q2[:0], q
		for _, prov := range q2 {
			for _, w := range s.locsByProv.Get(prov.Center) {
				if ok(w, score, prov == start) {
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
	l := &myLocator{cn: b.cn, fairLoc: b.loc}
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

	plan := p.Plan(l, prev, NewMyLocatedSet(b.m, b.cn, workers, b.locSet, b.locsByProv), targets, scores)
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
		ant.Path = b.cn.Path(ant.Loc(b.m.Turn()), assign.Target)
		fmt.Fprintf(os.Stderr, "path: %v\n", ant.Path)
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

	b.loc.Add(b.m.NewCells...)
	b.loc.Update()
	b.perf.Log("Fair locator update")

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
				ant.Path = b.cn.Path(prov.Center, ant.Target)
			}
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
