package config

import (
	"flag"
	"os"
)

const (
	defaultServerAddress string = "localhost:8080"
)

type Config struct {
	ServerAddress string
}

func Initialize() *Config {
	config := Config{
		ServerAddress: defaultServerAddress,
	}

	argValue := flag.String("a", defaultServerAddress, "HTTP server endpoint")
	flag.Parse()
	if argValue != nil && *argValue != "" {
		config.ServerAddress = *argValue
	}

	envValue, ok := os.LookupEnv("ADDRESS")
	if ok && envValue != "" {
		config.ServerAddress = envValue
	}

	return &config
}
