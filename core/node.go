package core

import (
	"fmt"
	"math"

	"gonum.org/v1/plot/plotter"
)

const (
	// Energy values measured in [J/byte].
	E_TX = 40e-9      // Transmission
	E_RX = 4e-9       // Receiving
	E_MP = 0.0104e-12 // Multipath fading
	E_FS = 80e-12     // Line of sight free space channel
)

type Node struct {
	ID    int64
	Ready bool

	// Energy levels.
	Energy       float64
	energyPoints plotter.XYs

	// Coordinates.
	X float64
	Y float64

	dataSent     int64
	dataReceived int64
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
	// Call destination to receive.
	n.dataSent += msg
	dst.Receive(msg) // Do not fetch error.

	fmt.Printf("node <%d> sends to node <%d>\n", n.ID, dst.ID)
	return nil
}

func (n *Node) Receive(msg int64) error {
	// Deduct cost of receiving.
	if err := n.consume(E_TX * float64(msg)); err != nil {
		return err
	}
	n.dataReceived += msg

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

func (n *Node) Info() string {
	return fmt.Sprintf("node <%d> tx: <%d> rx: <%d>", n.ID, n.dataSent, n.dataReceived)
}
