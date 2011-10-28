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
)

const MaxRadius = 6

type Country struct {
	m *Map

	// Centers of the provinces.
	// Initially, my hills are the centers of the first provinces
	centers []Location

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
	index := len(cn.centers)
	cn.centers = append(cn.centers, center)
	cn.cells[center] = index
}

func (cn *Country) IsCenter(loc Location) bool {
	return cn.IsOwn(loc) && cn.centers[cn.cells[loc]] == loc
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

func (cn *Country) AddCell(loc Location) {
	if cn.IsOwn(loc) {
		return
	}
	if !cn.CanOwn(loc) {
		return
	}
	minDist := -1
	bestProvince := -1
	for _, cell := range cn.m.Neighbours(loc) {
		if cn.IsOwn(cell) && (cn.dist[cell]+1 < minDist || minDist == -1) {
			bestProvince = cn.cells[cell]
			minDist = cn.dist[cell] + 1
		}
	}
	if bestProvince == -1 || minDist > MaxRadius {
		cn.cells[loc] = len(cn.centers)
		cn.centers = append(cn.centers, loc)
		cn.updateDist(loc)
	} else {
		cn.dist[loc] = minDist
		cn.cells[loc] = bestProvince
	}
	cn.borders = append(cn.borders, loc)
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
