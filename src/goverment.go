package main

type TurnReport struct {
	MyLiveAnts []*MyAnt
	Food       []Location
	Enemy      []Location
}

type Goverment struct {
	TurnRep []TurnReport
	cn      *Country
	m       *Map
}

func NewGoverment(cn *Country, m *Map) *Goverment {
	return &Goverment{cn: cn, m: m}
}

func (g *Goverment) Update() {
	g.TurnRep = make([]TurnReport, g.cn.ProvCount())

	for _, ant := range g.m.MyLiveAnts {
		provInd := g.cn.Prov(ant.Loc(g.m.Turn())).Ind
		g.TurnRep[provInd].MyLiveAnts = append(g.TurnRep[provInd].MyLiveAnts, ant)
	}
	for _, loc := range g.m.Food() {
		prov := g.cn.Prov(loc)
		if prov == nil {
			// This food is not reachable
			continue
		}
		g.TurnRep[prov.Ind].Food = append(g.TurnRep[prov.Ind].Food, loc)
	}
	for _, loc := range g.m.Enemy() {
		prov := g.cn.Prov(loc)
		if prov == nil {
			continue
		}

		g.TurnRep[prov.Ind].Enemy = append(g.TurnRep[prov.Ind].Enemy, loc)
	}
}
