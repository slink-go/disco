package registry

import (
	"github.com/ws-slink/disco/common/api"
	"time"
)

// for plugin support

type BackendInitializer interface {
	Init(pingInterval time.Duration) api.Registry
}

type InMemBackendInitializer struct {
}

func (bi *InMemBackendInitializer) Init(pingInterval time.Duration) api.Registry {
	return NewInMemRegistry(pingInterval)
}
