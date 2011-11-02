package main

// This file defines country as a set of provinces.
//
// Distance between two cells is the length of the
// shortest path between the given cells
//
// Region is a connected set of cells
//
// Observable region is a region that has at least one cell (center)
// from which other cells of this region are observable
// (i.e. the squared Cartesian distance from the center to any node
// is less or equal to ViewRadius2)
//
// Radius of a region is the maximum distance
// between the center and other cells.
//
// All discovered part of the map reachable from our hills is
// divided by provinces -- a set of observable regions with
// the radius smaller than ProvinceRadius which
// have no common cells.
//
// Provinces are connected if they have at least one pair
// of connected cells.

import (
	"fmt"
	"os"
	"sort"
)

const MaxRadius = 10
const JoinSize = 8

type Province struct {
	Ind    int
	Center Location
	Size   int
	Conn   []int
	Dist   int
}

func (p *Province) ConnectedWith(ind int) bool {
	for _, v := range p.Conn {
		if v == ind {
			return true
		}
	}
	return false
}

func (p *Province) Live() bool {
	return p.Size > 0
}

type Country struct {
	T Torus

	m *Map

	// Centers of the provinces.
	// Initially, my hills are the centers of the first provinces
	prov []Province

	// A mapping from location to the index of province.
	// -1 is for 'no province'
	cells []int

	// Distances from the cell to the center of the province
	// It's 0 for cells outside of provinces
	dist []int

	// The list of cells which are on the border of the area
	// that can be reached from hills
	// This list is used to discover newly connected cells
	// and adding them to the provinces
	borders []Location

	pathSlow_used     LocIntMap
	provPath_provUsed LocIntMap
}

// Creates an empty country with initial provinces with centers in my hills
func NewCountry(m *Map) (cn *Country) {
	cn = &Country{
		T:                 m.T,
		m:                 m,
		cells:             make([]int, m.T.Size()),
		dist:              make([]int, m.T.Size()),
		pathSlow_used:     NewLocIntMap(m.T.Size()),
		provPath_provUsed: NewLocIntMap(m.T.Size()),
	}
	for i := range cn.cells {
		cn.cells[i] = -1
	}
	for _, hill := range m.MyHills() {
		cn.AddCell(hill.Loc)
	}
	cn.Update()
	return
}

func (cn *Country) ProvCount() int {
	return len(cn.prov)
}

func (cn *Country) addProvince(center Location) {
	index := len(cn.prov)
	cn.prov = append(cn.prov, Province{Center: center, Size: 1, Ind: index})
	cn.cells[center] = index
}

func (cn *Country) IsCenter(loc Location) bool {
	return cn.IsOwn(loc) && cn.Prov(loc).Center == loc
}

func (cn *Country) Prov(loc Location) (prov *Province) {
	if !cn.IsOwn(loc) {
		return nil
	}
	prov = &cn.prov[cn.cells[loc]]
	for !prov.Live() {
		prov = &cn.prov[cn.cells[prov.Center]]

		// This is intended: we hope to cut the length of chain by 2, not by 1
		cn.cells[loc] = cn.cells[prov.Center]
	}
	return
}

func (cn *Country) updateDist(at Location) {
	q := []Location{at}
	for len(q) > 0 {
		cur := q
		q = nil
		for _, loc := range cur {
			dist := cn.dist[loc]
			for _, cell := range cn.T.Neighbours(loc) {
				if cn.IsOwn(cell) && cn.dist[cell] > dist+1 && !cn.IsCenter(cell) {
					cn.dist[cell] = dist + 1
					cn.cells[cell] = cn.cells[loc]
					q = append(q, cell)
				}
			}
		}
	}
}

// Suboptimal implementation
func removeDuplicates(a []int, what, with int) (b []int) {
	for i := 0; i < len(a); i++ {
		if a[i] == what {
			a[i] = with
		}
	}
	sort.Ints(a)
	for i := 0; i < len(a); i++ {
		if i > 0 && a[i] == a[i-1] {
			continue
		}
		b = append(b, a[i])
	}
	return
}

