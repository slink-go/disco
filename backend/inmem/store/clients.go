package store

import (
	"github.com/slink-go/disco/common/api"
	"sync"
)

//type Clients interface {
//	Size() int
//	Get(key string) api.Client
//	Set(key string, value api.Client)
//	Keys() []string
//	List() []api.Client
//}

type ClientsSync struct {
	sync.RWMutex
	clients map[string]api.Client
}

func CreateClients() *ClientsSync {
	return &ClientsSync{
		clients: make(map[string]api.Client),
	}
}
func (t *ClientsSync) Size() int {
	t.RLock()
	var result = 0
	if t.clients != nil {
		result = len(t.clients)
	}
	t.RUnlock()
	return result
}
func (t *ClientsSync) Get(key string) api.Client {
	t.RLock()
	result := t.clients[key]
	t.RUnlock()
	return result
}
func (t *ClientsSync) Set(key string, value api.Client) {
	t.Lock()
	t.clients[key] = value
	t.Unlock()
}
func (t *ClientsSync) Delete(key string) {
	t.Lock()
	delete(t.clients, key)
	t.Unlock()
}
func (t *ClientsSync) List() []api.Client {
	t.Lock()
	var result []api.Client
	for _, v := range t.clients {
		result = append(result, v)
	}
	t.Unlock()
	return result
}
