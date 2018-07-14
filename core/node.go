package core

import (
	"fmt"
	"math"

	"github.com/keadwen/msc_project/proto"
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
	Conf    config.Node
	Ready   bool
	nextHop *Node

	// Energy levels.
	Energy       float64
	energyPoints plotter.XYs

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
	err := dst.Receive(msg)
	if err != nil {
		fmt.Printf("destination node <%d> failed: %v\n", dst.Conf.GetId(), err)
	}

	fmt.Printf("node <%d> sends to node <%d>\n", n.Conf.GetId(), dst.Conf.GetId())
	return nil
}

func (n *Node) Receive(msg int64) error {
	// Deduct cost of receiving.
	if err := n.consume(E_RX * float64(msg)); err != nil {
		return err
	}
	n.dataReceived += msg

	fmt.Printf("node <%d> receive message <%d>\n", n.Conf.GetId(), msg)
	return nil
}

func (n *Node) distance(dst *Node) float64 {
	x := math.Abs(n.Conf.GetLocation().GetX() - dst.Conf.GetLocation().GetX())
	y := math.Abs(n.Conf.GetLocation().GetY() - dst.Conf.GetLocation().GetY())
	return math.Hypot(x, y)
}

func (n *Node) consume(e float64) error {
	if n.Energy-e < 0 {
		n.Ready = false
		n.Energy = 0
		return fmt.Errorf("node <%d> no more energy!", n.Conf.GetId())
	}
	n.Energy -= e
	return nil
}

func (n *Node) Info() string {
	return fmt.Sprintf("node <%d> tx: <%d> rx: <%d>", n.Conf.GetId(), n.dataSent, n.dataReceived)
}