func (cn *Country) join(what, with *Province) {
	whatInd := cn.cells[what.Center]
	withInd := cn.cells[with.Center]
	what.Size += with.Size
	with.Size = 0
	with.Center = what.Center
	conn := what.Conn
	what.Conn = nil
	conn = append(conn, with.Conn...)
	with.Conn = nil
	for i := 0; i < len(conn); i++ {
		c := cn.Prov(cn.prov[conn[i]].Center)
		if c == with || c == what {
			conn[i] = conn[len(conn)-1]
			conn = conn[:len(conn)-1]
		}
	}
	what.Conn = removeDuplicates(conn, -1, -1)
	for _, ind := range what.Conn {
		cn.prov[ind].Conn = removeDuplicates(cn.prov[ind].Conn, whatInd, withInd)
	}
	if what.Dist > with.Dist {
		what.Dist = with.Dist
	}
}

func (cn *Country) updateConnections(loc Location) {
	prov := cn.Prov(loc)
	for _, cell := range cn.T.Neighbours(loc) {
		cur := cn.Prov(cell)
		if cur == nil {
			continue
		}
		if prov == cur {
			continue
		}
		if prov.ConnectedWith(cn.cells[cur.Center]) {
			continue
		}
		if prov.Size < JoinSize && cur.Size < JoinSize {
			if prov.Size < cur.Size {
				cn.join(cur, prov)
			} else {
				cn.join(prov, cur)
			}
		} else {
			prov.Conn = append(prov.Conn, cn.cells[cur.Center])
			cur.Conn = append(cur.Conn, cn.cells[prov.Center])
		}
	}
}

func (cn *Country) AddCell(loc Location) {
	if cn.IsOwn(loc) {
		return
	}
	if !cn.CanOwn(loc) {
		return
	}
	minDist := -1
	var bestProv *Province
	for _, cell := range cn.T.Neighbours(loc) {
		if cn.IsOwn(cell) && (cn.dist[cell]+1 < minDist || bestProv == nil) {
			bestProv = cn.Prov(cell)
			minDist = cn.dist[cell] + 1
		}
	}
	if bestProv == nil || minDist > MaxRadius {
		cn.addProvince(loc)
	} else {
		cn.dist[loc] = minDist
		cn.cells[loc] = cn.cells[bestProv.Center]
		bestProv.Size++
	}
	cn.borders = append(cn.borders, loc)
	cn.updateConnections(loc)

	// Update province distance to hill province
	// This can be wrong in case of merging areas from different hills
	// But it's somehow fine.
	prov := cn.Prov(loc)
	dist := -1
	for _, ind := range prov.Conn {
		another := cn.prov[ind]
		if dist == -1 || another.Dist < dist {
			dist = another.Dist
		}
	}
	prov.Dist = dist + 1
}

// Returns a province that's closer to a hill than the given one\
// Returns itself, if it's a hill province
func (cn *Country) CloserProv(prov *Province) *Province {
	if prov == nil || prov.Dist == 0 {
		return prov
	}

	bestDist := -1
	var bestProv *Province
	for _, ind := range prov.Conn {
		another := cn.prov[ind]
		if bestDist == -1 || another.Dist < bestDist {
			bestDist = another.Dist
			bestProv = prov
		}
	}
	return bestProv
}

func (cn *Country) ProvByIndex(ind int) *Province {
	if ind < 0 || ind >= len(cn.prov) {
		return nil
	}
	return &cn.prov[ind]
}

func (cn *Country) PathSlow(from, to Location) (p Path) {
	p = NewPath(cn.T, from)
	if to == from {
		return
	}
	a := cn.pathSlow_used
	a.Clear()
	a.Add(to, 1)
	q := []Location{to}
	var q2 []Location
	for len(q) > 0 {
		tmp := q2
		q2 = q
		q = tmp[:0]
		for _, loc := range q2 {
			for _, cell := range cn.m.LandNeighbours(loc) {
				if a.Get(cell) == 0 {
					a.Add(cell, a.Get(loc)+1)
					q = append(q, cell)
				}
				if cell == from {
					break
				}
			}
		}
	}
	if a.Get(from) == 0 {
		return nil
	}
	// Now, collect the path
	cur := from
	for cur != to {
		for _, cell := range cn.m.LandNeighbours(cur) {
			if a.Get(cell) == a.Get(cur)-1 {
				p.Append(cn.T.GuessDir(cur, cell))
				cur = cell
				break
			}
		}
	}
	return
}

