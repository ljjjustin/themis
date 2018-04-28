package client

import (
	"fmt"
	"strings"
	"time"
)

type ThemisClient struct {
	// BaseUrl specify themis server's ip and port.
	BaseUrl string

	// HTTPClient allows users to interject arbitrary http, https, or other transit behaviors.
	http HTTPClient
}

func NewThemisClient(url string) *ThemisClient {
	return &ThemisClient{
		BaseUrl: strings.TrimSuffix(url, "/"),
	}
}

type Host struct {
	// ID uniquely identifies this host amongst all other hosts.
	ID int `json:"id"`

	// Name contains the human-readable name for the host.
	Name string `json:"name"`

	// Status contains the current status of the host.
	Status string `json:"status"`

	// Disabled contains information about whether the host is disabled.
	Disabled bool `json:"disabled"`

	// UpdatedAt contains timestamps of when the state of the host last changed.
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *ThemisClient) ListHosts() ([]Host, error) {
	var hosts []Host

	url := fmt.Sprintf("%s/hosts", c.BaseUrl)
	result := c.http.Get(url, nil)
	err := result.ExtractIntoSlicePtr(&hosts, "")
	return hosts, err
}

func (c *ThemisClient) ShowHost(id int) (Host, error) {
	var host Host

	url := fmt.Sprintf("%s/hosts/%d", c.BaseUrl, id)
	result := c.http.Get(url, nil)
	err := result.ExtractInto(&host)

	return host, err
}

func (c *ThemisClient) AddHost(h *Host) (Host, error) {
	var host Host

	url := fmt.Sprintf("%s/hosts", c.BaseUrl)
	result := c.http.Post(url, h, nil)
	err := result.ExtractInto(&host)

	return host, err
}

func (c *ThemisClient) DeleteHost(id int) error {

	url := fmt.Sprintf("%s/hosts/%d", c.BaseUrl, id)
	return c.http.Delete(url, nil)
}

func (c *ThemisClient) EnableHost(id int) (Host, error) {
	var host Host

	url := fmt.Sprintf("%s/hosts/%d/enable", c.BaseUrl, id)
	result := c.http.Post(url, nil, nil)
	err := result.ExtractInto(&host)

	return host, err
}

func (c *ThemisClient) DisableHost(id int) (Host, error) {
	var host Host

	url := fmt.Sprintf("%s/hosts/%d/disable", c.BaseUrl, id)
	result := c.http.Post(url, nil, nil)
	err := result.ExtractInto(&host)

	return host, err
}

type Fencer struct {
	// ID uniquely identifies this fencer amongst all other fencers.
	ID int `json:"id"`

	// HostId uniquely identifies host ID associated with this fencer.
	HostId int `json:"host_id"`

	// Type identifies fencer type, such as "IPMI".
	Type string `json:"type"`

	// Remote host name for IPMI LAN interface
	Host string `json:"host"`

	// Remote RMCP port
	Port int `json:"port"`

	// Remote session username
	Username string `json:"username"`

	// Remote session password
	Password string `json:"password"`
}

func (c *ThemisClient) ListFencers() ([]Fencer, error) {
	var fencers []Fencer

	url := fmt.Sprintf("%s/fencers", c.BaseUrl)
	result := c.http.Get(url, nil)
	err := result.ExtractIntoSlicePtr(&fencers, "")
	return fencers, err
}

func (c *ThemisClient) ShowFencer(id int) (Fencer, error) {
	var fencer Fencer

	url := fmt.Sprintf("%s/fencers/%d", c.BaseUrl, id)
	result := c.http.Get(url, nil)
	err := result.ExtractInto(&fencer)

	return fencer, err
}

func (c *ThemisClient) AddFencer(f *Fencer) (Fencer, error) {
	var fencer Fencer

	url := fmt.Sprintf("%s/fencers", c.BaseUrl)
	result := c.http.Post(url, f, nil)
	err := result.ExtractInto(&fencer)

	return fencer, err
}

func (c *ThemisClient) DeleteFencer(id int) error {

	url := fmt.Sprintf("%s/fencers/%d", c.BaseUrl, id)
	return c.http.Delete(url, nil)
}

type ElectionRecord struct {
	Id           uint32    `json:"id"`
	ElectionName string    `json:"election_name"`
	LeaderName   string    `json:"leader_name"`
	LastUpdate   time.Time `json:"last_update"`
}

func (c *ThemisClient) GetLeader() (ElectionRecord, error) {

	var leader ElectionRecord

	url := fmt.Sprintf("%s/leader", c.BaseUrl)

	result := c.http.Get(url, nil)
	err := result.ExtractInto(&leader)

	return leader, err
}