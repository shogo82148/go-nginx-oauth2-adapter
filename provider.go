package adapter

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

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

func init() {
	RegisterProvider("development", providerDevelopment{})
}

type providerDevelopment struct{}
type providerConfigDevelopment struct {
	listener net.Listener
}

func (_ providerDevelopment) ParseConfig(configFile map[string]interface{}) (ProviderConfig, error) {
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
		fmt.Fprintln(w, "{}")
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
	return "", map[string]interface{}{}, nil
}
