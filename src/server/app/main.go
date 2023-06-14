package main

import (
	"github.com/ws-slink/disco/server/app/controller/rest"
	"github.com/ws-slink/disco/server/app/jwt"
	"github.com/ws-slink/disco/server/app/registry"
)

func main() {
	j, err := jwt.Init("cF8Kj6GN5zrWffjeSdIMBtXBTkPbSYWI")
	if err != nil {
		panic(err)
	}
	r := registry.NewInMemRegistry()
	restSvc, err := rest.Init(j, r)
	if err != nil {
		panic(err)
	}
	restSvc.Run(":8080")
}
