package main

import (
	"os"
	"log"
)

func main() {
	var s State
	err := s.Start()
	if err != nil {
		log.Panicf("Start() failed (%s)", err)
	}
	mb := NewBot(&s)
	if err := s.Loop(mb); err != nil {
		if err != os.EOF {
			log.Panicf("Loop: %s", err)
		}
	}
}
