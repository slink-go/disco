package store

import (
	"github.com/slink-go/disco/common/api"
	"sync"
)

type tenant struct {
	name    string
	clients *ClientsSync
}

func (t *tenant) Name() string {
	return t.name
}
func (t *tenant) Set(clientId string, value api.Client) {
	t.clients.Set(clientId, value)
}
func (t *tenant) Get(clientId string) api.Client {
	return t.clients.Get(clientId)
}
func (t *tenant) Delete(clientId string) {
	t.clients.Delete(clientId)
}
func (t *tenant) Clients() []api.Client {
	return t.clients.List()
}

type TenantsSync struct {
	sync.RWMutex
	tenants map[string]api.Tenant
}

func CreateTenants() *TenantsSync {
	return &TenantsSync{
		tenants: make(map[string]api.Tenant),
	}
}
func CreateTenant(name string) api.Tenant {
	return &tenant{
		name:    name,
		clients: CreateClients(),
	}
}

func (t *TenantsSync) Size() int {
	t.RLock()
	var result = 0
	if t.tenants != nil {
		result = len(t.tenants)
	}
	t.RUnlock()
	return result
}
func (t *TenantsSync) Get(key string) api.Tenant {
	t.RLock()
	result := t.tenants[key]
	t.RUnlock()
	return result
}
func (t *TenantsSync) Set(key string, value api.Tenant) {
	t.Lock()
	t.tenants[key] = value
	t.Unlock()
}
func (t *TenantsSync) List() []api.Tenant {
	t.Lock()
	var result []api.Tenant
	for _, v := range t.tenants {
		result = append(result, v)
	}
	t.Unlock()
	return result
}
