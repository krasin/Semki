package main

import (
	"fmt"
	"testing"
)

type fairLocatorTest struct {
	n    int
	adj  [][]int    // just for row < col
	dist [][]int    // just for tow < col
	add  []Location // add order
}

func (t *fairLocatorTest) Conn(loc Location) (res []Location) {
	fmt.Printf("Conn, loc: %d\n", loc)
	ind := int(loc)
	for i := 0; i < t.n; i++ {
		if i == ind {
			continue
		}
		from := i
		to := ind
		if from > to {
			from, to = to, from
		}
		fmt.Printf("from: %d, to: %d, t.adj: %v\n", from, to, t.adj)
		if t.adj[from][to-from-1] == 1 {
			res = append(res, Location(i))
		}
	}
	return
}

var fairLocatorTests = []fairLocatorTest{
	{
		n:    1,
		adj:  [][]int{},
		dist: [][]int{},
		add:  []Location{0},
	},
	{
		n: 2,
		adj: [][]int{
			[]int{1},
		},
		dist: [][]int{
			[]int{1},
		},
		add: []Location{0, 1},
	},
	{
		n: 2,
		adj: [][]int{
			[]int{0},
		},
		dist: [][]int{
			[]int{NoPath},
		},
		add: []Location{0, 1},
	},
	{
		n: 2,
		adj: [][]int{
			[]int{1},
		},
		dist: [][]int{
			[]int{1},
		},
		add: []Location{1, 0},
	},
	{
		n: 2,
		adj: [][]int{
			[]int{0},
		},
		dist: [][]int{
			[]int{NoPath},
		},
		add: []Location{1, 0},
	},
	{
		n: 2,
		adj: [][]int{
			[]int{1},
		},
		dist: [][]int{
			[]int{1},
		},
		add: []Location{1, 0, 1, 0, 1, 0},
	},
	{
		n: 2,
		adj: [][]int{
			[]int{0},
		},
		dist: [][]int{
			[]int{NoPath},
		},
		add: []Location{1, 0, 1, 0, 1, 0},
	},
	{
		n: 3,
		adj: [][]int{
			[]int{0, 1},
			[]int{1},
		},
		dist: [][]int{
			[]int{2, 1},
			[]int{1},
		},
		add: []Location{0, 1, 2},
	},
	{
		n: 3,
		adj: [][]int{
			[]int{0, 1},
			[]int{1},
		},
		dist: [][]int{
			[]int{2, 1},
			[]int{1},
		},
		add: []Location{2, 1, 0},
	},
	{
		n: 3,
		adj: [][]int{
			[]int{0, 1},
			[]int{1},
		},
		dist: [][]int{
			[]int{2, 1},
			[]int{1},
		},
		add: []Location{1, 0, 2},
	},
	{
		n: 4,
		adj: [][]int{
			[]int{1, 1, 0},
			[]int{0, 1},
			[]int{1},
		},
		dist: [][]int{
			[]int{1, 1, 2},
			[]int{2, 1},
			[]int{1},
		},
		add: []Location{0, 1, 2, 3},
	},
}

func cleanBig() {
	for i := range big {
		big[i] = 0
	}
}

func TestFairLocator(t *testing.T) {
	for testInd, test := range fairLocatorTests {
		cleanBig()
		l := NewFairPathLocator(&test, big)
		for _, loc := range test.add {
			l.Add(loc)
			for l.NeedUpdate() {
				l.UpdateStep()
			}
		}
		for i := 0; i < test.n; i++ {
			for j := 0; j < i; j++ {
				want := test.dist[j][i-j-1]
				got := l.Dist(Location(j), Location(i))
				if want != got {
					t.Errorf("test #%d: %d = test.dist[%d][%d-%d-1] != l.Dist(%d, %d) = %d", testInd, want, j, i, j, j, i, got)
				}
				got = l.Dist(Location(i), Location(j))
				if want != got {
					t.Errorf("test #%d: %d = test.dist[%d][%d-%d-1] != l.Dist(%d, %d) = %d", testInd, want, j, i, j, i, j, got)
				}
			}
		}
	}
}
