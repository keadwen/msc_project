package core

import (
	"fmt"
	"sync"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
)

type Network struct {
	BaseStation *Node
	Nodes       sync.Map

	round int64

	PlotRound *plot.Plot
}

func (net *Network) AddNode(n *Node) error {
	if v, ok := net.Nodes.Load(n.ID); ok {
		fmt.Errorf("node ID <%d> already exists: %+v", n.ID, v)
	}
	n.energyPoints = plotter.XYs{}
	n.tx_data = 0
	n.rx_data = 0
	net.Nodes.Store(n.ID, n)
	return nil
}

func (net *Network) Simulate() {
	net.round = 0
	for net.CheckNodes() > 0 {
		net.round++
		fmt.Printf("=== Round %d ===\n", net.round)

		// Check if nodes are alive.
		if net.CheckNodes() == 0 {
			fmt.Printf("Simulation stopped. No active nodes.\n")
			break
		}

		var wg sync.WaitGroup
		net.Nodes.Range(func(_, n interface{}) bool {
			if !n.(*Node).Ready {
				return true
			}

			wg.Add(1)
			go func(n *Node) {
				defer wg.Done()
				// Simply send the message to Base Station.
				if err := n.Transmit(int64(n.X), net.BaseStation); err != nil {
					fmt.Println(err)
				}
			}(n.(*Node))
			return true
		})
		wg.Wait() // Wait for all nodes to finish before plot.
		net.PopulateEnergyPoints()
	}
	// Recollect all plot data.
	net.PlotEnergy()

	// Display the final count of TX/RX data per node.
	fmt.Println("=== Final ===")
	fmt.Println(net.BaseStation.Info())
	net.Nodes.Range(func(_, n interface{}) bool {
		fmt.Println(n.(*Node).Info())
		return true
	})
}

func (net *Network) CheckNodes() int {
	count := 0
	net.Nodes.Range(func(_, n interface{}) bool {
		if n.(*Node).Energy > 0 {
			count++
		}
		return true
	})
	return count
}

func (net *Network) PopulateEnergyPoints() {
	net.Nodes.Range(func(_, n interface{}) bool {
		n.(*Node).energyPoints = append(n.(*Node).energyPoints, plotter.XYs{{
			X: float64(net.round),
			Y: n.(*Node).Energy,
		}}...)
		// fmt.Printf("Node: %v\n", n.(*Node).energyPoints)
		return true
	})
}

func (net *Network) PlotEnergy() {
	net.Nodes.Range(func(_, n interface{}) bool {
		if err := plotutil.AddLinePoints(
			net.PlotRound,
			fmt.Sprintf("Node <%d>", n.(*Node).ID),
			n.(*Node).energyPoints,
		); err != nil {
			fmt.Printf("failed to AddLinePoints(): %v", err)
		}
		return true
	})
}
