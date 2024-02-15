package env

import (
	"os"
	"strconv"
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

// FloatFromEnv returns the float number for the given key
// and falls back to the given defaultValue if not set
func FloatFromEnv(key string, defaultValue float64) float64 {
	str := ""
	if str = os.Getenv(key); str != "" {
		str = strings.TrimSpace(str)
	}
	if s, err := strconv.ParseFloat(str, 64); err == nil {
		return s
	}
	return defaultValue
}

// IntFromEnv returns the int number for the given key
// and falls back to the given defaultValue if not set
func IntFromEnv(key string, defaultValue int64) int64 {
	str := ""
	if str = os.Getenv(key); str != "" {
		str = strings.TrimSpace(str)
	}
	if s, err := strconv.ParseInt(str, 10, 64); err == nil {
		return s
	}
	return defaultValue
}
