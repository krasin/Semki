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

type MyEstimator struct {
	cn *Country
}

func (est *MyEstimator) Estimate(w *Worker, loc Location) int {
	return 1
}

func (est *MyEstimator) Prov(loc Location) int {
	return est.cn.Prov(loc).Ind
}

func (est *MyEstimator) Conn(prov int) (res []int) {
	for _, connProv := range est.cn.ProvByIndex(prov).Conn {
		res = append(res, connProv)
	}
	return
}

func (b *MyBot) Plan() []Assignment {
	est := &MyEstimator{cn: b.cn}
	p := NewPlanner(est, b.cn.ProvCount())
	for i, ant := range b.m.MyLiveAnts {
		loc := ant.Loc(b.m.Turn())
		p.AddWorker(&Worker{Loc: loc, LiveInd: i, Prov: b.cn.Prov(loc).Ind})
	}
	for _, food := range b.m.Food() {
		if b.cn.IsOwn(food) {
			p.AddTarget(food, FoodScore)
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
			p.AddTarget(prov.Center, score)
		}
	}
	return p.MakePlan()
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
	plan := b.Plan()

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

	fmt.Fprintf(os.Stderr, "turn: %d, plan: %v\n", turn, plan)
	for _, assignment := range plan {
		ant := b.m.MyLiveAnts[assignment.Worker.LiveInd]
		path := b.cn.Path(ant.Loc(turn), assignment.Target.Loc)
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
