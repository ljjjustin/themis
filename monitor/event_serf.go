package monitor

import (
	"github.com/hashicorp/serf/client"
)

type SerfMonitor struct {
	rpcAddr string
}

func NewSerfMonitor(rpcaddr string) *SerfMonitor {
	return &SerfMonitor{rpcAddr: rpcaddr}
}

func (m *SerfMonitor) Start() (chan Events, error) {

	eventCh := make(chan Events)

	go getMembers(m.rpcAddr, eventCh)

	return eventCh, nil
}

func getMembers(address string, eventCh chan Events) {
	rpc, err := client.NewRPCClient(address)
	if err != nil {
		plog.Noticef("Create Serf RPC client failed: %s", err)
		return
	}
	defer rpc.Close()

	members, err := rpc.Members()
	if err != nil {
		plog.Noticef("Query members failed: %s", err)
		return
	}
	var events Events
	for _, member := range members {
		// convert status
		var status string
		if member.Status == "alive" {
			status = "active"
		} else if member.Status == "failed" {
			status = "failed"
		}
		event := &Event{
			Hostname:   member.Name,
			NetworkTag: member.Tags["network"],
			Status:     status,
		}
		events = append(events, event)
	}
	eventCh <- events
	close(eventCh)
}
