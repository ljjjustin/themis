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

	go getFailedMembers(m.rpcAddr, eventCh)

	return eventCh, nil
}

func getFailedMembers(address string, eventCh chan Events) {
	rpc, err := client.NewRPCClient(address)
	if err != nil {
		plog.Noticef("Create Serf RPC client failed: %s", err)
		return
	}
	defer rpc.Close()

	members, err := rpc.MembersFiltered(nil, "failed", "")
	if err != nil {
		plog.Noticef("Query member failed: %s", err)
		return
	}
	var events Events
	for _, member := range members {
		event := &Event{
			Hostname:   member.Name,
			NetworkTag: member.Tags["network"],
		}
		events = append(events, event)
	}
	eventCh <- events
	close(eventCh)
}
