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
	distances []int

	// The list of cells which are on the border of the area
	// that can be reached from hills
	// This list is used to discover newly connected cells
	// and adding them to the provinces
	borders []Location
}

// Creates an empty country with initial provinces with centers in my hills
func NewCountry(m *Map) (cn *Country) {
	cn = &Country{
		m:         m,
		cells:     make([]int, m.Rows*m.Cols),
		distances: make([]int, m.Rows*m.Cols),
	}
	for i := range cn.cells {
		cn.cells[i] = -1
	}
	for _, hill := range m.MyHills() {
		cn.tryAddBorder(hill.Loc)
	}
	cn.Update()
	return
}

func (cn *Country) addProvince(center Location) {
	index := len(cn.centers)
	cn.centers = append(cn.centers, center)
	cn.cells[center] = index
}

func (cn *Country) isBorder(loc Location) bool {
	if cn.m.Terrain[loc] != Land {
		return false
	}
	for _, cell := range cn.m.Neighbours(loc) {
		if cn.cells[cell] == -1 && cn.m.Terrain[cell] != Water {
			return true
		}
	}
	return false
}

func (cn *Country) updateCell(loc Location) bool {
	if cn.cells[loc] >= 0 {
		return false
	}
	minDist := -1
	bestProvince := -1
	for _, cell := range cn.m.Neighbours(loc) {
		curDist := cn.distances[cell] + 1
		curProvince := cn.cells[cell]
		if curProvince >= 0 && curDist < minDist || minDist == -1 {
			minDist = curDist
			bestProvince = curProvince
		}
	}
	if minDist == -1 {
		cn.addProvince(loc)
		return true
	}
	cn.cells[loc] = bestProvince
	cn.distances[loc] = minDist
	return true
}

func (cn *Country) tryAddBorder(loc Location) bool {
	if !cn.isBorder(loc) {
		return false
	}
	if cn.updateCell(loc) {
		cn.borders = append(cn.borders, loc)
		return true
	}
	return false
}                  This is completely broken....

func (cn *Country) Update() {
	changed := true
	for changed {
		changed = false
		borders := cn.borders
		cn.borders = nil
		for _, loc := range borders {
			for _, cell := range cn.m.Neighbours(loc) {
				if cn.cells[cell] == -1 && cn.m.Terrain[cell] == Land {
					if cn.tryAddBorder(cell) {
						changed = true
					}
				}
			}

		}
	}
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
				case cn.distances[cell] == 0:
					ch = '*'
				case cn.isBorder(cell):
					ch = '+'
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
