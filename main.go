package adapter

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

func Main() {
	var configFile string
	flag.StringVar(&configFile, "c", "", "configuration file")
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	c := NewConfig()
	if err := c.LoadEnv(); err != nil {
		panic(err)
	}
	if configFile != "" {
		if err := c.LoadYaml(configFile); err != nil {
			panic(err)
		}
	}
	fmt.Println(c)
	s, err := NewServer(*c)
	if err != nil {
		panic(err)
	}

	s.ListenAndServe()
}
