package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const LagMs int = 30

type Order struct {
	Row, Col int
	Dir      Direction
}

type Input struct {
	What  int
	Row   int
	Col   int
	Owner int
}

type Bot interface {
	Init(p Params) os.Error
	DoTurn(input []Input) (orders []Order, err os.Error)
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

func doTurn(p Params, b Bot, input []Input) (err os.Error) {
	var orders []Order
	if orders, err = b.DoTurn(input); err != nil {
		return
	}
	for _, order := range orders {
		line := fmt.Sprintf("o %d %d %c\n", order.Row, order.Col, order.Dir)
		os.Stdout.Write([]byte(line))
	}
	return
}

func Loop(p Params, b Bot) (err os.Error) {
	//indicate we're ready
	os.Stdout.Write([]byte("go\n"))

	var input []Input
	isNewTurn := true
	for {
		line, err := stdin.ReadString('\n')
		if err != nil {
			if err == os.EOF && isNewTurn {
				return nil
			}
			return fmt.Errorf("ReadString: %v", err)
		}
		isNewTurn = false
		if line = strings.TrimSpace(line); line == "" {
			continue
		}

		if line == "go" {
			if err = doTurn(p, b, input); err != nil {
				return fmt.Errorf("doTurn: %v", err)
			}
			os.Stdout.Write([]byte("go\n"))
			input = nil
			isNewTurn = true
			continue
		}

		if line == "end" {
			break
		}

		words := strings.SplitN(line, " ", 5)
		if words[0] == "turn" {
			continue
		}
		if len(words) < 3 {
			return fmt.Errorf("Invalid command format: \"%s\"", line)
		}
		var in Input
		in.What = int(words[0][0])
		in.Row, _ = strconv.Atoi(words[1])
		in.Col, _ = strconv.Atoi(words[2])
		if in.What == Hill || in.What == Ant || in.What == DeadAnt {
			in.Owner, _ = strconv.Atoi(words[3])
		}
		input = append(input, in)
	}

	return nil
}
