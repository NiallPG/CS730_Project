// Package partition assigns mega-cells to agents.
//
// Voronoi performs a multi-source BFS on the mega-cell graph: every agent's
// starting mega-cell seeds its own wavefront, and all wavefronts expand one
// step per round. A mega-cell is claimed by whichever agent's wavefront
// reaches it first; ties (equal geodesic distance) are broken by lower
// agent ID for determinism.
//
// This produces geodesic-Voronoi regions — distance is measured by BFS hops
// through the free mega-cell graph, so partitions correctly route around
// obstacles. The partition is not balanced; that's DARP's job.
package partition

import (
	"fmt"

	"cs730_project/grid" // adjust to match the module path in go.mod
)

// Voronoi partitions the free mega-cells of g among agents whose fine-cell
// start positions are given by starts (indexed by agent ID). Returns a map
// from agent ID to the list of mega-cells that agent owns. Every free
// mega-cell reachable from at least one agent's start appears in exactly
// one agent's list.
//
// Panics if any start is blocked, two agents start in the same mega-cell,
// or no agents are provided.
func Voronoi(g *grid.Grid, starts []grid.Position) map[int][]grid.Position {
	if len(starts) == 0 {
		panic("Voronoi: need at least one agent")
	}

	// Seed: each agent's starting mega-cell is claimed at distance 0.
	owner := make(map[grid.Position]int)
	queue := make([]grid.Position, 0, len(starts))
	for id, s := range starts {
		if !g.Free(s.Row, s.Col) {
			panic(fmt.Sprintf("Voronoi: agent %d start (%d,%d) is blocked",
				id, s.Row, s.Col))
		}
		m := grid.MegaCellOf(s.Row, s.Col)
		if existing, taken := owner[m]; taken {
			panic(fmt.Sprintf("Voronoi: agents %d and %d share mega-cell %v",
				existing, id, m))
		}
		owner[m] = id
		queue = append(queue, m)
	}

	// Multi-source BFS. All wavefronts advance in lockstep because we
	// process the queue in FIFO order — every cell at distance d is
	// fully expanded before any cell at distance d+1 is touched. Ties
	// resolve to the agent whose wavefront enqueued the cell first,
	// which (given the seed order above) is the lower agent ID.
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, nb := range g.MegaNeighbors(curr) {
			if _, claimed := owner[nb]; claimed {
				continue
			}
			owner[nb] = owner[curr]
			queue = append(queue, nb)
		}
	}

	// Invert owner map into per-agent cell lists.
	out := make(map[int][]grid.Position, len(starts))
	for id := range starts {
		out[id] = nil
	}
	for cell, id := range owner {
		out[id] = append(out[id], cell)
	}
	return out
}