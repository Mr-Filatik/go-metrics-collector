package config

import (
	"flag"
	"os"
	"strconv"
)

const (
	defaultServerAddress  string = "localhost:8080"
	defaultHashKey        string = ""
	defaultPollInterval   int64  = 2
	defaultReportInterval int64  = 10
	defaultRateLimit      int64  = 1
)

type Config struct {
	ServerAddress  string
	HashKey        string
	PollInterval   int64
	ReportInterval int64
	RateLimit      int64
}

func Initialize() *Config {
	config := Config{
		ServerAddress:  "http://" + defaultServerAddress,
		HashKey:        defaultHashKey,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		RateLimit:      defaultRateLimit,
	}

	config.getFlags()
	config.getEnvironments()

	return &config
}

func (c *Config) getFlags() {
	argEndpValue := flag.String("a", defaultServerAddress, "HTTP server endpoint")
	argKeyValue := flag.String("k", defaultHashKey, "Hash key")
	argRepValue := flag.Int64("r", defaultReportInterval, "Report interval")
	argPollValue := flag.Int64("p", defaultPollInterval, "Poll interval")
	argLimitValue := flag.Int64("l", defaultPollInterval, "Rate limit")

	flag.Parse()

	if argEndpValue != nil && *argEndpValue != "" {
		c.ServerAddress = "http://" + *argEndpValue
	}
	if argKeyValue != nil && *argKeyValue != "" {
		c.HashKey = *argKeyValue
	}
	if argRepValue != nil && *argRepValue != 0 {
		c.ReportInterval = *argRepValue
	}
	if argPollValue != nil && *argPollValue != 0 {
		c.PollInterval = *argPollValue
	}
	if argLimitValue != nil && *argLimitValue != 0 {
		c.RateLimit = *argLimitValue
	}
}

func (c *Config) getEnvironments() {
	envEndpValue, ok := os.LookupEnv("ADDRESS")
	if ok && envEndpValue != "" {
		c.ServerAddress = "http://" + envEndpValue
	}

	envKeyValue, ok := os.LookupEnv("KEY")
	if ok && envKeyValue != "" {
		c.HashKey = envKeyValue
	}

	envRepValue, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok && envRepValue != "" {
		if val, err := strconv.ParseInt(envRepValue, 10, 64); err == nil {
			c.ReportInterval = val
		}
	}

	envPollValue, ok := os.LookupEnv("POLL_INTERVAL")
	if ok && envPollValue != "" {
		if val, err := strconv.ParseInt(envPollValue, 10, 64); err == nil {
			c.PollInterval = val
		}
	}

	envLimitValue, ok := os.LookupEnv("RATE_LIMIT")
	if ok && envLimitValue != "" {
		if val, err := strconv.ParseInt(envLimitValue, 10, 64); err == nil {
			c.RateLimit = val
		}
	}
}
