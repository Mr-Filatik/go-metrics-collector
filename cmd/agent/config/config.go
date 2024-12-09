package config

import (
	"flag"
	"os"
	"strconv"
)

const (
	defaultServerAddress  string = "localhost:8080"
	defaultPollInterval   int64  = 2
	defaultReportInterval int64  = 10
)

type Config struct {
	ServerAddress  string
	PollInterval   int64
	ReportInterval int64
}

func Initialize() *Config {
	config := Config{
		ServerAddress:  "http://" + defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
	}

	argEndpValue := flag.String("a", defaultServerAddress, "HTTP server endpoint")
	argRepValue := flag.Int64("r", defaultReportInterval, "Report interval")
	argPollValue := flag.Int64("p", defaultPollInterval, "Poll interval")
	flag.Parse()
	if argEndpValue != nil && *argEndpValue != "" {
		config.ServerAddress = "http://" + *argEndpValue
	}
	if argRepValue != nil && *argRepValue != 0 {
		config.ReportInterval = *argRepValue
	}
	if argPollValue != nil && *argPollValue != 0 {
		config.PollInterval = *argPollValue
	}

	envEndpValue, ok := os.LookupEnv("ADDRESS")
	if ok && envEndpValue != "" {
		config.ServerAddress = "http://" + envEndpValue
	}
	envRepValue, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok && envRepValue != "" {
		if val, err := strconv.ParseInt(envRepValue, 10, 64); err == nil {
			config.ReportInterval = val
		}
	}
	envPollValue, ok := os.LookupEnv("POLL_INTERVAL")
	if ok && envPollValue != "" {
		if val, err := strconv.ParseInt(envPollValue, 10, 64); err == nil {
			config.PollInterval = val
		}
	}

	return &config
}
