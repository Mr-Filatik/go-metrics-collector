package config

import "flag"

// configFlags - структура, содержащая основные флаги приложения.
type configFlags struct {
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

// initializeFlags получает значения из аргументов командной строки.
func initializeFlags() *configFlags {
	config := &configFlags{}

	argC := flag.String("c", "", "Path to JSON config file")
	argConfig := flag.String("config", "", "Path to JSON config file")
	argCryptoKey := flag.String("crypto-key", defaultCryptoKeyPath, "Public crypto key path")
	argK := flag.String("k", defaultHashKey, "Hash key")
	argA := flag.String("a", defaultServerAddress, "HTTP server endpoint")
	argP := flag.Int64("p", defaultPollInterval, "Poll interval")
	argR := flag.Int64("r", defaultReportInterval, "Report interval")
	argL := flag.Int64("l", defaultRateLimit, "Rate limit")

	flag.Parse()

	if argC != nil && *argC != "" {
		config.configPath = *argC
		config.configPathIsValue = true
	}
	if argConfig != nil && *argConfig != "" {
		config.configPath = *argConfig
		config.configPathIsValue = true
	}
	if argCryptoKey != nil && *argCryptoKey != "" {
		config.cryptoKeyPath = *argCryptoKey
		config.cryptoKeyPathIsValue = true
	}
	if argK != nil && *argK != "" {
		config.hashKey = *argK
		config.hashKeyIsValue = true
	}
	if argA != nil && *argA != "" {
		config.serverAddress = *argA
		config.serverAddressIsValue = true
	}
	if argP != nil && *argP != 0 {
		config.pollInterval = *argP
		config.pollIntervalIsValue = true
	}
	if argR != nil && *argR != 0 {
		config.reportInterval = *argR
		config.reportIntervalIsValue = true
	}
	if argL != nil && *argL != 0 {
		config.rateLimit = *argL
		config.rateLimitIsValue = true
	}

	return config
}

// overrideConfigFromFlags переопределяет основной конфиг новыми значениями.
func (c *Config) overrideConfigFromFlags(conf *configFlags) {
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
