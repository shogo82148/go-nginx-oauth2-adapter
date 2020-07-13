package adapter

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

// ErrProviderConfigNotFound is the error which provider configure is not found.
var ErrProviderConfigNotFound = errors.New("shogo82148/go-nginx-oauth2-adapter: provider configure not found")

// ErrForbidden is the error which the access is forbidden.
var ErrForbidden = errors.New("shogo82148/go-nginx-oauth2-adapter/provider: access forbidden")

// Server is the go-nginx-oauth2-adapter server.
type Server struct {
	Config             Config
	DefaultPrivider    string
	ProviderConfigs    map[string]ProviderConfig
	SessionStore       sessions.Store
	AppRefreshInterval time.Duration
}

func init() {
	gob.Register(time.Time{})
}

// NewServer returns a new go-nginx-oauth2-adapter server.
func NewServer(config Config) (*Server, error) {
	s := &Server{
		Config:          config,
		ProviderConfigs: map[string]ProviderConfig{},
	}

	for name, provider := range providers {
		var conf map[string]interface{}
		var ok bool
		if config.Providers == nil {
			continue
		}
		conf, ok = config.Providers[name]
		if !ok {
			continue
		}
		providerConfig, err := provider.ParseConfig(conf)
		if err != nil && err != ErrProviderConfigNotFound {
			return nil, err
		}
		s.ProviderConfigs[name] = providerConfig
		if s.DefaultPrivider != "" {
			return nil, errors.New("shogo82148/go-nginx-oauth2-adapter: multiple providers are not supported")
		}
		s.DefaultPrivider = name
	}

	if s.DefaultPrivider == "" {
		return nil, ErrProviderConfigNotFound
	}

	// Decode secrets and set up the session store.
	secrets := make([][]byte, len(s.Config.Secrets))
	for i, secret := range s.Config.Secrets {
		isAuthKey := i%2 == 0
		if secret == nil {
			if isAuthKey {
				logrus.Warn("The session authentication key is empty. you should set secure random hex string.")
				secrets[i] = []byte("dummy-session-authentication-key")
				if s.Config.ConfigTest {
					return nil, errors.New("shogo82148/go-nginx-oauth2-adapter: the session authentication key is empty")
				}
			}
			continue
		}
		var err error
		if isAuthKey {
			switch len(*secret) {
			case 32, 64:
				secrets[i] = []byte(*secret)
			case 128:
				secrets[i], err = hex.DecodeString(*secret)
			default:
				err = errors.New("shogo82148/go-nginx-oauth2-adapter: invalid session authentication key length")
			}
		} else {
			switch len(*secret) {
			case 16, 24, 32:
				secrets[i] = []byte(*secret)
			case 48, 64:
				secrets[i], err = hex.DecodeString(*secret)
			default:
				err = errors.New("shogo82148/go-nginx-oauth2-adapter: invalid session authentication key length")
			}
		}
		if err != nil {
			return nil, err
		}
	}
	if len(secrets) == 0 {
		logrus.Warn("The session authentication key is empty. you should set secure random hex string.")
		secrets = append(secrets, []byte("dummy-session-authentication-key"))
		if s.Config.ConfigTest {
			return nil, errors.New("adapter: the session authentication key is empty")
		}
	}
	store := sessions.NewCookieStore(secrets...)
	store.Options = config.Cookie.Options()
	s.SessionStore = store

	if s.Config.AppRefreshInterval == "" {
		s.AppRefreshInterval = 24 * time.Hour
	} else {
		var err error
		s.AppRefreshInterval, err = time.ParseDuration(s.Config.AppRefreshInterval)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer context.Clear(r)

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
	session, err := s.SessionStore.Get(r, s.Config.SessionName)
	if err != nil {
		// session is broken. retrigger authorization for fix it
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Info("session is broken. trigger reauthorization for fix it.")

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// check when the user has started the session.
	var val interface{}
	var ok bool
	var loggedInAt time.Time
	val = session.Values["logged_in_at"]
	if loggedInAt, ok = val.(time.Time); !ok {
		logrus.WithFields(logrus.Fields{
			"err": "logged_in_at is not found",
		}).Info("session is broken. trigger reauthorization for fix it.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if time.Now().Sub(loggedInAt) > s.AppRefreshInterval {
		logrus.Info("session is expired. trigger reauthorization for fix it.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// send the user information to the application server.
	var provider string
	val = session.Values["provider"]
	if provider, ok = val.(string); !ok {
		logrus.WithFields(logrus.Fields{
			"err": "provider is not found",
		}).Info("session is broken. trigger reauthorization for fix it.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.Header().Add("x-ngx-omniauth-provider", provider)

	var uid string
	val = session.Values["uid"]
	if uid, ok = val.(string); !ok {
		logrus.WithFields(logrus.Fields{
			"err": "uid is not found",
		}).Info("session is broken. trigger reauthorization for fix it.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.Header().Add("x-ngx-omniauth-user", uid)

	var info string
	val = session.Values["info"]
	if info, ok = val.(string); !ok {
		logrus.WithFields(logrus.Fields{
			"err": "info is not found",
		}).Info("session is broken. trigger reauthorization for fix it.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.Header().Add("x-ngx-omniauth-info", info)

	var b []byte
	b, err = base64.StdEncoding.DecodeString(info)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": "info is not base64 encoded",
		}).Info("session is broken. trigger reauthorization for fix it.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var j map[string]interface{}
	if err := json.Unmarshal(b, &j); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": "info is invalid json",
		}).Info("session is broken. trigger reauthorization for fix it.")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if email, ok := j["email"]; ok {
		if email, ok := email.(string); ok {
			w.Header().Add("x-ngx-omniauth-email", email)
		}
	}

	fmt.Fprint(w, "")
}

// HandlerInitiate redirects to authorization page.
func (s *Server) HandlerInitiate(w http.ResponseWriter, r *http.Request) {
	// ignore error bacause we don't need privious session values.
	session, _ := s.SessionStore.Get(r, s.Config.SessionName)

	conf := s.ProviderConfigs[s.DefaultPrivider].Config()
	callback := r.Header.Get("x-ngx-omniauth-initiate-callback")
	next := r.Header.Get("x-ngx-omniauth-initiate-back-to")
	state := generateNewState()

	conf.RedirectURL = callback
	session.Values = map[interface{}]interface{}{}
	session.Values["provider"] = s.DefaultPrivider
	session.Values["callback"] = callback
	session.Values["next"] = next
	session.Values["state"] = state
	if err := session.Save(r, w); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Error("failed to save session")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

// HandlerCallback validates the user infomation, set to cookie
func (s *Server) HandlerCallback(w http.ResponseWriter, r *http.Request) {
	session, err := s.SessionStore.Get(r, s.Config.SessionName)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Error("failed to get session")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var val interface{}
	var ok bool

	var provider string
	val = session.Values["provider"]
	if provider, ok = val.(string); !ok {
		logrus.WithFields(logrus.Fields{
			"err": "provider is not found",
		}).Info("session is broken.")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var callback string
	val = session.Values["callback"]
	if callback, ok = val.(string); !ok {
		logrus.WithFields(logrus.Fields{
			"err": "callback is not found",
		}).Info("session is broken.")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var next string
	val = session.Values["next"]
	if next, ok = val.(string); !ok {
		logrus.WithFields(logrus.Fields{
			"err": "next is not found",
		}).Info("session is broken.")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var state string
	val = session.Values["state"]
	if state, ok = val.(string); !ok {
		logrus.WithFields(logrus.Fields{
			"err": "state is not found",
		}).Info("session is broken.")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	providerConfig, ok := s.ProviderConfigs[provider]
	if !ok {
		logrus.WithFields(logrus.Fields{
			"err": "provider is not found",
		}).Info("session is broken.")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	conf := providerConfig.Config()
	conf.RedirectURL = callback

	query := r.URL.Query()

	if state != query.Get("state") {
		logrus.WithFields(logrus.Fields{
			"err": "state is not correct",
		}).Info("session is broken.")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	code := query.Get("code")
	t, err := conf.Exchange(r.Context(), code)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Info("oauth verification faild")
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	var uid string
	var info map[string]interface{}
	if infoctx, ok := providerConfig.(ProviderInfoContext); ok {
		uid, info, err = infoctx.InfoContext(r.Context(), &conf, t)
	} else {
		uid, info, err = providerConfig.Info(&conf, t)
	}
	if err != nil {
		if err == ErrForbidden {
			logrus.WithFields(logrus.Fields{
				"err": err.Error(),
			}).Warn("access forbidden")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			logrus.WithFields(logrus.Fields{
				"err": err.Error(),
			}).Warn("user info cannot get")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	jsonInfo, encodedInfo, err := encodeInfo(info)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Info("failed encoding info")
	}

	session.Values["uid"] = uid
	session.Values["info"] = encodedInfo
	session.Values["logged_in_at"] = time.Now()

	if err := session.Save(r, w); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err.Error(),
		}).Error("failed to save session")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	logrus.WithFields(logrus.Fields{
		"user": uid,
		"info": jsonInfo,
	}).Info("user login")

	http.Redirect(w, r, next, http.StatusFound)
}

// generateNewState generate secure random state
func generateNewState() string {
	var data [32]byte
	if _, err := crand.Read(data[:]); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(data[:])
}

// encodeInfo encodes the user information for embeding to http header.
func encodeInfo(info map[string]interface{}) (string, string, error) {
	data, err := json.Marshal(info)
	if err != nil {
		return "", "", err
	}
	return string(data), base64.StdEncoding.EncodeToString(data), nil
}
