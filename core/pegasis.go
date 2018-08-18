package core

import (
	"math/rand"
	"sort"
)

type PEGASIS struct {
	Clusters int // A number of clusters in the network.
	Nodes    int // A number of nodes in the network.
}

type clusterHead struct {
	hid            int64
	distanceToBase float64
}

// Setup implements Protocol.Setup.
func (p *PEGASIS) Setup(net *Network) ([]int64, error) {
	r := rand.New(rand.NewSource(seed))

	// Clean previous cluster heads election.
	net.Nodes.Range(func(_, n interface{}) bool {
		n.(*Node).nextHop = net.BaseStation
		n.(*Node).receiveQueue = 0
		n.(*Node).transmitQueue = 0
		return true
	})

	// Election of cluster heads.
	var heads []clusterHead
	for len(heads) < p.Clusters {
		net.Nodes.Range(func(_, n interface{}) bool {
			// Skip dead nodes.
			if !n.(*Node).Ready {
				return true
			}

			// Nominate the node to be a cluster head.
			ur := r.Float64()
			if ur*float64(p.Nodes) < float64(p.Nodes/p.Clusters) {
				heads = append(heads, clusterHead{
					hid:            n.(*Node).Conf.GetId(),
					distanceToBase: n.(*Node).distance(net.BaseStation),
				})
			}
			return true
		})
	}
	// Shrink the slice to maximum amount of clusters.
	heads = heads[:p.Clusters]
	if len(heads) == 0 {
		return []int64{}, nil
	}

	// Sort cluster heads by distance (from higher to lower).
	sort.Slice(heads, func(i, j int) bool { return heads[i].distanceToBase > heads[j].distanceToBase })

	// PEGASIS algorithm.
	// For each cluster head (starting from the most remote from Base Station)
	// find new nexthop to relay the data. The new nexthop must also be a cluster head
	// and it must be located closer to the Base Station.
	for index, s := range heads {
		no, _ := net.Nodes.Load(s.hid)
		src := no.(*Node)

		// To prevent the loops, do not look back in previous clusters.
		for _, d := range heads[index:] {
			if s.hid == d.hid {
				continue
			}
			no, _ = net.Nodes.Load(d.hid)
			dst := no.(*Node)

			// Cluster head S has closer to cluster D than Base Station.
			if src.distance(dst) < s.distanceToBase {
				src.nextHop = dst
			}
		}
	}

	var hids []int64
	for _, h := range heads {
		hids = append(hids, h.hid)
	}
	return hids, nil
}

// SetNodes implements Protocol.SetNodes.
func (p *PEGASIS) SetNodes(v int) {
	p.Nodes = v
}

// SetClusters implements Protocol.SetClusters.
func (p *PEGASIS) SetClusters(v int) {
	p.Clusters = v
	if v == 0 { // Leach must have at least one cluster.
		p.Clusters = 1
	}
}
