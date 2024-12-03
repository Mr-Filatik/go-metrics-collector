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

	endpointEnv := os.Getenv("ADDRESS")
	var endpoint *string
	if endpointEnv == "" {
		endpoint = flag.String("a", defaultServerAddress, "HTTP server endpoint")
	} else {
		endpoint = &endpointEnv
	}

	reportIntervalEnv := os.Getenv("REPORT_INTERVAL")
	rnum, rerr := strconv.ParseInt(reportIntervalEnv, 10, 64)
	var reportInterval *int64
	if reportIntervalEnv == "" || rerr != nil {
		reportInterval = flag.Int64("r", defaultReportInterval, "Report interval")
	} else {
		reportInterval = &rnum
	}

	pollIntervalEnv := os.Getenv("POLL_INTERVAL")
	pnum, perr := strconv.ParseInt(pollIntervalEnv, 10, 64)
	var pollInterval *int64
	if pollIntervalEnv == "" || perr != nil {
		pollInterval = flag.Int64("p", defaultPollInterval, "Poll interval")
	} else {
		pollInterval = &pnum
	}
	flag.Parse()

	config := Config{
		ServerAddress:  "http://" + *endpoint,
		PollInterval:   *pollInterval,
		ReportInterval: *reportInterval,
	}
	return &config
}
