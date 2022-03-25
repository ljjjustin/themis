package monitor

import "themis/database"

type FencerInterface interface {
	Fence() error
}

func NewFencer(fencer *database.HostFencer) FencerInterface {
	return NewIPMIFencer(fencer)
}
