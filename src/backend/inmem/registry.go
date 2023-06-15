package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/ws-slink/disco/backend/common"
	"github.com/ws-slink/disco/common/api"
	"github.com/ws-slink/disco/common/util/logger"
	"github.com/ws-slink/disco/server/config"
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
	mutex        sync.RWMutex
}

func newInMemRegistry(cfg *config.AppConfig) api.Registry {
	registry := inMemRegistry{
		tenants:      map[string]*common.Tenant{},
		clients:      map[string]api.Client{},
		pingInterval: api.Duration{Duration: cfg.PingDuration},
	}
	registry.run(cfg)
	return &registry
}

func (rs *inMemRegistry) Join(ctx context.Context, request api.JoinRequest) (*api.JoinResponse, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	clientId := rs.createClientId()
	c, err := common.NewClient(clientId, request.ServiceId, request.Endpoints, request.Meta)
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
				rs.tenants[tenant] = &common.Tenant{
					Name:    tenant,
					Clients: make(map[string]api.Client),
				}
			}
			rs.tenants[tenant].Clients[clientId] = c
		}
	}
	return &api.JoinResponse{
		ClientId:     clientId,
		PingInterval: rs.pingInterval,
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
	result := []api.Client{}
	clients := rs.clients
	if ctx.Value(api.TenantKey) != nil {
		tenant := ctx.Value(api.TenantKey).(string)
		if tenant != "" {
			tnt, ok := rs.tenants[tenant]
			if ok {
				clients = tnt.Clients
			}
		}
	}
	for _, t := range clients {
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
	// if v.needsUpdate return PongTypeChanged
	return api.Pong{
		Response: api.PongTypeOk,
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
func (rs *inMemRegistry) run(cfg *config.AppConfig) {
	//defer func() {
	//	logger.Warning("run exit")
	//}()
	go func(r api.Registry) {
		//defer func() {
		//	logger.Warning("checker exit")
		//}()
		for {
			//logger.Info("check clients")
			time.Sleep(time.Second)
			for _, c := range r.List(context.Background()) {
				interval := time.Now().Sub(c.LastSeen())
				if time.Duration(cfg.RemoveThreshold)*cfg.PingDuration < interval {
					if c.State() != api.ClientStateRemoved {
						c.SetState(api.ClientStateRemoved)
						rs.remove(c.ClientId())
						logger.Info("removing client %s", c.ClientId())
					}
				} else if time.Duration(cfg.DownThreshold)*cfg.PingDuration < interval {
					if c.State() != api.ClientStateDown {
						c.SetState(api.ClientStateDown)
						logger.Info("client %s down", c.ClientId())
					}
				} else if time.Duration(cfg.FailingThreshold)*cfg.PingDuration < interval {
					if c.State() != api.ClientStateFailing {
						c.SetState(api.ClientStateFailing)
						logger.Info("client %s failing", c.ClientId())
					}
				} else {
					// skip
				}
			}
		}
	}(rs)
}
func (rs *inMemRegistry) remove(clientId string) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	delete(rs.clients, clientId)
	for _, t := range rs.tenants {
		_, ok := t.Clients[clientId]
		if ok {
			delete(t.Clients, clientId)
			return
		}
	}
}
