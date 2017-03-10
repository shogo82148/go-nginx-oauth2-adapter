package main

import (
	"os"

	"github.com/shogo82148/go-nginx-oauth2-adapter"
	_ "github.com/shogo82148/go-nginx-oauth2-adapter/provider"
)

func main() {
	os.Exit(adapter.Main(os.Args))
}
