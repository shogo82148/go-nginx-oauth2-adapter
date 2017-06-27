class OAuth2Adapter

  def initialize(config={})
    config = {
      :auth_server => "http://127.0.0.1:18081",
      :auth_callback => "/_auth/callback",
    }.merge(config)
    @auth_server = config[:auth_server]
    @auth_callback = config[:auth_callback]
    raise "auth_server must not be nil" if @auth_server.nil?
    raise "auth_callback must not be nil" if @auth_callback.nil?
  end

  def call(env)
    if env["PATH_INFO"] == @auth_callback
      req = http_request(
        "#{@auth_server}/callback?#{env["QUERY_STRING"]}",
        headers: {
          "COOKIE" => env["HTTP_COOKIE"],
        }
      )
      return req.join
    end

    req = http_request(
      "#{@auth_server}/test",
      headers: {
        "COOKIE"                      => env["HTTP_COOKIE"],
        "X-NGX-OMNIAUTH-ORIGINAL-URI" => env["rack.url_scheme"] + "://" + env["HTTP_HOST"] + env["PATH_INFO"],
      })
    status, testheaders, body = req.join

    if status == 401
      # not login, redirect to authorization page
      req = http_request(
        "#{@auth_server}/initiate",
        headers: {
          "COOKIE"                           => env["HTTP_COOKIE"],
          "X-NGX-OMNIAUTH-INITIATE-BACK-TO"  => env["rack.url_scheme"] + "://" + env["HTTP_HOST"] + env["PATH_INFO"],
          "X-NGX-OMNIAUTH-INITIATE-CALLBACK" => env["rack.url_scheme"] + "://" + env["HTTP_HOST"] + @auth_callback,
        })
      return req.join
    elsif status < 200 || 300 <= status
      return [403, {'content-type' => 'text/plain'}, ["Forbidden\n"]]
    end

    return [399, {}, []]
  end

end
