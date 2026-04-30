package stc

import (
	"fmt"
	"testing"

	"cs730_project/grid"
)

// --- core test cases ---

func TestSTC_Empty4x4(t *testing.T) {
	g := grid.New(4, 4, 0, 0)
	start := grid.Position{Row: 0, Col: 0}
	validateCircuit(t, g, start, STC(g, start))
}

func TestSTC_Empty6x6(t *testing.T) {
	g := grid.New(6, 6, 0, 0)
	start := grid.Position{Row: 0, Col: 0}
	validateCircuit(t, g, start, STC(g, start))
}

func TestSTC_Empty8x8(t *testing.T) {
	g := grid.New(8, 8, 0, 0)
	start := grid.Position{Row: 0, Col: 0}
	validateCircuit(t, g, start, STC(g, start))
}

// Tall thin grid: a single row of mega-cells. Stress-tests circumnavigation
// when the spanning tree degenerates to a chain.
func TestSTC_Tall2x10(t *testing.T) {
	g := grid.New(2, 10, 0, 0)
	start := grid.Position{Row: 0, Col: 0}
	validateCircuit(t, g, start, STC(g, start))
}

// Same grid, multiple start positions. Each must produce a valid circuit.
func TestSTC_VariousStarts(t *testing.T) {
	g := grid.New(8, 8, 0, 0)
	starts := []grid.Position{
		{Row: 0, Col: 0},
		{Row: 3, Col: 5},
		{Row: 7, Col: 7},
		{Row: 4, Col: 0},
	}
	for _, s := range starts {
		t.Run(fmt.Sprintf("start_%d_%d", s.Row, s.Col), func(t *testing.T) {
			validateCircuit(t, g, s, STC(g, s))
		})
	}
}

// Grids with random mega-cell obstacles. Skips seeds that produce a blocked
// start or a disconnected free region (STC's preconditions don't hold there).
func TestSTC_RandomGridSeeded(t *testing.T) {
	for _, seed := range []int64{1, 7, 42, 100, 2024} {
		t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
			g := grid.New(10, 10, 0.1, seed)
			start := grid.Position{Row: 0, Col: 0}
			if !g.Free(start.Row, start.Col) {
				t.Skipf("seed %d: start (0,0) is blocked", seed)
			}
			if !megaConnected(g, grid.MegaCellOf(0, 0)) {
				t.Skipf("seed %d: free mega-region is disconnected", seed)
			}
			validateCircuit(t, g, start, STC(g, start))
		})
	}
}

// --- helpers ---

// validateCircuit asserts the four acceptance criteria for a Hamiltonian
// coverage circuit:
//  1. length == free-cell count
//  2. starts at the given start cell
//  3. each entry is unique, free, and in-bounds
//  4. consecutive entries (and the last → start closing edge) are 4-adjacent
func validateCircuit(t *testing.T, g *grid.Grid, start grid.Position, path []grid.Position) {
	t.Helper()
	free := g.FreeCells()
	if len(path) != len(free) {
		t.Fatalf("path length = %d, want %d (free cell count)", len(path), len(free))
	}
	if path[0] != start {
		t.Fatalf("path[0] = %v, want start = %v", path[0], start)
	}
	seen := make(map[grid.Position]bool, len(path))
	for i, p := range path {
		if !g.Free(p.Row, p.Col) {
			t.Fatalf("path[%d] = %v is blocked or out of bounds", i, p)
		}
		if seen[p] {
			t.Fatalf("path[%d] = %v is a duplicate", i, p)
		}
		seen[p] = true
		if i > 0 && !adjacent(path[i-1], p) {
			t.Fatalf("path[%d]=%v not 4-adjacent to path[%d]=%v",
				i, p, i-1, path[i-1])
		}
	}
	if !adjacent(path[len(path)-1], start) {
		t.Fatalf("circuit does not close: last %v not 4-adjacent to start %v",
			path[len(path)-1], start)
	}
}

func adjacent(a, b grid.Position) bool {
	dr, dc := a.Row-b.Row, a.Col-b.Col
	if dr < 0 {
		dr = -dr
	}
	if dc < 0 {
		dc = -dc
	}
	return dr+dc == 1
}

// megaConnected reports whether every free mega-cell is reachable from root
// via 4-connectivity. Used to gate tests against random grids that happen
// to produce disconnected free regions.
func megaConnected(g *grid.Grid, root grid.Position) bool {
	visited := map[grid.Position]bool{root: true}
	queue := []grid.Position{root}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, nb := range g.MegaNeighbors(curr) {
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}
	return len(visited) == len(g.FreeMegaCells())
}
