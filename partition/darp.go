package partition

import (
	"math"

	"cs730_project/grid"
)

// DARPConfig holds tunable parameters for the DARP algorithm.
type DARPConfig struct {
	// MaxIterations caps the iterative refinement loop.
	MaxIterations int
	// LearningRate is the step size c in m_i ← m_i + c·(k_i - fair).
	// Smaller values are more stable on symmetric configurations but
	// converge more slowly on imbalanced ones.
	LearningRate float64
	// BalanceTolerance is the largest |k_i - fair| (in mega-cells) that
	// counts as "balanced enough" for early termination. The paper's
	// theoretical guarantee is ≤ 1 mega-cell difference between any two
	// agents, so 1 is a reasonable default.
	BalanceTolerance int
}

// DefaultDARPConfig returns reasonable defaults for grids up to ~100×100.
func DefaultDARPConfig() DARPConfig {
	return DARPConfig{
		MaxIterations:    500,
		LearningRate:     0.01,
		BalanceTolerance: 1,
	}
}

// DARP partitions the free mega-cells of g among the agents using the
// Divide Areas based on Robots' initial Positions algorithm of Kapoutsis,
// Chatzichristofis & Kosmatopoulos (2017).
//
// Each starts[i] is a fine-cell position; the corresponding mega-cell seeds
// agent i. Output is a map from agent ID (0-indexed) to the list of mega-cell
// positions assigned to that agent.
//
// Algorithm sketch: each agent i carries a scalar correction factor m_i and
// (when its region is disconnected) a per-cell connectivity matrix C_i. The
// effective evaluation matrix is E_i(c) = m_i · baseDist_i(c) · C_i(c), where
// baseDist_i is the geodesic (BFS) distance from agent i's start. Each cell
// is assigned by argmin_i E_i(c). Each iteration: (1) compute the assignment,
// (2) adjust m_i additively toward the fair share, (3) recompute C_i to
// reward cells near the connected component containing agent i's start and
// penalize cells in disconnected fragments. Termination is when all agents
// are within BalanceTolerance of fair AND every agent's region is connected,
// or MaxIterations is reached. On non-convergence, the lowest-cost
// assignment seen is returned (with disconnected assignments penalized so a
// connected fallback is preferred).
func DARP(g *grid.Grid, starts []grid.Position, config DARPConfig) map[int][]grid.Position {
	nr := len(starts)
	if nr == 0 {
		return map[int][]grid.Position{}
	}

	// Convert fine starts to mega-cell starts.
	megaStarts := make([]grid.Position, nr)
	for i, s := range starts {
		megaStarts[i] = grid.Position{Row: s.Row / 2, Col: s.Col / 2}
	}

	// Single-agent shortcut: claim everything reachable.
	if nr == 1 {
		return map[int][]grid.Position{0: darpReachable(g, megaStarts[0])}
	}

	// Geodesic base distances from each start to every reachable mega-cell.
	baseDist := make([]map[grid.Position]float64, nr)
	for i := 0; i < nr; i++ {
		baseDist[i] = darpBFSDist(g, megaStarts[i])
	}

	// Universe = mega-cells reachable from at least one start.
	cellSet := make(map[grid.Position]bool)
	for i := 0; i < nr; i++ {
		for c := range baseDist[i] {
			cellSet[c] = true
		}
	}
	cells := make([]grid.Position, 0, len(cellSet))
	for c := range cellSet {
		cells = append(cells, c)
	}

	fair := float64(len(cells)) / float64(nr)
	tol := float64(config.BalanceTolerance)

	// Per-agent state: scalar multiplier and (optional) per-cell correction.
	// correction[i] == nil means "all ones" (region is connected, no penalty).
	m := make([]float64, nr)
	correction := make([]map[grid.Position]float64, nr)
	for i := 0; i < nr; i++ {
		m[i] = 1.0
	}

	var bestAssign map[grid.Position]int
	bestCost := math.Inf(1)

	for iter := 0; iter < config.MaxIterations; iter++ {
		// (1) Build assignment: each cell goes to argmin_i (m_i · baseDist_i · C_i).
		assign := make(map[grid.Position]int, len(cells))
		for _, c := range cells {
			best := -1
			bestVal := math.Inf(1)
			for i := 0; i < nr; i++ {
				d, ok := baseDist[i][c]
				if !ok {
					continue
				}
				cval := 1.0
				if correction[i] != nil {
					if v, ok := correction[i][c]; ok {
						cval = v
					}
				}
				e := m[i] * (d + 1) * cval // +1 so distance-0 cells aren't trivially zero
				if e < bestVal {
					bestVal = e
					best = i
				}
			}
			assign[c] = best
		}
		// Force start ownership (Definition 3, condition 5).
		for i := 0; i < nr; i++ {
			assign[megaStarts[i]] = i
		}

		// (2) Counts and imbalance cost.
		counts := make([]int, nr)
		for _, owner := range assign {
			counts[owner]++
		}
		cost := 0.0
		for i := 0; i < nr; i++ {
			d := float64(counts[i]) - fair
			cost += d * d
		}
		cost *= 0.5

		// (3) Connected components per agent.
		agentCells := make([][]grid.Position, nr)
		for c, owner := range assign {
			agentCells[owner] = append(agentCells[owner], c)
		}
		allConnected := true
		comps := make([][][]grid.Position, nr)
		for i := 0; i < nr; i++ {
			comps[i] = darpComponents(g, agentCells[i])
			if len(comps[i]) > 1 {
				allConnected = false
			}
		}

		// Track best feasible-ish assignment (penalize disconnection heavily).
		adjCost := cost
		if !allConnected {
			adjCost += 1e9
		}
		if adjCost < bestCost {
			bestCost = adjCost
			bestAssign = make(map[grid.Position]int, len(assign))
			for k, v := range assign {
				bestAssign[k] = v
			}
		}

		// (4) Termination check.
		balanced := true
		for i := 0; i < nr; i++ {
			if math.Abs(float64(counts[i])-fair) > tol {
				balanced = false
				break
			}
		}
		if balanced && allConnected {
			return darpGroup(assign, nr)
		}

		// (5) Update m_i additively toward fair share, with bounds.
		for i := 0; i < nr; i++ {
			m[i] += config.LearningRate * (float64(counts[i]) - fair)
			if m[i] < 0.01 {
				m[i] = 0.01
			} else if m[i] > 100 {
				m[i] = 100
			}
		}

		// (6) Update connectivity matrices C_i for disconnected agents.
		for i := 0; i < nr; i++ {
			if len(comps[i]) <= 1 {
				correction[i] = nil
				continue
			}
			// R_i = component containing megaStarts[i]; Q_i = the rest.
			var rCells, qCells []grid.Position
			startMega := megaStarts[i]
			for _, comp := range comps[i] {
				containsStart := false
				for _, c := range comp {
					if c == startMega {
						containsStart = true
						break
					}
				}
				if containsStart {
					rCells = append(rCells, comp...)
				} else {
					qCells = append(qCells, comp...)
				}
			}
			// C_i(c) = min_dist(c, R_i) - min_dist(c, Q_i), Euclidean per the
			// paper. Then shift to keep all values positive (preserves rank,
			// keeps multiplication well-behaved against positive baseDist).
			raw := make(map[grid.Position]float64, len(cells))
			minVal := math.Inf(1)
			for _, c := range cells {
				dr := darpMinEuclid(c, rCells)
				dq := darpMinEuclid(c, qCells)
				v := dr - dq
				raw[c] = v
				if v < minVal {
					minVal = v
				}
			}
			shift := 1.0 - minVal // ensures min becomes 1.0
			corr := make(map[grid.Position]float64, len(cells))
			for c, v := range raw {
				corr[c] = v + shift
			}
			correction[i] = corr
		}
	}

	return darpGroup(bestAssign, nr)
}

