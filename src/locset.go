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
     return &locSet{ a: make([]int, size) }
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
