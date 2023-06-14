package registry

import "github.com/ws-slink/disco/common/api"

type Tenant struct {
	Name    string
	Clients map[string]api.Client
}
