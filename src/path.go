package main

type Path interface {
	Advance(hops int) bool
	Append(dir Direction)
	Len() int
	Dir(ind int) Direction
}

// Encodes up to 27 moves
type storedPath uint64

func (p storedPath) Len() int {
	return int(p & 0xFF)
}

func pathCodeToDirection(code int) Direction {
	switch code {
	case 0:
		return North
	case 1:
		return East
	case 2:
		return South
	case 3:
		return West
	}
	panic("PathCodeToDirection: unreachable")
}

func (p storedPath) Dir(ind int, cols int) Direction {
	code := int((p >> (8 + uint64(2*ind))) & 0x3)
	return pathCodeToDirection(code)
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
