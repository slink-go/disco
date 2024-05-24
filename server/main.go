package main

import (
	_ "embed"
	"flag"
	"fmt"
	"github.com/slink-go/disco/server/config"
	"github.com/slink-go/disco/server/controller/rest"
	"github.com/slink-go/disco/server/jwt"
	"github.com/slink-go/disco/server/registry"
	"github.com/slink-go/logging"
	"github.com/xhit/go-str2duration/v2"
	"time"
)

//go:embed logo.txt
var discoLogo string
var logger logging.Logger

func main() {

	if generateToken() {
		return
	}

	cfg, j := prepare()
	logger = logging.GetLogger("main")

	// Print Version
	fmt.Println("")
	fmt.Println(discoLogo)
	fmt.Println("")

	logger.Info("[cfg] monitoring port: %v", cfg.MonitoringPort)
	logger.Info("[cfg] service port: %v", cfg.ServicePort)
	logger.Info("[cfg] service secured: %v", cfg.Secured)
	logger.Info("[cfg] certificate file: %v", cfg.SslCertFile)
	logger.Info("[cfg] certificate key: %v", cfg.SslCertKey)
	logger.Info("[cfg] ping duration: %v", str2duration.String(cfg.PingDuration))
	logger.Info("[cfg] failing threshold: %v", cfg.FailingThreshold)
	logger.Info("[cfg] down threshold: %v", cfg.DownThreshold)
	logger.Info("[cfg] remove threshold: %v", cfg.RemoveThreshold)
	logger.Info("[cfg] max clients: %v", cfg.MaxClients)
	logger.Info("[cfg] rate limit: %v", cfg.RequestRate)
	logger.Info("[cfg] burst limit: %v", cfg.RequestBurst)
	logger.Info("[cfg] registered users: %v", cfg.Users())
	//logger.Info("[cfg] secret key: %v", cfg.SecretKey)
	logger.Info("[cfg] backend type: %v", cfg.BackendType)
	logger.Info("[cfg] plugin dir: %v", cfg.PluginDir)
	logger.Info("[cfg] static file path: %v", config.StaticFilePath())

	b, err := registry.LoadBackend(cfg.PluginDir, cfg.BackendType)
	if err != nil {
		panic(err)
	}
	r := b.Init(cfg)

	restSvc, err := rest.NewDiscoService(j, r, cfg)
	if err != nil {
		panic(err)
	}
	restSvc.Run()

}
func generateToken() bool {
	tokenPtr := flag.Bool("token", false, "generate token")
	tenantPtr := flag.String("tenant", "", "use provided tenant name for token generation")
	durPtr := flag.String("duration", "", "use provided duration for token generation")
	flag.Parse()
	if tokenPtr != nil && *tokenPtr {
		durationStr := "1d"
		tenant := "tenant"
		if tenantPtr != nil && *tenantPtr != "" {
			tenant = *tenantPtr
		}
		if durPtr != nil && *durPtr != "" {
			durationStr = *durPtr
		}
		var err error
		var duration time.Duration
		duration, err = str2duration.ParseDuration(durationStr)
		if err != nil {
			logger.Warning("could not parse duration: %s", err.Error())
		} else {
			_, j := prepare()
			var token string
			token, err = j.Generate("disco", tenant, duration)
			if err != nil {
				logger.Warning("could not generate token: %s", err.Error())
			} else {
				fmt.Println(token)
			}
		}
		return true
	}
	return false
}
func prepare() (*config.AppConfig, jwt.Jwt) {
	cfg := config.Load()
	if cfg.SecretKey != "" {
		j, err := jwt.Init(cfg.SecretKey)
		if err != nil {
			panic(err)
		}
		return cfg, j
	}
	return cfg, nil
}
