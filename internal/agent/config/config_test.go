package config

import (
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAndOverrideConfig(t *testing.T) {
	mockFileConf := &configJSONs{
		CryptoKeyPath:         "file crypto path",
		cryptoKeyPathIsValue:  true,
		ReportInterval:        88,
		reportIntervalIsValue: true,
		PollInterval:          44,
		pollIntervalIsValue:   true,
		ServerAddress:         "file address", // не указано true в serverAddressIsValue
	}
	mockFlagsConf := &configFlags{
		cryptoKeyPath:         "flags crypto path",
		cryptoKeyPathIsValue:  true,
		reportInterval:        120,
		reportIntervalIsValue: true,
		serverAddress:         "flags address", // не указано true в serverAddressIsValue
	}
	mockEnvsConf := &configEnvs{
		cryptoKeyPath:        "envs crypto path",
		cryptoKeyPathIsValue: true,
		serverAddress:        "envs address", // не указано true в serverAddressIsValue
	}

	config := createAndOverrideConfig(mockFileConf, mockFlagsConf, mockEnvsConf)

	assert.Equal(t, defaultServerAddress, config.ServerAddress)
	assert.Equal(t, defaultHashKey, config.HashKey)
	assert.Equal(t, "envs crypto path", config.CryptoKeyPath)
	assert.Equal(t, int64(44), config.PollInterval)
	assert.Equal(t, int64(120), config.ReportInterval)
	assert.Equal(t, defaultRateLimit, config.RateLimit)
}

func TestGetEnvsConfig(t *testing.T) {
	tests := []struct {
		name           string
		env            map[string]string
		expectedConfig configEnvs
	}{
		{
			name: "full values",
			env: map[string]string{
				"CONFIG":          "/config.json",
				"CRYPTO_KEY":      "/keys/public.pem",
				"KEY":             "myhashkey",
				"ADDRESS":         "example.com:8080",
				"POLL_INTERVAL":   "3",
				"REPORT_INTERVAL": "15",
				"RATE_LIMIT":      "5",
			},
			expectedConfig: configEnvs{
				configPath:            "/config.json",
				configPathIsValue:     true,
				cryptoKeyPath:         "/keys/public.pem",
				cryptoKeyPathIsValue:  true,
				hashKey:               "myhashkey",
				hashKeyIsValue:        true,
				serverAddress:         "example.com:8080",
				serverAddressIsValue:  true,
				pollInterval:          3,
				pollIntervalIsValue:   true,
				reportInterval:        15,
				reportIntervalIsValue: true,
				rateLimit:             5,
				rateLimitIsValue:      true,
			},
		},
		{
			name: "partial values",
			env: map[string]string{
				"ADDRESS":       "localhost:9090",
				"POLL_INTERVAL": "2",
			},
			expectedConfig: configEnvs{
				configPath:            "",
				configPathIsValue:     false,
				cryptoKeyPath:         "",
				cryptoKeyPathIsValue:  false,
				hashKey:               "",
				hashKeyIsValue:        false,
				serverAddress:         "localhost:9090",
				serverAddressIsValue:  true,
				pollInterval:          2,
				pollIntervalIsValue:   true,
				reportInterval:        0,
				reportIntervalIsValue: false,
				rateLimit:             0,
				rateLimitIsValue:      false,
			},
		},
		{
			name: "empty values",
			env:  map[string]string{},
			expectedConfig: configEnvs{
				configPath:            "",
				configPathIsValue:     false,
				cryptoKeyPath:         "",
				cryptoKeyPathIsValue:  false,
				hashKey:               "",
				hashKeyIsValue:        false,
				serverAddress:         "",
				serverAddressIsValue:  false,
				pollInterval:          0,
				pollIntervalIsValue:   false,
				reportInterval:        0,
				reportIntervalIsValue: false,
				rateLimit:             0,
				rateLimitIsValue:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEnv := func(key string) (string, bool) {
				val, ok := tt.env[key]
				return val, ok
			}

			config := getEnvsConfig(mockEnv)

			assert.Equal(t, tt.expectedConfig.configPath, config.configPath)
			assert.Equal(t, tt.expectedConfig.configPathIsValue, config.configPathIsValue)

			assert.Equal(t, tt.expectedConfig.cryptoKeyPath, config.cryptoKeyPath)
			assert.Equal(t, tt.expectedConfig.cryptoKeyPathIsValue, config.cryptoKeyPathIsValue)

			assert.Equal(t, tt.expectedConfig.hashKey, config.hashKey)
			assert.Equal(t, tt.expectedConfig.hashKeyIsValue, config.hashKeyIsValue)

			assert.Equal(t, tt.expectedConfig.serverAddress, config.serverAddress)
			assert.Equal(t, tt.expectedConfig.serverAddressIsValue, config.serverAddressIsValue)

			assert.Equal(t, tt.expectedConfig.pollInterval, config.pollInterval)
			assert.Equal(t, tt.expectedConfig.pollIntervalIsValue, config.pollIntervalIsValue)

			assert.Equal(t, tt.expectedConfig.reportInterval, config.reportInterval)
			assert.Equal(t, tt.expectedConfig.reportIntervalIsValue, config.reportIntervalIsValue)

			assert.Equal(t, tt.expectedConfig.rateLimit, config.rateLimit)
			assert.Equal(t, tt.expectedConfig.rateLimitIsValue, config.rateLimitIsValue)
		})
	}
}

func TestGetFlagsConfig(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedConfig configFlags
	}{
		{
			name: "full values",
			args: []string{
				"-c", "/config.json",
				"-crypto-key", "/keys/public.pem",
				"-k", "myhashkey",
				"-a", "example.com:8080",
				"-p", "3",
				"-r", "15",
				"-l", "5",
			},
			expectedConfig: configFlags{
				configPath:            "/config.json",
				configPathIsValue:     true,
				cryptoKeyPath:         "/keys/public.pem",
				cryptoKeyPathIsValue:  true,
				hashKey:               "myhashkey",
				hashKeyIsValue:        true,
				serverAddress:         "example.com:8080",
				serverAddressIsValue:  true,
				pollInterval:          3,
				pollIntervalIsValue:   true,
				reportInterval:        15,
				reportIntervalIsValue: true,
				rateLimit:             5,
				rateLimitIsValue:      true,
			},
		},
		{
			name: "partial values",
			args: []string{
				"-a", "localhost:9090",
				"-p", "2",
				"-l", "1",
			},
			expectedConfig: configFlags{
				configPath:            "",
				configPathIsValue:     false,
				cryptoKeyPath:         "",
				cryptoKeyPathIsValue:  false,
				hashKey:               "",
				hashKeyIsValue:        false,
				serverAddress:         "localhost:9090",
				serverAddressIsValue:  true,
				pollInterval:          2,
				pollIntervalIsValue:   true,
				reportInterval:        0,
				reportIntervalIsValue: false,
				rateLimit:             1,
				rateLimitIsValue:      true,
			},
		},
		{
			name: "empty values",
			args: []string{},
			expectedConfig: configFlags{
				configPath:            "",
				configPathIsValue:     false,
				cryptoKeyPath:         "",
				cryptoKeyPathIsValue:  false,
				hashKey:               "",
				hashKeyIsValue:        false,
				serverAddress:         "",
				serverAddressIsValue:  false,
				pollInterval:          0,
				pollIntervalIsValue:   false,
				reportInterval:        0,
				reportIntervalIsValue: false,
				rateLimit:             0,
				rateLimitIsValue:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)

			config, err := getFlagsConfig(fs, tt.args)
			require.NoError(t, err)
			require.NotNil(t, config)

			assert.Equal(t, tt.expectedConfig.configPath, config.configPath)
			assert.Equal(t, tt.expectedConfig.configPathIsValue, config.configPathIsValue)

			assert.Equal(t, tt.expectedConfig.cryptoKeyPath, config.cryptoKeyPath)
			assert.Equal(t, tt.expectedConfig.cryptoKeyPathIsValue, config.cryptoKeyPathIsValue)

			assert.Equal(t, tt.expectedConfig.hashKey, config.hashKey)
			assert.Equal(t, tt.expectedConfig.hashKeyIsValue, config.hashKeyIsValue)

			assert.Equal(t, tt.expectedConfig.serverAddress, config.serverAddress)
			assert.Equal(t, tt.expectedConfig.serverAddressIsValue, config.serverAddressIsValue)

			assert.Equal(t, tt.expectedConfig.pollInterval, config.pollInterval)
			assert.Equal(t, tt.expectedConfig.pollIntervalIsValue, config.pollIntervalIsValue)

			assert.Equal(t, tt.expectedConfig.reportInterval, config.reportInterval)
			assert.Equal(t, tt.expectedConfig.reportIntervalIsValue, config.reportIntervalIsValue)

			assert.Equal(t, tt.expectedConfig.rateLimit, config.rateLimit)
			assert.Equal(t, tt.expectedConfig.rateLimitIsValue, config.rateLimitIsValue)
		})
	}
}

