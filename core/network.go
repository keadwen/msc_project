package core

import (
	"context"
	"fmt"
	"sync"
)

type Message struct {
	souceID       int64
	destinationID int64
	data          int64
}

type Network struct {
	BaseStation *Node
	Nodes       map[int64]*Node

	round int64
}

func (net *Network) AddNode(n *Node) error {
	if net.Nodes == nil {
		net.Nodes = make(map[int64]*Node)
	}
	if v, ok := net.Nodes[n.ID]; ok {
		fmt.Errorf("node ID <%d> already exists: %+v", n.ID, v)
	}
	net.Nodes[n.ID] = n
	return nil
}

func (net *Network) Simulate() {
	var wg sync.WaitGroup

	for r := 0; r < 10; r++ {
		net.round = int64(r)

		// Check if nodes are alive.
		if net.CheckNodes() == 0 {
			fmt.Printf("Simulation stopped. No active nodes.\n")
			break
		}

		for _, n := range net.Nodes {
			if !n.Ready {
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				n.Run()
			}()
		}
		wg.Wait() // Wait for all nodes to finish.
	}
}

func (net *Network) CheckNodes() int {
	count := 0
	for _, n := range net.Nodes {
		if n.Energy > 0 {
			count++
		}
	}
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

func (n *Node) Listen(ctx context.Context) {
}

func (n *Node) Transmit(msg Message) error {
	return nil
}

func (n *Node) Receive(msg Message) error {
	return nil
}

func (n *Node) Run() error {
	// Simply send a message to the base station.
	return nil
}
