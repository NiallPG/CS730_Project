










package partition

import (
	"fmt"

	"cs730_project/grid"
)









func Voronoi(g *grid.Grid, starts []grid.Position) map[int][]grid.Position {
	if len(starts) == 0 {
		panic("Voronoi: need at least one agent")
	}


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


	out := make(map[int][]grid.Position, len(starts))
	for id := range starts {
		out[id] = nil
	}
	for cell, id := range owner {
		out[id] = append(out[id], cell)
	}
	return out
}