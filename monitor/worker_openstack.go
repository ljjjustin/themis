package monitor

import (
	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
)


const (
	flagManage  uint = 1 << 2
	flagStorage uint = 1 << 1
	flagNetwork uint = 1 << 0
)

var (
	flagTagMap = map[string]uint{
		"manage":  flagManage,
		"storage": flagStorage,
		"network": flagNetwork,
	}
	doFenceStatus = HostFailedStatus

	openstackDecisionMatrix = []bool{
		/* +------------+----------+-----------+--------------+ */
		/* | Management | Storage  | Network   |    Fence     | */
		/* +------------+----------+-----------+--------------+ */
		/* | good       | good     | good      | */ false, /* | */
		/* | good       | good     | bad       | */ true, /*  | */
		/* | good       | bad      | good      | */ true, /*  | */
		/* | good       | bad      | bad       | */ true, /*  | */
		/* | bad        | good     | good      | */ false, /* | */
		/* | bad        | good     | bad       | */ true, /*  | */
		/* | bad        | bad      | good      | */ true, /*  | */
		/* | bad        | bad      | bad       | */ true, /*  | */
		/* +-----------------------------------+--------------+ */
	}
)

type OpenstackWorker struct {
	config         *config.ThemisConfig
	decisionMatrix []bool
}

func NewOpenstackWorker(config *config.ThemisConfig) *OpenstackWorker{

	return &OpenstackWorker{
		config: config,
		decisionMatrix: openstackDecisionMatrix,
	}
}

func (w *OpenstackWorker) GetDecision(host *database.Host, states []*database.HostState) bool {

	if host.Disabled {
		return false
	}

	statusDecision := false
	if host.Status == doFenceStatus {
		statusDecision = true
	}

	var decision uint = 0
	for _, s := range states {
		// judge if one network is down.
		if s.FailedTimes >= 6 {
			decision |= flagTagMap[s.Tag]
		}
	}

	return statusDecision && w.decisionMatrix[decision]
}

func (w *OpenstackWorker) FenceHost(host *database.Host, states []*database.HostState) {
	defer func() {
		if err := recover(); err != nil {
			plog.Warning("unexpected error during HandleEvents: ", err)
		}
	}()

	// check if we have disabled fence operation globally
	if w.config.Fence.DisableFenceOps {
		plog.Info("fence operation have been disabled.")
		return
	}

	plog.Infof("Begin fence host %s", host.Name)
	// update host status
	host.Status = HostFencingStatus
	saveHost(host)

	// execute power off through IPMI
	fencers, err := database.FencerGetByHost(host.Id)
	if err != nil || len(fencers) < 1 {
		plog.Warning("Can't find fencers with given host: ", host.Name)
		return
	}

	var IPMIFencers []FencerInterface
	for _, fencer := range fencers {
		IPMIFencers = append(IPMIFencers, NewFencer(fencer))
	}

	plog.Debug("Begin execute fence operation")
	for _, fencer := range IPMIFencers {
		if err := fencer.Fence(); err != nil {
			plog.Warningf("Fence operation failed on host %s", host.Name)
			continue
		}
		plog.Infof("Fence operation successed on host: %s", host.Name)
		break
	}

	// evacuate all virtual machine on that host
	nova, err := NewNovaClient(&w.config.Openstack)
	if err != nil {
		plog.Warning("Can't create nova client: ", err)
		return
	}

	services, err := nova.ListServices()
	if err != nil {
		plog.Warning("Can't get service list", err)
		return
	}
	for _, service := range services {
		if host.Name == service.Host && service.Binary == "nova-compute" {
			nova.ForceDownService(service)
			nova.DisableService(service, "disabled by themis monitor")
		}
	}

	servers, err := nova.ListServers(host.Name)
	if err != nil {
		plog.Warning("Can't get service list: ", err)
		return
	}
	for _, server := range servers {
		id := server.ID
		plog.Infof("Try to evacuate instance: %s", id)
		nova.Evacuate(id)
	}

	// disable host status
	host.Status = HostFencedStatus
	host.Disabled = true
	saveHost(host)
}