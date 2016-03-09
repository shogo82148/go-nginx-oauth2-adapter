package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/shogo82148/go-nginx-oauth2-adapter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type providerGoogle struct{}
type providerConfigGoogle struct {
	baseConfig     oauth2.Config
	enabledProfile bool
	restrictions   []string
}
type profileGoole struct {
	Gender        string `json:"gender"`
	Name          string `json:"name"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	Kind          string `json:"kind"`
	Sub           string `json:"sub"`
	Profile       string `json:"profile"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
}
type idTypeGoole struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	HD    string `json:"hd"`
}

func init() {
	adapter.RegisterProvider("google_oauth2", providerGoogle{})
}

func (_ providerGoogle) ParseConfig(configFile map[string]interface{}) (adapter.ProviderConfig, error) {
	strScopes := getConfigString(configFile, "scopes", "NGX_OMNIAUTH_GOOGLE_SCOPES")
	if strScopes == "" {
		strScopes = "email,profile"
	}
	scopes := strings.Split(strScopes, ",")

	var c providerConfigGoogle
	c.baseConfig = oauth2.Config{
		Endpoint:     google.Endpoint,
		ClientID:     getConfigString(configFile, "client_id", "NGX_OMNIAUTH_GOOGLE_KEY"),
		ClientSecret: getConfigString(configFile, "client_secret", "NGX_OMNIAUTH_GOOGLE_SECRET"),
		Scopes:       scopes,
	}

	for _, s := range scopes {
		switch s {
		case "profile":
			c.enabledProfile = true
		}
	}

	if c.baseConfig.ClientID == "" || c.baseConfig.ClientSecret == "" {
		return nil, adapter.ErrProviderConfigNotFound
	}

	if irestrictions, ok := configFile["restrictions"].([]interface{}); ok {
		restrictions := make([]string, 0, len(irestrictions))
		for _, r := range restrictions {
			if restriction, ok := r.(string); ok {
				restrictions = append(restrictions, restriction)
			}
		}
		c.restrictions = restrictions
	}

	return c, nil
}

func (pc providerConfigGoogle) Config() oauth2.Config {
	return pc.baseConfig
}

func (pc providerConfigGoogle) Info(c *oauth2.Config, t *oauth2.Token) (string, map[string]interface{}, error) {
	info := map[string]interface{}{}

	// parse id_token
	extra, ok := t.Extra("id_token").(string)
	if !ok {
		return "", nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: id_token is not found")
	}

	keys := strings.Split(extra, ".")
	if len(keys) < 2 {
		return "", nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: invalid id_token")
	}

	data, err := base64Decode(keys[1])
	if err != nil {
		return "", nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: invalid id_token")
	}

	var idType idTypeGoole
	if err := json.Unmarshal(data, &idType); err != nil {
		return "", nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: invalid id_token")
	}
	info["email"] = idType.Email

	if len(pc.restrictions) > 0 {
		valid := false
		for _, r := range pc.restrictions {
			if strings.Contains(r, "@") {
				if r == idType.Email {
					valid = true
					break
				}
			} else {
				if strings.HasSuffix(idType.Email, "@"+r) {
					valid = true
					break
				}
			}
		}
		if !valid {
			return "", nil, adapter.ErrForbidden
		}
	}

	// get detail of profile
	if pc.enabledProfile {
		client := c.Client(oauth2.NoContext, t)
		resp, err := client.Get("https://www.googleapis.com/plus/v1/people/me/openIdConnect")
		if err != nil {
			return "", nil, err
		}
		defer resp.Body.Close()

		var profile profileGoole
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&profile); err != nil {
			fmt.Println(err)
			return "", nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: invaid profile")
		}
		info["name"] = profile.Name
		info["first_name"] = profile.GivenName
		info["last_name"] = profile.FamilyName
		info["image"] = profile.Picture
		info["urls"] = map[string]string{
			"Google": profile.Profile,
		}
	}

	return idType.Sub, info, nil
}
