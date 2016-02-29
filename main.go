package adapter

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lestrrat/go-server-starter/listener"
	"github.com/shogo82148/go-gracedown"
)

func Main() {
	log.Printf("start pid %d\n", os.Getpid())
	rand.Seed(time.Now().UnixNano())

	c := parseConfig()
	startWatchSignal()
	l, err := getListener(c)
	if err != nil {
		panic(err)
	}

	s, err := NewServer(*c)
	if err != nil {
		panic(err)
	}

	gracedown.Serve(l, s)
}

func parseConfig() *Config {
	var configFile string
	flag.StringVar(&configFile, "c", "", "configuration file")
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.Parse()

	c := NewConfig()
	if err := c.LoadEnv(); err != nil {
		panic(err)
	}
	if configFile != "" {
		if err := c.LoadYaml(configFile); err != nil {
			panic(err)
		}
	}

	return c
}

func startWatchSignal() {
	signal_chan := make(chan os.Signal)
	signal.Notify(signal_chan, syscall.SIGTERM)
	go func() {
		for {
			s := <-signal_chan
			if s == syscall.SIGTERM {
				log.Printf("SIGTERM!!!!\n")
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
