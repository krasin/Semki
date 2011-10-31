package main

type Target struct {
	Loc   Location
	Score int
}

type Worker struct {
	loc Location
}

type Estimator interface {
	Estimate(w *Worker, loc Location) int
}

type Planner struct {
	estimator Estimator
	worker    []*Worker
	target    []Target
	assign    []internalAssignment
	score     int
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

func NewPlanner(estimator Estimator) *Planner {
	return &Planner{estimator: estimator}
}

func (p *Planner) AddWorker(w *Worker) {
	p.worker = append(p.worker, w)
}

func (p *Planner) AddTarget(loc Location, score int) {
	p.target = append(p.target, Target{Loc: loc, Score: score})
}

func (p *Planner) scoreAssignment(w *Worker, loc Location, score int) int {
	estimate := p.estimator.Estimate(w, loc)
	if estimate == -1 {
		return -1
	}
	return score / (1 + estimate)
}

func (p *Planner) AddAssignedWorker(w *Worker, loc Location, score int) {
	wInd := len(p.worker)
	p.worker = append(p.worker, w)

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

func (p *Planner) MakePlan() (res []Assignment) {
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
