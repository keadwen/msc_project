package core

import (
	"fmt"
	"sync"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
)

type Network struct {
	Protocol    Protocol
	BaseStation *Node
	Nodes       sync.Map

	round int64

	PlotRound        *plot.Plot
	PlotNodes        *plot.Plot
	nodesAlivePoints plotter.XYs
}

func (net *Network) AddNode(n *Node) error {
	if v, ok := net.Nodes.Load(n.Conf.GetId()); ok {
		fmt.Errorf("node ID <%d> already exists: %+v", n.Conf.GetId(), v)
	}
	n.Energy = n.Conf.GetInitialEnergy()
	n.energyPoints = plotter.XYs{{X: float64(0), Y: float64(n.Energy)}}
	n.nextHop = net.BaseStation
	n.dataSent = 0
	n.dataReceived = 0
	net.Nodes.Store(n.Conf.GetId(), n)
	return nil
}

func (net *Network) Simulate() error {
	net.round = 0
	for net.CheckNodes() > 0 {
		net.round++
		fmt.Printf("=== Round %d ===\n", net.round)
		// Perform data collection before the round.
		net.PopulateEnergyPoints()
		net.PopulateNodesAlivePoints()

		// Setup routing protocol.
		heads, err := net.Protocol.Setup(net)
		if err != nil {
			return fmt.Errorf("failed to setup: %v", err)
		}

		// Run leaf nodes (not cluster heads).
		var wg sync.WaitGroup
		net.Nodes.Range(func(_, n interface{}) bool {
			if !n.(*Node).Ready {
				return true
			}
			for _, h := range heads {
				if h == n.(*Node).Conf.GetId() {
					return true // Skip cluster heads.
				}
			}

			wg.Add(1)
			go func(n *Node) {
				defer wg.Done()
				// Send the transmit queue to next hop.
				if err := n.Transmit(DEFAULT_MSG, n.nextHop); err != nil {
					fmt.Println(err)
				}
			}(n.(*Node))
			return true
		})
		wg.Wait()

		// Run cluster head nodes.
		net.Nodes.Range(func(_, n interface{}) bool {
			if !n.(*Node).Ready {
				return true
			}
			var head bool
			for _, h := range heads {
				if h == n.(*Node).Conf.GetId() {
					head = true
				}
			}
			if !head {
				return true // Skip leaf nodes.
			}

			wg.Add(1)
			go func(n *Node) {
				defer wg.Done()
				// Read the receiving queue and move to transmit queue.
				n.transmitQueue = n.receiveQueue
				if err := n.Transmit(DEFAULT_MSG+n.transmitQueue, n.nextHop); err != nil {
					fmt.Println(err)
				}
				n.receiveQueue = 0
			}(n.(*Node))
			return true
		})
		wg.Wait() // Wait for all nodes to finish before plot.
	}
	// Final data collection and plotting.
	net.PopulateEnergyPoints()
	net.PopulateNodesAlivePoints()
	net.PlotAggregatedEnergy()
	net.PlotNodesAlive()

	// Display the final count of TX/RX data per node.
	fmt.Println("=== Final ===")
	fmt.Println(net.BaseStation.Info())
	net.Nodes.Range(func(_, n interface{}) bool {
		fmt.Println(n.(*Node).Info())
		return true
	})
	return nil
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
		return true
	})
}

func (net *Network) PopulateNodesAlivePoints() {
	net.nodesAlivePoints = append(net.nodesAlivePoints, plotter.XYs{{
		X: float64(net.round),
		Y: float64(net.CheckNodes()),
	}}...)
}

func (net *Network) PlotEnergy() {
	net.Nodes.Range(func(_, n interface{}) bool {
		if err := plotutil.AddLinePoints(
			net.PlotRound,
			fmt.Sprintf("Node <%d>", n.(*Node).Conf.GetId()),
			n.(*Node).energyPoints,
		); err != nil {
			fmt.Printf("failed to AddLinePoints(): %v", err)
		}
		return true
	})
}

func (net *Network) PlotNodesAlive() {
	if err := plotutil.AddLinePoints(
		net.PlotNodes,
		fmt.Sprintf(""),
		net.nodesAlivePoints,
	); err != nil {
		fmt.Printf("failed to AddLinePoints(): %v", err)
	}
}

func (net *Network) PlotAggregatedEnergy() {
	var aggEnergy plotter.XYs
	for r := 1; r < int(net.round); r++ {
		var e float64
		net.Nodes.Range(func(_, n interface{}) bool {
			e += n.(*Node).energyPoints[r].Y
			return true
		})
		aggEnergy = append(aggEnergy, plotter.XYs{{X: float64(r), Y: e}}...)
	}

	if err := plotutil.AddLinePoints(
		net.PlotRound,
		fmt.Sprintf(""),
		aggEnergy,
	); err != nil {
		fmt.Printf("failed to AddLinePoints(): %v", err)
	}
}
