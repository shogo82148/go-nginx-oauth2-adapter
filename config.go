package adapter

import (
	"io/ioutil"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"gopkg.in/yaml.v2"
)

// Config is a configration for go-nginx-oauth2-adapter.
type Config struct {
	Address            string                            `yaml:"address" json:"address"`
	Secrets            []*string                         `yaml:"secrets" json:"secrets"`
	SessionName        string                            `yaml:"session_name" json:"session_name"`
	Providers          map[string]map[string]interface{} `yaml:"providers" json:"providers"`
	AppRefreshInterval string                            `yaml:"app_refresh_interval" json:"app_refresh_interval"`

	// set with -configtest option.
	ConfigTest bool `yaml:"-" json:"-"`

	// Fields are a subset of http.Cookie fields.
	Cookie *CookieConfig `yaml:"cookie" json:"cookie"`
}

// CookieConfig is a configration for the cookie of HTTP.
type CookieConfig struct {
	Path   string `yaml:"path" json:"path"`
	Domain string `yaml:"domain" json:"domain"`
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int  `yaml:"max_age" json:"max_age"`
	Secure   bool `yaml:"secure" json:"secure"`
	HTTPOnly bool `yaml:"http_only" json:"http_only"`
}

// NewConfig returns a new config.
func NewConfig() *Config {
	return &Config{
		Address:            ":18081",
		Secrets:            nil,
		SessionName:        "go-nginx-oauth2-session",
		Providers:          map[string]map[string]interface{}{},
		AppRefreshInterval: "24h",
		Cookie: &CookieConfig{
			Path:   "/",
			MaxAge: 60 * 60 * 24 * 3,
		},
	}
}

// LoadYaml loads the config from yaml file.
func (c *Config) LoadYaml(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

// LoadEnv loads the config from the environment values.
func (c *Config) LoadEnv() error {
	if v := os.Getenv("NGX_OMNIAUTH_SESSION_COOKIE_TIMEOUT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		c.Cookie.MaxAge = i
	}

	if v := os.Getenv("NGX_OMNIAUTH_SESSION_COOKIE_SECURE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		c.Cookie.Secure = b
	}

	if v := os.Getenv("NGX_OMNIAUTH_SESSION_COOKIE_HTTP_ONLY"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		c.Cookie.HTTPOnly = b
	}

	if v := os.Getenv("NGX_OMNIAUTH_SESSION_SECRET"); v != "" {
		c.Secrets = []*string{&v}
	}

	if v := os.Getenv("NGX_OMNIAUTH_SESSION_COOKIE_NAME"); v != "" {
		c.SessionName = v
	}

	if v := os.Getenv("NGX_OMNIAUTH_APP_REFRESH_INTERVAL"); v != "" {
		c.AppRefreshInterval = v
	}

	if v := os.Getenv("NGX_OMNIAUTH_ADDRESS"); v != "" {
		c.Address = v
	}

	return nil
}

// Options returns the sesseion config.
func (c *CookieConfig) Options() *sessions.Options {
	if c == nil {
		return &sessions.Options{}
	}
	return &sessions.Options{
		Path:     c.Path,
		Domain:   c.Domain,
		MaxAge:   c.MaxAge,
		Secure:   c.Secure,
		HttpOnly: c.HTTPOnly,
	}
}
