


package grid

import ("math/rand"
	"strings"
)



type Position struct {
	Row, Col int
}



type Grid struct {
	Rows, Cols         int
	MegaRows, MegaCols int
	blocked            [][]bool
}















func New(rows, cols int, density float64, seed int64) *Grid {
	if rows%2 != 0 {
		rows++
	}
	if cols%2 != 0 {
		cols++
	}
	megaRows, megaCols := rows/2, cols/2
	rng := rand.New(rand.NewSource(seed))

	blocked := make([][]bool, rows)
	for r := range blocked {
		blocked[r] = make([]bool, cols)
	}

	for mr := 0; mr < megaRows; mr++ {
		for mc := 0; mc < megaCols; mc++ {
			if rng.Float64() < density {

				for dr := 0; dr < 2; dr++ {
					for dc := 0; dc < 2; dc++ {
						blocked[mr*2+dr][mc*2+dc] = true
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




func (g *Grid) Free(r, c int) bool {
	return r >= 0 && r < g.Rows && c >= 0 && c < g.Cols && !g.blocked[r][c]
}


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




func (g *Grid) MegaFree(mr, mc int) bool {
	return mr >= 0 && mr < g.MegaRows && mc >= 0 && mc < g.MegaCols && !g.blocked[mr*2][mc*2]
}


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


func MegaCellOf(r, c int) Position {
	return Position{r / 2, c / 2}
}


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