package main

import (
	//	"fmt"
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
	b.m = NewMap(p.Rows, p.Cols)
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
		}
	}
	r.Iterate(RichmanCount)
	return r
}

func (b *MyBot) DoTurn(input []Input) (orders []Order, err os.Error) {
	dirs := []Direction{North, East, South, West}
	b.m.Update(input)
	r := b.GenerateRichman()
	r.Dump("/tmp/richman.txt")

	for _, ant := range b.m.MyAnts() {
		bestD := Direction(North)
		bestR := r.Val(ant.Loc)
		for _, d := range dirs {
			if b.m.CanMove(ant.Loc, d) {
				newR := r.Val(b.m.NewLoc(ant.Loc, d))
				if newR > bestR {
					bestD = d
					bestR = newR
				}
			}
		}

		if b.m.CanMove(ant.Loc, bestD) {
			orders = append(orders,
				Order{
					Row: b.m.Row(ant.Loc),
					Col: b.m.Col(ant.Loc),
					Dir: bestD,
				})
			b.m.Move(ant.Loc, bestD)
		}
	}
	//	ioutil.WriteFile(fmt.Sprintf("/tmp/turn.%d.txt", b.m.Turn()),
	//		[]byte(fmt.Sprintf("Orders: %v\n", orders)),
	//		0644,
	//	)
	return
}
