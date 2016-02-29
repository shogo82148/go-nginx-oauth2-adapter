package adapter

import (
	"flag"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lestrrat/go-server-starter/listener"
	"github.com/shogo82148/go-gracedown"
)

func Main() {
	rand.Seed(time.Now().UnixNano())

	c, err := parseConfig()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Fatal("error while parsing configure")
		os.Exit(1)
	}

	startWatchSignal()
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
	}

	gracedown.Serve(l, LoggingHandler(s))
}

func parseConfig() (*Config, error) {
	var configFile string
	flag.StringVar(&configFile, "c", "", "configuration file")
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.Parse()

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

func startWatchSignal() {
	signal_chan := make(chan os.Signal)
	signal.Notify(signal_chan, syscall.SIGTERM)
	go func() {
		for {
			s := <-signal_chan
			if s == syscall.SIGTERM {
				logrus.Info("received SIGTERM")
				gracedown.Close()
			}
		}
	}()
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
	return net.Listen("tcp", net.JoinHostPort(c.Host, c.Port))
}
