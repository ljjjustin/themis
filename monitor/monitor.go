package monitor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

	policyEngine := NewPolicyEngine()

	var collectors []*EventCollector
	for tag, monitor := range config.Monitors {
		collector := NewEventCollector(tag, &monitor)
		collectors = append(collectors, collector)
	}

	return &ThemisMonitor{
		config:          config,
		context:         context,
		election:        election,
		policyEngine:    policyEngine,
		eventCollectors: collectors,
	}
}

func (m *ThemisMonitor) Start() {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// start API server
	plog.Info("Starting API server.")
	apiCtx, apiCancel := context.WithCancel(m.context)
	go startAPIServer(apiCtx, m)

	// start monitoring server
	plog.Info("Starting monitoring routines.")
	monitorCtx, monitorCancel := context.WithCancel(m.context)
	go startMonitoring(monitorCtx, m)

	// handler os signals
	select {
	case s := <-signals:
		plog.Infof("Received system signal %s", s)
		// stop api server
		apiCancel()
		// stop monitoring routines
		monitorCancel()
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
	monitorCtx, monitorCancel := context.WithCancel(ctx)

	for {
		leaderName := m.election.LeaderName
		plog.Infof("%s start campaign leader.", leaderName)
		// wait until we become a leader or a error occour
		electionCtx, electionCancel := context.WithCancel(monitorCtx)
		electionErr := m.election.Campaign(electionCtx)
		select {
		case err := <-electionErr:
			plog.Warning(err)
			electionCancel()
			break
		default:
			plog.Infof("%s became leader.", leaderName)
		}

		// start policy engine who will handle events and make decision
		//policyEngineErr := startPolicyEngine(ctx, m)

		// event collectors who will collect events and deliver to policy engine
		collectorCtx, collectorCancel := context.WithCancel(monitorCtx)
		collectorErr := startEventCollectors(collectorCtx, m)

		for {
			select {
			case err := <-electionErr:
				// perform monitoring if we are still leader
				plog.Info(err)
				monitorCancel()
				break
			case err := <-collectorErr:
				plog.Info(err)
				collectorCancel()
				break
			case <-ctx.Done():
				// stop policy engine & event collectors
				monitorCancel()
				return
			}
		}
	}
}

func startEventCollectors(ctx context.Context, m *ThemisMonitor) <-chan error {
	quit := make(chan error, 1)

	plog.Info("Starting event collectors.")
	go func() {
		for {
			for _, collector := range m.eventCollectors {
				collector.Start()
			}

			time.Sleep(5 * time.Second)

			for _, collector := range m.eventCollectors {
				events, err := collector.DrainEvents()
				if err != nil {
					quit <- errors.New(err.Error())
				}
				for _, event := range events {
					plog.Infof("Host %s with tag %s became Failed.",
						event.Hostname, event.NetworkTag)
				}
			}
		}
	}()
	return quit
}

func (m *ThemisMonitor) Stop() {
}
