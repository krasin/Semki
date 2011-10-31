package main

import (
	//"fmt"
	//	"io/ioutil"
	"os"
	"rand"
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
