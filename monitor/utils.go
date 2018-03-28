package monitor

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

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

func hasBindAddress(ip string) bool {
	ifaceAddrs := getInterfaceAddrs()

	for _, addrs := range ifaceAddrs {
		for _, addr := range addrs {
			if ip == addr {
				return true
			}
		}
	}
	return false
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
