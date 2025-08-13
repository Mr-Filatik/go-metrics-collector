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
		CryptoKeyPath:        "file crypto path",
		cryptoKeyPathIsValue: true,
		StoreInterval:        88,
		storeIntervalIsValue: true,
		ConnString:           "my database",
		connStringIsValue:    true,
		ServerAddress:        "file address", // не указано true в serverAddressIsValue
	}
	mockFlagsConf := &configFlags{
		cryptoKeyPath:        "flags crypto path",
		cryptoKeyPathIsValue: true,
		storeInterval:        120,
		storeIntervalIsValue: true,
		serverAddress:        "flags address", // не указано true в serverAddressIsValue
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
	assert.Equal(t, "my database", config.ConnectionString)
	assert.Equal(t, int64(120), config.StoreInterval)
	assert.Equal(t, defaultFileStoragePath, config.FileStoragePath)
	assert.Equal(t, defaultRestore, config.Restore)
}

func TestGetEnvsConfig(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected configEnvs
	}{
		{
			name: "full values",
			env: map[string]string{
				"CONFIG":            "/config.json",
				"CRYPTO_KEY":        "/keys/public.pem",
				"DATABASE_DSN":      "my database",
				"KEY":               "myhashkey",
				"ADDRESS":           "example.com:8080",
				"FILE_STORAGE_PATH": "my storage",
				"TRUSTED_SUBNET":    "mylocalhost",
				"STORE_INTERVAL":    "15",
				"RESTORE":           "true",
			},
			expected: configEnvs{
				configPath:           "/config.json",
				configPathIsValue:    true,
				connString:           "my database",
				connStringIsValue:    true,
				cryptoKeyPath:        "/keys/public.pem",
				cryptoKeyPathIsValue: true,
				hashKey:              "myhashkey",
				hashKeyIsValue:       true,
				serverAddress:        "example.com:8080",
				serverAddressIsValue: true,
				storagePath:          "my storage",
				storagePathIsValue:   true,
				trustedSubnet:        "mylocalhost",
				trustedSubnetIsValue: true,
				storeInterval:        15,
				storeIntervalIsValue: true,
				restore:              true,
				restoreIsValue:       true,
			},
		},
		{
			name: "partial values",
			env: map[string]string{
				"ADDRESS":        "example.com:8080",
				"DATABASE_DSN":   "my database",
				"STORE_INTERVAL": "15",
				"TRUSTED_SUBNET": "mylocalhost",
			},
			expected: configEnvs{
				configPath:           "",
				configPathIsValue:    false,
				connString:           "my database",
				connStringIsValue:    true,
				cryptoKeyPath:        "",
				cryptoKeyPathIsValue: false,
				hashKey:              "",
				hashKeyIsValue:       false,
				serverAddress:        "example.com:8080",
				serverAddressIsValue: true,
				storagePath:          "",
				trustedSubnet:        "mylocalhost",
				trustedSubnetIsValue: true,
				storagePathIsValue:   false,
				storeInterval:        15,
				storeIntervalIsValue: true,
				restore:              false,
				restoreIsValue:       false,
			},
		},
		{
			name: "empty values",
			env:  map[string]string{},
			expected: configEnvs{
				configPath:           "",
				configPathIsValue:    false,
				connString:           "",
				connStringIsValue:    false,
				cryptoKeyPath:        "",
				cryptoKeyPathIsValue: false,
				hashKey:              "",
				hashKeyIsValue:       false,
				serverAddress:        "",
				serverAddressIsValue: false,
				storagePath:          "",
				storagePathIsValue:   false,
				trustedSubnet:        "",
				trustedSubnetIsValue: false,
				storeInterval:        0,
				storeIntervalIsValue: false,
				restore:              false,
				restoreIsValue:       false,
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

			assert.Equal(t, tt.expected.configPath, config.configPath)
			assert.Equal(t, tt.expected.configPathIsValue, config.configPathIsValue)

			assert.Equal(t, tt.expected.cryptoKeyPath, config.cryptoKeyPath)
			assert.Equal(t, tt.expected.cryptoKeyPathIsValue, config.cryptoKeyPathIsValue)

			assert.Equal(t, tt.expected.hashKey, config.hashKey)
			assert.Equal(t, tt.expected.hashKeyIsValue, config.hashKeyIsValue)

			assert.Equal(t, tt.expected.serverAddress, config.serverAddress)
			assert.Equal(t, tt.expected.serverAddressIsValue, config.serverAddressIsValue)

			assert.Equal(t, tt.expected.connString, config.connString)
			assert.Equal(t, tt.expected.connStringIsValue, config.connStringIsValue)

			assert.Equal(t, tt.expected.storagePath, config.storagePath)
			assert.Equal(t, tt.expected.storagePathIsValue, config.storagePathIsValue)

			assert.Equal(t, tt.expected.storeInterval, config.storeInterval)
			assert.Equal(t, tt.expected.storeIntervalIsValue, config.storeIntervalIsValue)

			assert.Equal(t, tt.expected.restore, config.restore)
			assert.Equal(t, tt.expected.restoreIsValue, config.restoreIsValue)
		})
	}
}

