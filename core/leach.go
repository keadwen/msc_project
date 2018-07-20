package core

import (
	"fmt"
	"math/rand"
	"time"
)

type LEACH struct {
	Clusters int // A number of clusters in the network.
	Nodes    int // A number of nodes in the network.
}

// Setup implements Protocol.Setup.
func (l *LEACH) Setup(net *Network) ([]int64, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var heads []int64
	for {
		net.Nodes.Range(func(_, n interface{}) bool {
			// Skip dead nodes.
			if !n.(*Node).Ready {
				return true
			}

			// Nominate the node to be a cluster head.
			ur := r.Float64()
			if ur < float64(l.Clusters)/float64(l.Nodes) {
				n.(*Node).nextHop = net.BaseStation
				heads = append(heads, n.(*Node).Conf.GetId())
			}
			return true
		})

		// Shrink the slice to maximum amount of clusters.
		// TODO(keadwen): Check if the shrinking is required.
		if len(heads) >= l.Clusters {
			heads = heads[:l.Clusters]
			break
		}
	}

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

		// Check distance to each cluster head.
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

	fmt.Printf("=== Validation ===\n")
	net.Nodes.Range(func(_, n interface{}) bool {
		fmt.Printf("====> N<%d>: nextHop <%d>\n", n.(*Node).Conf.GetId(), n.(*Node).nextHop.Conf.GetId())
		return true
	})

	return heads, nil
}
