package provider

import (
	"github.com/shogo82148/go-nginx-oauth2-adapter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type providerGoogle struct{}
type providerConfigGoogle struct {
	baseConfig oauth2.Config
}

func init() {
	adapter.RegisterProvider("google_oauth2", providerGoogle{})
}

func (_ providerGoogle) ParseConfig(configFile map[string]interface{}) (adapter.ProviderConfig, error) {
	var c providerConfigGoogle
	c.baseConfig = oauth2.Config{
		Endpoint:     google.Endpoint,
		ClientID:     getConfigString(configFile, "client_id", "NGX_OMNIAUTH_GOOGLE_KEY"),
		ClientSecret: getConfigString(configFile, "client_secret", "NGX_OMNIAUTH_GOOGLE_SECRET"),
		Scopes:       []string{"email"},
	}

	if c.baseConfig.ClientID == "" || c.baseConfig.ClientSecret == "" {
		return nil, adapter.ErrProviderConfigNotFound
	}

	return c, nil
}

func (c providerConfigGoogle) Config() oauth2.Config {
	return c.baseConfig
}