func TestGetFlagsConfig(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected configFlags
	}{
		{
			name: "full values",
			args: []string{
				"-c", "/config.json",
				"-crypto-key", "/keys/public.pem",
				"-k", "myhashkey",
				"-a", "example.com:8080",
				"-d", "my database",
				"-f", "my storage",
				"-t", "mylocalhost",
				"-i", "15",
				"-r", "true",
			},
			expected: configFlags{
				configPath:           "/config.json",
				configPathIsValue:    true,
				connString:           "my database",
				connStringIsValue:    true,
				cryptoKeyPath:        "/keys/public.pem",
				cryptoKeyPathIsValue: true,
				hashKey:              "myhashkey",
				hashKeyIsValue:       true,
				serverAddress:        "example.com:8080",
				serverAddressIsValue: true,
				storagePath:          "my storage",
				storagePathIsValue:   true,
				trustedSubnet:        "mylocalhost",
				trustedSubnetIsValue: true,
				storeInterval:        15,
				storeIntervalIsValue: true,
				restore:              true,
				restoreIsValue:       true,
			},
		},
		{
			name: "partial values",
			args: []string{
				"-a", "example.com:8080",
				"-d", "my database",
				"-i", "15",
				"-t", "mylocalhost",
			},
			expected: configFlags{
				configPath:           "",
				configPathIsValue:    false,
				connString:           "my database",
				connStringIsValue:    true,
				cryptoKeyPath:        "",
				cryptoKeyPathIsValue: false,
				hashKey:              "",
				hashKeyIsValue:       false,
				serverAddress:        "example.com:8080",
				serverAddressIsValue: true,
				storagePath:          "",
				storagePathIsValue:   false,
				trustedSubnet:        "mylocalhost",
				trustedSubnetIsValue: true,
				storeInterval:        15,
				storeIntervalIsValue: true,
				restore:              false,
				restoreIsValue:       false,
			},
		},
		{
			name: "empty values",
			args: []string{},
			expected: configFlags{
				configPath:           "",
				configPathIsValue:    false,
				connString:           "",
				connStringIsValue:    false,
				cryptoKeyPath:        "",
				cryptoKeyPathIsValue: false,
				hashKey:              "",
				hashKeyIsValue:       false,
				serverAddress:        "",
				serverAddressIsValue: false,
				storagePath:          "",
				storagePathIsValue:   false,
				trustedSubnet:        "",
				trustedSubnetIsValue: false,
				storeInterval:        0,
				storeIntervalIsValue: false,
				restore:              false,
				restoreIsValue:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)

			config, err := getFlagsConfig(fs, tt.args)
			require.NoError(t, err)
			require.NotNil(t, config)

			assert.Equal(t, tt.expected.configPath, config.configPath)
			assert.Equal(t, tt.expected.configPathIsValue, config.configPathIsValue)

			assert.Equal(t, tt.expected.cryptoKeyPath, config.cryptoKeyPath)
			assert.Equal(t, tt.expected.cryptoKeyPathIsValue, config.cryptoKeyPathIsValue)

			assert.Equal(t, tt.expected.hashKey, config.hashKey)
			assert.Equal(t, tt.expected.hashKeyIsValue, config.hashKeyIsValue)

			assert.Equal(t, tt.expected.serverAddress, config.serverAddress)
			assert.Equal(t, tt.expected.serverAddressIsValue, config.serverAddressIsValue)

			assert.Equal(t, tt.expected.connString, config.connString)
			assert.Equal(t, tt.expected.connStringIsValue, config.connStringIsValue)

			assert.Equal(t, tt.expected.storagePath, config.storagePath)
			assert.Equal(t, tt.expected.storagePathIsValue, config.storagePathIsValue)

			assert.Equal(t, tt.expected.storeInterval, config.storeInterval)
			assert.Equal(t, tt.expected.storeIntervalIsValue, config.storeIntervalIsValue)

			assert.Equal(t, tt.expected.restore, config.restore)
			assert.Equal(t, tt.expected.restoreIsValue, config.restoreIsValue)
		})
	}
}

