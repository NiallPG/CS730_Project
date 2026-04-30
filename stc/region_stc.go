package stc

import (
	"fmt"

	"cs730_project/grid"
)

// RegionSTC returns a coverage path that visits every fine cell of the given
// region exactly once, starting at start. region is a list of mega-cells
// (typically one agent's Voronoi partition); any cell outside region is
// treated as blocked. The returned path has length 4 * len(region) and the
// last entry is 4-adjacent to start.
//
// Preconditions:
//   - start is a free fine cell.
//   - start's mega-cell is in region.
//   - region is 4-connected through mega-cell adjacency.
//
// Voronoi partitions always satisfy the third condition by construction
// (multi-source BFS produces connected regions). Other partitioners may not.
func RegionSTC(g *grid.Grid, region []grid.Position, start grid.Position) []grid.Position {
	if !g.Free(start.Row, start.Col) {
		panic(fmt.Sprintf("RegionSTC: start (%d,%d) is blocked or out of bounds",
			start.Row, start.Col))
	}
	if len(region) == 0 {
		panic("RegionSTC: region is empty")
	}
	inRegion := make(map[grid.Position]bool, len(region))
	for _, m := range region {
		inRegion[m] = true
	}
	startMega := grid.MegaCellOf(start.Row, start.Col)
	if !inRegion[startMega] {
		panic(fmt.Sprintf("RegionSTC: start mega-cell %v not in region", startMega))
	}

	t, reached := buildRegionTree(g, startMega, inRegion)
	if len(reached) != len(region) {
		panic(fmt.Sprintf(
			"RegionSTC: region is disconnected (reached %d of %d mega-cells)",
			len(reached), len(region)))
	}
	return regionCircumnavigate(g, t, start, inRegion)
}

// buildRegionTree is buildSpanningTree restricted to mega-cells in region.
func buildRegionTree(g *grid.Grid, root grid.Position,
	region map[grid.Position]bool) (tree, map[grid.Position]bool) {

	t := tree{}
	visited := map[grid.Position]bool{root: true}
	stack := []grid.Position{root}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		foundNew := false
		for _, nb := range g.MegaNeighbors(curr) {
			if !region[nb] {
				continue
			}
			if !visited[nb] {
				t[curr] = append(t[curr], nb)
				t[nb] = append(t[nb], curr)
				visited[nb] = true
				stack = append(stack, nb)
				foundNew = true
				break
			}
		}
		if !foundNew {
			stack = stack[:len(stack)-1]
		}
	}
	return t, visited
}

// regionCircumnavigate is circumnavigate with mega-cells outside region
// treated as blocked. dirs and turnOffsets are reused from stc.go.
func regionCircumnavigate(g *grid.Grid, t tree, start grid.Position,
	region map[grid.Position]bool) []grid.Position {

	totalFree := len(region) * 4
	path := make([]grid.Position, 0, totalFree)
	path = append(path, start)

	r, c := start.Row, start.Col
	heading := 1

	maxSteps := totalFree * 4
	for step := 0; step < maxSteps; step++ {
		moved := false
		for _, off := range turnOffsets {
			h := (heading + off) % 4
			nr, nc := r+dirs[h][0], c+dirs[h][1]
			if !regionCanMove(g, t, r, c, nr, nc, region) {
				continue
			}
			heading = h
			r, c = nr, nc
			if r == start.Row && c == start.Col {
				return path
			}
			path = append(path, grid.Position{Row: r, Col: c})
			moved = true
			break
		}
		if !moved {
			panic(fmt.Sprintf("RegionSTC: stuck at (%d,%d) after %d steps", r, c, step))
		}
	}
	panic(fmt.Sprintf("RegionSTC: %d steps without returning to start", maxSteps))
}

func regionCanMove(g *grid.Grid, t tree, r1, c1, r2, c2 int,
	region map[grid.Position]bool) bool {

	if !g.Free(r2, c2) {
		return false
	}
	m2 := grid.MegaCellOf(r2, c2)
	if !region[m2] {
		return false
	}
	m1 := grid.MegaCellOf(r1, c1)
	if m1 == m2 {
		return true
	}
	return t.hasEdge(m1, m2)
}