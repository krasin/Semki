package main

import (
	"os"
)

type MyBot struct {
	p Params
}

func (b *MyBot) Init(p Params) (err os.Error) {
	b.p = p
	return nil
}

//DoTurn is where you should do your bot's actual work.
func (b *MyBot) DoTurn(turn int, input []Input, orders <-chan Order) (err os.Error) {
	/*	dirs := []Direction{North, East, South, West}
		for loc, ant := range s.Map.Ants {
			if ant != MY_ANT {
				continue
			}

			//try each direction in a random order
			p := rand.Perm(4)
			for _, i := range p {
				d := dirs[i]

				loc2 := s.Map.Move(loc, d)
				if s.Map.SafeDestination(loc2) {
					s.IssueOrderLoc(loc, d)
					//there's also an s.IssueOrderRowCol if you don't have a Location handy
					break
				}
			}
		}*/
	//returning an error will halt the whole program!
	return nil
}
