package main

import (
	"fmt"
	"github.com/ws-slink/disco/server/app/config"
	"github.com/ws-slink/disco/server/app/controller/rest"
	"github.com/ws-slink/disco/server/app/jwt"
	"github.com/ws-slink/disco/server/app/registry"
)

func main() {

	cfg := config.Load()

	j, err := jwt.Init(cfg.SecretKey)
	if err != nil {
		panic(err)
	}

	// for plugin support we should import BackendInitializer
	// from external libraries
	bi := registry.InMemBackendInitializer{}
	r := bi.Init(cfg.PingDuration)

	restSvc, err := rest.Init(j, r, cfg.MonitoringEnabled)
	if err != nil {
		panic(err)
	}
	restSvc.Run(fmt.Sprintf(":%d", cfg.ServicePort))

}
