package monitor

import (
	"time"

	"github.com/ljjjustin/themis/database"
)

const (
	flagManagement uint = 1 << 2
	flagStorage    uint = 1 << 1
	flagNetwork    uint = 1 << 0
)

var (
	flagTagMap = map[string]uint{
		"management": flagManagement,
		"storage":    flagStorage,
		"network":    flagNetwork,
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
	decisionMatrix []bool
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		decisionMatrix: openstackDecisionMatrix,
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
		var host *database.Host
		host, err := database.HostGetByName(hostname)
		if err != nil {
			plog.Warningf("Can't find Host %s.", hostname)
			return
		} else if host == nil {
			// save to database
			host = &database.Host{Name: hostname}
			if err := database.HostInsert(host); err != nil {
				plog.Warning("Save host failed", err)
			}
		}

		// update host states
		for tag, status := range tags {

			state, err := database.StateGetByTag(host.Id, tag)
			if err != nil {
				plog.Warning("Can't find Host states by tag")
				continue
			} else if state == nil {
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
			updateState(state, status)
			database.StateUpdate(state.Id, state)
		}

		var states []*database.HostState
		states, err = database.StateGetAll(host.Id)
		if err != nil {
			plog.Warning("Can't find Host states")
			return
		}

		// judge if a host is down
		if p.getDecision(host, states) {
			go fenceHost(host)
		}
	}
}

func (p *PolicyEngine) getDecision(host *database.Host,
	states []*database.HostState) bool {

	var decision uint = 0
	for _, s := range states {
		if isDown(s) {
			decision |= flagTagMap[s.Tag]
		}
	}
	return p.decisionMatrix[decision]
}

func updateState(state *database.HostState, status string) {
	if status == "active" && state.FailedTimes > 0 {
		state.FailedTimes -= 1
	} else if status == "failed" {
		state.FailedTimes += 1
	}
}

func isDown(state *database.HostState) bool {
	// judge if one network is down.
	duration := time.Since(state.UpdatedAt)
	if duration.Seconds() >= 60 && state.FailedTimes >= 6 {
		return true
	} else {
		return false
	}
}

func fenceHost(host *database.Host) {
	plog.Infof("fence host %s", host.Name)

	// execute power off through IPMI
	fencers, err := database.FencerGetAll(host.Id)
	if err != nil || len(fencers) < 1 {
		plog.Warning("Can't find fencers with given host: ", host.Name)
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

	// TODO: evacuate all virtual machine on that host
}
