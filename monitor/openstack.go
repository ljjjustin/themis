package monitor

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/evacuate"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/services"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"

	"github.com/ljjjustin/themis/config"
)

type NovaClient struct {
	client *gophercloud.ServiceClient
}

func NewNovaClient(cfg *config.OpenstackConfig) (*NovaClient, error) {
	authOptions := gophercloud.AuthOptions{
		IdentityEndpoint: cfg.AuthURL,
		Username:         cfg.Username,
		Password:         cfg.Password,
		TenantName:       cfg.ProjectName,
		DomainName:       cfg.DomainName,
	}

	provider, err := openstack.AuthenticatedClient(authOptions)
	if err != nil {
		plog.Warning("Unable to authenticat with openstack", err)
		return nil, err
	}

	endpointOptions := gophercloud.EndpointOpts{}
	if len(cfg.RegionName) > 0 {
		endpointOptions = gophercloud.EndpointOpts{Region: cfg.RegionName}
	}

	client, err := openstack.NewComputeV2(provider, endpointOptions)
	if err != nil {
		plog.Warning("Unable to create nova client", err)
		return nil, err
	}

	return &NovaClient{client: client}, nil
}

func (nova *NovaClient) ListServices() ([]services.Service, error) {
	pages, err := services.List(nova.client).AllPages()
	if err != nil {
		plog.Warning("Can't list services", err)
		return nil, err
	}
	return services.ExtractServices(pages)
}

type ServiceUpdateOpts struct {
	// The name of the host.
	Host string `json:"host"`

	// The binary name of the service.
	Binary string `json:"binary"`

	// The reason for disabling a service.
	DisabledReason string `json:"disabled_reason,omitempty"`

	// Force down the service.
	ForcedDown bool `json:"forced_down,omitempty"`
}

func (nova *NovaClient) ForceDownService(s services.Service) (r gophercloud.ErrResult) {
	url := nova.client.ServiceURL("os-services", "force-down")
	updateOpts := ServiceUpdateOpts{
		Host:       s.Host,
		Binary:     s.Binary,
		ForcedDown: true,
	}
	reqBody, err := gophercloud.BuildRequestBody(updateOpts, "")
	if err != nil {
		plog.Warning("Build request body failed", err)
		return
	}
	requestOpts := &gophercloud.RequestOpts{
		MoreHeaders: map[string]string{
			"X-OpenStack-Nova-API-Version": "2.37",
		},
	}
	_, r.Err = nova.client.Put(url, reqBody, nil, requestOpts)
	return
}

func (nova *NovaClient) DisableService(s services.Service, reason string) (r gophercloud.ErrResult) {
	var url string
	if len(reason) > 0 {
		url = nova.client.ServiceURL("os-services", "disable-log-reason")
	} else {
		url = nova.client.ServiceURL("os-services", "disable")
	}
	updateOpts := ServiceUpdateOpts{
		Host:           s.Host,
		Binary:         s.Binary,
		DisabledReason: reason,
	}
	reqBody, err := gophercloud.BuildRequestBody(updateOpts, "")
	if err != nil {
		plog.Warning("Build request body failed", err)
		return
	}

	_, r.Err = nova.client.Put(url, reqBody, nil, nil)
	return
}

func (nova *NovaClient) ListServers(hostname string) ([]servers.Server, error) {
	listOpts := servers.ListOpts{Host: hostname, AllTenants: true}
	pages, err := servers.List(nova.client, listOpts).AllPages()
	if err != nil {
		return nil, err
	}
	return servers.ExtractServers(pages)
}

func (nova *NovaClient) Evacuate(id string) {
	evacuateOpts := evacuate.EvacuateOpts{
		OnSharedStorage: true,
	}
	result := evacuate.Evacuate(nova.client, id, evacuateOpts)
	if result.Err != nil {
		plog.Warning("Execute evacuate failed", result.Err)
	} else {
		plog.Infof("evacuate instance: %s successfully.", id)
	}
}
