package main

import (
	"fmt"
)

type Path interface {
	Advance(hops int) bool
	Append(dir Direction)
	Len() int
	Dir(ind int) Direction
}

type PathFinder interface {
	Path(from, to Location) Path
}

type pathFinder struct {
	t Torus
	c Connector
	l Locator
}

func NewPathFinder(t Torus, c Connector, l Locator) PathFinder {
	return &pathFinder{t: t, c: c, l: l}
}

func (f *pathFinder) Path(from, to Location) (p Path) {
	dist := f.l.Dist(from, to)
	if dist == NoPath {
		return
	}
	p = NewPath(f.t, from)
	if from == to {
		return
	}
	cur := from
	for cur != to {
		found := false
		for _, conn := range f.c.Conn(cur) {
			if f.l.Dist(conn, to) <= dist-1 {
				p.Append(f.t.GuessDir(cur, conn))
				cur = conn
				found = true
				dist--
				break
			}
		}
		if !found {
			panic(fmt.Sprintf("pathFinder.Path(%d, %d): !found", from, to))
		}
	}
	return
}

type path struct {
	t Torus
	l []Location
}

func NewPath(t Torus, from Location) Path {
	return &path{
		t: t,
		l: []Location{from},
	}
}

func (p *path) Advance(hops int) bool {
	if hops < 0 || hops > p.Len() {
		return false
	}
	p.l = p.l[hops:]
	return true
}

func (p *path) Append(dir Direction) {
	loc := p.t.NewLoc(p.l[len(p.l)-1], dir)
	p.l = append(p.l, loc)
}

func (p *path) Len() int {
	if len(p.l) == 0 {
		return 0
	}
	return len(p.l) - 1
}

func (p *path) Dir(ind int) Direction {
	return p.t.GuessDir(p.l[ind], p.l[ind+1])
}

func AppendPath(dest, source Path) {
	if source == nil {
		return
	}
	for i := 0; i < source.Len(); i++ {
		dest.Append(source.Dir(i))
	}
}
