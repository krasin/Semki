package main

const bigN = 20000

const NoPath = (1 << 31) - 1

type locPair struct {
	a Location
	b Location
}

type fairPathLocator struct {
	conn    Connector
	big     []int16
	loc2ind []int
	ind2loc []Location

	// indices to update
	toUpdate []locPair
	buf      []locPair
}

func NewFairPathLocator(conn Connector, big []int16) *fairPathLocator {
	return &fairPathLocator{
		conn:    conn,
		big:     big,
		loc2ind: make([]int, 40000),
	}
}

func (l *fairPathLocator) hasLoc(loc Location) bool {
	return l.loc2ind[int(loc)] > 0
}

func (l *fairPathLocator) bigIndex(from, to Location) int {
	fromInd := l.loc2ind[int(from)] - 1
	toInd := l.loc2ind[int(to)] - 1
	if fromInd == -1 || toInd == -1 {
		return -1
	}
	if fromInd == toInd {
		return 0
	}
	if fromInd > toInd {
		fromInd, toInd = toInd, fromInd
	}
	s := bigN*fromInd - (fromInd*(fromInd+1))/2
	return s + (toInd - fromInd - 1)
}

func (l *fairPathLocator) set(from, to Location, dist int) {
	//	fmt.Printf("set(%d, %d, dist=%d)\n", from, to, dist)
	ind := l.bigIndex(from, to)
	//	fmt.Printf("ind: %d\n", ind)
	l.big[ind] = int16(dist)
}

// Update distances for loc looking at from
func (l *fairPathLocator) updatePair(loc, from Location) {
	was := false
	if loc == from {
		return
	}
	for i := 0; i < len(l.ind2loc); i++ {
		to := l.ind2loc[i]
		if to == loc || to == from || !l.hasLoc(to) {
			continue
		}
		curDist := l.Dist(loc, to)
		newDist := l.Dist(from, to)
		if newDist != NoPath && newDist+1 < curDist {
			l.set(loc, to, newDist+1)
			was = true
		}
	}
	if was {
		for _, conn := range l.conn.Conn(loc) {
			if conn != from && l.hasLoc(conn) {
				l.toUpdate = append(l.toUpdate, locPair{conn, loc})
			}
		}
	}
}

func (l *fairPathLocator) NeedUpdate() bool {
	return len(l.toUpdate) > 0
}

func (l *fairPathLocator) Add(locs ...Location) {
	//	fmt.Printf("Add(%d)\n", loc)
	for _, loc := range locs {
		if l.hasLoc(loc) {
			continue
		}
		ind := len(l.ind2loc)
		l.ind2loc = append(l.ind2loc, loc)
		l.loc2ind[int(loc)] = ind + 1

		for _, conn := range l.conn.Conn(loc) {
			if !l.hasLoc(conn) {
				continue
			}
			l.set(loc, conn, 1)
			l.toUpdate = append(l.toUpdate, locPair{loc, conn})
			l.toUpdate = append(l.toUpdate, locPair{conn, loc})
		}
	}
}

func (l *fairPathLocator) UpdateStep() {
	l.toUpdate, l.buf = l.buf[:0], l.toUpdate
	for _, pair := range l.buf {
		l.updatePair(pair.a, pair.b)
	}
}

func (l *fairPathLocator) Update() {
	for l.NeedUpdate() {
		l.UpdateStep()
	}
}

func (l *fairPathLocator) Dist(from, to Location) int {
	if from == to {
		return 0
	}
	ind := l.bigIndex(from, to)
	if ind == -1 {
		return NoPath
	}
	val := int(l.big[ind])
	if val == 0 {
		return NoPath
	}
	return val
}
