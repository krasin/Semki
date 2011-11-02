package main

type LocSet interface {
	Add(loc Location)
	Has(loc Location) bool
	Clear()
}

type locSet struct {
	a []int
	b int
}

func NewLocSet(size int) LocSet {
	return &locSet{a: make([]int, size), b: 1}
}

func (s *locSet) Clear() {
	s.b++
}

func (s *locSet) Add(loc Location) {
	s.a[loc] = s.b
}

func (s *locSet) Has(loc Location) bool {
	return s.a[loc] == s.b
}

type LocListMap interface {
	Add(at Location, value Location)
	Get(at Location) []Location
	Clear()
}

type locListMap struct {
	l [][]Location
	s LocSet
}

func NewLocListMap(size int) LocListMap {
	return &locListMap{
		l: make([][]Location, size),
		s: NewLocSet(size),
	}
}

func (s *locListMap) Add(at Location, value Location) {
	if !s.s.Has(at) {
		s.l[at] = nil
		s.s.Add(at)
	}
	s.l[at] = append(s.l[at], value)
}

func (s *locListMap) Get(at Location) []Location {
	if s.s.Has(at) {
		return s.l[at]
	}
	return nil
}

func (s *locListMap) Clear() {
	s.s.Clear()
}