func TestGetJSONConfig(t *testing.T) {
	tests := []struct {
		name            string
		jsonInput       string
		expectedAddress string
		expectedCrypto  string
		expectedPoll    int64
		expectedReport  int64
		reportIsSet     bool
		addressIsSet    bool
		pollIsSet       bool
		cryptoIsSet     bool
	}{
		{
			name: "full config",
			jsonInput: `{
				"server_address": "localhost:8080", 
				"crypto_key": "/keys/public.pem", 
				"poll_interval": 5, 
				"report_interval": 10
			}`,
			expectedAddress: "localhost:8080",
			addressIsSet:    true,
			expectedCrypto:  "/keys/public.pem",
			cryptoIsSet:     true,
			expectedPoll:    5,
			pollIsSet:       true,
			expectedReport:  10,
			reportIsSet:     true,
		},
		{
			name: "partial config",
			jsonInput: `{
				"server_address": "localhost:8080",
				"poll_interval": 5}`,
			expectedAddress: "localhost:8080",
			addressIsSet:    true,
			expectedCrypto:  "",
			cryptoIsSet:     false,
			expectedPoll:    5,
			pollIsSet:       true,
			expectedReport:  0,
			reportIsSet:     false,
		},
		{
			name:            "empty config",
			jsonInput:       `{}`,
			expectedAddress: "",
			addressIsSet:    false,
			expectedCrypto:  "",
			cryptoIsSet:     false,
			expectedPoll:    0,
			pollIsSet:       false,
			expectedReport:  0,
			reportIsSet:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := getJSONConfig(strings.NewReader(tt.jsonInput))
			require.NoError(t, err)

			assert.Equal(t, tt.expectedAddress, config.ServerAddress)
			assert.Equal(t, tt.addressIsSet, config.serverAddressIsValue)

			assert.Equal(t, tt.expectedCrypto, config.CryptoKeyPath)
			assert.Equal(t, tt.cryptoIsSet, config.cryptoKeyPathIsValue)

			assert.Equal(t, tt.expectedPoll, config.PollInterval)
			assert.Equal(t, tt.pollIsSet, config.pollIntervalIsValue)

			assert.Equal(t, tt.expectedReport, config.ReportInterval)
			assert.Equal(t, tt.reportIsSet, config.reportIntervalIsValue)
		})
	}
}

func TestStripHTTPPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with http",
			input:    "http://localhost:8080",
			expected: "localhost:8080",
		},
		{
			name:     "with https",
			input:    "https://example.com",
			expected: "example.com",
		},
		{
			name:     "without prefix",
			input:    "localhost:8080",
			expected: "localhost:8080",
		},
		{
			name:     "only http://",
			input:    "http://",
			expected: "",
		},
		{
			name:     "only https://",
			input:    "https://",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "with path",
			input:    "http://localhost:8080/metrics",
			expected: "localhost:8080/metrics",
		},
		{
			name:     "with query",
			input:    "https://api.com/v1?token=abc",
			expected: "api.com/v1?token=abc",
		},
		{
			name:     "contains http but not prefix",
			input:    "myhttp://example.com",
			expected: "myhttp://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTTPPrefix(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
