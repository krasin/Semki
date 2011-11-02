package main

import (
	"fmt"
	//	"io/ioutil"
	"os"
	"rand"
)

const FoodScore = 1000000
const VisitScore = 1000
const NeverVisitedScore = 100000
const EnemyWithdrawalScore = 2000

type MyBot struct {
	p   Params
	t   Torus
	m   *Map
	cn  *Country
	gov *Goverment
}

func (b *MyBot) Init(p Params) (err os.Error) {
	rand.Seed(p.PlayerSeed)
	b.p = p
	b.t = Torus{Rows: p.Rows, Cols: p.Cols}
	b.m = NewMap(b.t, p.ViewRadius2)
	return nil
}

type myLocator struct {
	cn *Country
}

func (l *myLocator) Dist(from, to Location) int {
	panic("myLocator.Dist is not implemented")
}

type MyLocatedSet struct {
	locs []Location
}

func (s *MyLocatedSet) All() (res []Location) {
	res = make([]Location, len(s.locs))
	copy(res, s.locs)
	return
}

func (s *MyLocatedSet) FindNear(at Location, ok func(Location) bool) (Location, bool) {
	for _, loc := range s.locs {
		if ok(loc) {
			return loc, true
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
	plan := p.Plan(l, prev, &MyLocatedSet{workers}, targets, scores)
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
	return
}

func (b *MyBot) DoTurn(input []Input) (orders []Order, err os.Error) {
	b.m.Update(input)
	if b.cn == nil {
		b.cn = NewCountry(b.m)
	} else {
		b.cn.Update()
	}
	if b.gov == nil {
		b.gov = NewGoverment(b.cn, b.m)
	} else {
		b.gov.Update()
	}
	b.cn.Dump("/tmp/country.txt")

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

	b.m.MoveAnts()

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
	return
}
