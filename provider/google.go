package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt"
	"github.com/mendsley/gojwk"
	adapter "github.com/shogo82148/go-nginx-oauth2-adapter"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleOpenIDConfigurationURL = "https://accounts.google.com/.well-known/openid-configuration"

type providerGoogle struct{}
type providerConfigGoogle struct {
	baseConfig     oauth2.Config
	enabledProfile bool
	restrictions   []string

	mu      sync.RWMutex
	jwksuri googleJWKSURI
}
type googleOpenIDConfiguration struct {
	JWKSURI string `json:"jwks_uri"`
}
type googleJWKSURI struct {
	Keys []gojwk.Key `json:"keys"`
}

type profileGoogle struct {
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
	EmailVerified bool   `json:"email_verified"`
}
type idTypeGoogle struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	HD    string `json:"hd"`
}

func (*idTypeGoogle) Valid() error {
	return nil
}

func init() {
	adapter.RegisterProvider("google_oauth2", providerGoogle{})
}

func (providerGoogle) ParseConfig(configFile map[string]interface{}) (adapter.ProviderConfig, error) {
	strScopes := getConfigString(configFile, "scopes", "NGX_OMNIAUTH_GOOGLE_SCOPES")
	if strScopes == "" {
		strScopes = "openid,email,profile"
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
		for _, r := range irestrictions {
			if restriction, ok := r.(string); ok {
				restrictions = append(restrictions, restriction)
			}
		}
		c.restrictions = restrictions
	}

	return &c, nil
}

func (pc *providerConfigGoogle) Config() oauth2.Config {
	return pc.baseConfig
}

func (pc *providerConfigGoogle) Info(c *oauth2.Config, t *oauth2.Token) (string, map[string]interface{}, error) {
	return pc.InfoContext(context.Background(), c, t)
}

func (pc *providerConfigGoogle) InfoContext(ctx context.Context, c *oauth2.Config, t *oauth2.Token) (string, map[string]interface{}, error) {
	info := map[string]interface{}{}

	// parse id_token
	extra, ok := t.Extra("id_token").(string)
	if !ok {
		return "", nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: id_token is not found")
	}

	jwksuri, err := pc.getJWKSURI(ctx)
	if err != nil {
		return "", nil, err
	}

	// parse id_token and validate.
	var idType idTypeGoogle
	_, err = jwt.ParseWithClaims(extra, &idType, func(token *jwt.Token) (interface{}, error) {
		ikid, ok := token.Header["kid"]
		if !ok {
			return nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: kid is not found")
		}
		kid, ok := ikid.(string)
		if !ok {
			return nil, fmt.Errorf("invalid kid type: %T", ikid)
		}
		for _, key := range jwksuri.Keys {
			if key.Kid == kid {
				return key.DecodePublicKey()
			}
		}
		return nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: kid is not found")
	})
	if err != nil {
		return "", nil, errors.New("shogo82148/go-nginx-oauth2-adapter/provider: fail to validate id_token")
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
		client := c.Client(ctx, t)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			return "", nil, err
		}
		defer resp.Body.Close()

		var profile profileGoogle
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&profile); err != nil {
			return "", nil, err
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

func (pc *providerConfigGoogle) getJWKSURI(ctx context.Context) (googleJWKSURI, error) {
	var conf googleOpenIDConfiguration
	if err := parseJSONFromURL(ctx, googleOpenIDConfigurationURL, &conf); err != nil {
		return googleJWKSURI{}, err
	}
	if err := parseJSONFromURL(ctx, conf.JWKSURI, &pc.jwksuri); err != nil {
		return googleJWKSURI{}, err
	}

	return pc.jwksuri, nil
}
