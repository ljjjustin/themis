package monitor

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"themis/config"
)

type ThemisAgent struct {
	config     *config.ThemisConfig
	context    context.Context
	cancelFunc context.CancelFunc
}

func NewThemisAgent(config *config.ThemisConfig) *ThemisAgent {
	// fast fail if we can not connect to all collectors.
	for _, monitor := range config.Monitors {
		ip := strings.Split(monitor.Address, ":")[0]
		if !hasBindAddress(ip) {
			plog.Fatalf("no interface with ip %s", ip)
		}
	}
	context, cancel := context.WithCancel(context.Background())

	return &ThemisAgent{
		config:     config,
		context:    context,
		cancelFunc: cancel,
	}
}

func isSerfAgent(pid int) bool {
	path := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdline, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	command := string(cmdline)
	cmd := strings.Replace(command, "\x00", " ", -1)
	if strings.Contains(cmd, "serf agent") {
		return true
	} else {
		return false
	}
}

func (agent *ThemisAgent) Start() {
	for _, monitor := range agent.config.Monitors {
		pid := getPidByAddress(monitor.Address)
		// fast failure if RPC address is already used by other program.
		if pid != 0 && !isSerfAgent(pid) {
			plog.Fatalf("%s is already used by other program.", monitor.Address)
		}
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	serfCtx, _ := context.WithCancel(agent.context)
	for tag, monitor := range agent.config.Monitors {
		go keepRunning(serfCtx, tag, monitor.Address)
	}

	// handler os signals
	select {
	case s := <-signals:
		plog.Infof("Received system signal %s", s)
		// stop monitoring routines
		agent.Stop()
	}
}

func keepRunning(ctx context.Context, tag, address string) {
	quit := make(chan struct{})

	for {
		// check if monitor already started.
		pid := getPidByAddress(address)

		if pid != 0 && isSerfAgent(pid) {
			plog.Info("serf agent already listening: ", address)
		} else {
			// start and monitor serf agent.
			iface := getInterfaceByIP(strings.Split(address, ":")[0])
			args := []string{
				fmt.Sprintf("systemd-run serf agent"),
				fmt.Sprintf("-iface=%s -discover=serf.%s", iface, tag),
				fmt.Sprintf("-rpc-addr=%s -tag network=%s", address, tag),
			}
			serfCmd := strings.Join(args, " ")

			plog.Info("start serf with: ", serfCmd)
			cmd := exec.Command("bash", "-c", serfCmd)
			if err := cmd.Run(); err != nil {
				plog.Warning("Can't start serf agent: ", err)
				break
			}
			time.Sleep(time.Second)
			pid = getPidByAddress(address)
		}
		plog.Infof("serf agent pid is: %d", pid)
		go monitorPid(pid, quit)

		select {
		case <-ctx.Done():
			return
		case <-quit:
			plog.Infof("serf agent %d exited.", pid)
		}
	}
}

func monitorPid(pid int, quit chan<- struct{}) {
	for {
		path := fmt.Sprintf("/proc/%d/cmdline", pid)
		_, err := ioutil.ReadFile(path)
		if err != nil {
			break
		}
		time.Sleep(time.Second)
	}
	quit <- struct{}{}
}

func (agent *ThemisAgent) Stop() {
	agent.cancelFunc()
}
