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

type MyBot struct {
	p    Params
	t    Torus
	m    *Map
	cn   *Country
	gov  *Goverment
	plan []MyAssignment
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

type MyAssignment struct {
	Ant    *MyAnt
	Target Location
	Score  int
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

func (b *MyBot) Plan(myPrev []MyAssignment) (myPlan []MyAssignment) {
	l := &myLocator{cn: b.cn}
	p := NewGreedyPlanner(b.t.Size())
	var workers []Location
	var targets []Location
	var scores []int

	for _, ant := range b.m.MyLiveAnts {
		workers = append(workers, ant.Loc(b.m.Turn()))
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
	var prev []Assignment
	for _, myAssign := range myPrev {
		if !myAssign.Ant.Alive {
			continue
		}
		loc := myAssign.Ant.Loc(b.m.Turn())
		if loc == myAssign.Target {
			continue
		}
		if b.m.MyLiveAntAt(loc) == nil {
			fmt.Fprintf(os.Stderr, "MyLiveAnts: %v\n", b.m.MyLiveAnts)
			fmt.Fprintf(os.Stderr, "b.m.MyLiveAntAt(%v): nil\n", loc)
		}
		prev = append(prev, Assignment{
			Worker: loc,
			Target: myAssign.Target,
			Score:  myAssign.Score,
		})
	}
	plan := p.Plan(l, prev, &MyLocatedSet{workers}, targets, scores)
	for _, assign := range plan {
		ant := b.m.MyLiveAntAt(assign.Worker)
		if ant == nil {
			panic("ant == nil")
		}
		myPlan = append(myPlan, MyAssignment{
			Ant:    ant,
			Target: assign.Target,
			Score:  assign.Score,
		})
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
	b.plan = b.Plan(b.plan)

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
				path := b.cn.Path(ant.Loc(turn), closerProv.Center)
				if path == nil {
					panic("Unable to find a path between an ant and the center of a closer prov")
				}
				if path.Len() == 0 {
					continue
				}
				dir := path.Dir(0)
				if b.m.CanMove(ant.Loc(turn), dir) {
					b.m.Move(ant, dir)
				}
			}
			continue
		}
	}

	fmt.Fprintf(os.Stderr, "turn: %d, plan: %v\n", turn, b.plan)
	for _, assign := range b.plan {
		ant := assign.Ant
		if ant == nil {
			panic("ant == nil")
		}
		path := b.cn.Path(ant.Loc(turn), assign.Target)
		fmt.Fprintf(os.Stderr, "Path: %v\n", path)
		if path.Len() == 0 {
			continue
		}
		dir := path.Dir(0)
		fmt.Fprintf(os.Stderr, "Dir: %c\n", dir)
		if b.m.CanMove(ant.Loc(turn), dir) {
			b.m.Move(ant, dir)
		} else {
			fmt.Fprintf(os.Stderr, "Can't move!\n")
		}
		continue
	}
	// Harvest
	//		if len(rep.Food) > 0 {
	//			ant := rep.MyLiveAnts[0]
	//			path := b.cn.Path(ant.Loc(turn), rep.Food[0])
	//			if path == nil {
	//				panic("Unable to find a path between an ant and food in the same prov")
	//			}
	//			if path.Len() > 0 {
	//				dir := path.Dir(0)
	//				if b.m.CanMove(ant.Loc(turn), dir) {
	//					b.m.Move(ant, dir)
	//				}
	//			}
	//			continue
	//		}
	// Discover: move to any adjacent province
	/*		if len(prov.Conn) > 0 {
				ant := rep.MyLiveAnts[0]
				toInd := prov.Conn[rand.Intn(len(prov.Conn))]
				toProv := b.cn.ProvByIndex(toInd)
				path := b.cn.Path(ant.Loc(turn), toProv.Center)
				if path.Len() > 0 {
					dir := path.Dir(0)
					if b.m.CanMove(ant.Loc(turn), dir) {
						b.m.Move(ant, dir)
						continue
					}
				}
			}
			// Discover: random move
			dir := Dirs[rand.Intn(4)]
			ant := rep.MyLiveAnts[0]
			if b.m.CanMove(ant.Loc(turn), dir) {
				b.m.Move(ant, dir)
			}*/

	//	}

	b.cn.Dump("/tmp/country.txt")

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
