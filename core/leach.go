package core

import (
	"math/rand"
	"time"
)

var (
	// seed
	seed = time.Now().UnixNano()
)

type LEACH struct {
	Clusters int // A number of clusters in the network.
	Nodes    int // A number of nodes in the network.
}

// Setup implements Protocol.Setup.
func (l *LEACH) Setup(net *Network) ([]int64, error) {
	r := rand.New(rand.NewSource(seed))

	// Clean previous cluster heads election.
	net.Nodes.Range(func(_, n interface{}) bool {
		n.(*Node).nextHop = net.BaseStation
		n.(*Node).receiveQueue = 0
		n.(*Node).transmitQueue = 0
		return true
	})

	// Election of cluster heads.
	var heads []int64
	for len(heads) < l.Clusters {
		net.Nodes.Range(func(_, n interface{}) bool {
			// Skip dead nodes.
			if !n.(*Node).Ready {
				return true
			}

			// Nominate the node to be a cluster head.
			ur := r.Float64()
			if ur*float64(l.Nodes) < float64(l.Nodes/l.Clusters) {
				heads = append(heads, n.(*Node).Conf.GetId())
			}
			return true
		})
	}
	// Shrink the slice to maximum amount of clusters.
	heads = heads[:l.Clusters]

	// Go through all nodes, as modify the base station to nearest cluster head.
	net.Nodes.Range(func(_, src interface{}) bool {
		// Skip dead nodes.
		if !src.(*Node).Ready {
			return true
		}
		// Omit the cluster head.
		for _, hid := range heads {
			if src.(*Node).Conf.GetId() == hid {
				return true
			}
		}

		// Assign to the base station. If more than one cluster heads,
		// choose the one with strongest signal (the smalles distance).
		nearest := src.(*Node).distance(net.BaseStation)
		for _, hid := range heads {
			dst, _ := net.Nodes.Load(hid)
			if d := src.(*Node).distance(dst.(*Node)); d < nearest {
				nearest = d
				// TODO(keadwen): Inform destinations about the next hop change.
				src.(*Node).nextHop = dst.(*Node)
			}
		}
		return true
	})

	// fmt.Printf("=== Validation ===\n")
	net.Nodes.Range(func(_, n interface{}) bool {
		// fmt.Printf("====> N<%d>: nextHop <%d>\n", n.(*Node).Conf.GetId(), n.(*Node).nextHop.Conf.GetId())
		return true
	})

	return heads, nil
}

// SetNodes implements Protocol.SetNodes.
func (l *LEACH) SetNodes(v int) {
	l.Nodes = v
}

// SetClusters implements Protocol.SetClusters.
func (l *LEACH) SetClusters(v int) {
	l.Clusters = v
	if v == 0 { // Leach must have at least one cluster.
		l.Clusters = 1
	}
}
