package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/slink-go/disco/backend/common"
	"github.com/slink-go/disco/common/api"
	"github.com/slink-go/disco/server/config"
	"github.com/slink-go/logger"
	"reflect"
	"sync"
	"time"
)

var Backend inMemBackendInitializer

type inMemBackendInitializer struct{}

func (bi *inMemBackendInitializer) Init(cfg *config.AppConfig) api.Registry {
	return newInMemRegistry(cfg)
}

type inMemRegistry struct {
	tenants      map[string]*common.Tenant
	clients      map[string]api.Client
	pingInterval api.Duration
	maxClients   int
	mutex        sync.RWMutex
}

func newInMemRegistry(cfg *config.AppConfig) api.Registry {
	registry := inMemRegistry{
		tenants:      map[string]*common.Tenant{},
		clients:      map[string]api.Client{},
		maxClients:   cfg.MaxClients,
		pingInterval: api.Duration{Duration: cfg.PingDuration},
	}
	registry.run(cfg)
	return &registry
}

func (rs *inMemRegistry) Join(ctx context.Context, request api.JoinRequest) (*api.JoinResponse, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	logger.Debug("[registry][join] client join")

	if len(rs.clients) >= rs.maxClients {
		return nil, api.NewMaxClientsReachedError(rs.maxClients)
	}

	tenant := ctx.Value(api.TenantKey).(string)

	clientId := rs.createClientId()
	c, err := common.NewClient(clientId, request.ServiceId, tenant, request.Endpoints, request.Meta)
	if err != nil {
		return nil, err
	}

	if rs.has(c) {
		return nil, api.NewAlreadyRegisteredError()
	}
	rs.clients[clientId] = c
	if rs.tenants[tenant] == nil {
		rs.tenants[tenant] = &common.Tenant{
			Name:    tenant,
			Clients: make(map[string]api.Client),
		}
	}
	rs.tenants[tenant].Clients[clientId] = c
	rs.update(c)
	logger.Debug("[registry][join] client %s joined", c.ClientId())
	return &api.JoinResponse{
		ClientId:     clientId,
		PingInterval: rs.pingInterval,
	}, nil
}
func (rs *inMemRegistry) Leave(ctx context.Context, clientId string) error {
	client, ok := rs.clients[clientId]
	if !ok {
		return api.NewClientNotFoundError(clientId)
	}
	logger.Debug("[registry][leave] remove client %s", clientId)
	rs.remove(client)
	return nil
	//rs.mutex.Lock()
	//defer rs.mutex.Unlock()
	//logger.Debug("[registry][leave] client %s leave", clientId)
	//_, ok := rs.clients[clientId]
	//if !ok {
	//	return api.NewClientNotFoundError(clientId)
	//}
	//logger.Debug("[registry][leave] remove client %s", clientId)
	//delete(rs.clients, clientId)
	//tenant := ctx.Value(api.TenantKey).(string)
	//if tenant != "" {
	//	_, ok = rs.tenants[tenant]
	//	if !ok {
	//		return api.NewTenantNotFoundError(tenant)
	//	}
	//	_, ok = rs.tenants[tenant].Clients[clientId]
	//	if !ok {
	//		return api.NewTenantsClientNotFoundError(clientId)
	//	}
	//	delete(rs.tenants[tenant].Clients, clientId)
	//}
	//logger.Debug("[registry][leave] client %s left", clientId)
	//return nil
}
func (rs *inMemRegistry) List(ctx context.Context) []api.Client {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	result := []api.Client{}
	var clients map[string]api.Client
	tenant := ctx.Value(api.TenantKey).(string)
	logger.Debug("[registry][list] list for %v", tenant)
	if tenant == api.TenantDefault || tenant == "" {
		logger.Debug("[list] list all")
		clients = rs.clients
	} else {
		logger.Debug("[list] list for tenant = %s", tenant)
		tnt, ok := rs.tenants[tenant]
		if ok {
			clients = tnt.Clients
		} else {
			clients = make(map[string]api.Client)
		}
	}
	for _, t := range clients {
		result = append(result, t)
	}
	logger.Debug("[registry][list] list for %v (%d)", tenant, len(result))
	return result
}
func (rs *inMemRegistry) Ping(clientId string) (api.Pong, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	v, ok := rs.clients[clientId]
	if !ok {
		return api.Pong{}, api.NewClientNotFoundError(clientId)
	}
	if v.Ping() {
		rs.update(v)
	}
	response := api.PongTypeOk
	if v.IsDirty() {
		v.SetDirty(false)
		response = api.PongTypeChanged
	}
	logger.Debug("[registry][ping] client '%s' ping: '%s'", clientId, response)
	return api.Pong{
		Response: response,
	}, nil
}

func (rs *inMemRegistry) createClientId() string {
	u, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return u.String()
}
func (rs *inMemRegistry) has(client api.Client) bool {
	t, ok := rs.tenants[client.Tenant()]
	if ok {
		for _, c := range t.Clients {
			if rs.equalClients(c, client) {
				return true
			}
		}
	}
	return false
}
func (rs *inMemRegistry) equalClients(a, b api.Client) bool {
	return a.ServiceId() == b.ServiceId() &&
		reflect.DeepEqual(a.Endpoints(), b.Endpoints()) &&
		reflect.DeepEqual(a.Meta(), b.Meta())
}

func (rs *inMemRegistry) run(cfg *config.AppConfig) {
	// TODO: one runner per tenant
	go func() {
		for {
			time.Sleep(time.Second)
			for _, t := range rs.tenants {
				go rs.runner(cfg, t)
			}
		}
	}()
}
func (rs *inMemRegistry) runner(cfg *config.AppConfig, tenant *common.Tenant) {
	for _, c := range tenant.Clients {
		interval := time.Now().Sub(c.LastSeen())
		if time.Duration(cfg.RemoveThreshold)*cfg.PingDuration < interval {
			if c.State() != api.ClientStateRemoved {
				rs.remove(c)
			}
		} else if time.Duration(cfg.DownThreshold)*cfg.PingDuration < interval {
			rs.down(c)
		} else if time.Duration(cfg.FailingThreshold)*cfg.PingDuration < interval {
			rs.failing(c)
		} else {
			// skip
		}
	}
}
func (rs *inMemRegistry) failing(client api.Client) {
	if client.State() != api.ClientStateFailing {
		client.SetState(api.ClientStateFailing)
		rs.update(client)
		logger.Info("client %s failing", client.ClientId())
	}
}
func (rs *inMemRegistry) down(client api.Client) {
	if client.State() != api.ClientStateDown {
		client.SetState(api.ClientStateDown)
		rs.update(client)
		logger.Info("client %s down", client.ClientId())
	}
}
func (rs *inMemRegistry) remove(client api.Client) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	logger.Info("removing client %s", client.ClientId())
	defer rs.update(client)
	delete(rs.clients, client.ClientId())
	for _, t := range rs.tenants {
		_, ok := t.Clients[client.ClientId()]
		if ok {
			delete(t.Clients, client.ClientId())
			return
		}
	}

}
func (rs *inMemRegistry) update(client api.Client) {
	t, ok := rs.tenants[client.Tenant()]
	if !ok {
		return
	}
	for _, c := range t.Clients {
		logger.Notice("set dirty %s", c.ClientId())
		c.SetDirty(true)
	}
}
