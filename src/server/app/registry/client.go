package registry

import (
	"github.com/ws-slink/disco/common/api"
	"github.com/ws-slink/disco/server/common/util/logger"
	"time"
)

type client struct {
	ClientId_  string          `json:"client_id"`
	ServiceId_ string          `json:"service_id"`
	Endpoints_ []api.Endpoint  `json:"endpoints,omitempty"`
	Meta_      map[string]any  `json:"meta,omitempty"`
	LastSeen_  time.Time       `json:"last_seen,omitempty"`
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
}
