package config

import (
	"github.com/joho/godotenv"
	"github.com/ws-slink/disco/common/config"
	"github.com/ws-slink/disco/common/util/logger"
	"github.com/xhit/go-str2duration/v2"
	"strings"
	"time"
)

type AppConfig struct {
	MonitoringEnabled bool
	ServicePort       uint16
	PingDuration      time.Duration
	SecretKey         string
	BackendType       string
	PluginDir         string
}

func Load() *AppConfig {
	_ = godotenv.Load(".env") // init env from .env (if found)

	cfg := AppConfig{
		MonitoringEnabled: config.ReadBooleanOrDefault("DISCO_MONITORING_ENABLED", false),
		ServicePort:       uint16(config.ReadIntOrDefault("DISCO_SERVICE_PORT", 8080)),
		PingDuration:      config.ReadDurationOrDefault("DISCO_PING_INTERVAL", 15*time.Second),
		SecretKey:         config.ReadString("DISCO_SECRET_KEY"),
		BackendType:       strings.ToLower(config.ReadString("DISCO_BACKEND_TYPE")),
		PluginDir:         config.ReadStringOrDefault("DISCO_PLUGIN_PATH", "."),
	}

	logger.Info("[cfg] monitoring enabled: %v", cfg.MonitoringEnabled)
	logger.Info("[cfg] service port: %v", cfg.ServicePort)
	logger.Info("[cfg] ping duration: %v", str2duration.String(cfg.PingDuration))
	logger.Info("[cfg] secret key: %v", cfg.SecretKey)

	return &cfg
}
