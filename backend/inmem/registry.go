package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/slink-go/disco/backend/common"
	"github.com/slink-go/disco/backend/inmem/store"
	"github.com/slink-go/disco/common/api"
	"github.com/slink-go/disco/server/config"
	"github.com/slink-go/logger"
	"reflect"
	"sort"
	"sync"
	"time"
)

var Backend inMemBackendInitializer

type inMemBackendInitializer struct{}

func (bi *inMemBackendInitializer) Init(cfg *config.AppConfig) api.Registry {
	return newInMemRegistry(cfg)
}

type inMemRegistry struct {
	sync.RWMutex
	tenants      *store.TenantsSync
	clients      *store.ClientsSync
	pingInterval api.Duration
	maxClients   int
}

func newInMemRegistry(cfg *config.AppConfig) api.Registry {
	registry := inMemRegistry{
		tenants:      store.CreateTenants(),
		clients:      store.CreateClients(),
		maxClients:   cfg.MaxClients,
		pingInterval: api.Duration{Duration: cfg.PingDuration},
	}
	registry.run(cfg)
	return &registry
}

func (rs *inMemRegistry) Join(ctx context.Context, request api.JoinRequest) (*api.JoinResponse, error) {
	rs.Lock()
	defer rs.Unlock()

	logger.Debug("[registry][join] client join")

	if rs.clients.Size() >= rs.maxClients {
		return nil, api.NewMaxClientsReachedError(rs.maxClients)
	}

	tnt := ctx.Value(api.TenantKey).(string)

	clientId := rs.createClientId()
	c, err := common.NewClient(clientId, request.ServiceId, tnt, request.Endpoints, request.Meta)
	if err != nil {
		return nil, err
	}

	if rs.has(c) {
		return nil, api.NewAlreadyRegisteredError()
	}
	rs.clients.Set(clientId, c)
	if rs.tenants.Get(tnt) == nil {
		rs.tenants.Set(tnt, store.CreateTenant(tnt))
	}
	rs.tenants.Get(tnt).Set(clientId, c)
	rs.update(c)
	logger.Debug("[registry][join] client %s joined", c.ClientId())
	return &api.JoinResponse{
		ClientId:     clientId,
		PingInterval: rs.pingInterval,
	}, nil
}
func (rs *inMemRegistry) Leave(ctx context.Context, clientId string) error {
	client := rs.clients.Get(clientId)
	if client == nil {
		return api.NewClientNotFoundError(clientId)
	}
	logger.Debug("[registry][leave] remove client %s", clientId)
	rs.remove(client)
	return nil
}
func (rs *inMemRegistry) List(ctx context.Context) []api.Client {
	rs.RLock()
	defer rs.RUnlock()
	var clients []api.Client
	tenant := ctx.Value(api.TenantKey).(string)
	if tenant == api.TenantDefault || tenant == "" {
		logger.Debug("[list] list all")
		clients = rs.clients.List()
	} else {
		tnt := rs.tenants.Get(tenant)
		if tnt != nil {
			clients = tnt.Clients()
		} else {
			clients = []api.Client{}
		}
	}
	logger.Debug("[registry][list] list for %v (%d)", tenant, len(clients))
	sort.Slice(clients, func(a, b int) bool {
		if clients[a].ServiceId() != clients[b].ServiceId() {
			return clients[a].ServiceId() < clients[b].ServiceId()
		} else {
			return clients[a].ClientId() < clients[b].ClientId()
		}
	})
	return clients
}
func (rs *inMemRegistry) Ping(clientId string) (api.Pong, error) {
	rs.Lock()
	defer rs.Unlock()
	v := rs.clients.Get(clientId)
	if v == nil {
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
	t := rs.tenants.Get(client.Tenant())
	if t != nil {
		for _, c := range t.Clients() {
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
			rs.RLock()
			for _, t := range rs.tenants.List() {
				go rs.runner(cfg, t)
			}
			rs.RUnlock()
		}
	}()
}
func (rs *inMemRegistry) runner(cfg *config.AppConfig, tenant api.Tenant) {
	for _, c := range tenant.Clients() {
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
		logger.Info("client %s (%s) failing", client.ClientId(), client.ServiceId())
	}
}
func (rs *inMemRegistry) down(client api.Client) {
	if client.State() != api.ClientStateDown {
		client.SetState(api.ClientStateDown)
		rs.update(client)
		logger.Info("client %s (%s) down", client.ClientId(), client.ServiceId())
	}
}
func (rs *inMemRegistry) remove(client api.Client) {
	rs.Lock()
	defer rs.Unlock()
	logger.Info("removing client %s (%s)", client.ClientId(), client.ServiceId())
	defer rs.update(client)
	rs.clients.Delete(client.ClientId())
	for _, t := range rs.tenants.List() {
		c := t.Get(client.ClientId())
		if c != nil {
			t.Delete(client.ClientId())
			return
		}
	}
}
func (rs *inMemRegistry) update(client api.Client) {
	t := rs.tenants.Get(client.Tenant())
	if t == nil {
		return
	}
	for _, c := range t.Clients() {
		c.SetDirty(true)
	}
}
