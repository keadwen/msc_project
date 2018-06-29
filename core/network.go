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
				n.Transmit(n.X, net.BaseStation)
			}(n.(*Node))
			return true
		})
		wg.Wait() // Wait for all nodes to finish.
	}
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

type Node struct {
	ID    int64
	Ready bool

	// Energy levels.
	Energy float64

	// Coordinates.
	X int64
	Y int64
}

func (n *Node) Transmit(msg int64, dst *Node) error {
	fmt.Printf("node <%d> sends to node <%d>\n", n.ID, dst.ID)
	dst.Receive(msg)
	return nil
}

func (n *Node) Receive(msg int64) error {
	fmt.Printf("node <%d> receive message <%d>\n", n.ID, msg)
	return nil
}
