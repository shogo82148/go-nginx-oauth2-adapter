package main

import (
	"os"

	adapter "github.com/shogo82148/go-nginx-oauth2-adapter"
	_ "github.com/shogo82148/go-nginx-oauth2-adapter/provider"
	"github.com/sirupsen/logrus"
)

func setLogLevel() {
	env := os.Getenv("NGX_OMNIAUTH_LOG_LEVEL")
	if env == "" {
		return
	}
	level, err := logrus.ParseLevel(env)
	if err != nil {
		logrus.Warnf("unknown log level: %s", env)
		return
	}
	logrus.SetLevel(level)
}

func setLogFormat() {
	env := os.Getenv("NGX_OMNIAUTH_LOG_FORMAT")
	switch env {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}

func main() {
	setLogLevel()
	setLogFormat()
	os.Exit(adapter.Main(os.Args))
}
