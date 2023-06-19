package common

import "github.com/slink-go/disco/common/api"

type Tenant struct {
	Name    string
	Clients map[string]api.Client
}
