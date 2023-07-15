package config

import (
	"github.com/joho/godotenv"
	"github.com/slink-go/disco/common/config"
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
	FailingThreshold  uint16
	DownThreshold     uint16
	RemoveThreshold   uint16
	MaxClients        int
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
		FailingThreshold:  uint16(config.ReadIntOrDefault("DISCO_CLIENT_FAILING_THRESHOLD", 2)),
		DownThreshold:     uint16(config.ReadIntOrDefault("DISCO_CLIENT_DOWN_THRESHOLD", 4)),
		RemoveThreshold:   uint16(config.ReadIntOrDefault("DISCO_CLIENT_REMOVE_THRESHOLD", 8)),
		MaxClients:        config.ReadIntOrDefault("DISCO_MAX_CLIENTS", 1024),
	}

	return &cfg
}
