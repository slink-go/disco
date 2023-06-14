package registry

import (
	"context"
	"github.com/google/uuid"
	"github.com/ws-slink/disco/common/api"
	"github.com/ws-slink/disco/server/common/util/logger"
	"reflect"
	"sync"
)

type inMemRegistry struct {
	tenants map[string]*Tenant
	clients map[string]api.Client
	mutex   sync.RWMutex
}

func NewInMemRegistry() api.Registry {
	return &inMemRegistry{
		tenants: map[string]*Tenant{},
		clients: map[string]api.Client{},
	}
}

func (rs *inMemRegistry) Join(ctx context.Context, request api.JoinRequest) (*api.JoinResponse, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	clientId := rs.createClientId()
	c, err := NewClient(clientId, request.ServiceId, request.Endpoints, request.Meta)
	if err != nil {
		return nil, err
	}
	if rs.has(c) {
		return nil, api.NewAlreadyRegisteredError()
	}
	rs.clients[clientId] = c
	if ctx.Value(api.TenantKey) != nil {
		tenant := ctx.Value(api.TenantKey).(string)
		if tenant != "" {
			if rs.tenants[tenant] == nil {
				rs.tenants[tenant] = &Tenant{
					Name:    tenant,
					Clients: make(map[string]api.Client),
				}
			}
			rs.tenants[tenant].Clients[clientId] = c
		}
	}
	return &api.JoinResponse{
		ClientId:     clientId,
		PingInterval: api.DefaultPingInterval,
	}, nil
}
func (rs *inMemRegistry) Leave(ctx context.Context, clientId string) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	_, ok := rs.clients[clientId]
	if !ok {
		return api.NewClientNotFoundError(clientId)
	}
	logger.Debug("remove client %s", clientId)
	delete(rs.clients, clientId)
	tenant := ""
	if ctx.Value(api.TenantKey) != nil {
		tenant = ctx.Value(api.TenantKey).(string)
	}
	if tenant != "" {
		_, ok = rs.tenants[tenant]
		if !ok {
			return api.NewTenantNotFoundError(tenant)
		}
		_, ok = rs.tenants[tenant].Clients[clientId]
		if !ok {
			return api.NewTenantsClientNotFoundError(clientId)
		}
		delete(rs.tenants[tenant].Clients, clientId)
	}
	return nil
}
func (rs *inMemRegistry) List(ctx context.Context) []api.Client {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	var result []api.Client
	clients := rs.clients
	if ctx.Value(api.TenantKey) != nil {
		tenant := ctx.Value(api.TenantKey).(string)
		if tenant != "" {
			clients = rs.tenants[tenant].Clients
		}
	}
	for _, t := range clients { // iterate over map (serviceId:client)
		result = append(result, t)
	}
	return result
}
func (rs *inMemRegistry) Ping(clientId string) (api.Pong, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	v, ok := rs.clients[clientId]
	if !ok {
		return api.Pong{}, api.NewClientNotFoundError(clientId)
	}
	v.Ping()
	// if v.needsUpdade return PongTypeUpdated
	return api.Pong{}, nil
}

func (rs *inMemRegistry) createClientId() string {
	u, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return u.String()
}
func (rs *inMemRegistry) has(client api.Client) bool {
	for _, c := range rs.clients {
		if rs.equalClients(c, client) {
			return true
		}
	}
	return false
}
func (rs *inMemRegistry) equalClients(a, b api.Client) bool {
	return a.ServiceId() == b.ServiceId() &&
		reflect.DeepEqual(a.Endpoints(), b.Endpoints()) &&
		reflect.DeepEqual(a.Meta(), b.Meta())
}
