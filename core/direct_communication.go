package core

type DirectCommunication struct {
}

// Setup implements Protocol.Setup.
func (d DirectCommunication) Setup(net *Network) ([]int64, error) {
	net.Nodes.Range(func(_, n interface{}) bool {
		n.(*Node).nextHop = net.BaseStation
		return true
	})
	return nil, nil // There are no cluster head ids to return.
}

// SetNodes implements Protocol.SetNodes.
func (_ DirectCommunication) SetNodes(_ int) {}

// SetClusters implements Protocol.SetClusters.
func (_ DirectCommunication) SetClusters(_ int) {}
