package core

import (
	"fmt"
	"sync"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
)

type Network struct {
	Protocol    Protocol
	BaseStation *Node
	Nodes       sync.Map

	Round     int64
	MaxRounds int64

	PlotTotalEnergy   *plot.Plot // An amount of total energy in the network per Round.
	PlotNodes         *plot.Plot // A number of alive nodes in the network per Round.
	NodesAlivePoints  plotter.XYs
	NodesEnergyPoints map[int64]plotter.XYs
}

func (net *Network) AddNode(n *Node) error {
	if net.NodesEnergyPoints == nil {
		net.NodesEnergyPoints = map[int64]plotter.XYs{}
	}
	if v, ok := net.Nodes.Load(n.Conf.GetId()); ok {
		fmt.Errorf("node ID <%d> already exists: %+v", n.Conf.GetId(), v)
	}
	n.Energy = n.Conf.GetInitialEnergy()
	n.nextHop = net.BaseStation
	n.dataSent = 0
	n.dataReceived = 0
	net.Nodes.Store(n.Conf.GetId(), n)
	// Include the initial state of the node in the plot.
	net.NodesEnergyPoints[n.Conf.GetId()] = plotter.XYs{{X: float64(0), Y: float64(n.Energy)}}
	return nil
}

func (net *Network) Simulate() error {
	net.Round = 0
	for net.CheckNodes() > 0 && net.Round < net.MaxRounds {
		net.Round++
		// Perform data collection before the Round.
		net.PopulateEnergyPoints()
		net.PopulateNodesAlivePoints()

		// Setup routing protocol.
		heads, err := net.Protocol.Setup(net)
		if err != nil {
			return fmt.Errorf("failed to setup: %v", err)
		}

		// Run leaf nodes (not cluster heads).
		net.Nodes.Range(func(_, ni interface{}) bool {
			n := ni.(*Node)

			if !n.Ready {
				return true
			}
			for _, h := range heads {
				if h == n.Conf.GetId() {
					return true // Skip cluster heads.
				}
			}

			// Send the transmit queue to next hop.
			if err := n.Transmit(DEFAULT_MSG, n.nextHop); err != nil {
				// If you receive a dead message from nexthop, send directly to base.
				n.Transmit(DEFAULT_MSG, net.BaseStation)
			}
			return true
		})

		// Run cluster head nodes.
		net.Nodes.Range(func(_, ni interface{}) bool {
			n := ni.(*Node)

			if !n.Ready {
				return true
			}

			// Check if a node is a cluster head.
			var head bool
			for _, h := range heads {
				if h == n.Conf.GetId() {
					head = true
					break
				}
			}
			if !head {
				return true // Skip non-cluster head nodes.
			}

			// Read the receiving queue and move to transmit queue.
			n.transmitQueue = n.receiveQueue
			if err := n.Transmit(DEFAULT_MSG+n.transmitQueue, n.nextHop); err != nil {
				// No action.
			}
			n.receiveQueue = 0
			return true
		})
	}
	// Final data collection and plotting.
	net.PopulateEnergyPoints()
	net.PopulateNodesAlivePoints()

	// Display the final count of TX/RX data per node.
	fmt.Printf("=== Final: %d\n%v\n\n", net.Round, net.BaseStation.Info())
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
		net.NodesEnergyPoints[n.(*Node).Conf.GetId()] = append(
			net.NodesEnergyPoints[n.(*Node).Conf.GetId()],
			plotter.XYs{{
				X: float64(net.Round),
				Y: n.(*Node).Energy,
			}}...)
		return true
	})
}

func (net *Network) PopulateNodesAlivePoints() {
	net.NodesAlivePoints = append(net.NodesAlivePoints, plotter.XYs{{
		X: float64(net.Round),
		Y: float64(net.CheckNodes()),
	}}...)
}
