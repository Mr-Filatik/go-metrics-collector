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
	clearErr := clearEnv("ADDRESS", "KEY", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE", "DATABASE_DSN")

	config := Initialize()

	require.NoError(t, clearErr)

	assert.Equal(t, defaultServerAddress, config.ServerAddress)
	assert.Equal(t, defaultHashKey, config.HashKey)
	assert.Equal(t, defaultStoreInterval, config.StoreInterval)
	assert.Equal(t, defaultFileStoragePath, config.FileStoragePath)
	assert.Equal(t, defaultRestore, config.Restore)
	assert.Equal(t, defaultConnectionString, config.ConnectionString)
}

func TestInitialize_Flags(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE", "DATABASE_DSN")

	os.Args = []string{
		"cmd",
		"-a", "localhost:9090",
		"-k", "mysecretkey",
		"-i", "10",
		"-f", "/tmp/metrics.json",
		"-r=true",
		"-d", "postgresql://localhost:5432/metrics",
	}

	config := Initialize()

	require.NoError(t, clearErr)

	assert.Equal(t, "localhost:9090", config.ServerAddress)
	assert.Equal(t, "mysecretkey", config.HashKey)
	assert.Equal(t, int64(10), config.StoreInterval)
	assert.Equal(t, "/tmp/metrics.json", config.FileStoragePath)
	assert.Equal(t, true, config.Restore)
	assert.Equal(t, "postgresql://localhost:5432/metrics", config.ConnectionString)
}

func TestInitialize_Environment(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE", "DATABASE_DSN")
	setupErr := setupEnv(map[string]string{
		"ADDRESS":           "localhost:8081",
		"KEY":               "envkey123",
		"STORE_INTERVAL":    "20",
		"FILE_STORAGE_PATH": "/data/metrics.json",
		"RESTORE":           "true",
		"DATABASE_DSN":      "sqlite:///app.db",
	})

	config := Initialize()

	require.NoError(t, clearErr)
	require.NoError(t, setupErr)

	assert.Equal(t, "localhost:8081", config.ServerAddress)
	assert.Equal(t, "envkey123", config.HashKey)
	assert.Equal(t, int64(20), config.StoreInterval)
	assert.Equal(t, "/data/metrics.json", config.FileStoragePath)
	assert.Equal(t, true, config.Restore)
	assert.Equal(t, "sqlite:///app.db", config.ConnectionString)
}

func TestInitialize_Priority(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("ADDRESS", "KEY", "STORE_INTERVAL", "FILE_STORAGE_PATH", "RESTORE", "DATABASE_DSN")
	setupErr := setupEnv(map[string]string{
		"ADDRESS": "env:8080",
		"KEY":     "envkey",
	})

	os.Args = []string{
		"cmd",
		"-a", "flag:9090",
		"-k", "flagkey",
		"-i", "5",
		"-r=false",
	}

	config := Initialize()

	require.NoError(t, clearErr)
	require.NoError(t, setupErr)

	assert.Equal(t, "env:8080", config.ServerAddress)
	assert.Equal(t, "envkey", config.HashKey)
	assert.Equal(t, int64(5), config.StoreInterval)
	assert.Equal(t, false, config.Restore)
	assert.Equal(t, defaultFileStoragePath, config.FileStoragePath)
	assert.Equal(t, defaultConnectionString, config.ConnectionString)
}

func TestInitialize_StoreInterval_Negative(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("STORE_INTERVAL")

	os.Args = []string{"cmd", "-i", "-10"}
	config := Initialize()

	require.NoError(t, clearErr)

	assert.Equal(t, defaultStoreInterval, config.StoreInterval)
}

func TestInitialize_Restore_Empty(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("RESTORE")
	setErr := setEnv("RESTORE", "")

	config := Initialize()

	require.NoError(t, clearErr)
	require.NoError(t, setErr)

	assert.Equal(t, defaultRestore, config.Restore)
}

func TestInitialize_InvalidStoreInterval(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("STORE_INTERVAL")
	setErr := setEnv("STORE_INTERVAL", "abc")

	config := Initialize()

	require.NoError(t, clearErr)
	require.NoError(t, setErr)

	assert.Equal(t, defaultStoreInterval, config.StoreInterval)
}

func TestInitialize_InvalidRestore(t *testing.T) {
	clearFlags()
	clearErr := clearEnv("RESTORE")
	setErr := setEnv("RESTORE", "maybe")

	config := Initialize()

	require.NoError(t, clearErr)
	require.NoError(t, setErr)

	assert.Equal(t, defaultRestore, config.Restore)
}
