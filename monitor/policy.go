package monitor

import (
	"time"

	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
)

const stateTransitionInterval = 60

type PolicyEngine struct {

	worker WorkerInterface
}

func NewPolicyEngine(config *config.ThemisConfig) *PolicyEngine {

	return &PolicyEngine{
		worker: NewWorker(config),
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
		// update host status
		plog.Debugf("update %s's FSM.", hostname)
		updateHostFSM(host, states)

		// judge if a host is down
		if p.worker.GetDecision(host, states) {
			p.worker.FenceHost(host, states)
		}
	}
}