func TestGetJSONConfig(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected configJSONs
	}{
		{
			name: "full config",
			jsonData: `{
				"address": "localhost:8080",
				"crypto_key": "/keys/public.pem",
				"store_file": "/path/to/file.db",
				"trusted_subnet": "mylocalhost",
				"store_interval": 12,
				"database_dsn": "my database",
				"restore": true
			}`,
			expected: configJSONs{
				ServerAddress:        "localhost:8080",
				serverAddressIsValue: true,
				CryptoKeyPath:        "/keys/public.pem",
				cryptoKeyPathIsValue: true,
				StoragePath:          "/path/to/file.db",
				storagePathIsValue:   true,
				TrustedSubnet:        "mylocalhost",
				trustedSubnetIsValue: true,
				StoreInterval:        12,
				storeIntervalIsValue: true,
				ConnString:           "my database",
				connStringIsValue:    true,
				Restore:              true,
				restoreIsValue:       true,
			},
		},
		{
			name: "partial config",
			jsonData: `{
				"address": "localhost:8080",
				"store_interval": 12,
				"trusted_subnet": "mylocalhost",
				"restore": true
			}`,
			expected: configJSONs{
				ServerAddress:        "localhost:8080",
				serverAddressIsValue: true,
				CryptoKeyPath:        "",
				cryptoKeyPathIsValue: false,
				StoragePath:          "",
				storagePathIsValue:   false,
				TrustedSubnet:        "mylocalhost",
				trustedSubnetIsValue: true,
				StoreInterval:        12,
				storeIntervalIsValue: true,
				ConnString:           "",
				connStringIsValue:    false,
				Restore:              true,
				restoreIsValue:       true,
			},
		},
		{
			name:     "empty config",
			jsonData: `{}`,
			expected: configJSONs{
				ServerAddress:        "",
				serverAddressIsValue: false,
				CryptoKeyPath:        "",
				cryptoKeyPathIsValue: false,
				StoragePath:          "",
				storagePathIsValue:   false,
				TrustedSubnet:        "",
				trustedSubnetIsValue: false,
				StoreInterval:        0,
				storeIntervalIsValue: false,
				ConnString:           "",
				connStringIsValue:    false,
				Restore:              false,
				restoreIsValue:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := getJSONConfig(strings.NewReader(tt.jsonData))
			require.NoError(t, err)

			assert.Equal(t, tt.expected.ServerAddress, config.ServerAddress)
			assert.Equal(t, tt.expected.serverAddressIsValue, config.serverAddressIsValue)

			assert.Equal(t, tt.expected.CryptoKeyPath, config.CryptoKeyPath)
			assert.Equal(t, tt.expected.cryptoKeyPathIsValue, config.cryptoKeyPathIsValue)

			assert.Equal(t, tt.expected.ConnString, config.ConnString)
			assert.Equal(t, tt.expected.connStringIsValue, config.connStringIsValue)

			assert.Equal(t, tt.expected.StoragePath, config.StoragePath)
			assert.Equal(t, tt.expected.storagePathIsValue, config.storagePathIsValue)

			assert.Equal(t, tt.expected.StoreInterval, config.StoreInterval)
			assert.Equal(t, tt.expected.storeIntervalIsValue, config.storeIntervalIsValue)

			assert.Equal(t, tt.expected.Restore, config.Restore)
			assert.Equal(t, tt.expected.restoreIsValue, config.restoreIsValue)
		})
	}
}
