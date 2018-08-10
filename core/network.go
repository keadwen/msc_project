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

	Round int64

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
	maxRounds := int64(25000) // TODO(keadwen): Put inside a config file.
	for net.CheckNodes() > 0 && net.Round < maxRounds {
		net.Round++
		// fmt.Printf("=== Round %d ===\n", net.Round)
		// Perform data collection before the Round.
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
					//fmt.Println(err)
					fmt.Println(n.Info())
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
					// No action.
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
