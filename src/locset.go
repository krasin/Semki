package main

type LocSet interface {
	Add(loc Location)
	Has(loc Location) bool
	All() []Location
	Clear()
}

type locSet struct {
	a []int
	l []Location
	b int
}

func NewLocSet(size int) LocSet {
	return &locSet{a: make([]int, size), b: 1}
}

func (s *locSet) Clear() {
	s.b++
	s.l = s.l[:0]
}

func (s *locSet) Add(loc Location) {
	if !s.Has(loc) {
		s.a[loc] = s.b
		s.l = append(s.l, loc)
	}
}

func (s *locSet) Has(loc Location) bool {
	return s.a[loc] == s.b
}

func (s *locSet) All() (res []Location) {
	res = make([]Location, len(s.l))
	copy(res, s.l)
	return
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

type LocIntMap interface {
	Add(at Location, value int)
	Get(at Location) int
	All() []Location
	Clear()
}

type locIntMap struct {
	a []int
	s LocSet
}

func NewLocIntMap(size int) LocIntMap {
	return &locIntMap{
		a: make([]int, size),
		s: NewLocSet(size),
	}
}

func (m *locIntMap) Add(at Location, value int) {
	m.s.Add(at)
	m.a[at] = value
}

func (m *locIntMap) Get(at Location) int {
	if !m.s.Has(at) {
		return 0
	}
	return m.a[at]
}

func (m *locIntMap) Clear() {
	m.s.Clear()
}

func (m *locIntMap) All() []Location {
	return m.s.All()
}

type LocLocMap interface {
	Add(at Location, value Location)
	Get(at Location) Location
	Clear()
}

type locLocMap struct {
	a []Location
	s LocSet
}

func NewLocLocMap(size int) LocLocMap {
	return &locLocMap{
		a: make([]Location, size),
		s: NewLocSet(size),
	}
}

func (m *locLocMap) Add(at Location, value Location) {
	m.s.Add(at)
	m.a[at] = value
}

func (m *locLocMap) Get(at Location) Location {
	if !m.s.Has(at) {
		return 0
	}
	return m.a[at]
}

func (m *locLocMap) Clear() {
	m.s.Clear()
}
