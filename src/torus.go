package main

import (
	"fmt"
)

type Torus struct {
	Rows int
	Cols int
}

func (t Torus) Loc(row, col int) Location {
	return Location(row*t.Cols + col)
}

func (t Torus) Row(loc Location) int {
	return int(loc) / t.Cols
}

func (t Torus) Col(loc Location) int {
	return int(loc) % t.Cols
}

func (t Torus) GuessDir(from, to Location) (d Direction) {
	fromRow := t.Row(from)
	fromCol := t.Col(from)
	toRow := t.Row(to)
	toCol := t.Col(to)

	if fromRow == toRow { // West or East
		if fromCol+1 == toCol || fromCol-1 != toCol && toCol == 0 {
			return East
		}
		return West
	}
	if fromRow+1 == toRow || fromRow-1 != toRow && toRow == 0 {
		return South
	}
	return North
}

func (t Torus) NewLoc(loc Location, d Direction) Location {
	row := t.Row(loc)
	col := t.Col(loc)
	switch d {
	case North:
		return t.Loc((row+t.Rows-1)%t.Rows, col)
	case South:
		return t.Loc((row+1)%t.Rows, col)
	case West:
		return t.Loc(row, (col+t.Cols-1)%t.Cols)
	case East:
		return t.Loc(row, (col+1)%t.Cols)

	}
	panic(fmt.Sprintf("Unknown direction: %d", d))
}

func (t Torus) ShiftLoc(loc Location, rowShift, colShift int) Location {
	row := t.Row(loc)
	col := t.Col(loc)
	return t.Loc((row+rowShift+t.Rows)%t.Rows, (col+colShift+t.Cols)%t.Cols)
}

func (t Torus) Neighbours(loc Location) []Location {
	return []Location{
		t.NewLoc(loc, North),
		t.NewLoc(loc, East),
		t.NewLoc(loc, South),
		t.NewLoc(loc, West),
	}
}

func (t Torus) Size() int {
	return t.Rows * t.Cols
}
