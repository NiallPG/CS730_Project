package stc

import (
	"fmt"
	"testing"

	"cs730_project/grid"
	"cs730_project/partition"
)

// Single mega-cell region: should produce a 4-cell circuit.
func TestRegionSTC_SingleMegaCell(t *testing.T) {
	g := grid.New(4, 4, 0, 0)
	region := []grid.Position{{Row: 0, Col: 0}}
	start := grid.Position{Row: 0, Col: 0}
	path := RegionSTC(g, region, start)
	validateRegionCircuit(t, g, region, start, path)
}

// Equivalence: STC on the full grid == RegionSTC with all mega-cells.
func TestRegionSTC_EqualsFullSTC(t *testing.T) {
	g := grid.New(8, 8, 0, 0)
	start := grid.Position{Row: 0, Col: 0}
	full := STC(g, start)
	region := g.FreeMegaCells()
	restricted := RegionSTC(g, region, start)
	if len(full) != len(restricted) {
		t.Fatalf("length mismatch: full=%d, restricted=%d",
			len(full), len(restricted))
	}
	for i := range full {
		if full[i] != restricted[i] {
			t.Fatalf("path[%d]: full=%v, restricted=%v",
				i, full[i], restricted[i])
		}
	}
}

// Realistic case: Voronoi splits an 8x8 between two corner agents,
// each agent runs RegionSTC on its partition.
func TestRegionSTC_VoronoiPartitionedCoverage(t *testing.T) {
	g := grid.New(8, 8, 0, 0)
	starts := []grid.Position{
		{Row: 0, Col: 0},
		{Row: 7, Col: 7},
	}
	parts := partition.Voronoi(g, starts)

	totalCovered := 0
	for id, region := range parts {
		path := RegionSTC(g, region, starts[id])
		validateRegionCircuit(t, g, region, starts[id], path)
		totalCovered += len(path)
	}
	if totalCovered != len(g.FreeCells()) {
		t.Fatalf("coverage union = %d cells, want %d",
			totalCovered, len(g.FreeCells()))
	}
}

// Same as above but on random obstacle-laden grids.
func TestRegionSTC_VoronoiOnRandomGrids(t *testing.T) {
	for _, seed := range []int64{1, 7, 42, 100, 2024} {
		t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
			g := grid.New(10, 10, 0.1, seed)
			starts := []grid.Position{
				{Row: 0, Col: 0},
				{Row: 9, Col: 9},
			}
			if !g.Free(starts[0].Row, starts[0].Col) ||
				!g.Free(starts[1].Row, starts[1].Col) {
				t.Skipf("seed %d: a start is blocked", seed)
			}
			parts := partition.Voronoi(g, starts)

			totalCovered := 0
			for id, region := range parts {
				path := RegionSTC(g, region, starts[id])
				validateRegionCircuit(t, g, region, starts[id], path)
				totalCovered += len(path)
			}
			if totalCovered != len(g.FreeCells()) {
				t.Fatalf("coverage union = %d cells, want %d",
					totalCovered, len(g.FreeCells()))
			}
		})
	}
}

// validateRegionCircuit asserts: length = 4*len(region); all cells are
// inside region; unique; 4-adjacent; closes back to start.
func validateRegionCircuit(t *testing.T, g *grid.Grid, region []grid.Position,
	start grid.Position, path []grid.Position) {
	t.Helper()

	want := len(region) * 4
	if len(path) != want {
		t.Fatalf("path length = %d, want %d (4 * region size)", len(path), want)
	}
	if path[0] != start {
		t.Fatalf("path[0] = %v, want start = %v", path[0], start)
	}
	inRegion := make(map[grid.Position]bool, len(region))
	for _, m := range region {
		inRegion[m] = true
	}
	seen := make(map[grid.Position]bool, len(path))
	for i, p := range path {
		if !g.Free(p.Row, p.Col) {
			t.Fatalf("path[%d] = %v is blocked or out of bounds", i, p)
		}
		if !inRegion[grid.MegaCellOf(p.Row, p.Col)] {
			t.Fatalf("path[%d] = %v is outside region", i, p)
		}
		if seen[p] {
			t.Fatalf("path[%d] = %v duplicate", i, p)
		}
		seen[p] = true
		if i > 0 {
			dr := p.Row - path[i-1].Row
			dc := p.Col - path[i-1].Col
			if dr < 0 {
				dr = -dr
			}
			if dc < 0 {
				dc = -dc
			}
			if dr+dc != 1 {
				t.Fatalf("path[%d]=%v not 4-adjacent to path[%d]=%v",
					i, p, i-1, path[i-1])
			}
		}
	}
	last := path[len(path)-1]
	dr := last.Row - start.Row
	dc := last.Col - start.Col
	if dr < 0 {
		dr = -dr
	}
	if dc < 0 {
		dc = -dc
	}
	if dr+dc != 1 {
		t.Fatalf("circuit does not close: last %v not 4-adjacent to start %v",
			last, start)
	}
}