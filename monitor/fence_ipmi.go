package monitor

import (
	ipmi "github.com/vmware/goipmi"

	"themis/database"
)

type IPMIFencer struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewIPMIFencer(fencer *database.HostFencer) *IPMIFencer {
	return &IPMIFencer{
		Host:     fencer.Host,
		Port:     fencer.Port,
		Username: fencer.Username,
		Password: fencer.Password,
	}
}

func (f *IPMIFencer) Fence() error {
	client, err := ipmi.NewClient(&ipmi.Connection{
		Hostname:  f.Host,
		Port:      f.Port,
		Username:  f.Username,
		Password:  f.Password,
		Interface: "lanplus",
	})
	if err != nil {
		plog.Warning("IPMI fencer failed to create client: ", err)
		return err
	}

	err = client.Open()
	if err != nil {
		plog.Warning("IPMI fencer failed to connect BMC server: ", err)
		return err
	}
	defer client.Close()

	err = client.Control(ipmi.ControlPowerDown)
	if err != nil {
		plog.Warning("IPMI fencer failed to set power down: ", err)
		return err
	}
	return nil
}
