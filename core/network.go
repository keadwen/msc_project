package core

import (
	"fmt"
	"math"
	"sync"
)

const (
	// Energy values measured in [J/byte].
	E_TX = 40e-9      // Transmission
	E_RX = 4e-9       // Receiving
	E_MP = 0.0104e-12 // Multipath fading
	E_FS = 80e-12     // Line of sight free space channel
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
				if err := n.Transmit(int64(n.X), net.BaseStation); err != nil {
					fmt.Println(err)
				}
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
	X float64
	Y float64
}

func (n *Node) Transmit(msg int64, dst *Node) error {
	// Deduct cost of transmission.
	var cost float64
	if d := n.distance(dst); d > math.Sqrt(E_FS/E_MP) {
		cost = E_TX + math.Pow(E_MP, 4)
	} else {
		cost = E_TX + math.Pow(E_MP, 2)
	}
	if err := n.consume(cost * float64(msg)); err != nil {
		return err
	}

	fmt.Printf("node <%d> sends to node <%d>\n", n.ID, dst.ID)
	return nil
}

func (n *Node) Receive(msg int64) error {
	// Deduct cost of receiving.
	if err := n.consume(E_TX * float64(msg)); err != nil {
		return err
	}

	fmt.Printf("node <%d> receive message <%d>\n", n.ID, msg)
	return nil
}

func (n *Node) distance(dst *Node) float64 {
	x := math.Abs(n.X - dst.X)
	y := math.Abs(n.Y - dst.Y)
	return math.Hypot(x, y)
}

func (n *Node) consume(e float64) error {
	if n.Energy-e < 0 {
		n.Ready = false
		n.Energy = 0
		return fmt.Errorf("node <%d> no more energy!", n.ID)
	}
	n.Energy -= e
	return nil
}
