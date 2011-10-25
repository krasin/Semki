package main

import (
	"testing"
)

type GuessDirTest struct {
	FromRow, FromCol int
	ToRow, ToCol     int
	Cols             int
	Dir              Direction
}

var guessDirTests = []GuessDirTest{
	{1, 1, 1, 2, 4, East},
	{1, 2, 1, 1, 4, West},
	{3, 3, 0, 3, 4, South},
	{0, 0, 3, 0, 4, North},
	{1, 0, 1, 3, 4, West},
	{31, 1, 31, 0, 32, West},
}

func TestGuessDir(t *testing.T) {
	for _, test := range guessDirTests {
		from := Loc(test.FromRow, test.FromCol, test.Cols)
		to := Loc(test.ToRow, test.ToCol, test.Cols)
		if d := GuessDir(from, to, test.Cols); d != test.Dir {
			t.Errorf("from: %v, to: %v, want: %c, got: %c", from, to, test.Dir, d)
		}
	}
}
