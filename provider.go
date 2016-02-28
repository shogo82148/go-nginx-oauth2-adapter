package adapter

import "golang.org/x/oauth2"

type Provider interface {
	ParseConfig(configFile map[string]interface{}) (ProviderConfig, error)
}

type ProviderConfig interface {
	Config() oauth2.Config
}

var providers map[string]Provider = map[string]Provider{}

func RegisterProvider(name string, provider Provider) {
	providers[name] = provider
}
