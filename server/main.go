package main

import (
	"fmt"
	"github.com/slink-go/disco/server/config"
	"github.com/slink-go/disco/server/controller/rest"
	"github.com/slink-go/disco/server/jwt"
	"github.com/slink-go/disco/server/registry"
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
