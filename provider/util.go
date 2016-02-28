package provider

import "os"

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
