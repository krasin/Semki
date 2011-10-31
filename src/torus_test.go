package main

import (
	"testing"
)

type GuessDirTest struct {
	T                Torus
	FromRow, FromCol int
	ToRow, ToCol     int
	Dir              Direction
}

var t4 = Torus{4, 4}
var t32 = Torus{32, 32}

var guessDirTests = []GuessDirTest{
	{t4, 1, 1, 1, 2, East},
	{t4, 1, 2, 1, 1, West},
	{t4, 3, 3, 0, 3, South},
	{t4, 0, 0, 3, 0, North},
	{t4, 1, 0, 1, 3, West},
	{t32, 31, 1, 31, 0, West},
}

func TestGuessDir(t *testing.T) {
	for _, test := range guessDirTests {
		from := test.T.Loc(test.FromRow, test.FromCol)
		to := test.T.Loc(test.ToRow, test.ToCol)
		if d := test.T.GuessDir(from, to); d != test.Dir {
			t.Errorf("from: %v, to: %v, want: %c, got: %c", from, to, test.Dir, d)
		}
	}
}
