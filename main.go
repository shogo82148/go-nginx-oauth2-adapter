package adapter

import (
	"math/rand"
	"time"
)

func Main() {
	rand.Seed(time.Now().UnixNano())

	c := NewConfig()
	if err := c.LoadEnv(); err != nil {
		panic(err)
	}
	s, err := NewServer(*c)
	if err != nil {
		panic(err)
	}

	s.ListenAndServe()
}
