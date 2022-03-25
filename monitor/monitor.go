package monitor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/pkg/capnslog"
	"themis/api"
	"themis/config"
	"themis/database"
)

const (
	defaultElectionProclaimInterval      = 6 * time.Second
	defaultEventCollectionInterval       = 6 * time.Second
	defaultEventCollectorMonitorInterval = 10 * time.Second
)

var plog = capnslog.NewPackageLogger("themis", "monitor")

type ThemisMonitor struct {
	config          *config.ThemisConfig
	context         context.Context
	cancelFunc      context.CancelFunc
	waitGroup       sync.WaitGroup
	election        *Election
	policyEngine    *PolicyEngine
	eventCollectors []*EventCollector
}

func NewThemisMonitor(config *config.ThemisConfig) *ThemisMonitor {
	leaderName, err := os.Hostname()
	if err != nil {
		plog.Fatal(err)
	}

	engine := database.Engine(&config.Database)
	election := NewElection(leaderName, engine)

	policyEngine := NewPolicyEngine(config)

	context, cancel := context.WithCancel(context.Background())

	return &ThemisMonitor{
		config:       config,
		context:      context,
		cancelFunc:   cancel,
		election:     election,
		policyEngine: policyEngine,
	}
}

func (m *ThemisMonitor) Start() {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// start API server
	plog.Info("Starting API server.")
	apiCtx, _ := context.WithCancel(m.context)
	go startAPIServer(apiCtx, m)

	// start monitoring server
	plog.Info("Starting monitoring routines.")
	monitorCtx, _ := context.WithCancel(m.context)
	go startMonitoring(monitorCtx, m)

	// handler os signals
	select {
	case s := <-signals:
		plog.Infof("Received system signal %s", s)
		m.Stop()
	}
}

func startAPIServer(ctx context.Context, m *ThemisMonitor) {
	listenAddrs := fmt.Sprintf("%s:%d",
		m.config.BindHost, m.config.BindPort)

	server := &http.Server{
		Addr:    listenAddrs,
		Handler: api.Router(),
	}

	go func() {
		m.waitGroup.Add(1)
		defer m.waitGroup.Done()

		server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		plog.Info("REST API exiting: ", ctx.Err())
		if err := server.Shutdown(ctx); err != nil {
			plog.Warningf("Shutdown REST API failed: %s", err)
		}
	}
}

func startMonitoring(ctx context.Context, m *ThemisMonitor) {
	m.waitGroup.Add(1)
	defer m.waitGroup.Done()

	leaderName := m.election.LeaderName

	for {
	StartMonitoring:
		// create monitoring top context
		monitorCtx, monitorCancel := context.WithCancel(ctx)

		// create event collectors
		plog.Info("Creating event collectors.")
		m.eventCollectors = make([]*EventCollector, 0)
		for tag, monitor := range m.config.Monitors {
			ip := strings.Split(monitor.Address, ":")[0]
			if !hasBindAddress(ip) {
				plog.Warningf("no interface with ip %s", ip)
				time.Sleep(defaultEventCollectorMonitorInterval)
				goto StartMonitoring
			}
			collector := NewEventCollector(tag, &monitor)
			m.eventCollectors = append(m.eventCollectors, collector)
		}
		// monitor event collectors
		plog.Info("Starting event collector monitors.")
		IPMonitorCtx, _ := context.WithCancel(monitorCtx)
		IPMonitorErr := startIPMonitor(IPMonitorCtx, m)

		plog.Infof("%s start campaign leader.", leaderName)
		electionCtx, _ := context.WithCancel(monitorCtx)
		electionErr := startCampaign(electionCtx, m)

		// start policy engine who will handle events and make decision
		plog.Info("Starting policy engine.")
		policyEngineCtx, _ := context.WithCancel(monitorCtx)
		policyEngineErr := startPolicyEngine(policyEngineCtx, m)

		for {
			select {
			case err := <-IPMonitorErr:
				plog.Info(err)
				monitorCancel()
				goto StartMonitoring
			case err := <-electionErr:
				plog.Info(err)
				monitorCancel()
				goto StartMonitoring
			case err := <-policyEngineErr:
				plog.Info(err)
				monitorCancel()
				goto StartMonitoring
			case <-ctx.Done():
				// stop all monitoring routines
				plog.Info("Monitoring exiting: ", ctx.Err())
				monitorCancel()
				return
			}
		}
	}
}

func startIPMonitor(ctx context.Context, m *ThemisMonitor) <-chan error {
	quit := make(chan error, 1)

	go func() {
		m.waitGroup.Add(1)
		defer m.waitGroup.Done()
		for {
			select {
			case <-ctx.Done():
				plog.Info("IP monitor exiting: ", ctx.Err())
				return
			case <-time.After(defaultEventCollectorMonitorInterval):
			}

			for _, monitor := range m.config.Monitors {
				ip := strings.Split(monitor.Address, ":")[0]
				if !hasBindAddress(ip) {
					msg := fmt.Sprintf("no interface with ip %s", ip)
					quit <- errors.New(msg)
				}
			}
		}
	}()

	return quit
}

func startCampaign(ctx context.Context, m *ThemisMonitor) <-chan error {
	quit := make(chan error, 1)

	leaderName := m.election.LeaderName
	electionErr := m.election.Campaign(ctx)
	// wait until we become a leader or a error occour
	select {
	case err := <-electionErr:
		plog.Warning(err)
		quit <- err
		return quit
	default:
		plog.Infof("%s became leader.", leaderName)
	}

	go func() {
		m.waitGroup.Add(1)
		defer m.waitGroup.Done()

		defer m.election.Quit()
		for {
			select {
			case <-ctx.Done():
				plog.Info("Proclaim exiting: ", ctx.Err())
				return
			case <-time.After(defaultElectionProclaimInterval):
			}
			plog.Debugf("%s updating term.", leaderName)
			succ, err := m.election.Proclaim()
			if err != nil {
				plog.Info("update term failed: ", err)
				quit <- err
				return
			} else if !succ {
				plog.Infof("%s proclaim failed, we are not leader now.", leaderName)
				quit <- errors.New("Leader changed.")
				return
			}
		}
	}()

	return quit
}

func startPolicyEngine(ctx context.Context, m *ThemisMonitor) <-chan error {
	quit := make(chan error, 1)

	go func() {
		m.waitGroup.Add(1)
		defer m.waitGroup.Done()

		for {
			// check if we should quit
			select {
			case <-ctx.Done():
				plog.Info("policy engine exiting: ", ctx.Err())
				return
			default:
			}

			for _, collector := range m.eventCollectors {
				collector.Start()
			}

			time.Sleep(defaultEventCollectionInterval)

			allEvents := make(Events, 0)
			for _, collector := range m.eventCollectors {
				events, err := collector.DrainEvents()
				if err != nil {
					quit <- errors.New(err.Error())
				}
				if events != nil {
					allEvents = append(allEvents, events...)
				}
			}
			m.policyEngine.HandleEvents(allEvents)
		}
	}()

	return quit
}

func (m *ThemisMonitor) Stop() {
	m.cancelFunc()
	m.waitGroup.Wait()
}
