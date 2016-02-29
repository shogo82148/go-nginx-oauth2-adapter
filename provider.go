package adapter

import "golang.org/x/oauth2"

type Provider interface {
	ParseConfig(configFile map[string]interface{}) (ProviderConfig, error)
}

type ProviderConfig interface {
	Config() oauth2.Config
	Info(c *oauth2.Config, t *oauth2.Token) (string, map[string]interface{}, error)
}

var providers map[string]Provider = map[string]Provider{}

func RegisterProvider(name string, provider Provider) {
	providers[name] = provider
}
