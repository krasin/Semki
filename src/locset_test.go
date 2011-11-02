package main

import (
	"testing"
)

const (
	SetAdd   = 0
	SetCheck = 1
	SetClear = 2
)

type locSetTestAction struct {
	Action int
	Loc    Location
	Res    bool
}

type locSetTest struct {
	Size    int
	Actions []locSetTestAction
}

var locSetTests = []locSetTest{
	{
		Size: 1,
		Actions: []locSetTestAction{
			{SetCheck, 0, false},
			{SetAdd, 0, false},
			{SetCheck, 0, true},
		},
	},
}

func TestSimple(t *testing.T) {
	for testInd, test := range locSetTests {
		s := NewLocSet(test.Size)
		for actionInd, action := range test.Actions {
			switch action.Action {
			case SetAdd:
				s.Add(action.Loc)
			case SetCheck:
				res := s.Has(action.Loc)
				if res != action.Res {
					t.Errorf("[Test %d, action %d]: check failed. Want: %v, got: %v",
						testInd, actionInd, action.Res, res)
				}
			case SetClear:
				s.Clear()
			}
		}
	}
}
