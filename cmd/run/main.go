// Command run is the experiment harness. It sweeps the comparison cube
// (algorithm × grid size × agent count × seed × target) and writes one CSV
// row per simulation.
//
// Output schema (one row per run):
//
//	algorithm, rows, cols, density, num_agents, seed, target_idx,
//	makespan, time_to_discovery, mean_utilization, min_utilization
//
// Conventions:
//   - Single-agent STC is run once per (size, seed, target); num_agents=1.
//   - Voronoi+STC is run for every (k, seed, target) with k from -agents.
//   - Targets are random free fine cells; the same set is reused across
//     algorithms within a (size, seed) so comparisons are paired.
//   - Agent starts are placed at the top-left fine cell of distinct random
//     free mega-cells. The same start sequence is used as a prefix for
//     every k, and the first start is also the single-agent start.
//   - Seeds whose grid is disconnected at the mega-cell level are skipped.
//
// Usage:
//
//	go run ./cmd/run/ -out results.csv -seeds 50 -targets 20 -sizes 50,100 -agents 2,3,5
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"cs730_project/grid"
	"cs730_project/partition"
	"cs730_project/sim"
	"cs730_project/stc"
)

func main() {
	var (
		outPath   = flag.String("out", "results.csv", "output CSV path")
		nSeeds    = flag.Int("seeds", 50, "grid seeds per condition")
		nTargets  = flag.Int("targets", 20, "random targets per grid")
		density   = flag.Float64("density", 0.1, "obstacle density (mega-cell fraction)")
		sizeFlag  = flag.String("sizes", "50,100", "comma-separated square grid side lengths")
		agentFlag = flag.String("agents", "2,3,5", "comma-separated multi-agent counts")
	)
	flag.Parse()

	sizes := mustParseInts(*sizeFlag)
	agentCounts := mustParseInts(*agentFlag)

	maxK := 1
	for _, k := range agentCounts {
		if k > maxK {
			maxK = k
		}
	}

	f, err := os.Create(*outPath)
	if err != nil {
		die("create %s: %v", *outPath, err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{
		"algorithm", "rows", "cols", "density", "num_agents", "seed",
		"target_idx", "makespan", "time_to_discovery",
		"mean_utilization", "min_utilization",
	}); err != nil {
		die("csv header: %v", err)
	}

	planned := 0
	for range sizes {
		planned += *nSeeds * *nTargets * (1 + 2*len(agentCounts))
	}

	t0 := time.Now()
	done, skipped := 0, 0

	for _, sz := range sizes {
		for s := int64(0); s < int64(*nSeeds); s++ {
			g := grid.New(sz, sz, *density, s)

			if !megaConnected(g) {
				skipped += *nTargets * (1 + len(agentCounts))
				continue
			}

			placeRNG := rand.New(rand.NewSource(s + 1_000_000))
			targets := pickFineCells(g, *nTargets, placeRNG)
			allStarts := pickMegaStarts(g, maxK, placeRNG)

			// --- single-agent STC ---
			path1, ok := safeSTC(g, allStarts[0])
			if !ok {
				skipped += *nTargets
			} else {
				paths1 := map[int][]grid.Position{0: path1}
				for ti, target := range targets {
					r := sim.Run(g, paths1, target)
					writeRow(w, "single_agent_stc", sz, sz, *density, 1, s, ti, r)
					done++
				}
			}

			// --- Voronoi+STC for each k ---
			for _, k := range agentCounts {
				starts := allStarts[:k]
				parts := partition.Voronoi(g, starts)

				paths, ok := safeRegionPaths(g, parts, starts)
				if !ok {
					skipped += *nTargets
					continue
				}
				for ti, target := range targets {
					r := sim.Run(g, paths, target)
					writeRow(w, "voronoi_stc", sz, sz, *density, k, s, ti, r)
					done++
				}
			}
			// --- DARP+STC for each k ---
			for _, k := range agentCounts {
				starts := allStarts[:k]
				parts := partition.DARP(g, starts, partition.DefaultDARPConfig())

				paths, ok := safeRegionPaths(g, parts, starts)
				if !ok {
					skipped += *nTargets
					continue
				}
				for ti, target := range targets {
					r := sim.Run(g, paths, target)
					writeRow(w, "darp_stc", sz, sz, *density, k, s, ti, r)
					done++
				}
			}

			if (s+1)%10 == 0 {
				fmt.Fprintf(os.Stderr,
					"size=%d seed=%d/%d  rows=%d  skipped=%d  elapsed=%s\n",
					sz, s+1, *nSeeds, done, skipped, time.Since(t0).Round(time.Second))
			}
			w.Flush()
		}
	}

	fmt.Fprintf(os.Stderr, "complete: %d rows written, %d skipped of %d planned, %s elapsed\n",
		done, skipped, planned, time.Since(t0).Round(time.Second))
}

// safeSTC catches panics so a bad seed doesn't kill the run.
func safeSTC(g *grid.Grid, start grid.Position) (path []grid.Position, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
			path = nil
		}
	}()
	return stc.STC(g, start), true
}

