package main

import (
	"fmt"
	"os"
	"sort"
)

const ReassignThresholdRatio = 1.5
const ReassignDist = 5

type Connector interface {
	Conn(loc Location) []Location
}

type Locator interface {
	Dist(from, to Location) int
}

type LocatedSet interface {
	All() []Location
	FindNear(loc Location, score int, ok func(worker Location, score int, sameProv bool) bool) (Location, bool)
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
	size                     int
	assignedTargets          LocSet
	assignedWorkers          LocIntMap
	assignedWorkersToTargets LocLocMap
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
	tmpTarget := s.targets[i]
	s.targets[i] = s.targets[j]
	s.targets[j] = tmpTarget

	tmpScore := s.scores[i]
	s.scores[i] = s.scores[j]
	s.scores[j] = tmpScore
}

func (p *greedyPlanner) Plan(l Locator, prev []Assignment, workerSet LocatedSet, targets []Location, scores []int) (res []Assignment) {
	workers := workerSet.All()

	// Remember assigned workers and targets
	// Greedy planner will not deassign workers from targets
	p.assignedTargets.Clear()
	p.assignedWorkers.Clear()
	p.assignedWorkersToTargets.Clear()
	for _, assign := range prev {
		p.assignedTargets.Add(assign.Target)
		p.assignedWorkers.Add(assign.Worker, assign.Score)
		p.assignedWorkersToTargets.Add(assign.Worker, assign.Target)
	}

	// Filter assigned workers
	for i := 0; i < len(workers); i++ {
		if p.assignedWorkers.Get(workers[i]) > 0 {
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

	for i, t := range targets {
		// Find closest unassigned worker
		w, found := workerSet.FindNear(t, scores[i], func(worker Location, score int, sameProv bool) bool {
			ascore := p.assignedWorkers.Get(worker)
			if ascore == 0 || int(float64(ascore)*ReassignThresholdRatio) < score {
				return true
			}
			newDist := l.Dist(worker, t)
			if (sameProv || newDist >= 0 && newDist < ReassignDist) && score >= ascore {
				curDist := l.Dist(worker, p.assignedWorkersToTargets.Get(worker))
				return newDist > 0 && (curDist == -1 || newDist < curDist)
			}
			return false
		})
		if !found {
			// There's no available workers for this target
			// Probably, it means that there's no available workers at all,
			// but it also can mean that we have no workers in the current area
			// (it can be several areas if more than one player hill exists
			// FIXME: stop planning if all workers are set
			continue
		}
		p.assignedWorkers.Add(w, scores[i])
		p.assignedTargets.Add(t)
		p.assignedWorkersToTargets.Add(w, t)
	}
	fmt.Fprintf(os.Stderr, "Assigned workers: %v\n", p.assignedWorkers.All())
	for _, w := range p.assignedWorkers.All() {
		res = append(res, Assignment{
			Worker: w,
			Target: p.assignedWorkersToTargets.Get(w),
			Score:  p.assignedWorkers.Get(w),
		})
	}

	return res
}

func NewGreedyPlanner(size int) Planner {
	return &greedyPlanner{size: size,
		assignedTargets:          NewLocSet(size),
		assignedWorkers:          NewLocIntMap(size),
		assignedWorkersToTargets: NewLocLocMap(size),
	}
}
