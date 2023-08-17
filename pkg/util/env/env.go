package env

import (
	"os"
	"strings"
)

// StringFromEnv returns the env variable for the given key
// and falls back to the given defaultValue if not set
func StringFromEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return strings.TrimSpace(v)
	}
	return defaultValue
}

func BoolFromEnv(key string) bool {
	v := os.Getenv(key)
	v = strings.ToLower(strings.TrimSpace(v))
	return v != "" && v != "0" && v != "false" && v != "no"
}
