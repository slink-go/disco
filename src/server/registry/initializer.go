package registry

import (
	"fmt"
	"github.com/ws-slink/disco/common/api"
	"plugin"
	"time"
)

type Backend interface {
	Init(pingInterval time.Duration) api.Registry
}

func LoadBackend(path, typ string) (Backend, error) {
	module := fmt.Sprintf("%s/%s.so", path, typ)
	p, err := plugin.Open(module)
	if err != nil {
		return nil, err
	}
	sym, err := p.Lookup("Backend")
	if err != nil {
		return nil, err
	}
	back, ok := sym.(Backend)
	if !ok {
		return nil, fmt.Errorf("that's not a Backend")
	}
	return back, nil
}
