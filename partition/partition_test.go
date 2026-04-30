package partition

import (
	"fmt"
	"testing"

	"cs730_project/grid" // adjust to match the module path in go.mod
)

func TestVoronoi_TwoAgentsEmpty8x8(t *testing.T) {
	g := grid.New(8, 8, 0, 0)
	starts := []grid.Position{
		{Row: 0, Col: 0},
		{Row: 7, Col: 7},
	}
	parts := Voronoi(g, starts)
	validatePartition(t, g, starts, parts)
}

func TestVoronoi_FourAgents(t *testing.T) {
	g := grid.New(10, 10, 0, 0)
	starts := []grid.Position{
		{Row: 0, Col: 0},
		{Row: 0, Col: 9},
		{Row: 9, Col: 0},
		{Row: 9, Col: 9},
	}
	parts := Voronoi(g, starts)
	validatePartition(t, g, starts, parts)
}

func TestVoronoi_SingleAgentClaimsEverything(t *testing.T) {
	g := grid.New(6, 6, 0, 0)
	starts := []grid.Position{{Row: 2, Col: 2}}
	parts := Voronoi(g, starts)
	validatePartition(t, g, starts, parts)
	if len(parts[0]) != len(g.FreeMegaCells()) {
		t.Fatalf("single agent claimed %d mega-cells, want %d",
			len(parts[0]), len(g.FreeMegaCells()))
	}
}

func TestVoronoi_RandomGridSeeded(t *testing.T) {
	for _, seed := range []int64{1, 7, 42, 100, 2024} {
		t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
			g := grid.New(10, 10, 0.1, seed)
			starts := []grid.Position{
				{Row: 0, Col: 0},
				{Row: 9, Col: 9},
			}
			if !g.Free(starts[0].Row, starts[0].Col) ||
				!g.Free(starts[1].Row, starts[1].Col) {
				t.Skipf("seed %d: a start cell is blocked", seed)
			}
			parts := Voronoi(g, starts)
			validatePartition(t, g, starts, parts)
		})
	}
}

// --- helpers ---

// validatePartition asserts:
//  1. every agent ID has an entry (possibly empty)
//  2. partitions are pairwise disjoint
//  3. union covers every free mega-cell reachable from any start
//  4. each agent's start mega-cell is in its own partition
func validatePartition(t *testing.T, g *grid.Grid, starts []grid.Position,
	parts map[int][]grid.Position) {
	t.Helper()

	if len(parts) != len(starts) {
		t.Fatalf("got %d partitions, want %d", len(parts), len(starts))
	}

	seen := make(map[grid.Position]int)
	for id, cells := range parts {
		for _, c := range cells {
			if !g.MegaFree(c.Row, c.Col) {
				t.Fatalf("agent %d owns blocked or out-of-bounds mega-cell %v", id, c)
			}
			if other, dup := seen[c]; dup {
				t.Fatalf("mega-cell %v claimed by both agent %d and agent %d",
					c, other, id)
			}
			seen[c] = id
		}
	}

	for id, s := range starts {
		m := grid.MegaCellOf(s.Row, s.Col)
		if seen[m] != id {
			t.Fatalf("agent %d's start mega-cell %v owned by agent %d",
				id, m, seen[m])
		}
	}

	// Every free mega-cell reachable from at least one start should be
	// claimed. On a connected grid, that's all of them.
	for _, m := range g.FreeMegaCells() {
		if _, claimed := seen[m]; !claimed {
			// Allow this only if m is unreachable from every start.
			reachable := false
			for _, s := range starts {
				if megaReachable(g, grid.MegaCellOf(s.Row, s.Col), m) {
					reachable = true
					break
				}
			}
			if reachable {
				t.Fatalf("free mega-cell %v reachable but unclaimed", m)
			}
		}
	}
}

func megaReachable(g *grid.Grid, from, to grid.Position) bool {
	if from == to {
		return true
	}
	visited := map[grid.Position]bool{from: true}
	queue := []grid.Position{from}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, nb := range g.MegaNeighbors(curr) {
			if nb == to {
				return true
			}
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}
	return false
}