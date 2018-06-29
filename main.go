package main

import "github.com/keadwen/msc_project/core"

func main() {
	// Create new space.
	net := &core.Network{}

	// Create base station node.
	net.BaseStation = &core.Node{
		ID:     0,
		Ready:  true,
		Energy: 10000.0,
		X:      0.0,
		Y:      0.0,
	}

	// Create 2 nodes.
	net.AddNode(&core.Node{1, true, 100.0e-7, 10.0, 0.0})
	net.AddNode(&core.Node{2, true, 100.0e-8, 20.0, 0.0})

	// Run simulation for 10 rounds.
	net.Simulate()

}
