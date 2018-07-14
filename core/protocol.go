package core

type Protocol interface {
	Setup(net *Network) error
}
