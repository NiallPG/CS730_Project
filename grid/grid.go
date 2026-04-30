// Package grid provides the shared grid representation and spatial utilities
// used by all other packages. Obstacles are generated at the mega-cell (2×2)
// level so that STC's circumnavigation invariant is always satisfied.
package grid

import ("math/rand"
	"strings"
)

// Position is a (row, col) coordinate. Used for both fine cells and mega-cells;
// the distinction is contextual (documented at each call site).
type Position struct {
	Row, Col int
}

// Grid is a 2D occupancy grid. Fine cells are the unit the robot walks on.
// Mega-cells are 2×2 blocks of fine cells used by STC.
type Grid struct {
	Rows, Cols         int // fine-grid dimensions (always even)
	MegaRows, MegaCols int // = Rows/2, Cols/2
	blocked            [][]bool
}

// T f f
// f f f
// f f T

// T T f f f f
// T T f f f f
// f f f f f f
// f f f f T T
// f f f f T T


// New creates a grid of the given size with randomly placed mega-cell-aligned
// obstacles. rows and cols are rounded up to even numbers. density ∈ [0,1) is
// the fraction of mega-cells that are blocked. seed controls the RNG.
func New(rows, cols int, density float64, seed int64) *Grid {
	if rows%2 != 0 { // we call new but increment rows/cols if the input is not even
		rows++
	}
	if cols%2 != 0 {
		cols++
	}
	megaRows, megaCols := rows/2, cols/2 // here we set mega rows/cols, just rows/cols div by 2
	rng := rand.New(rand.NewSource(seed)) // we get a new seed

	blocked := make([][]bool, rows) // we make blocked a grid of booleans, at first all is false
	for r := range blocked {
		blocked[r] = make([]bool, cols)
	}

	for mr := 0; mr < megaRows; mr++ {
		for mc := 0; mc < megaCols; mc++ {
			if rng.Float64() < density { // we supply the density as an arg
				// but this ^ flips a coin essentially. if density = .1, then about 10% of the mega cells become obstacles
				for dr := 0; dr < 2; dr++ { // these hit the fine cells in the mega cell, which is a 2x2
					for dc := 0; dc < 2; dc++ {
						blocked[mr*2+dr][mc*2+dc] = true // index max to set that 
					}
				}
			}
		}
	}

	return &Grid{
		Rows: rows, Cols: cols,
		MegaRows: megaRows, MegaCols: megaCols,
		blocked: blocked,
	}
}

// --- Fine-cell queries ---

// reports whether fine cell (r, c) is in-bounds and unblocked
func (g *Grid) Free(r, c int) bool {
	return r >= 0 && r < g.Rows && c >= 0 && c < g.Cols && !g.blocked[r][c]
}

// returns every free fine cell
func (g *Grid) FreeCells() []Position {
	out := make([]Position, 0, g.Rows*g.Cols)
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if !g.blocked[r][c] {
				out = append(out, Position{r, c})
			}
		}
	}
	return out
}

// returns the 4-connected free neighbors of a fine cell
func (g *Grid) FineNeighbors(p Position) []Position {
	dirs := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	out := make([]Position, 0, 4)
	for _, d := range dirs {
		if g.Free(p.Row+d[0], p.Col+d[1]) {
			out = append(out, Position{p.Row + d[0], p.Col + d[1]})
		}
	}
	return out
}

// --- Mega-cell queries ---

// reports whether mega-cell (mr, mc) is in-bounds and unblocked
func (g *Grid) MegaFree(mr, mc int) bool {
	return mr >= 0 && mr < g.MegaRows && mc >= 0 && mc < g.MegaCols && !g.blocked[mr*2][mc*2]
}

// returns every free mega-cell position
func (g *Grid) FreeMegaCells() []Position {
	out := make([]Position, 0, g.MegaRows*g.MegaCols)
	for mr := 0; mr < g.MegaRows; mr++ {
		for mc := 0; mc < g.MegaCols; mc++ {
			if g.MegaFree(mr, mc) {
				out = append(out, Position{mr, mc})
			}
		}
	}
	return out
}

// returns 4-connected free mega-cell neighbors
func (g *Grid) MegaNeighbors(p Position) []Position {
	dirs := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	out := make([]Position, 0, 4)
	for _, d := range dirs {
		if g.MegaFree(p.Row+d[0], p.Col+d[1]) {
			out = append(out, Position{p.Row + d[0], p.Col + d[1]})
		}
	}
	return out
}

// returns the mega-cell containing fine cell (r, c)
func MegaCellOf(r, c int) Position {
	return Position{r / 2, c / 2}
}

// returns the 4 fine cells of mega-cell (mr, mc)
func FineCellsOf(mr, mc int) [4]Position {
	return [4]Position{
		{mr * 2, mc * 2},
		{mr * 2, mc*2 + 1},
		{mr*2 + 1, mc * 2},
		{mr*2 + 1, mc*2 + 1},
	}
}

func (g *Grid) String() string {
	var b strings.Builder
	b.Grow((g.Cols + 1) * g.Rows)
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if g.blocked[r][c] {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}