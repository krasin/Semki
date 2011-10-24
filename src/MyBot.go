package main

import (
	//	"fmt"
	//	"io/ioutil"
	"os"
	"rand"
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

func (b *MyBot) DoTurn(input []Input) (orders []Order, err os.Error) {
	dirs := []Direction{North, East, South, West}
	b.m.Update(input)
	for _, ant := range b.m.MyAnts() {
		p := rand.Perm(4)
		for _, i := range p {
			d := dirs[i]
			if b.m.CanMove(ant.Loc, d) {
				orders = append(orders,
					Order{
						Row: b.m.Row(ant.Loc),
						Col: b.m.Col(ant.Loc),
						Dir: d,
					})
				b.m.Move(ant.Loc, d)
				break
			}
		}
	}
	//	ioutil.WriteFile(fmt.Sprintf("/tmp/turn.%d.txt", b.m.Turn()),
	//		[]byte(fmt.Sprintf("Orders: %v\n", orders)),
	//		0644,
	//	)
	return
}
