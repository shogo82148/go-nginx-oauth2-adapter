package adapter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func testStartH2O() (*exec.Cmd, error) {
	wd, _ := os.Getwd()
	os.Chdir(filepath.Join(wd, "examples", "h2o"))
	defer os.Chdir(wd)
	cmd := exec.Command("h2o", "-c", "h2o.conf")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd, cmd.Start()
}

func testStopH2O(cmd *exec.Cmd) {
	cmd.Process.Signal(syscall.SIGTERM)
	cmd.Wait()
}

func TestH2O(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Log("SKIP in not CI environment. if you want this test, execute `CI=1 go test .`.")
		return
	}
	h2o, err := testStartH2O()
	if err != nil {
		t.Error(err)
		return
	}
	defer testStopH2O(h2o)
	fmt.Fprintln(os.Stderr, "start h2o")

	// wait for h2o is ready
	time.Sleep(time.Second)
	fmt.Fprintln(os.Stderr, "awake")

	c := NewConfig()
	c.Providers = map[string]map[string]interface{}{
		"development": {},
	}
	c.Cookie = &CookieConfig{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 3,
		Secure:   false,
		HTTPOnly: true,
		SameSite: "lax",
	}
	s, err := NewServer(*c)
	if err != nil {
		t.Error(err)
		return
	}
	go http.ListenAndServe(":18081", s)

	go http.ListenAndServe(":18082", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// XXX: it seems that h2o does not support to modify requests :(
		// if got, expected := r.Header.Get("x-ngx-omniauth-provider"), "development"; got != expected {
		// 	t.Errorf("want %s, got %s", expected, git)
		// }
		// if got, expected := r.Header.Get("x-ngx-omniauth-user"), "developer"; got != expected {
		// 	t.Errorf("want %s, got %s", expected, git)
		// }
		// if r.Header.Get("x-ngx-omniauth-info") == "" {
		// 	t.Errorf("want x-ngx-omniauth-info is set, but empty")
		// }

		fmt.Fprintln(w, "Hello, client")
	}))

	jar, _ := cookiejar.New(nil)
	client := http.Client{
		Jar: jar,

		// Note:
		// it takes a long time to shutdown gracefully when keep-alives is enabled.
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	resp, err := client.Get("http://ngx-auth-test.loopback.shogo82148.com:18080/")
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Fprintln(os.Stderr, string(b))
	if string(b) != "Hello, client\n" {
		t.Errorf("want Hello, client, got %s", string(b))
	}
}
