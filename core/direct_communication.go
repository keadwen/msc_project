package core

type DirectCommunication struct{}

// Setup implements Protocol.Setup.
func (d DirectCommunication) Setup(net *Network) error {
	net.Nodes.Range(func(_, n interface{}) bool {
		n.(*Node).nextHop = net.BaseStation
		return true
	})
	return nil
}
