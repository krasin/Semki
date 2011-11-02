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
		if t.adj[from][to] == 1 {
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
}

func cleanBig() {
	for i := range big {
		big[i] = 0
	}
}

func TestFairLocator(t *testing.T) {
	for _, test := range fairLocatorTests {
		cleanBig()
		l := NewFairPathLocator(&test, big)
		for _, loc := range test.add {
			l.Add(loc)
		}
		for i := 0; i < test.n; i++ {
			for j := 0; j < i; j++ {
				want := test.dist[j][i]
				got := l.Dist(Location(j), Location(i))
				if want != got {
					t.Errorf("%d = test.dist[%d][%d] != l.Dist(%d, %d) = %d", want, j, i, j, i, got)
				}
				got = l.Dist(Location(i), Location(j))
				if want != got {
					t.Errorf("%d = test.dist[%d][%d] != l.Dist(%d, %d) = %d", want, j, i, i, j, got)
				}
			}
		}
	}
}
