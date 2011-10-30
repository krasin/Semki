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
	"os"
	"sort"
)

const MaxRadius = 6
const JoinSize = 8

type Province struct {
	Center Location
	Size   int
	Conn   []int
}

func (p *Province) ConnectedWith(ind int) bool {
	for _, v := range p.Conn {
		if v == ind {
			return true
		}
	}
	return false
}

type Country struct {
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
}

// Creates an empty country with initial provinces with centers in my hills
func NewCountry(m *Map) (cn *Country) {
	cn = &Country{
		m:     m,
		cells: make([]int, m.Rows*m.Cols),
		dist:  make([]int, m.Rows*m.Cols),
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

func (cn *Country) addProvince(center Location) {
	index := len(cn.prov)
	cn.prov = append(cn.prov, Province{Center: center, Size: 1})
	cn.cells[center] = index
}

func (cn *Country) IsCenter(loc Location) bool {
	return cn.IsOwn(loc) && cn.prov[cn.cells[loc]].Center == loc
}

func (cn *Country) Prov(loc Location) (prov *Province) {
	if !cn.IsOwn(loc) {
		return nil
	}
	prov = &cn.prov[cn.cells[loc]]
	for prov.Size == 0 {
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
			for _, cell := range cn.m.Neighbours(loc) {
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
}

func (cn *Country) updateConnections(loc Location) {
	prov := cn.Prov(loc)
	for _, cell := range cn.m.Neighbours(loc) {
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
	for _, cell := range cn.m.Neighbours(loc) {
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
			for _, cell := range cn.m.Neighbours(loc) {
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

func (cn *Country) Dump(filename string) (err os.Error) {
	var f *os.File
	if f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return
	}
	defer f.Close()
	for row := 0; row < cn.m.Rows; row++ {
		for col := 0; col < cn.m.Cols; col++ {
			ch := 'E'
			cell := cn.m.Loc(row, col)
			switch cn.m.Terrain[cell] {
			case Land:
				switch {
				case cn.IsCenter(cell):
					ch = '*'
				case cn.IsOwn(cell):
					ch = '+'
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