// safeRegionPaths runs RegionSTC on every partition. If any panic, the whole
// k-agent configuration for this seed is skipped.
func safeRegionPaths(g *grid.Grid, parts map[int][]grid.Position,
	starts []grid.Position) (paths map[int][]grid.Position, ok bool) {

	defer func() {
		if r := recover(); r != nil {
			ok = false
			paths = nil
		}
	}()

	paths = make(map[int][]grid.Position, len(starts))
	for id := range starts {
		paths[id] = stc.RegionSTC(g, parts[id], starts[id])
	}
	return paths, true
}

func writeRow(w *csv.Writer, alg string, rows, cols int, density float64,
	k int, seed int64, ti int, r sim.Result) {

	meanU, minU := summarizeUtil(r.Utilization)
	_ = w.Write([]string{
		alg,
		strconv.Itoa(rows),
		strconv.Itoa(cols),
		strconv.FormatFloat(density, 'f', 3, 64),
		strconv.Itoa(k),
		strconv.FormatInt(seed, 10),
		strconv.Itoa(ti),
		strconv.Itoa(r.Makespan),
		strconv.Itoa(r.TimeToDiscovery),
		strconv.FormatFloat(meanU, 'f', 4, 64),
		strconv.FormatFloat(minU, 'f', 4, 64),
	})
}

func summarizeUtil(u map[int]float64) (mean, min float64) {
	if len(u) == 0 {
		return 0, 0
	}
	min = 1.0
	sum := 0.0
	for _, v := range u {
		sum += v
		if v < min {
			min = v
		}
	}
	return sum / float64(len(u)), min
}

func megaConnected(g *grid.Grid) bool {
	free := g.FreeMegaCells()
	if len(free) == 0 {
		return true
	}
	visited := map[grid.Position]bool{free[0]: true}
	queue := []grid.Position{free[0]}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, nb := range g.MegaNeighbors(c) {
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}
	return len(visited) == len(free)
}

func pickFineCells(g *grid.Grid, n int, rng *rand.Rand) []grid.Position {
	free := g.FreeCells()
	rng.Shuffle(len(free), func(i, j int) { free[i], free[j] = free[j], free[i] })
	if n > len(free) {
		n = len(free)
	}
	return free[:n]
}

// pickMegaStarts returns k distinct mega-cell starts, each given as the
// top-left fine cell of its mega-cell.
func pickMegaStarts(g *grid.Grid, k int, rng *rand.Rand) []grid.Position {
	free := g.FreeMegaCells()
	rng.Shuffle(len(free), func(i, j int) { free[i], free[j] = free[j], free[i] })
	if k > len(free) {
		die("not enough free mega-cells for %d agents (have %d)", k, len(free))
	}
	out := make([]grid.Position, k)
	for i := 0; i < k; i++ {
		m := free[i]
		out[i] = grid.Position{Row: m.Row * 2, Col: m.Col * 2}
	}
	return out
}

func mustParseInts(s string) []int {
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			die("parse %q: %v", p, err)
		}
		out = append(out, v)
	}
	return out
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}