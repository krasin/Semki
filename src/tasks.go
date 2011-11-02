package main

import (
	"sort"
)

type Locator interface {
	Dist(from, to Location) int
}

type LocatedSet interface {
	All() []Location
	FindNear(loc Location, ok func(loc Location) bool) (Location, bool)
}

type Assignment struct {
	Worker Location
	Target Location
	Score  int
}

type Planner interface {
	Plan(l Locator, prev []Assignment, workerSet LocatedSet, targets []Location, scores []int) []Assignment
}

type greedyPlanner struct {
	size            int
	assignedTargets LocSet
	assignedWorkers LocSet
}

type combineSortInterface struct {
	targets []Location
	scores  []int
}

type sorter struct {
	targets []Location
	scores  []int
}

func (s *sorter) Len() int {
	return len(s.scores)
}

func (s *sorter) Less(i, j int) bool {
	return s.scores[i] > s.scores[j]
}

func (s *sorter) Swap(i, j int) {
	s.targets[i], s.targets[j] = s.targets[i], s.targets[j]
	s.scores[i], s.scores[j] = s.scores[j], s.scores[i]
}

func (p *greedyPlanner) Plan(l Locator, prev []Assignment, workerSet LocatedSet, targets []Location, scores []int) (res []Assignment) {
	workers := workerSet.All()

	// Remember assigned workers and targets
	// Greedy planner will not deassign workers from targets
	p.assignedTargets.Clear()
	p.assignedWorkers.Clear()
	for _, assign := range prev {
		p.assignedTargets.Add(assign.Target)
		p.assignedWorkers.Add(assign.Worker)
		res = append(res, assign)
	}

	// Filter assigned workers
	for i := 0; i < len(workers); i++ {
		if p.assignedWorkers.Has(workers[i]) {
			workers[i] = workers[len(workers)-1]
			workers = workers[0 : len(workers)-1]
			i--
		}
	}
	// Filter assigned targets
	for i := 0; i < len(targets); i++ {
		if p.assignedTargets.Has(targets[i]) {
			targets[i] = targets[len(targets)-1]
			scores[i] = scores[len(scores)-1]
			targets = targets[0 : len(targets)-1]
			scores = scores[0 : len(scores)-1]
			i--
		}
	}

	// Sort targets desc
	sort.Sort(&sorter{targets, scores})

	for _, t := range targets {
		// Find closest unassigned worker
		w, found := workerSet.FindNear(t, func(loc Location) bool {
			return !p.assignedWorkers.Has(loc)
		})
		if !found {
			// There's no available workers for this target
			// Probably, it means that there's no available workers at all,
			// but it also can mean that we have no workers in the current area
			// (it can be several areas if more than one player hill exists
			// FIXME: stop planning if all workers are set
			continue
		}
		res = append(res, Assignment{Worker: w, Target: t})
		p.assignedWorkers.Add(w)
		p.assignedTargets.Add(t)
	}

	return res
}

func NewGreedyPlanner(size int) Planner {
	return &greedyPlanner{size: size,
		assignedTargets: NewLocSet(size),
		assignedWorkers: NewLocSet(size),
	}
}
