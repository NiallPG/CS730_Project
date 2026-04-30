// Package stc implements Spanning Tree Coverage (Gabriely & Rimon, 2001).
//
// STC builds a DFS spanning tree on the mega-cell graph and circumnavigates
// it on the fine grid using the right-hand rule, producing a Hamiltonian
// circuit that visits every free fine cell exactly once and returns to start.
package stc

import (
	"fmt"

	"cs730_project/grid"
)

// STC returns a coverage path that visits every free fine cell of g exactly
// once, starting at start. The closing edge back to start is implicit: the
// last entry of the returned slice is 4-adjacent to start, but start is not
// duplicated at the end. Length of the returned slice equals g's free-cell
// count.
//
// Preconditions:
//   - start is a free fine cell.
//   - The free mega-cell graph is connected (single component containing
//     start's mega-cell).
//
// Violation of either precondition causes a panic with a descriptive message.
func STC(g *grid.Grid, start grid.Position) []grid.Position {
	if !g.Free(start.Row, start.Col) {
		panic(fmt.Sprintf("STC: start (%d,%d) is blocked or out of bounds",
			start.Row, start.Col))
	}

	startMega := grid.MegaCellOf(start.Row, start.Col)
	t, reached := buildSpanningTree(g, startMega)

	if got, want := len(reached), len(g.FreeMegaCells()); got != want {
		panic(fmt.Sprintf(
			"STC: free mega-cell graph is disconnected (reached %d of %d); "+
				"STC requires a connected free region", got, want))
	}

	return circumnavigate(g, t, start)
}

// tree is an undirected adjacency list over mega-cell positions.
type tree map[grid.Position][]grid.Position

func (t tree) hasEdge(a, b grid.Position) bool {
	for _, nb := range t[a] {
		if nb == b {
			return true
		}
	}
	return false
}

// buildSpanningTree constructs a DFS spanning tree on the mega-cell graph
// rooted at root. Returns the tree and the set of mega-cells reached (used
// for the connectivity precondition check).
func buildSpanningTree(g *grid.Grid, root grid.Position) (tree, map[grid.Position]bool) {
	t := tree{}
	visited := map[grid.Position]bool{root: true}
	stack := []grid.Position{root}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		foundNew := false
		for _, nb := range g.MegaNeighbors(curr) {
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

// Heading codes: 0=Up, 1=Right, 2=Down, 3=Left. Index into dirs.
var dirs = [4][2]int{
	{-1, 0}, // 0: Up
	{0, 1},  // 1: Right
	{1, 0},  // 2: Down
	{0, -1}, // 3: Left
}

// turnOffsets implements the right-hand rule. Applied to current heading
// in this order: right turn, straight, left turn, back. Tree edges sit on
// the robot's left as it traverses.
var turnOffsets = [4]int{1, 0, 3, 2}

// circumnavigate walks the fine grid using the right-hand rule. Within a
// mega-cell every move is allowed; between mega-cells, only tree edges are
// allowed. Stops when the robot would step back onto start.
func circumnavigate(g *grid.Grid, t tree, start grid.Position) []grid.Position {
	totalFree := len(g.FreeCells())
	path := make([]grid.Position, 0, totalFree)
	path = append(path, start)

	r, c := start.Row, start.Col
	heading := 1 // start facing right

	maxSteps := totalFree * 4 // safety net against algorithmic bugs
	for step := 0; step < maxSteps; step++ {
		moved := false
		for _, off := range turnOffsets {
			h := (heading + off) % 4
			nr, nc := r+dirs[h][0], c+dirs[h][1]
			if !canMove(g, t, r, c, nr, nc) {
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
			panic(fmt.Sprintf("STC: stuck at (%d,%d) after %d steps; "+
				"likely a spanning-tree or circumnavigation bug", r, c, step))
		}
	}
	panic(fmt.Sprintf("STC: %d steps without returning to start", maxSteps))
}

// canMove reports whether the robot can step from (r1,c1) to (r2,c2). The
// destination must be a free fine cell, and any mega-cell crossing must
// correspond to a spanning-tree edge.
func canMove(g *grid.Grid, t tree, r1, c1, r2, c2 int) bool {
	if !g.Free(r2, c2) {
		return false
	}
	m1 := grid.MegaCellOf(r1, c1)
	m2 := grid.MegaCellOf(r2, c2)
	if m1 == m2 {
		return true
	}
	return t.hasEdge(m1, m2)
}
