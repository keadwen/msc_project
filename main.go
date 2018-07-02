package main

import (
	"fmt"
	"time"

	"github.com/keadwen/msc_project/core"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

func main() {
	// Create a plot for rounds.
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "Round"
	p.Y.Label.Text = "Energy"

	// Create new space.
	net := &core.Network{
		PlotRound: p,
	}

	// Create base station node.
	net.BaseStation = &core.Node{
		ID:     0,
		Ready:  true,
		Energy: 10000.0,
		X:      0.0,
		Y:      0.0,
	}

	// Create 2 nodes.
	net.AddNode(&core.Node{ID: 1, Ready: true, Energy: 100.0e-6, X: 100.0, Y: 0.0})
	net.AddNode(&core.Node{ID: 2, Ready: true, Energy: 200.0e-6, X: 200.0, Y: 0.0})

	// Run simulation for 10 rounds.
	net.Simulate()

	// Save the plot to a PNG file.
	if err := p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("graphs/rounds-%d.png", time.Now().Nanosecond())); err != nil {
		panic(err)
	}
}
