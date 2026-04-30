package sim

import (
	"math"
	"testing"

	"cs730_project/grid" // adjust to match the module path in go.mod
)

func TestRun_SingleAgentTargetMidway(t *testing.T) {
	g := grid.New(4, 4, 0, 0)
	// Path: row 0 left-to-right, then row 1 right-to-left, etc.
	path := []grid.Position{
		{0, 0}, {0, 1}, {0, 2}, {0, 3},
		{1, 3}, {1, 2}, {1, 1}, {1, 0},
	}
	target := grid.Position{Row: 1, Col: 2} // index 5 in the path

	r := Run(g, map[int][]grid.Position{0: path}, target)

	if r.Makespan != 7 {
		t.Errorf("Makespan = %d, want 7", r.Makespan)
	}
	if r.TimeToDiscovery != 5 {
		t.Errorf("TimeToDiscovery = %d, want 5", r.TimeToDiscovery)
	}
	if r.Utilization[0] != 1.0 {
		t.Errorf("Utilization[0] = %v, want 1.0", r.Utilization[0])
	}
}

func TestRun_TargetAtStart(t *testing.T) {
	g := grid.New(4, 4, 0, 0)
	path := []grid.Position{{0, 0}, {0, 1}, {0, 2}}

	r := Run(g, map[int][]grid.Position{0: path}, grid.Position{Row: 0, Col: 0})

	if r.TimeToDiscovery != 0 {
		t.Errorf("TimeToDiscovery = %d, want 0", r.TimeToDiscovery)
	}
}

func TestRun_TargetNotFound(t *testing.T) {
	g := grid.New(4, 4, 0, 0)
	path := []grid.Position{{0, 0}, {0, 1}, {0, 2}}

	r := Run(g, map[int][]grid.Position{0: path}, grid.Position{Row: 3, Col: 3})

	if r.TimeToDiscovery != -1 {
		t.Errorf("TimeToDiscovery = %d, want -1", r.TimeToDiscovery)
	}
}

func TestRun_TwoAgentsImbalanced(t *testing.T) {
	g := grid.New(4, 4, 0, 0)
	// Agent 0 takes 8 ticks, agent 1 takes 4.
	long := []grid.Position{
		{0, 0}, {0, 1}, {0, 2}, {0, 3},
		{1, 3}, {1, 2}, {1, 1}, {1, 0},
	}
	short := []grid.Position{{3, 3}, {3, 2}, {3, 1}, {3, 0}}
	target := grid.Position{Row: 3, Col: 1} // agent 1 reaches at tick 2

	r := Run(g, map[int][]grid.Position{0: long, 1: short}, target)

	if r.Makespan != 7 {
		t.Errorf("Makespan = %d, want 7 (longer path)", r.Makespan)
	}
	if r.TimeToDiscovery != 2 {
		t.Errorf("TimeToDiscovery = %d, want 2", r.TimeToDiscovery)
	}
	// Agent 0: moves all 7 ticks → 7/7 = 1.0
	// Agent 1: moves on ticks 1-3, idles on 4-7 → 3/7
	if r.Utilization[0] != 1.0 {
		t.Errorf("Utilization[0] = %v, want 1.0", r.Utilization[0])
	}
	if want := 3.0 / 7.0; math.Abs(r.Utilization[1]-want) > 1e-9 {
		t.Errorf("Utilization[1] = %v, want %v", r.Utilization[1], want)
	}
}

// Earliest-tick tiebreak: if two agents both visit the target, discovery
// is the smaller tick.
func TestRun_DiscoveryEarliestTick(t *testing.T) {
	g := grid.New(4, 4, 0, 0)
	a := []grid.Position{{0, 0}, {0, 1}, {0, 2}, {0, 3}} // hits (0,3) at t=3
	b := []grid.Position{{1, 3}, {0, 3}}                 // hits (0,3) at t=1
	target := grid.Position{Row: 0, Col: 3}

	r := Run(g, map[int][]grid.Position{0: a, 1: b}, target)

	if r.TimeToDiscovery != 1 {
		t.Errorf("TimeToDiscovery = %d, want 1", r.TimeToDiscovery)
	}
}