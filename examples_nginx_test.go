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

func testStartNginx() (*exec.Cmd, error) {
	wd, _ := os.Getwd()
	cmd := exec.Command("nginx", "-c", filepath.Join(wd, "examples", "nginx", "nginx.conf"))
	return cmd, cmd.Start()
}

func testStopNginx(cmd *exec.Cmd) {
	cmd.Process.Signal(syscall.SIGTERM)
	cmd.Wait()
}

func TestNginx(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Log("SKIP in not CI environment. if you want this test, execute `CI=1 go test .`.")
		return
	}

	nginx, err := testStartNginx()
	if err != nil {
		t.Error(err)
		return
	}
	defer testStopNginx(nginx)

	// wait for nginx is ready
	time.Sleep(time.Second)

	c := NewConfig()
	c.Providers = map[string]map[string]interface{}{
		"development": map[string]interface{}{},
	}
	s, err := NewServer(*c)
	if err != nil {
		t.Error(err)
		return
	}
	go http.ListenAndServe(":18081", s)

	go http.ListenAndServe(":18082", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))

	jar, _ := cookiejar.New(nil)
	client := http.Client{Jar: jar}
	resp, err := client.Get("http://ngx-auth-test.127.0.0.1.xip.io:18080/")
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	if string(b) != "Hello, client\n" {
		t.Error("want Hello, client, got %s", string(b))
	}
}
