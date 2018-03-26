package monitor

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ljjjustin/themis/config"
)

type ThemisAgent struct {
	config  *config.ThemisConfig
	context context.Context
}

func NewThemisAgent(config *config.ThemisConfig) *ThemisAgent {
	// fast fail if we can not connect to all collectors.
	for _, monitor := range config.Monitors {
		ip := strings.Split(monitor.Address, ":")[0]
		checkBindAddress(ip)
	}
	context := context.Background()

	return &ThemisAgent{config: config, context: context}
}

func getInterfaceAddrs() map[string][]string {

	ifaceAddrsMap := map[string][]string{}

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		ips := []string{}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				log.Fatal(err)
			}
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
		if len(ips) > 0 {
			ifaceAddrsMap[iface.Name] = ips
		}
	}
	return ifaceAddrsMap
}

func checkBindAddress(ip string) {
	ifaceAddrs := getInterfaceAddrs()

	for _, addrs := range ifaceAddrs {
		for _, addr := range addrs {
			if ip == addr {
				return
			}
		}
	}
	plog.Fatalf("no interface with ip %s", ip)
}

func getInterfaceByIP(ip string) string {
	ifaceAddrs := getInterfaceAddrs()

	for iface, addrs := range ifaceAddrs {
		for _, addr := range addrs {
			if ip == addr {
				return iface
			}
		}
	}
	plog.Fatalf("no interface with ip %s", ip)
	return ""
}

func getPidByAddress(address string) int {
	cmd := fmt.Sprintf("netstat -ntlp | grep -w '%s' | awk '{print $NF}'", address)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil || len(string(out)) < 1 {
		return 0
	}
	stdout := string(out)
	result := strings.Split(stdout, "/")[0]
	pid, err := strconv.Atoi(result)
	if err != nil {
		plog.Fatal("Get pid failed: %s", err.Error())
		return 0
	} else {
		return pid
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

	serfCtx, serfCancel := context.WithCancel(agent.context)
	for tag, monitor := range agent.config.Monitors {
		go keepRunning(serfCtx, tag, monitor.Address)
	}

	// handler os signals
	select {
	case s := <-signals:
		plog.Infof("Received system signal %s", s)
		// stop monitoring routines
		serfCancel()
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
		time.Sleep(1 * time.Second)
	}
	quit <- struct{}{}
}

func (agent *ThemisAgent) Stop() {
}
