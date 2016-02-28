package adapter

import (
	"net/http"

	"golang.org/x/oauth2"
)

var conf oauth2.Config

func Main() {
	provider := providers["google_oauth2"]
	providerConfig, _ := provider.ParseConfig(map[string]interface{}{})
	conf = providerConfig.Config()

	http.HandleFunc("/test", HandlerTest)
	http.HandleFunc("/initiate", HandlerInitiate)
	http.HandleFunc("/callback", HandlerCallback)
	http.ListenAndServe(":8081", nil)
}

// HandlerTest validates the session.
func HandlerTest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

// HandlerInitiate redirects to authorization page.
func HandlerInitiate(w http.ResponseWriter, r *http.Request) {
	conf.RedirectURL = r.Header.Get("x-ngx-omniauth-initiate-callback")
	state := r.Header.Get("x-ngx-omniauth-initiate-back-to")
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

// HandlerCallback validates the user infomation, set to cookie
func HandlerCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	t, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	_ = t
	//http.Redirect(w, r, "/", http.StatusFound)
}
