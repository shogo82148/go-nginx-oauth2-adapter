package adapter

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/sessions"

	"golang.org/x/oauth2"
)

const DefaultSessionName = "go-nginx-oauth2-session"

type Config struct {
	Host        string                            `yaml:"host", json:"host"`
	Port        string                            `yaml:"port", json:"port"`
	Secret      string                            `yaml:"secret", json:"scret"`
	SessionName string                            `yaml:"session_name", json:"session_name"`
	Providers   map[string]map[string]interface{} `yaml:"providers", json:"providers"`
}

type Server struct {
	Config          Config
	DefaultPrivider string
	ProviderConfigs map[string]ProviderConfig
	SessionStore    sessions.Store
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

	s.SessionStore = sessions.NewCookieStore([]byte(config.Secret))

	if s.Config.SessionName == "" {
		s.Config.SessionName = DefaultSessionName
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
	session, err := s.SessionStore.Get(r, s.Config.SessionName)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	conf := s.ProviderConfigs[s.DefaultPrivider].Config()
	callback := r.Header.Get("x-ngx-omniauth-initiate-callback")
	next := r.Header.Get("x-ngx-omniauth-initiate-back-to")

	conf.RedirectURL = callback
	session.Values["callback"] = callback
	session.Values["next"] = next
	session.Save(r, w)

	// TODO: state is recommended
	http.Redirect(w, r, conf.AuthCodeURL(""), http.StatusFound)
}

// HandlerCallback validates the user infomation, set to cookie
func (s *Server) HandlerCallback(w http.ResponseWriter, r *http.Request) {
	session, err := s.SessionStore.Get(r, s.Config.SessionName)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	conf := s.ProviderConfigs[s.DefaultPrivider].Config()

	var val interface{}
	var ok bool

	var callback string
	val = session.Values["callback"]
	if callback, ok = val.(string); !ok {
		fmt.Println("callback is not set")
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	conf.RedirectURL = callback

	var next string
	val = session.Values["next"]
	if next, ok = val.(string); !ok {
		fmt.Println("next is not set")
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	code := r.URL.Query().Get("code")
	t, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	_ = t

	session.Save(r, w)
	http.Redirect(w, r, next, http.StatusFound)
}
