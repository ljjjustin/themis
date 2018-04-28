package fence

import "github.com/ljjjustin/themis/database"

type FencerInterface interface {
	Fence() error
}

func NewFencer(fencer *database.HostFencer) FencerInterface {
	return NewIPMIFencer(fencer)
}