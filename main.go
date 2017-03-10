package adapter

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lestrrat/go-server-starter/listener"
)

// Version is the version of go-nginx-oauth2-adapter.
const Version = "0.1.0"

// Main starts the go-nginx-oauth2-adapter server.
func Main() {
	rand.Seed(time.Now().UnixNano())

	var configFile string
	var configtest bool
	var showVersion bool
	flag.StringVar(&configFile, "c", "", "configuration file")
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.BoolVar(&configtest, "t", false, "test configuration and exit")
	flag.BoolVar(&configtest, "configtest", false, "test configuration and exit")
	flag.BoolVar(&showVersion, "v", false, "show version information")
	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.Parse()

	if showVersion {
		fmt.Println("go-nginx-oauth2-adapter", Version)
		return
	}

	c, err := parseConfig(configFile)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Fatal("error while parsing configure")
		os.Exit(1)
	}

	l, err := getListener(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Fatal("listen error")
		os.Exit(1)
	}

	s, err := NewServer(*c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Fatal("init server error")
		os.Exit(1)
	} else {
		if configtest {
			os.Exit(0)
		}
	}

	server := &http.Server{
		Handler: LoggingHandler(s),
	}
	go func() {
		err := server.Serve(l)
		if err == http.ErrServerClosed {
			return
		}
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err.Error(),
			}).Fatal("serve error")
		}
	}()

	waitSignal()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = server.Shutdown(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Error("shutdown server error")
	}
}

func parseConfig(configFile string) (*Config, error) {
	c := NewConfig()
	if err := c.LoadEnv(); err != nil {
		return nil, err
	}
	if configFile != "" {
		if err := c.LoadYaml(configFile); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func waitSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	for {
		s := <-signalChan
		if s == syscall.SIGTERM {
			logrus.Info("received SIGTERM")
			return
		}
	}
}

func getListener(c *Config) (net.Listener, error) {
	listeners, err := listener.ListenAll()
	if err != nil && err != listener.ErrNoListeningTarget {
		panic(err)
	}
	if err != listener.ErrNoListeningTarget {
		return listeners[0], nil
	}

	// Fallback if not running under Server::Starter
	return net.Listen("tcp", c.Address)
}
