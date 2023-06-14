package logger

import (
	"github.com/joho/godotenv"
	"github.com/op/go-logging"
	"os"
)

// https://www.honeybadger.io/blog/golang-logging/
// https://github.com/op/go-logging/blob/master/examples/example.go
var log *logging.Logger

func init() {
	log = logging.MustGetLogger("logger")
	format := `%{color} %{time:2006-01-02 15:04:05.000} %{level:.4s} %{message}%{color:reset}` // %{id:03x}
	formatter, err := logging.NewStringFormatter(format)
	if err != nil {
		log.Fatal(err.Error())
	}
	logging.SetFormatter(formatter)
	var level logging.Level
	_ = godotenv.Load(".env") // init env from .env (if found)
	level, err = logging.LogLevel(os.Getenv("LOGGING_LEVEL"))
	if err != nil {
		level = logging.INFO
	}
	backend := logging.NewLogBackend(os.Stdout, "", 0)

	var backendLeveled logging.LeveledBackend
	backendLeveled = logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(level, "")

	logging.SetBackend(backendLeveled)
}

func Debug(message string, args ...interface{}) {
	if mode == PrintModeNormal {
		log.Debugf(message, args...)
	} else {
		log.Debugf(message+"\r", args...)
	}
}
func Info(message string, args ...interface{}) {
	if mode == PrintModeNormal {
		log.Infof(message, args...)
	} else {
		log.Infof(message+"\r", args...)
	}
}
func Notice(message string, args ...interface{}) {
	if mode == PrintModeNormal {
		log.Noticef(message, args...)
	} else {
		log.Noticef(message+"\r", args...)
	}
}
func Warning(message string, args ...interface{}) {
	if mode == PrintModeNormal {
		log.Warningf(message, args...)
	} else {
		log.Warningf(message+"\r", args...)
	}
}
func Error(message string, args ...interface{}) {
	if mode == PrintModeNormal {
		log.Errorf(message, args...)
	} else {
		log.Errorf(message+"\r", args...)
	}
}
func Critical(message string, args ...interface{}) {
	if mode == PrintModeNormal {
		log.Criticalf(message, args...)
	} else {
		log.Criticalf(message+"\r", args...)
	}
}
func Fatal(message string, args ...interface{}) {
	if mode == PrintModeNormal {
		log.Fatalf(message, args...)
	} else {
		log.Fatalf(message+"\r", args...)
	}
}
