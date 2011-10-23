package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const LagMs int = 20

type Order struct {
	Row, Col, Dir int
	Live          bool
}

type Input struct {
	What  string
	Row   int
	Col   int
	Owner int
}

type Bot interface {
	Init(p Params) os.Error
	DoTurn(turn int, input []Input, orders <-chan Order) os.Error
}

var stdin = bufio.NewReader(os.Stdin)

type Params struct {
	LoadTime      int   //in milliseconds
	TurnTime      int   //in milliseconds
	Rows          int   //number of rows in the map
	Cols          int   //number of columns in the map
	Turns         int   //maximum number of turns in the game
	ViewRadius2   int   //view radius squared
	AttackRadius2 int   //battle radius squared
	SpawnRadius2  int   //spawn radius squared
	PlayerSeed    int64 //random player seed
}

func ReadParams() (p Params, err os.Error) {
	for {
		line, err := stdin.ReadString('\n')
		if err != nil {
			return p, err
		}
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if line == "ready" {
			break
		}

		words := strings.SplitN(line, " ", 2)
		if len(words) != 2 {
			return p, fmt.Errorf("Invalid command format: %s", line)
		}

		param, _ := strconv.Atoi(words[1])

		switch words[0] {
		case "loadtime":
			p.LoadTime = param
		case "turntime":
			p.TurnTime = param
		case "rows":
			p.Rows = param
		case "cols":
			p.Cols = param
		case "turns":
			p.Turns = param
		case "viewradius2":
			p.ViewRadius2 = param
		case "attackradius2":
			p.AttackRadius2 = param
		case "spawnradius2":
			p.SpawnRadius2 = param
		case "player_seed":
			param64, _ := strconv.Atoi64(words[1])
			p.PlayerSeed = param64
		case "turn":
		case "ready":
			return
		default:
			return p, fmt.Errorf("unknown command: %s", line)
		}
	}
	return
}

func doTurn(p Params, b Bot, turn int, input []Input) (err os.Error) {
	ready := make(chan os.Error)
	defer close(ready)
	can := make(chan bool, 1)
	can <- true
	orders := make(chan Order, 100)
	go func() {
		err := b.DoTurn(turn, input, orders)
		close(orders)
		_, ok := <-can
		if ok {
			ready <- err
		}
	}()
	go func() {
		time.Sleep(1000 * int64(p.TurnTime-LagMs))
		_, ok := <-can
		if ok {
			ready <- nil
		}
	}()
	for {
		var order Order
		select {
		case err = <-ready:
			close(can)
			return
		case order = <-orders:
			// This should be some bug of select.
			if !order.Live {
				continue
			}
			line := fmt.Sprintf("o %d %d %d\n", order.Row, order.Col, order.Dir)
			os.Stdout.Write([]byte(line))
		}
	}
	return
}

func Loop(p Params, b Bot) (err os.Error) {
	//indicate we're ready
	os.Stdout.Write([]byte("go\n"))

	var turn int
	var input []Input
	for {
		line, err := stdin.ReadString('\n')
		if err != nil {
			if err == os.EOF {
				return err
			}
			return fmt.Errorf("ReadString: %v", err)
		}
		if line = strings.TrimSpace(line); line == "" {
			continue
		}

		if line == "go" {
			if err = doTurn(p, b, turn, input); err != nil {
				return fmt.Errorf("doTurn: %v", err)
			}
			os.Stdout.Write([]byte("go\n"))
			input = nil
			continue
		}

		if line == "end" {
			break
		}

		words := strings.SplitN(line, " ", 5)
		var in Input
		in.What = words[0]
		if in.What == "turn" {
			turn, _ = strconv.Atoi(words[1])
			continue
		}
		if len(words) < 3 {
			return fmt.Errorf("Invalid command format: \"%s\"", line)
		}
		in.Row, _ = strconv.Atoi(words[1])
		in.Col, _ = strconv.Atoi(words[2])
		if in.What == "h" || in.What == "a" || in.What == "d" {
			in.Owner, _ = strconv.Atoi(words[3])
		}
		input = append(input, in)
	}

	return nil
}
