package config

import (
	"github.com/slink-go/disco/common/util/logger"
	"github.com/xhit/go-str2duration/v2"
	"os"
	"strconv"
	"time"
)

const errTemplate = "could not find environment variable %s"

func preprocessKey(key string) string {

	//k := strings.ReplaceAll(key, ".", "_")
	//k = strings.ToUpper(k)
	//logger.Info("processed key: %s -> %s", key, k)

	return key
}

func ReadBooleanOrDefault(key string, def bool) bool {
	k := preprocessKey(key)
	env := os.Getenv(k)
	if env == "" {
		logger.Debug(errTemplate, k)
		return def
	}
	v, err := strconv.ParseBool(env)
	if err != nil {
		logger.Debug("could not parse boolean environment variable %s: %s", k, err.Error())
		return def
	}
	return v
}
func ReadIntOrDefault(key string, def int) int {
	k := preprocessKey(key)
	env := os.Getenv(k)
	if env == "" {
		logger.Debug(errTemplate, k)
		return def
	}
	v, err := strconv.ParseInt(env, 10, 64)
	if err != nil {
		logger.Debug("could not parse int from environment variable %s: %s", k, err.Error())
		return def
	}
	return int(v)
}
func ReadDurationOrDefault(key string, def time.Duration) time.Duration {
	k := preprocessKey(key)
	env := os.Getenv(k)
	if env == "" {
		logger.Debug(errTemplate, k)
		return def
	}
	v, err := str2duration.ParseDuration(env)
	if err != nil {
		logger.Debug("could not parse duration from environment variable %s: %s", k, err.Error())
		return def
	}
	return v
}
func ReadStringOrDefault(key, def string) string {
	k := preprocessKey(key)
	env := os.Getenv(k)
	if env == "" {
		logger.Debug(errTemplate, k)
		return def
	}
	return env
}

func ReadString(key string) string {
	k := preprocessKey(key)
	env := os.Getenv(k)
	if env == "" {
		logger.Debug(errTemplate, k)
	}
	return env
}
