package adapter

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

// Provider is an OAuth provider.
type Provider interface {
	ParseConfig(configFile map[string]interface{}) (ProviderConfig, error)
}

// ProviderConfig is a config for an OAuth provider.
type ProviderConfig interface {
	Config() oauth2.Config
	Info(c *oauth2.Config, t *oauth2.Token) (string, map[string]interface{}, error)
}

// ProviderInfoContext is for support context.Context.
type ProviderInfoContext interface {
	InfoContext(ctx context.Context, c *oauth2.Config, t *oauth2.Token) (string, map[string]interface{}, error)
}

var providers = map[string]Provider{}

// RegisterProvider registers the OAuth provider.
func RegisterProvider(name string, provider Provider) {
	providers[name] = provider
}

func init() {
	RegisterProvider("development", providerDevelopment{})
}

type providerDevelopment struct{}
type providerConfigDevelopment struct {
	listener net.Listener
}

// ParseConfig parses the config for the development provider.
func (providerDevelopment) ParseConfig(configFile map[string]interface{}) (ProviderConfig, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	c := providerConfigDevelopment{
		listener: l,
	}
	go http.Serve(l, c)
	return c, nil
}

func (pc providerConfigDevelopment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth":
		v := url.Values{}
		v.Add("state", r.FormValue("state"))
		v.Add("code", "development-code")
		http.Redirect(w, r, r.FormValue("redirect_uri")+"?"+v.Encode(), http.StatusFound)
	case "/token":
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"access_token":"hogehoge"}`)
	default:
		http.NotFound(w, r)
	}
}

func (pc providerConfigDevelopment) Config() oauth2.Config {
	l := pc.listener
	return oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://" + l.Addr().String() + "/auth",
			TokenURL: "http://" + l.Addr().String() + "/token",
		},
		ClientID:     "development@example.com",
		ClientSecret: "development-client-secret",
		Scopes:       []string{},
	}
}

func (pc providerConfigDevelopment) Info(c *oauth2.Config, t *oauth2.Token) (string, map[string]interface{}, error) {
	return "developer", map[string]interface{}{}, nil
}
