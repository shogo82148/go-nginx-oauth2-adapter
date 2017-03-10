lambda do |env|
  auth_server = "http://127.0.0.1:18081"
  auth_callback = "/_auth/callback"

  if env["PATH_INFO"] == auth_callback
    req = http_request(
      "#{auth_server}/callback?#{env["QUERY_STRING"]}",
      headers: {
        "COOKIE" => env["HTTP_COOKIE"],
      }
    )
    return req.join
  end

  req = http_request(
    "#{auth_server}/test",
    headers: {
      "COOKIE"                      => env["HTTP_COOKIE"],
      "X-NGX-OMNIAUTH-ORIGINAL-URI" => env["rack.url_scheme"] + "://" + env["HTTP_HOST"] + env["PATH_INFO"],
    })
  status, testheaders, body = req.join

  if status == 401
    # not login, redirect to authorization page
    req = http_request(
      "#{auth_server}/initiate",
      headers: {
        "COOKIE"                           => env["HTTP_COOKIE"],
        "X-NGX-OMNIAUTH-INITIATE-BACK-TO"  => env["rack.url_scheme"] + "://" + env["HTTP_HOST"] + env["PATH_INFO"],
        "X-NGX-OMNIAUTH-INITIATE-CALLBACK" => env["rack.url_scheme"] + "://" + env["HTTP_HOST"] + auth_callback,
      })
    return req.join
  elsif status < 200 || 300 <= status
    return [403, {'content-type' => 'text/plain'}, ["Forbidden\n"]]
  end

  return [399, {
    "x-ngx-omniauth-provider" => testheaders["x-ngx-omniauth-provider"],
    "x-ngx-omniauth-user"     => testheaders["x-ngx-omniauth-user"],
    "x-ngx-omniauth-info"     => testheaders["x-ngx-omniauth-info"],
  }, []]
end
