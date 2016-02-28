package adapter

import (
	"math/rand"
	"time"
)

func Main() {
	rand.Seed(time.Now().UnixNano())

	s, err := NewServer(Config{})
	if err != nil {
		panic(err)
	}

	s.ListenAndServe()
}