// darpReachable returns all mega-cells reachable from start via 4-connected
// mega-cell moves through free mega-cells.
func darpReachable(g *grid.Grid, start grid.Position) []grid.Position {
	visited := map[grid.Position]bool{start: true}
	queue := []grid.Position{start}
	out := []grid.Position{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, n := range g.MegaNeighbors(cur) {
			if !visited[n] {
				visited[n] = true
				queue = append(queue, n)
				out = append(out, n)
			}
		}
	}
	return out
}

// darpBFSDist computes geodesic distance (in mega-cell steps) from start to
// every reachable free mega-cell.
func darpBFSDist(g *grid.Grid, start grid.Position) map[grid.Position]float64 {
	dist := map[grid.Position]float64{start: 0}
	queue := []grid.Position{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		d := dist[cur]
		for _, n := range g.MegaNeighbors(cur) {
			if _, seen := dist[n]; !seen {
				dist[n] = d + 1
				queue = append(queue, n)
			}
		}
	}
	return dist
}

// darpComponents returns the connected components of a set of mega-cells
// under 4-connected mega-cell adjacency.
func darpComponents(g *grid.Grid, cells []grid.Position) [][]grid.Position {
	if len(cells) == 0 {
		return nil
	}
	inSet := make(map[grid.Position]bool, len(cells))
	for _, c := range cells {
		inSet[c] = true
	}
	visited := make(map[grid.Position]bool, len(cells))
	var comps [][]grid.Position
	for _, c := range cells {
		if visited[c] {
			continue
		}
		var comp []grid.Position
		queue := []grid.Position{c}
		visited[c] = true
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			comp = append(comp, cur)
			for _, n := range g.MegaNeighbors(cur) {
				if inSet[n] && !visited[n] {
					visited[n] = true
					queue = append(queue, n)
				}
			}
		}
		comps = append(comps, comp)
	}
	return comps
}

// darpMinEuclid returns the minimum Euclidean distance from p to any cell in
// set, or +Inf if set is empty.
func darpMinEuclid(p grid.Position, set []grid.Position) float64 {
	best := math.Inf(1)
	for _, q := range set {
		dr := float64(p.Row - q.Row)
		dc := float64(p.Col - q.Col)
		d := math.Sqrt(dr*dr + dc*dc)
		if d < best {
			best = d
		}
	}
	return best
}

// darpGroup inverts a cell→agent map into agent→cells, ensuring every agent
// has an entry (possibly nil).
func darpGroup(assign map[grid.Position]int, nr int) map[int][]grid.Position {
	out := make(map[int][]grid.Position, nr)
	for i := 0; i < nr; i++ {
		out[i] = nil
	}
	for c, owner := range assign {
		out[owner] = append(out[owner], c)
	}
	return out
}