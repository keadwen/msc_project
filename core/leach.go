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

func (l *LEACH) Setup(net *Network) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var heads []int64
	fmt.Println("== Starting LEACH ==")
	for {
		net.Nodes.Range(func(_, n interface{}) bool {
			// Skip dead nodes.
			if !n.(*Node).Ready {
				return true
			}

			// Nominate the node to be a cluster head.
			ur := r.Float64()
			fmt.Printf("====> %v < %v\n", ur, float64(l.Clusters)/float64(l.Nodes))
			if ur < float64(l.Clusters)/float64(l.Nodes) {
				n.(*Node).nextHop = net.BaseStation
				heads = append(heads, n.(*Node).Conf.GetId())
			}

			return true
		})
		fmt.Printf("====> Heads: %v\n", heads)

		// Shrink the slice to maximum amount of clusters.
		if len(heads) >= l.Clusters {
			heads = heads[:l.Clusters]
			break
		}
	}

	fmt.Printf("==> Final heads: %v\n", heads)
	// Go through all nodes, as modify the base station to nearest cluster head.
	net.Nodes.Range(func(_, src interface{}) bool {
		// Skip dead nodes.
		if !src.(*Node).Ready {
			fmt.Printf("====> Ready: False <%d>\n", src.(*Node).Conf.GetId())
			return true
		}
		// Omit the cluster head.
		for _, hid := range heads {
			if src.(*Node).Conf.GetId() == hid {
				fmt.Printf("====> Node Head <%d>\n", src.(*Node).Conf.GetId())
				return true
			}
		}

		nearest := src.(*Node).distance(net.BaseStation)
		// Check distance to each cluster head.
		for _, hid := range heads {
			fmt.Printf("===> BS distance %v\n", nearest)
			dst, _ := net.Nodes.Load(hid)
			if d := src.(*Node).distance(dst.(*Node)); d < nearest {
				fmt.Printf("====> New NH: <%d>\n", src.(*Node).Conf.GetId())
				nearest = d
				// TODO(keadwen): Inform destinations about the next hop change.
				src.(*Node).nextHop = dst.(*Node)
			}
		}
		return true
	})

	return nil
}
