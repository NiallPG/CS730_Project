// Package sim runs a coverage simulation: given precomputed per-agent paths
// and a target cell, it steps every agent one fine-cell per tick in parallel
// and reports search metrics.
//
// Time conventions:
//   - Tick 0 = initial state. Each agent occupies path[0]. The target may be
//     visible immediately (TimeToDiscovery = 0).
//   - Tick t > 0 = each still-active agent has just moved to path[t]. An
//     agent that has reached the end of its path stops moving but does not
//     "disappear"; the simulation continues until every agent has finished.
//   - Makespan = the tick at which the last agent finishes its path. Equal
//     to max(len(path)) - 1 across all agents.
//   - TimeToDiscovery = the earliest tick at which any agent occupies the
//     target cell. -1 if no agent ever visits the target.
package sim

import (
	"fmt"

	"cs730_project/grid" // adjust to match the module path in go.mod
)

// Result holds the metrics of a single simulation run.
type Result struct {
	Makespan        int         // ticks until the last agent finishes
	TimeToDiscovery int         // earliest tick the target is occupied; -1 if never
	Utilization     map[int]float64 // agent ID -> fraction of [0, Makespan] spent moving
}

// Run simulates the given paths against the target. Paths are indexed by
// agent ID, with path[i][t] being the agent's fine-cell location at tick t.
// All paths must be non-empty. Target must be a free fine cell of g.
func Run(g *grid.Grid, paths map[int][]grid.Position, target grid.Position) Result {
	if !g.Free(target.Row, target.Col) {
		panic(fmt.Sprintf("sim: target (%d,%d) is blocked or out of bounds",
			target.Row, target.Col))
	}
	if len(paths) == 0 {
		panic("sim: need at least one agent path")
	}

	// Makespan is determined entirely by path lengths; no need to step
	// per-tick to compute it. Discovery requires a per-tick scan.
	makespan := 0
	for id, p := range paths {
		if len(p) == 0 {
			panic(fmt.Sprintf("sim: agent %d has empty path", id))
		}
		if last := len(p) - 1; last > makespan {
			makespan = last
		}
	}

	// Find earliest tick any agent sits on the target. Scan tick-by-tick
	// across all agents so ties resolve to the smallest t.
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

	// Utilization: fraction of ticks [1, makespan] during which the agent
	// was still moving (i.e., t < len(path)). Tick 0 is the initial state
	// and isn't counted. If makespan is 0, every agent is trivially 100%.
	util := make(map[int]float64, len(paths))
	for id, p := range paths {
		if makespan == 0 {
			util[id] = 1.0
			continue
		}
		// Agent moves on ticks 1..len(p)-1, idles on ticks len(p)..makespan.
		moving := len(p) - 1
		util[id] = float64(moving) / float64(makespan)
	}

	return Result{
		Makespan:        makespan,
		TimeToDiscovery: discovery,
		Utilization:     util,
	}
}