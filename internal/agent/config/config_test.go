package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func setEnv(key, value string) error {
	if err := os.Setenv(key, value); err != nil {
		return fmt.Errorf("incorrect setup enviroment: %w", err)
	}
	return nil
}

func clearEnv(keys ...string) error {
	var errs []error

	for _, k := range keys {
		if err := os.Unsetenv(k); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("incorrect clear enviroments: %w", errors.Join(errs...))
	}
	return nil
}

func setupEnv(env map[string]string) error {
	var errs []error

	for k, v := range env {
		if err := setEnv(k, v); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("incorrect setup enviroments: %w", errors.Join(errs...))
	}
	return nil
}

func TestInitialize_Defaults(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "REPORT_INTERVAL", "POLL_INTERVAL", "RATE_LIMIT")

	config := Initialize()

	require.NoError(t, clearErr)

	assert.Equal(t, "http://"+defaultServerAddress, config.ServerAddress)
	assert.Equal(t, defaultHashKey, config.HashKey)
	assert.Equal(t, defaultPollInterval, config.PollInterval)
	assert.Equal(t, defaultReportInterval, config.ReportInterval)
	assert.Equal(t, defaultRateLimit+1, config.RateLimit)
}

func TestInitialize_Flags(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "REPORT_INTERVAL", "POLL_INTERVAL", "RATE_LIMIT")

	os.Args = []string{
		"cmd",
		"-a", "example.com:8080",
		"-k", "mykey123",
		"-r", "30",
		"-p", "5",
		"-l", "10",
	}

	config := Initialize()

	require.NoError(t, clearErr)

	assert.Equal(t, "http://example.com:8080", config.ServerAddress)
	assert.Equal(t, "mykey123", config.HashKey)
	assert.Equal(t, int64(30), config.ReportInterval)
	assert.Equal(t, int64(5), config.PollInterval)
	assert.Equal(t, int64(10), config.RateLimit)
}

func TestInitialize_Environment(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "REPORT_INTERVAL", "POLL_INTERVAL", "RATE_LIMIT")
	setupErr := setupEnv(map[string]string{
		"ADDRESS":         "api.myapp.com:8080",
		"KEY":             "envkey",
		"REPORT_INTERVAL": "25",
		"POLL_INTERVAL":   "3",
		"RATE_LIMIT":      "7",
	})

	config := Initialize()

	require.NoError(t, clearErr)
	require.NoError(t, setupErr)

	assert.Equal(t, "http://api.myapp.com:8080", config.ServerAddress)
	assert.Equal(t, "envkey", config.HashKey)
	assert.Equal(t, int64(25), config.ReportInterval)
	assert.Equal(t, int64(3), config.PollInterval)
	assert.Equal(t, int64(7), config.RateLimit)
}

func TestInitialize_Priority(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "REPORT_INTERVAL", "POLL_INTERVAL", "RATE_LIMIT")
	setupErr := setupEnv(map[string]string{
		"ADDRESS":         "env.com:8080",
		"REPORT_INTERVAL": "100",
		"POLL_INTERVAL":   "20",
	})

	os.Args = []string{
		"cmd",
		"-a", "flag.com:9090",
		"-r", "10",
	}

	config := Initialize()

	require.NoError(t, clearErr)
	require.NoError(t, setupErr)

	assert.Equal(t, "http://env.com:8080", config.ServerAddress)
	assert.Equal(t, int64(100), config.ReportInterval)
	assert.Equal(t, int64(20), config.PollInterval)
	assert.Equal(t, defaultHashKey, config.HashKey)
	assert.Equal(t, defaultRateLimit+1, config.RateLimit)
}

func TestInitialize_ZeroValues(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "REPORT_INTERVAL", "POLL_INTERVAL", "RATE_LIMIT")

	os.Args = []string{
		"cmd",
		"-r", "0",
		"-p", "0",
		"-l", "0",
	}

	config := Initialize()

	require.NoError(t, clearErr)

	assert.Equal(t, defaultReportInterval, config.ReportInterval)
	assert.Equal(t, defaultPollInterval, config.PollInterval)
	assert.Equal(t, defaultRateLimit, config.RateLimit)
}

func TestInitialize_InvalidValues(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "REPORT_INTERVAL", "POLL_INTERVAL", "RATE_LIMIT")
	setupErr := setupEnv(map[string]string{
		"REPORT_INTERVAL": "abc",
		"POLL_INTERVAL":   "-5",
		"RATE_LIMIT":      "xyz",
	})

	config := Initialize()

	require.NoError(t, clearErr)
	require.NoError(t, setupErr)

	assert.Equal(t, defaultReportInterval, config.ReportInterval)
	assert.Equal(t, int64(-5), config.PollInterval)
	assert.Equal(t, defaultRateLimit, config.RateLimit)
}

func TestInitialize_HttpPrefix(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "REPORT_INTERVAL", "POLL_INTERVAL", "RATE_LIMIT")

	os.Args = []string{"cmd", "-a", "custom.com:8080"}
	config := Initialize()

	require.NoError(t, clearErr)

	assert.Equal(t, "http://custom.com:8080", config.ServerAddress)

	clearFlags()
	os.Args = []string{"cmd"}
	config = Initialize()

	assert.Equal(t, "http://"+defaultServerAddress, config.ServerAddress)
}
