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
	configPathIsValue     bool
	cryptoKeyPathIsValue  bool
	hashKeyIsValue        bool
	serverAddressIsValue  bool
	pollIntervalIsValue   bool
	reportIntervalIsValue bool
	rateLimitIsValue      bool
}

// initializeEnvs получает значения из переменных окружения.
func initializeEnvs() *configEnvs {
	config := &configEnvs{}

	envConfig, ok := os.LookupEnv("CONFIG")
	if ok && envConfig != "" {
		config.configPath = envConfig
		config.configPathIsValue = true
	}

	envCryptoKey, ok := os.LookupEnv("CRYPTO_KEY")
	if ok && envCryptoKey != "" {
		config.cryptoKeyPath = envCryptoKey
		config.cryptoKeyPathIsValue = true
	}

	envKey, ok := os.LookupEnv("KEY")
	if ok && envKey != "" {
		config.hashKey = envKey
		config.hashKeyIsValue = true
	}

	envAddress, ok := os.LookupEnv("ADDRESS")
	if ok && envAddress != "" {
		config.serverAddress = envAddress
		config.serverAddressIsValue = true
	}

	envPollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if ok && envPollInterval != "" {
		if val, err := strconv.ParseInt(envPollInterval, 10, 64); err == nil {
			config.pollInterval = val
			config.pollIntervalIsValue = true
		}
	}

	envReportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok && envReportInterval != "" {
		if val, err := strconv.ParseInt(envReportInterval, 10, 64); err == nil {
			config.reportInterval = val
			config.reportIntervalIsValue = true
		}
	}

	envRateLimit, ok := os.LookupEnv("RATE_LIMIT")
	if ok && envRateLimit != "" {
		if val, err := strconv.ParseInt(envRateLimit, 10, 64); err == nil {
			config.rateLimit = val
			config.rateLimitIsValue = true
		}
	}

	return config
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
}
