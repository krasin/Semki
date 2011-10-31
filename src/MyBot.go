package main

import (
	"fmt"
	//	"io/ioutil"
	"os"
	"rand"
)

const (
	RichmanUnknown   = 10000
	RichmanEnemy     = 1000000
	RichmanFood      = 2000000
	RichmanMyHill    = 2
	RichmanEnemyHill = 3000000

	RichmanCount = 100
)

type MyBot struct {
	p   Params
	m   *Map
	cn  *Country
	gov *Goverment
}

func (b *MyBot) Init(p Params) (err os.Error) {
	rand.Seed(p.PlayerSeed)
	b.p = p
	b.m = NewMap(p.Rows, p.Cols, p.ViewRadius2)
	return nil
}

func (b *MyBot) GenerateRichman() (r *Richman) {
	r = NewRichman(b.m.Rows, b.m.Cols)
	for loc, terrain := range b.m.Terrain {
		switch terrain {
		case Water:
			r.Remove(Location(loc))
		case Unknown:
			r.PinVal(Location(loc), RichmanUnknown)
		}
	}
	for _, item := range b.m.Items[b.m.Turn()].All {
		switch item.What {
		case Ant:
			if item.Owner != Me {
				r.PinVal(item.Loc, RichmanEnemy)
			}
		case Food:
			r.PinVal(item.Loc, RichmanFood)
		case Hill:
			if item.Owner == Me {
				r.PinVal(item.Loc, RichmanMyHill)
			} else {
				r.PinVal(item.Loc, RichmanEnemyHill)
			}
		case DeadAnt:
		default:
			panic(fmt.Sprintf("Unknown item: %c", item.What))
		}
	}
	r.Iterate(RichmanCount)
	return r
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
				dir := path.Dir(0, b.m.Cols)
				if b.m.CanMove(ant.Loc(turn), dir) {
					b.m.Move(ant, dir)
				}
			}
			continue
		}
		// Harvest
		if len(rep.Food) > 0 {
			ant := rep.MyLiveAnts[0]
			path := b.cn.Path(ant.Loc(turn), rep.Food[0])
			if path == nil {
				panic("Unable to find a path between an ant and food in the same prov")
			}
			if path.Len() > 0 {
				dir := path.Dir(0, b.m.Cols)
				if b.m.CanMove(ant.Loc(turn), dir) {
					b.m.Move(ant, dir)
				}
			}
			continue
		}
		// Discover: move to any adjacent province
		if len(prov.Conn) > 0 {
			ant := rep.MyLiveAnts[0]
			toInd := prov.Conn[rand.Intn(len(prov.Conn))]
			toProv := b.cn.ProvByIndex(toInd)
			path := b.cn.Path(ant.Loc(turn), toProv.Center)
			if path.Len() > 0 {
				dir := path.Dir(0, b.m.Cols)
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
		}

	}

	b.cn.Dump("/tmp/country.txt")

	//	r := b.GenerateRichman()

	/*	for _, ant := range b.m.MyLiveAnts {
		bestD := Direction(North)
		bestR := r.Val(ant.Loc(turn))
		for _, d := range dirs {
			if b.m.CanMove(ant.Loc(turn), d) {
				newR := r.Val(b.m.NewLoc(ant.Loc(turn), d))
				if newR > bestR {
					bestD = d
					bestR = newR
				}
			}
		}

		if b.m.CanMove(ant.Loc(turn), bestD) {
			b.m.Move(ant, bestD)
		}
	}*/
	for _, ant := range b.m.MyLiveAnts {
		if ant.HasLoc(turn + 1) {
			// This ant has been moved
			dir := GuessDir(ant.Loc(turn), ant.Loc(turn+1), b.m.Cols)
			orders = append(orders,
				Order{
					Row: b.m.Row(ant.Loc(turn)),
					Col: b.m.Col(ant.Loc(turn)),
					Dir: dir,
				})
		}
	}
	return
}