func (cn *Country) ProvPath(fromProv, toProv *Province) (res []*Province) {
	fmt.Fprintf(os.Stderr, "ProvPath, 0\n")
	if fromProv == toProv {
		fmt.Fprintf(os.Stderr, "ProvPath, exit, trivial case\n")
		return
	}
	used := cn.provPath_provUsed
	used.Clear()
	used.Add(toProv.Center, 1)
	q := []Location{toProv.Center}
	var q2 []Location
	for len(q) > 0 {
		q, q2 = q2[:0], q
		for _, loc := range q2 {
			if loc == fromProv.Center {
				// Collect path
				fmt.Fprintf(os.Stderr, "Collecting path...\n")
				res = append(res, fromProv)
				cur := fromProv
				val := used.Get(loc)
				if val == 0 {
					panic("val == 0")
				}

				for cur != toProv {
					found := false
					for _, ind := range cur.Conn {
						cell := cn.ProvByIndex(ind).Center
						if used.Get(cell) == val-1 {
							found = true
							cur = cn.Prov(cell)
							res = append(res, cur)
							val--
							break
						}
					}
					if !found {
						panic("not found")
					}
				}
				fmt.Fprintf(os.Stderr, "ProvPath, exit, found\n")
				return
			}
			for _, ind := range cn.Prov(loc).Conn {
				cell := cn.ProvByIndex(ind).Center
				if used.Get(cell) == 0 {
					q = append(q, cell)
					used.Add(cell, used.Get(loc)+1)
				}
			}
		}
	}
	panic("ProvPath: not reachable")
}

// Path returns an approximately shortest path between two location.
// Returns nil if there's no path found
func (cn *Country) Path(from, to Location) Path {
	fromProv := cn.Prov(from)
	toProv := cn.Prov(to)
	if fromProv == nil || toProv == nil {
		return nil
	}
	if fromProv == toProv {
		// These locations are in one province.
		// It's faster to find the real shortest path
		p := cn.PathSlow(from, to)
		if p == nil {
			// This is a bug: we can't find a path between two cells in one province
			panic("PathSlow is unable to find a path between two cells in one province. This is a programmer's mistake!")
		}
		return p
	}

	provPath := cn.ProvPath(fromProv, toProv)
	if len(provPath) == 0 {
		panic("len(provPath) == 0")
	}
	if len(provPath) == 1 {
		panic("len(provPath) == 1")

	}
	path := cn.PathSlow(from, provPath[1].Center)
	fmt.Fprintf(os.Stderr, "First chunk: %v\n", path)
	provPath = provPath[1:]
	for i := range provPath {
		if i == 0 {
			continue
		}
		next := cn.PathSlow(provPath[i-1].Center, provPath[i].Center)
		fmt.Fprintf(os.Stderr, "Next: %v\n", next)
		AppendPath(path, next)
		fmt.Fprintf(os.Stderr, "Appended: %v\n", path)
	}
	AppendPath(path, cn.PathSlow(toProv.Center, to))
	fmt.Fprintf(os.Stderr, "Path(%d, %d): %v\n", from, to, path)
	return path
	//	panic("Path not implemented")
}

func (cn *Country) IsOwn(loc Location) bool {
	return cn.cells[loc] != -1
}

func (cn *Country) CanOwn(loc Location) bool {
	return cn.m.Terrain[loc] == Land
}

func (cn *Country) Update() {
	var old []Location
	for len(cn.borders) > 0 {
		borders := cn.borders
		cn.borders = nil

		for _, loc := range borders {
			was := false
			for _, cell := range cn.T.Neighbours(loc) {
				switch cn.m.Terrain[cell] {
				case Unknown:
					was = true
				case Water:
				case Land:
					if cn.IsOwn(cell) {
						continue
					}
					cn.AddCell(cell)
				}
			}
			if was {
				// This location is connected with unknown territory.
				old = append(old, loc)
			}
		}
	}
	cn.borders = old
}

var colors = []int{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n'}

func (cn *Country) Dump(filename string) (err os.Error) {
	var f *os.File
	if f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return
	}
	defer f.Close()
	for row := 0; row < cn.T.Rows; row++ {
		for col := 0; col < cn.T.Cols; col++ {
			ch := 'E'
			cell := cn.T.Loc(row, col)
			switch cn.m.Terrain[cell] {
			case Land:
				switch {
				case cn.IsCenter(cell):
					ch = '*'
				case cn.IsOwn(cell):
					provInd := cn.Prov(cell).Ind
					if provInd < len(colors) {
						ch = colors[provInd]
					} else {
						ch = '+'
					}
					//				case cn.isBorder(cell):
					//					ch = '+'
				default:
					ch = '.'
				}
			case Water:
				ch = '%'
			case Unknown:
				ch = '?'

			}
			f.Write([]byte{byte(ch)})
		}
		f.Write([]byte{'\n'})
	}
	return
}
