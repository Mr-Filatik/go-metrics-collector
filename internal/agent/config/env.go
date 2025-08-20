package config

import (
	"os"
	"strconv"
)

// configEnvs - структура, содержащая основные переменные окружения для приложения.
type configEnvs struct {
	configPath            string // путь до JSON конфига
	cryptoKeyPath         string // путь до публичного ключа
	hashKey               string // ключ хэширования
	serverAddress         string // адрес сервера
	pollInterval          int64  // интервал опроса (в секундах)
	reportInterval        int64  // интервал отправки данных (в секундах)
	rateLimit             int64  // лимит запросов для агента
	grpcEnabled           bool   // включать ли поддержку gRPC
	configPathIsValue     bool
	cryptoKeyPathIsValue  bool
	hashKeyIsValue        bool
	serverAddressIsValue  bool
	pollIntervalIsValue   bool
	reportIntervalIsValue bool
	rateLimitIsValue      bool
	grpcEnabledIsValue    bool
}

// envReader — интерфейс для чтения переменных окружения.
type envReader func(key string) (string, bool)

// getEnvsConfig получает значения из универсального хранилища.
func getEnvsConfig(getenv envReader) *configEnvs {
	config := &configEnvs{}

	envConfig, ok := getenv("CONFIG")
	if ok && envConfig != "" {
		config.configPath = envConfig
		config.configPathIsValue = true
	}

	envCryptoKey, ok := getenv("CRYPTO_KEY")
	if ok && envCryptoKey != "" {
		config.cryptoKeyPath = envCryptoKey
		config.cryptoKeyPathIsValue = true
	}

	envKey, ok := getenv("KEY")
	if ok && envKey != "" {
		config.hashKey = envKey
		config.hashKeyIsValue = true
	}

	envAddress, ok := getenv("ADDRESS")
	if ok && envAddress != "" {
		config.serverAddress = envAddress
		config.serverAddressIsValue = true
	}

	envPollInterval, ok := getenv("POLL_INTERVAL")
	if ok && envPollInterval != "" {
		if val, err := strconv.ParseInt(envPollInterval, 10, 64); err == nil {
			config.pollInterval = val
			config.pollIntervalIsValue = true
		}
	}

	envReportInterval, ok := getenv("REPORT_INTERVAL")
	if ok && envReportInterval != "" {
		if val, err := strconv.ParseInt(envReportInterval, 10, 64); err == nil {
			config.reportInterval = val
			config.reportIntervalIsValue = true
		}
	}

	envRateLimit, ok := getenv("RATE_LIMIT")
	if ok && envRateLimit != "" {
		if val, err := strconv.ParseInt(envRateLimit, 10, 64); err == nil {
			config.rateLimit = val
			config.rateLimitIsValue = true
		}
	}

	envGrpcEnabled, ok := getenv("GRPC_ENABLED")
	if ok && envGrpcEnabled != "" {
		if val, err := strconv.ParseBool(envGrpcEnabled); err == nil {
			config.grpcEnabled = val
			config.grpcEnabledIsValue = true
		}
	}

	return config
}

// getEnvsConfigFromOS получает значения из переменных окружения.
func getEnvsConfigFromOS() *configEnvs {
	return getEnvsConfig(func(key string) (string, bool) {
		value, ok := os.LookupEnv(key)
		return value, ok
	})
}

// overrideConfigFromEnvs переопределяет основной конфиг новыми значениями.
func (c *Config) overrideConfigFromEnvs(conf *configEnvs) {
	if conf == nil {
		return
	}

	if conf.cryptoKeyPathIsValue {
		c.CryptoKeyPath = conf.cryptoKeyPath
	}
	if conf.hashKeyIsValue {
		c.HashKey = conf.hashKey
	}
	if conf.serverAddressIsValue {
		c.ServerAddress = conf.serverAddress
	}
	if conf.pollIntervalIsValue {
		c.PollInterval = conf.pollInterval
	}
	if conf.reportIntervalIsValue {
		c.ReportInterval = conf.reportInterval
	}
	if conf.rateLimitIsValue {
		c.RateLimit = conf.rateLimit
	}
	if conf.grpcEnabledIsValue {
		c.GrpcEnabled = conf.grpcEnabled
	}
}
