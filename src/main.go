package main

import (
	"log"
)

func main() {
	var p Params
	p, err := ReadParams()
	if err != nil {
		log.Panicf("ReadParams: %v", err)
	}
	bot := new(MyBot)
	if err = bot.Init(p); err != nil {
		log.Panicf("bot.Init: %v", err)
	}
	if err := Loop(p, bot); err != nil {
		log.Panicf("Loop: %s", err)
	}
}
