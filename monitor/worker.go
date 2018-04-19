package monitor

import (
	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
)

type WorkerInterface interface {
	GetDecision(host *database.Host, states []*database.HostState) bool
	FenceHost(host *database.Host, states []*database.HostState)
}

func NewWorker(config *config.ThemisConfig) WorkerInterface {

	return NewOpenstackWorker(config)
}