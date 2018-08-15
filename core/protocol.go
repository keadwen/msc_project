package core

import "time"

var (
	// seed
	seed = time.Now().UnixNano()
)

type Protocol interface {
	Setup(net *Network) ([]int64, error)
	SetNodes(int)
	SetClusters(int)
}
