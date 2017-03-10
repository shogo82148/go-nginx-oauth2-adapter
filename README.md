[![Circle CI](https://circleci.com/gh/shogo82148/go-nginx-oauth2-adapter.svg?style=svg)](https://circleci.com/gh/shogo82148/go-nginx-oauth2-adapter)

# go-nginx-oauth2-adapter

a golang port for [sorah/nginx_omniauth_adapter](https://github.com/sorah/nginx_omniauth_adapter)

## PREREQUISITE

- nginx with ngx_http_auth_request_module, or h2o with mruby

## USAGE

``` bash
$ go get github.com/shogo82148/go-nginx-oauth2-adapter/cli/go-nginx-oauth2-adapter
$ go-nginx-oauth2-adapter
```

## CONFIGURATION

The example of configuration file.

``` yaml
address: ":18081" # listen address

# secret tokens to authenticate/encrypt cookie.
# see http://www.gorillatoolkit.org/pkg/sessions for more detail
secrets:
  - new-authentication-key
  - new-encryption-key
  - old-authentication-key
  - old-encryption-key
session_name: go-nginx-oauth2-session
app_refresh_interval: 24h

# cookie settings for saving session
cookie:
  path: /
  domain:
  max_age: 0
  secure: false
  http_only: false

providers:
  # development: {} # For test.
  google_oauth2:
    client_id: YOUR_CLIENT_ID
    client_secret: YOUR_CLIENT_SECRET
    scopes: "email,profile" # default: "email,profile"
    restrictions:
      - example.com # domain of your Google App
      - specific.user@example.com
```

## LICENSE

This software is released under the MIT License, see LICENSE.md.

## SEE ALSO

- [sorah/nginx_omniauth_adapter](https://github.com/sorah/nginx_omniauth_adapter)
- [nginx で omniauth を利用してアクセス制御を行う(written in Japanese)](http://techlife.cookpad.com/entry/2015/10/16/080000)
