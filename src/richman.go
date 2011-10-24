package main

import (
	"fmt"
	"os"
)

type Richman struct {
	Vals       []int
	Rows, Cols int
}

func NewRichman(rows, cols int) *Richman {
	return &Richman{
		Vals: make([]int, rows*cols),
		Rows: rows,
		Cols: cols,
	}
}

func (r *Richman) Remove(loc Location) {
	r.Vals[loc] = -1
}

// val should be positive
func (r *Richman) PinVal(loc Location, val int) {
	r.Vals[loc] = -val
}

func (r *Richman) Val(loc Location) int {
	if r.Vals[loc] < 0 && r.Vals[loc] != -1 {
		return -r.Vals[loc]
	}
	return r.Vals[loc]
}

func (r *Richman) NewLoc(loc Location, d Direction) Location {
	return NewLoc(loc, d, r.Rows, r.Cols)
}

func (r *Richman) GetAdjVal(loc Location, d Direction) int {
	return r.Val(r.NewLoc(loc, d))
}

func min2(a, b int) int {
	if a < b && a >= 0 {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (r *Richman) Iterate(count int) {
	var l []Location
	var cur []Location
	for i, v := range r.Vals {
		if v < 0 {
			continue
		}
		l = append(l, Location(i))
	}
	used := make([]bool, len(r.Vals))

	for i := 0; i < count; i++ {
		for _, loc := range l {
			used[loc] = false
		}
		tmp := cur[:0]
		cur = l
		l = tmp
		for _, loc := range cur {
			if r.Vals[loc] < 0 {
				continue
			}
			vals := [4]int{
				r.GetAdjVal(loc, North),
				r.GetAdjVal(loc, South),
				r.GetAdjVal(loc, West),
				r.GetAdjVal(loc, East),
			}
			mx := max(max(vals[0], vals[1]), max(vals[2], vals[3]))
			mn := min2(min2(vals[0], vals[1]), min2(vals[2], vals[3]))
			if mx < 0 || mn < 0 {
				continue
			}
			val := (mx + mn) / 2
			if r.Vals[loc] != val {
				r.Vals[loc] = val
				adjLoc := r.NewLoc(loc, North)
				if !used[adjLoc] {
					used[adjLoc] = true
					l = append(l, adjLoc)
				}
				adjLoc = r.NewLoc(loc, South)
				if !used[adjLoc] {
					used[adjLoc] = true
					l = append(l, adjLoc)
				}
				adjLoc = r.NewLoc(loc, West)
				if !used[adjLoc] {
					used[adjLoc] = true
					l = append(l, adjLoc)
				}
				adjLoc = r.NewLoc(loc, East)
				if !used[adjLoc] {
					used[adjLoc] = true
					l = append(l, adjLoc)
				}

			}
		}

	}
}

func (r *Richman) Dump(filename string) {
	f, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	for i := 0; i < r.Rows; i++ {
		for j := 0; j < r.Cols; j++ {
			fmt.Fprintf(f, "%d ", r.Vals[Loc(i, j, r.Cols)])
		}
		fmt.Fprintf(f, "\n")
	}
	f.Close()
}
