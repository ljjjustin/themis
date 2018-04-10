package monitor

import (
	"time"

	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
)

const (
	flagManage  uint = 1 << 2
	flagStorage uint = 1 << 1
	flagNetwork uint = 1 << 0

	// state
	stateTransitionInterval = 60
)

var (
	flagTagMap = map[string]uint{
		"manage":  flagManage,
		"storage": flagStorage,
		"network": flagNetwork,
	}
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

type PolicyEngine struct {
	config         *config.ThemisConfig
	decisionMatrix []bool
}

func NewPolicyEngine(config *config.ThemisConfig) *PolicyEngine {
	return &PolicyEngine{
		config:         config,
		decisionMatrix: openstackDecisionMatrix,
	}
}

func saveHost(host *database.Host) {
	host.UpdatedAt = time.Now()
	database.HostUpdateFields(host, "status", "disabled", "updated_at")
}

func isAllActive(states []*database.HostState) bool {
	allActive := true
	for _, state := range states {
		if state.FailedTimes > 0 {
			allActive = false
			break
		}
	}
	return allActive
}

func hasAnyFailure(states []*database.HostState) bool {
	hasFailure := false
	for _, state := range states {
		if state.FailedTimes > 0 {
			hasFailure = true
			break
		}
	}
	return hasFailure
}

func hasFatalFailure(states []*database.HostState) bool {
	keyStates := make([]*database.HostState, 0)
	for _, s := range states {
		if s.Tag == "network" || s.Tag == "storage" {
			keyStates = append(keyStates, s)
		}
	}

	hasFailure := false
	for _, state := range keyStates {
		if state.FailedTimes > 0 {
			hasFailure = true
			break
		}
	}
	return hasFailure
}

func updateHostFSM(host *database.Host, states []*database.HostState) {

	duration := time.Since(host.UpdatedAt).Seconds()
	switch host.Status {
	case HostActiveStatus:
		if hasAnyFailure(states) {
			host.Status = HostCheckingStatus
			saveHost(host)
		}
	case HostInitialStatus:
		if duration >= stateTransitionInterval {
			if isAllActive(states) {
				host.Status = HostActiveStatus
				saveHost(host)
			}
		}
	case HostCheckingStatus:
		if duration >= stateTransitionInterval {
			if isAllActive(states) {
				host.Status = HostActiveStatus
				saveHost(host)
			} else if hasFatalFailure(states) {
				host.Status = HostFailedStatus
				saveHost(host)
			}
		}
	}
}

func (p *PolicyEngine) HandleEvents(events Events) {

	// group by hostname
	hostTags := map[string]map[string]string{}
	for _, e := range events {
		tags := hostTags[e.Hostname]
		if tags != nil {
			tags[e.NetworkTag] = e.Status
		} else {
			tags = map[string]string{
				e.NetworkTag: e.Status,
			}
		}
		hostTags[e.Hostname] = tags
	}

	for hostname, tags := range hostTags {
		plog.Debugf("Handle %s's events.", hostname)

		var host *database.Host
		host, err := database.HostGetByName(hostname)
		if err != nil {
			plog.Warningf("Can't find Host %s.", hostname)
			return
		} else if host == nil {
			// save to database
			host = &database.Host{
				Name:     hostname,
				Status:   HostInitialStatus,
				Disabled: false,
			}
			if err := database.HostInsert(host); err != nil {
				plog.Warning("Save host failed", err)
				continue
			}
		}

		// update host states
		var states []*database.HostState
		states, err = database.StateGetAll(host.Id)
		if err != nil {
			plog.Warning("Can't find Host states")
			continue
		}
		for tag, status := range tags {
			var state *database.HostState
			for i := range states {
				if states[i].Tag == tag {
					state = states[i]
					break
				}
			}
			if state == nil { // if we don't find matched state
				state = &database.HostState{
					HostId:      host.Id,
					Tag:         tag,
					FailedTimes: 0,
				}
				if err := database.StateInsert(state); err != nil {
					plog.Warning("Save host state failed", err)
					continue
				}
			}
			if !host.Disabled {
				if status == "active" && state.FailedTimes > 0 {
					state.FailedTimes -= 1
				} else if status == "failed" {
					state.FailedTimes += 1
				}
			}
			database.StateUpdateFields(state, "failed_times")
		}

		states, err = database.StateGetAll(host.Id)
		if err != nil {
			plog.Warning("Can't find Host states")
			return
		}
		for _, state := range states {
			plog.Debugf("%d failed times: %d", state.HostId, state.FailedTimes)
		}

		// update host status
		plog.Debugf("update %s's status.", hostname)
		updateHostFSM(host, states)

		// judge if a host is down
		if !host.Disabled && p.getDecision(host, states) {
			p.fenceHost(host, states)
		}
	}
}

func (p *PolicyEngine) getDecision(host *database.Host, states []*database.HostState) bool {

	if host.Disabled {
		return false
	}

	var decision uint = 0
	for _, s := range states {
		// judge if one network is down.
		if s.FailedTimes >= 6 {
			decision |= flagTagMap[s.Tag]
		}
	}
	return p.decisionMatrix[decision]
}

func (p *PolicyEngine) fenceHost(host *database.Host, states []*database.HostState) {
	defer func() {
		if err := recover(); err != nil {
			plog.Warning("unexpected error during HandleEvents: ", err)
		}
	}()

	// check if we have disabled fence operation globally
	if p.config.Fence.DisableFenceOps {
		plog.Info("fence operation have been disabled.")
		return
	}

	plog.Infof("Begin fence host %s", host.Name)
	// update host status
	host.Status = HostFencingStatus
	saveHost(host)

	// execute power off through IPMI
	fencers, err := database.FencerGetAll(host.Id)
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
	nova, err := NewNovaClient(&p.config.Openstack)
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
