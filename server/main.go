package main

import (
	"fmt"
	"github.com/ws-slink/disco/server/config"
	"github.com/ws-slink/disco/server/controller/rest"
	"github.com/ws-slink/disco/server/jwt"
	"github.com/ws-slink/disco/server/registry"
)

func main() {

	cfg := config.Load()

	j, err := jwt.Init(cfg.SecretKey)
	if err != nil {
		panic(err)
	}

	b, err := registry.LoadBackend(cfg.PluginDir, cfg.BackendType)
	if err != nil {
		panic(err)
	}
	r := b.Init(cfg)

	restSvc, err := rest.Init(j, r, cfg.MonitoringEnabled)
	if err != nil {
		panic(err)
	}
	restSvc.Run(fmt.Sprintf(":%d", cfg.ServicePort))

}
