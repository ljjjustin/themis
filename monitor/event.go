package monitor

import (
	"themis/config"
)

type Event struct {
	Hostname   string
	NetworkTag string
	Status     string
}

type Events []*Event

type MonitorInterface interface {
	Start() (chan Events, error)
}

func NewEventMonitor(cfg *config.MonitorConfig) MonitorInterface {

	// TODO: create monitor according to monitor type
	return NewSerfMonitor(cfg.Address)
}

type EventCollector struct {
	Tag       string
	EventChan chan Events
	Monitor   MonitorInterface
}

func NewEventCollector(tag string, cfg *config.MonitorConfig) *EventCollector {
	monitor := NewEventMonitor(cfg)
	return &EventCollector{
		Tag:     tag,
		Monitor: monitor,
	}
}

func (c *EventCollector) Start() error {
	eventCh, err := c.Monitor.Start()
	if err != nil {
		return err
	}
	c.EventChan = eventCh
	return nil
}

func (c *EventCollector) DrainEvents() (Events, error) {
	select {
	case events := <-c.EventChan:
		return events, nil
	default:
		return nil, nil
	}
}
