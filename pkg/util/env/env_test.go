package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringFromEnv(t *testing.T) {
	// mock
	_ = os.Setenv("test_string_from_env", "test_string_from_env_value")
	_ = os.Setenv("test_string_from_env_empty", "")
	_ = os.Setenv("test_string_from_env_trim", " test string from env trim ")

	// Test default value
	assert.Equal(t, "default", StringFromEnv("not_exists", "default"))
	assert.Equal(t, "default", StringFromEnv("test_string_from_env_empty", "default"))

	// Test not default value
	assert.Equal(t, "test_string_from_env_value", StringFromEnv("test_string_from_env", "default"))

	// Test trim space
	assert.Equal(t, "test string from env trim", StringFromEnv("test_string_from_env_trim", "default"))
}

func TestBoolFromEnv(t *testing.T) {
	// mock
	_ = os.Setenv("key_0", "0")
	_ = os.Setenv("key_1", "1")
	_ = os.Setenv("key_true", "true")
	_ = os.Setenv("key_false", "false")
	_ = os.Setenv("key_yes", "yes")
	_ = os.Setenv("key_no", "no")
	_ = os.Setenv("key_empty", "")
	_ = os.Setenv("key_space", " ")

	assert.Equal(t, false, BoolFromEnv("not_exists"))
	assert.Equal(t, false, BoolFromEnv("key_empty"))
	assert.Equal(t, false, BoolFromEnv("key_space"))
	assert.Equal(t, false, BoolFromEnv("key_0"))
	assert.Equal(t, true, BoolFromEnv("key_1"))
	assert.Equal(t, true, BoolFromEnv("key_true"))
	assert.Equal(t, false, BoolFromEnv("key_false"))
	assert.Equal(t, true, BoolFromEnv("key_yes"))
	assert.Equal(t, false, BoolFromEnv("key_no"))
}
