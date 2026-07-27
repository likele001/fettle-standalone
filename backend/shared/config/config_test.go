package config

import (
	"os"
	"testing"
)

func TestGetEnv_Exists(t *testing.T) {
	os.Setenv("TEST_GET_ENV_EXISTS", "hello")
	defer os.Unsetenv("TEST_GET_ENV_EXISTS")

	result := GetEnv("TEST_GET_ENV_EXISTS", "default")
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestGetEnv_Missing(t *testing.T) {
	os.Unsetenv("TEST_GET_ENV_MISSING")

	result := GetEnv("TEST_GET_ENV_MISSING", "default")
	if result != "default" {
		t.Errorf("expected 'default', got '%s'", result)
	}
}

func TestGetEnv_Empty(t *testing.T) {
	os.Setenv("TEST_GET_ENV_EMPTY", "")
	defer os.Unsetenv("TEST_GET_ENV_EMPTY")

	result := GetEnv("TEST_GET_ENV_EMPTY", "default")
	if result != "default" {
		t.Errorf("expected 'default' for empty value, got '%s'", result)
	}
}

func TestGetEnvInt_Exists(t *testing.T) {
	os.Setenv("TEST_GET_ENV_INT_EXISTS", "42")
	defer os.Unsetenv("TEST_GET_ENV_INT_EXISTS")

	result := GetEnvInt("TEST_GET_ENV_INT_EXISTS", 0)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestGetEnvInt_Missing(t *testing.T) {
	os.Unsetenv("TEST_GET_ENV_INT_MISSING")

	result := GetEnvInt("TEST_GET_ENV_INT_MISSING", 100)
	if result != 100 {
		t.Errorf("expected 100, got %d", result)
	}
}

func TestGetEnvInt_Invalid(t *testing.T) {
	os.Setenv("TEST_GET_ENV_INT_INVALID", "not-a-number")
	defer os.Unsetenv("TEST_GET_ENV_INT_INVALID")

	result := GetEnvInt("TEST_GET_ENV_INT_INVALID", 50)
	if result != 50 {
		t.Errorf("expected 50 for invalid value, got %d", result)
	}
}

func TestGetEnvBool_True(t *testing.T) {
	os.Setenv("TEST_GET_ENV_BOOL_TRUE", "true")
	defer os.Unsetenv("TEST_GET_ENV_BOOL_TRUE")

	result := GetEnvBool("TEST_GET_ENV_BOOL_TRUE", false)
	if !result {
		t.Error("expected true")
	}
}

func TestGetEnvBool_False(t *testing.T) {
	os.Setenv("TEST_GET_ENV_BOOL_FALSE", "false")
	defer os.Unsetenv("TEST_GET_ENV_BOOL_FALSE")

	result := GetEnvBool("TEST_GET_ENV_BOOL_FALSE", true)
	if result {
		t.Error("expected false")
	}
}

func TestGetEnvBool_Missing(t *testing.T) {
	os.Unsetenv("TEST_GET_ENV_BOOL_MISSING")

	result := GetEnvBool("TEST_GET_ENV_BOOL_MISSING", true)
	if !result {
		t.Error("expected true (default)")
	}
}
