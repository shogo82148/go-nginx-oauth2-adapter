package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gregjones/httpcache"
)

var cacheTransport = httpcache.NewMemoryCacheTransport()

func getConfigString(configFile map[string]interface{}, key string, envName string) string {
	// load a value from config file
	if v, ok := configFile[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}

	// load from the environment if there is no value in config file
	return os.Getenv(envName)
}

func parseJSONFromURL(ctx context.Context, u string, v interface{}) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	resp, err := cacheTransport.Client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}
