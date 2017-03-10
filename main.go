package adapter

import (
	"flag"
	"fmt"
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

const Version = "0.1.0"

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
	} else {
		if configtest {
			os.Exit(0)
		}
	}

	gracedown.Serve(l, LoggingHandler(s))
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
	return net.Listen("tcp", c.Address)
}
