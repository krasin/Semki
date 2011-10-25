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
	p Params
	m *Map
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
	dirs := []Direction{North, East, South, West}
	b.m.Update(input)
	turn := b.m.Turn()
	r := b.GenerateRichman()

	for _, ant := range b.m.MyLiveAnts {
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
	}
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
