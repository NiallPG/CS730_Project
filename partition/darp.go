package partition

import (
	"math"

	"cs730_project/grid"
)


type DARPConfig struct {

	MaxIterations int



	LearningRate float64




	BalanceTolerance int
}


func DefaultDARPConfig() DARPConfig {
	return DARPConfig{
		MaxIterations:    500,
		LearningRate:     0.01,
		BalanceTolerance: 1,
	}
}





















func DARP(g *grid.Grid, starts []grid.Position, config DARPConfig) map[int][]grid.Position {
	nr := len(starts)
	if nr == 0 {
		return map[int][]grid.Position{}
	}


	megaStarts := make([]grid.Position, nr)
	for i, s := range starts {
		megaStarts[i] = grid.Position{Row: s.Row / 2, Col: s.Col / 2}
	}


	if nr == 1 {
		return map[int][]grid.Position{0: darpReachable(g, megaStarts[0])}
	}


	baseDist := make([]map[grid.Position]float64, nr)
	for i := 0; i < nr; i++ {
		baseDist[i] = darpBFSDist(g, megaStarts[i])
	}


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



	m := make([]float64, nr)
	correction := make([]map[grid.Position]float64, nr)
	for i := 0; i < nr; i++ {
		m[i] = 1.0
	}

	var bestAssign map[grid.Position]int
	bestCost := math.Inf(1)

	for iter := 0; iter < config.MaxIterations; iter++ {

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
				e := m[i] * (d + 1) * cval
				if e < bestVal {
					bestVal = e
					best = i
				}
			}
			assign[c] = best
		}

		for i := 0; i < nr; i++ {
			assign[megaStarts[i]] = i
		}


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


		for i := 0; i < nr; i++ {
			m[i] += config.LearningRate * (float64(counts[i]) - fair)
			if m[i] < 0.01 {
				m[i] = 0.01
			} else if m[i] > 100 {
				m[i] = 100
			}
		}


		for i := 0; i < nr; i++ {
			if len(comps[i]) <= 1 {
				correction[i] = nil
				continue
			}

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
			shift := 1.0 - minVal
			corr := make(map[grid.Position]float64, len(cells))
			for c, v := range raw {
				corr[c] = v + shift
			}
			correction[i] = corr
		}
	}

	return darpGroup(bestAssign, nr)
}



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