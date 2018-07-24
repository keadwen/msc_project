package core

type Protocol interface {
	Setup(net *Network) ([]int64, error)
	SetNodes(int)
	SetClusters(int)
}
