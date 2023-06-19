package registry

import (
	"fmt"
	"github.com/slink-go/disco/common/api"
	"github.com/slink-go/disco/server/config"
	"plugin"
)

type Backend interface {
	Init(cfg *config.AppConfig) api.Registry
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
