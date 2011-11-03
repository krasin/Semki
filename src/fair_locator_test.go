package main

import (
	//	"fmt"
	"rand"
	"testing"
)

type fairLocatorTest struct {
	n                  int
	adj                [][]int      // just for row < col
	dist               [][]int      // just for tow < col
	run                [][]Location // add order
	skipDistValidation bool         // this is set by pseudo-random tests until we would get a way to test them
}

func (t *fairLocatorTest) Conn(loc Location) (res []Location) {
	//	fmt.Printf("Conn, loc: %d\n", loc)
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
		//		fmt.Printf("from: %d, to: %d, t.adj: %v\n", from, to, t.adj)
		if t.adj[from][to-from-1] == 1 {
			res = append(res, Location(i))
		}
	}
	return
}

func pseudoRandomTest(n int, seed int64, p int) (test fairLocatorTest) {
	test.n = n
	test.run = [][]Location{make([]Location, n)}
	for i := 0; i < n; i++ {
		test.run[0][i] = Location(i)
	}
	rnd := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		var line []int
		for j := i + 1; j < n; j++ {
			adj := 0
			if rnd.Intn(100) < p {
				adj = 1
			}
			line = append(line, adj)
		}
		test.adj = append(test.adj, line)
	}
	test.skipDistValidation = true
	return
}

var fairLocatorTests = []fairLocatorTest{
	{
		n:    1,
		adj:  [][]int{},
		dist: [][]int{},
		run:  [][]Location{[]Location{0}},
	},
	{
		n: 2,
		adj: [][]int{
			[]int{1},
		},
		dist: [][]int{
			[]int{1},
		},
		run: [][]Location{
			[]Location{0, 1},
			[]Location{1, 0},
			[]Location{1, 0, 1, 0, 1, 0},
		},
	},
	{
		n: 2,
		adj: [][]int{
			[]int{0},
		},
		dist: [][]int{
			[]int{NoPath},
		},
		run: [][]Location{
			[]Location{0, 1},
			[]Location{1, 0},
			[]Location{1, 0, 1, 0, 1, 0},
		},
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
		run: [][]Location{
			[]Location{0, 1, 2},
			[]Location{2, 1, 0},
			[]Location{1, 0, 2},
		},
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
		run: [][]Location{
			[]Location{0, 1, 2, 3},
			[]Location{3, 0, 1, 2},
		},
	},
	pseudoRandomTest(3, 0, 50),
	pseudoRandomTest(10, 0, 50),
	pseudoRandomTest(100, 0, 50),
	pseudoRandomTest(200, 0, 50),
	pseudoRandomTest(200, 0, 20),
	pseudoRandomTest(200, 0, 70),
	pseudoRandomTest(200, 0, 10),
}

func cleanBig() {
	for i := range big {
		big[i] = 0
	}
}

func TestFairLocator(t *testing.T) {
	for testInd, test := range fairLocatorTests {
		for runInd, run := range test.run {
			cleanBig()
			l := NewFairPathLocator(&test, big)
			for _, loc := range run {
				l.Add(loc)
				for l.NeedUpdate() {
					l.UpdateStep()
				}
			}
			if test.skipDistValidation {
				continue
			}
			for i := 0; i < test.n; i++ {
				for j := 0; j < i; j++ {
					want := test.dist[j][i-j-1]
					got := l.Dist(Location(j), Location(i))
					if want != got {
						t.Errorf("test #%d, run #%d: %d = test.dist[%d][%d-%d-1] != l.Dist(%d, %d) = %d", testInd, runInd, want, j, i, j, j, i, got)
					}
					got = l.Dist(Location(i), Location(j))
					if want != got {
						t.Errorf("test #%d: run #d: %d = test.dist[%d][%d-%d-1] != l.Dist(%d, %d) = %d", testInd, runInd, want, j, i, j, i, j, got)
					}
				}
			}
		}
	}
}
