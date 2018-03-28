package monitor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/pkg/capnslog"
	"github.com/ljjjustin/themis/api"
	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "monitor")

type ThemisMonitor struct {
	config          *config.ThemisConfig
	context         context.Context
	cancelFunc      context.CancelFunc
	election        *Election
	policyEngine    *PolicyEngine
	eventCollectors []*EventCollector
}

func NewThemisMonitor(config *config.ThemisConfig) *ThemisMonitor {
	context := context.Background()

	leaderName, err := os.Hostname()
	if err != nil {
		plog.Fatal(err)
	}

	engine := database.Engine(&config.Database)
	election := NewElection(leaderName, engine)

	policyEngine := NewPolicyEngine(config)

	return &ThemisMonitor{
		config:       config,
		context:      context,
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
		if err := server.ListenAndServe(); err != nil {
			plog.Fatalf("Failed to start REST API: %s", err)
		}
	}()

	select {
	case <-ctx.Done():
		if err := server.Shutdown(ctx); err != nil {
			plog.Warningf("Failed to shutdown REST API: %s", err)
		}
		plog.Info("REST API exiting.")
	}
}

func startMonitoring(ctx context.Context, m *ThemisMonitor) {
	leaderName := m.election.LeaderName

	for {
	StartMonitoring:
		m.eventCollectors = make([]*EventCollector, 0)
		for tag, monitor := range m.config.Monitors {
			ip := strings.Split(monitor.Address, ":")[0]
			if !hasBindAddress(ip) {
				plog.Warningf("no interface with ip %s", ip)
				time.Sleep(20 * time.Second)
				goto StartMonitoring
			}
			collector := NewEventCollector(tag, &monitor)
			m.eventCollectors = append(m.eventCollectors, collector)
		}

		monitorCtx, monitorCancel := context.WithCancel(ctx)
		plog.Infof("%s start campaign leader.", leaderName)
		// wait until we become a leader or a error occour
		electionCtx, _ := context.WithCancel(monitorCtx)
		electionErr := m.election.Campaign(electionCtx)
		select {
		case err := <-electionErr:
			plog.Warning(err)
			monitorCancel()
			goto StartMonitoring
		default:
			plog.Infof("%s became leader.", leaderName)
		}

		// start policy engine who will handle events and make decision
		policyEngineCtx, _ := context.WithCancel(monitorCtx)
		policyEngineErr := startPolicyEngine(policyEngineCtx, m)

		for {
			select {
			case err := <-electionErr:
				// perform monitoring if we are still leader
				plog.Info(err)
				monitorCancel()
				goto StartMonitoring
			case err := <-policyEngineErr:
				plog.Info(err)
				monitorCancel()
				goto StartMonitoring
			case <-ctx.Done():
				// stop all monitoring routines
				monitorCancel()
				return
			}
		}
	}
}

func startPolicyEngine(ctx context.Context, m *ThemisMonitor) <-chan error {
	quit := make(chan error, 1)
	plog.Info("Starting policy engine.")

	go func() {
		for {
			for _, collector := range m.eventCollectors {
				collector.Start()
			}

			time.Sleep(5 * time.Second)

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
}
