package common

import (
	"github.com/slink-go/disco/common/api"
	"github.com/slink-go/logger"
	"time"
)

type client struct {
	ClientId_  string          `json:"client_id"`
	ServiceId_ string          `json:"service_id"`
	Tenant_    string          `json:"tenant,omitempty"`
	Endpoints_ []api.Endpoint  `json:"endpoints,omitempty"`
	Meta_      map[string]any  `json:"meta,omitempty"`
	LastSeen_  time.Time       `json:"-"`
	State_     api.ClientState `json:"state"`
	Dirty_     bool            `json:"-"`
}

func NewClient(clientId, serviceId, tenant string, endpoints []string, meta map[string]any) (api.Client, error) {
	var ep []api.Endpoint
	for _, u := range endpoints {
		v, err := api.NewEndpoint(u)
		if err != nil {
			logger.Warning("invalid endpoint %s", err.Error())
			return nil, err
		}
		ep = append(ep, v)
	}
	return &client{
		ClientId_:  clientId,
		ServiceId_: serviceId,
		Endpoints_: ep,
		Meta_:      meta,
		LastSeen_:  time.Now(),
		State_:     api.ClientStateStarting,
		Tenant_:    tenant,
		Dirty_:     true,
	}, nil
}

func (c *client) ClientId() string {
	return c.ClientId_
}
func (c *client) ServiceId() string {
	return c.ServiceId_
}
func (c *client) Tenant() string {
	return c.Tenant_
}
func (c *client) Endpoints() []api.Endpoint {
	return c.Endpoints_
}
func (c *client) Meta() map[string]any {
	return c.Meta_
}
func (c *client) Ping() bool {
	c.LastSeen_ = time.Now()
	if c.State() != api.ClientStateUp {
		c.SetState(api.ClientStateUp)
		logger.Info("client %s (%s) up", c.ClientId(), c.ServiceId())
		return true
	}
	return false
}
func (c *client) LastSeen() time.Time {
	return c.LastSeen_
}
func (c *client) State() api.ClientState {
	return c.State_
}
func (c *client) SetState(state api.ClientState) {
	c.State_ = state
}
func (c *client) SetDirty(value bool) {
	c.Dirty_ = value
}
func (c *client) IsDirty() bool {
	return c.Dirty_
}
