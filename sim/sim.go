













package sim

import (
	"fmt"

	"cs730_project/grid"
)


type Result struct {
	Makespan        int
	TimeToDiscovery int
	Utilization     map[int]float64
}




func Run(g *grid.Grid, paths map[int][]grid.Position, target grid.Position) Result {
	if !g.Free(target.Row, target.Col) {
		panic(fmt.Sprintf("sim: target (%d,%d) is blocked or out of bounds",
			target.Row, target.Col))
	}
	if len(paths) == 0 {
		panic("sim: need at least one agent path")
	}



	makespan := 0
	for id, p := range paths {
		if len(p) == 0 {
			panic(fmt.Sprintf("sim: agent %d has empty path", id))
		}
		if last := len(p) - 1; last > makespan {
			makespan = last
		}
	}



	discovery := -1
	for t := 0; t <= makespan && discovery == -1; t++ {
		for _, p := range paths {
			pos := p[len(p)-1]
			if t < len(p) {
				pos = p[t]
			}
			if pos == target {
				discovery = t
				break
			}
		}
	}




	util := make(map[int]float64, len(paths))
	for id, p := range paths {
		if makespan == 0 {
			util[id] = 1.0
			continue
		}

		moving := len(p) - 1
		util[id] = float64(moving) / float64(makespan)
	}

	return Result{
		Makespan:        makespan,
		TimeToDiscovery: discovery,
		Utilization:     util,
	}
}