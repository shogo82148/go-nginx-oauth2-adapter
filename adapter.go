package adapter

import (
	"fmt"
	"net"
	"net/http"

	"golang.org/x/oauth2"
)

type Config struct {
	Host      string                            `yaml:"host", json:"host"`
	Port      string                            `yaml:"port", json:"port"`
	Secret    string                            `yaml:"secret", json:"scret"`
	Providers map[string]map[string]interface{} `yaml:"providers", json:"providers"`
}

type Server struct {
	Config          Config
	DefaultPrivider string
	ProviderConfigs map[string]ProviderConfig
}

func Main() {
	s, err := NewServer(Config{})
	if err != nil {
		panic(err)
	}

	s.ListenAndServe()
}

func NewServer(config Config) (*Server, error) {
	s := &Server{
		Config:          config,
		ProviderConfigs: map[string]ProviderConfig{},
	}

	for name, provider := range providers {
		var conf map[string]interface{}
		var ok bool
		if config.Providers != nil {
			conf, ok = config.Providers[name]
			if !ok {
				conf = map[string]interface{}{}
			}
		} else {
			conf = map[string]interface{}{}
		}
		providerConfig, err := provider.ParseConfig(conf)
		if err != nil {
			return nil, err
		}
		s.ProviderConfigs[name] = providerConfig
		s.DefaultPrivider = name
	}

	return s, nil
}

func (s *Server) ListenAndServe() error {
	host := s.Config.Host
	port := s.Config.Port
	if port == "" {
		port = "18080"
	}
	return http.ListenAndServe(net.JoinHostPort(host, port), s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/test":
		s.HandlerTest(w, r)
	case "/initiate":
		s.HandlerInitiate(w, r)
	case "/callback":
		s.HandlerCallback(w, r)
	default:
		http.NotFound(w, r)
	}
}

// HandlerTest validates the session.
func (s *Server) HandlerTest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

// HandlerInitiate redirects to authorization page.
func (s *Server) HandlerInitiate(w http.ResponseWriter, r *http.Request) {
	conf := s.ProviderConfigs[s.DefaultPrivider].Config()
	conf.RedirectURL = r.Header.Get("x-ngx-omniauth-initiate-callback")
	state := r.Header.Get("x-ngx-omniauth-initiate-back-to")
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

// HandlerCallback validates the user infomation, set to cookie
func (s *Server) HandlerCallback(w http.ResponseWriter, r *http.Request) {
	conf := s.ProviderConfigs[s.DefaultPrivider].Config()
	code := r.URL.Query().Get("code")
	t, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	_ = t
	//http.Redirect(w, r, "/", http.StatusFound)
}
