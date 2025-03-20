package config

import (
	"flag"
	"os"
	"strconv"
)

const (
	defaultServerAddress   string = "localhost:8080"
	defaultStoreInterval   int64  = 300
	defaultFileStoragePath string = "../../temp_metrics.json"
	defaultRestore         bool   = false
)

type Config struct {
	ServerAddress   string
	FileStoragePath string
	StoreInterval   int64
	Restore         bool
}

func Initialize() *Config {
	config := Config{
		ServerAddress:   defaultServerAddress,
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: defaultFileStoragePath,
		Restore:         defaultRestore,
	}

	config.getFlags()
	config.getEnvironments()

	return &config
}

func (c *Config) getFlags() {
	argEndpValue := flag.String("a", defaultServerAddress, "HTTP server endpoint")
	argIntervalValue := flag.Int64("i", defaultStoreInterval, "Interval in seconds to save data")
	argFileValue := flag.String("f", defaultFileStoragePath, "Path to file")
	argRestoreValue := flag.Bool("r", defaultRestore, "Loading data when the application starts")

	flag.Parse()

	if argEndpValue != nil && *argEndpValue != "" {
		c.ServerAddress = *argEndpValue
	}
	if argIntervalValue != nil && *argIntervalValue >= 0 {
		c.StoreInterval = *argIntervalValue
	}
	if argFileValue != nil && *argFileValue != "" {
		c.FileStoragePath = *argFileValue
	}
	if argRestoreValue != nil {
		c.Restore = *argRestoreValue
	}
}

func (c *Config) getEnvironments() {
	envEndpValue, ok := os.LookupEnv("ADDRESS")
	if ok && envEndpValue != "" {
		c.ServerAddress = envEndpValue
	}

	envStoreValue, ok := os.LookupEnv("STORE_INTERVAL")
	if ok && envStoreValue != "" {
		if val, err := strconv.ParseInt(envStoreValue, 10, 64); err == nil {
			c.StoreInterval = val
		}
	}

	envFileValue, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok && envFileValue != "" {
		c.FileStoragePath = envFileValue
	}

	envRestoreValue, ok := os.LookupEnv("RESTORE")
	if ok && envRestoreValue != "" {
		if val, err := strconv.ParseBool(envRestoreValue); err == nil {
			c.Restore = val
		}
	}
}
