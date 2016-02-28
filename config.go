package adapter

import (
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Host               string                            `yaml:"host", json:"host"`
	Port               string                            `yaml:"port", json:"port"`
	Secret             string                            `yaml:"secret", json:"scret"`
	SessionName        string                            `yaml:"session_name", json:"session_name"`
	Providers          map[string]map[string]interface{} `yaml:"providers", json:"providers"`
	AppRefreshInterval string                            `yaml:"app_refresh_interval", json:"app_refresh_interval"`

	// Fields are a subset of http.Cookie fields.
	Cookie CookieConfig `yaml:"cookie", json:"cookie"`
}

type CookieConfig struct {
	Path   string `yaml:"path", json:"path"`
	Domain string `yaml:"domain", json:"domain"`
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	MaxAge   int  `yaml:"max_age", json:"max_age"`
	Secure   bool `yaml:"secure", json:"secure"`
	HttpOnly bool `yaml:"http_only", json:"http_only"`
}

func NewConfig() *Config {
	return &Config{
		Host:               "",
		Port:               "18080",
		Secret:             "ngx_omniauth_secret_dev",
		SessionName:        "go-nginx-oauth2-session",
		Providers:          map[string]map[string]interface{}{},
		AppRefreshInterval: "24h",
		Cookie: CookieConfig{
			Path:   "/",
			MaxAge: 60 * 60 * 24 * 3,
		},
	}
}

func (c *Config) LoadYaml(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

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
		c.Cookie.HttpOnly = b
	}

	if v := os.Getenv("NGX_OMNIAUTH_SESSION_SECRET"); v != "" {
		c.Secret = v
	}

	if v := os.Getenv("NGX_OMNIAUTH_SESSION_COOKIE_NAME"); v != "" {
		c.SessionName = v
	}

	if v := os.Getenv("NGX_OMNIAUTH_APP_REFRESH_INTERVAL"); v != "" {
		c.AppRefreshInterval = v
	}

	return nil
}
