package fence

import (
	ipmi "github.com/vmware/goipmi"

	"github.com/ljjjustin/themis/database"
	"github.com/coreos/pkg/capnslog"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "fence")

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

	client, err := f.getClient()
	if err != nil {
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

func (f *IPMIFencer) Ping() error {

	client, err := f.getClient()
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.DeviceID()
	if err != nil {
		plog.Warning("IPMI connect failed: ", err)
		return err
	}

	return nil
}

func (f *IPMIFencer) getClient() (*ipmi.Client, error) {

	client, err := ipmi.NewClient(&ipmi.Connection{
		Hostname:  f.Host,
		Port:      f.Port,
		Username:  f.Username,
		Password:  f.Password,
		Interface: "lanplus",
	})
	if err != nil {
		plog.Warning("IPMI fencer failed to create client: ", err)
		return nil, err
	}

	err = client.Open()
	if err != nil {
		plog.Warning("IPMI fencer failed to connect BMC server: ", err)
		return nil, err
	}

	return client, nil
}