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

	endpointEnv := os.Getenv("ADDRESS")
	var endpoint *string
	if endpointEnv == "" {
		endpoint = flag.String("a", defaultServerAddress, "HTTP server endpoint")
	} else {
		endpoint = &endpointEnv
	}
	flag.Parse()

	config := Config{
		ServerAddress: *endpoint,
	}
	return &config
}
