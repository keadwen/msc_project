package core

import (
	"fmt"
	"sync"
)

type Network struct {
	BaseStation *Node
	Nodes       sync.Map

	round int64
}

func (net *Network) AddNode(n *Node) error {
	if v, ok := net.Nodes.Load(n.ID); ok {
		fmt.Errorf("node ID <%d> already exists: %+v", n.ID, v)
	}
	n.tx_data = 0
	n.rx_data = 0
	net.Nodes.Store(n.ID, n)
	return nil
}

func (net *Network) Simulate() {
	for r := 0; r < 2; r++ {
		fmt.Printf("=== Round %d ===\n", r)
		net.round = int64(r)

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
		wg.Wait() // Wait for all nodes to finish.
	}

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
