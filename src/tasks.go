package main

import (
	"sort"
)

type Target struct {
	Loc   Location
	Score int
}

type Worker struct {
	Loc     Location
	LiveInd int
	Prov    int
}

type Estimator interface {
	Estimate(w *Worker, loc Location) int
	Prov(loc Location) int
	Conn(prov int) []int
}

type Planner struct {
	est          Estimator
	worker       []*Worker
	workerByProv [][]int
	target       []Target
	assign       []internalAssignment
	score        int
}

type Assignment struct {
	Worker *Worker
	Target Target
}

type internalAssignment struct {
	worker int
	target int
	score  int
}

func NewPlanner(est Estimator, provCount int) *Planner {
	return &Planner{est: est, workerByProv: make([][]int, provCount)}
}

func (p *Planner) AddWorker(w *Worker) {
	ind := len(p.worker)
	p.worker = append(p.worker, w)
	p.workerByProv[w.Prov] = append(p.workerByProv[w.Prov], ind)
}

func (p *Planner) AddTarget(loc Location, score int) {
	p.target = append(p.target, Target{Loc: loc, Score: score})
}

func (p *Planner) scoreAssignment(w *Worker, loc Location, score int) int {
	estimate := p.est.Estimate(w, loc)
	if estimate == -1 {
		return -1
	}
	return score / (1 + estimate)
}

func (p *Planner) AddAssignedWorker(w *Worker, loc Location, score int) {
	wInd := len(p.worker)
	p.worker = append(p.worker, w)
	// TODO: enable when the proper support of assigned works is implemented
	//	p.workerByProv[w.Prov] = append(p.workerByProv[w.Prov], wInd)

	assignScore := p.scoreAssignment(w, loc, score)
	if assignScore == -1 {
		// Impossible to achieve
		return
	}

	tInd := len(p.target)

	// We don't want to reassign this target to anyone else, so score=0
	// Rationale: it's likely that we have a duplicate of the same target
	// with some score. This fake target is to represent assigned worker with 
	// some specific score to this assignment
	p.target = append(p.target, Target{Loc: loc, Score: 0})

	p.assign = append(p.assign, internalAssignment{worker: wInd, target: tInd, score: assignScore})
}

type TargetsDescSlice struct {
	p           *Planner
	targetsDesc []int
}

func (tds *TargetsDescSlice) Len() int {
	return len(tds.p.target)
}

func (tds *TargetsDescSlice) Less(i, j int) bool {
	return tds.p.target[tds.targetsDesc[i]].Score > tds.p.target[tds.targetsDesc[j]].Score
}

func (tds *TargetsDescSlice) Swap(i, j int) {
	tmp := tds.targetsDesc[i]
	tds.targetsDesc[i] = tds.targetsDesc[j]
	tds.targetsDesc[j] = tmp
}

func (p *Planner) assignAny(prov int, tInd int) bool {
	if len(p.workerByProv[prov]) == 0 {
		return false
	}
	p.assign = append(p.assign, internalAssignment{
		worker: p.workerByProv[prov][0],
		target: tInd,
		score:  p.target[tInd].Score,
	})
	p.workerByProv[prov] = p.workerByProv[prov][1:]
	return true
}

func (p *Planner) MakePlan() (res []Assignment) {
	// A simple greedy algorithm:
	// 1. Find the most important target
	// 2. Find the most appropriate ant
	// 3. Assign that ant to the target
	// 4. Go to 1
	// Rule: don't touch assigned workers at all.
	targetsDesc := make([]int, len(p.target))
	for i := 0; i < len(targetsDesc); i++ {
		targetsDesc[i] = i
	}
	sort.Sort(&TargetsDescSlice{p: p, targetsDesc: targetsDesc})
	// Linear time assignments of tasks from the same province
	for _, sliceInd := range targetsDesc {
		tInd := targetsDesc[sliceInd]
		prov := p.est.Prov(p.target[tInd].Loc)
		if p.assignAny(prov, tInd) {
			continue
		}
		for _, connProv := range p.est.Conn(prov) {
			if p.assignAny(connProv, tInd) {
				break
			}
		}
	}

	for _, assign := range p.assign {
		res = append(res, Assignment{
			Worker: p.worker[assign.worker],
			Target: Target{
				p.target[assign.target].Loc,
				assign.score,
			},
		})
	}
	return
}
