package common

import (
	"github.com/ws-slink/disco/common/api"
	"github.com/ws-slink/disco/common/util/logger"
	"time"
)

type client struct {
	ClientId_  string          `json:"client_id"`
	ServiceId_ string          `json:"service_id"`
	Endpoints_ []api.Endpoint  `json:"endpoints,omitempty"`
	Meta_      map[string]any  `json:"meta,omitempty"`
	LastSeen_  time.Time       `json:"-"`
	State_     api.ClientState `json:"state"`
}

func NewClient(clientId, serviceId string, endpoints []string, meta map[string]any) (api.Client, error) {
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
	}, nil
}

func (c *client) ClientId() string {
	return c.ClientId_
}
func (c *client) ServiceId() string {
	return c.ServiceId_
}
func (c *client) Endpoints() []api.Endpoint {
	return c.Endpoints_
}
func (c *client) Meta() map[string]any {
	return c.Meta_
}
func (c *client) Ping() {
	c.LastSeen_ = time.Now()
	if c.State() != api.ClientStateUp {
		c.SetState(api.ClientStateUp)
		logger.Info("client %s up", c.ClientId())
	}
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
