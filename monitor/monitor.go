package monitor

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coreos/pkg/capnslog"
	"github.com/go-xorm/xorm"
	"github.com/ljjjustin/themis/config"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "monitor")

type ThemisMonitor struct {
	cfg *config.ThemisConfig
}

func NewThemisMonitor(cfg *config.ThemisConfig) *ThemisMonitor {
	return &ThemisMonitor{cfg: cfg}
}

func (m *ThemisMonitor) DbSync() {
	sc := m.cfg.Storage
	engine, err := xorm.NewEngine("mysql", sc.Connection)
	if err != nil {
		plog.Fatal(err)
	}
	defer engine.Close()

	err = engine.Sync2(new(ElectionRecord))
	if err != nil {
		log.Fatal(err)
	}
}

func (m *ThemisMonitor) Start() {
	sc := m.cfg.Storage
	engine, err := xorm.NewEngine("mysql", sc.Connection)
	if err != nil {
		plog.Fatal(err)
	}
	defer engine.Close()

	leader_name, err := os.Hostname()
	if err != nil {
		plog.Fatal(err)
	}

	election := NewElection(leader_name, engine)
	ctx, cancel := context.WithCancel(context.Background())

	for {
		// wait until we become a leader or a error occour
		if err := election.Campaign(ctx); err != nil {
			fmt.Println(err)
			cancel()
			break
		}

		// start API server
		//plog.Info("Starting API server.")
		//apiCtx, apiCancel := context.WithCancel(ctx)

		// do monitoring and fence operation if any hypervisor go down
		for {
			plog.Info("Starting monitoring routines.")

			// perform monitoring if we are still leader
			isLeader, err := election.IsLeader(ctx)
			if err != nil {
				cancel()
				plog.Fatal(err)
			}
			if !isLeader {
				break
			}

			var collectors []*EventCollector
			for tag, cfg := range m.cfg.Monitors {
				collector := NewEventCollector(tag, &cfg)
				collectors = append(collectors, collector)
				collector.Start()
			}

			for _, collector := range collectors {
				events, err := collector.DrainEvents()
				if err != nil {
					plog.Noticef("Some error occur: %s", err)
				}
				for _, event := range events {
					plog.Noticef("Host %s with tag %s became Failed.",
						event.Hostname, event.NetworkTag)
				}
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func (m *ThemisMonitor) Stop() {
}