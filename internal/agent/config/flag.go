package config

import (
	"flag"
	"fmt"
	"os"
)

// configFlags - структура, содержащая основные флаги приложения.
type configFlags struct {
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

// getFlagsConfig получает конфиг из указанных аргументов.
func getFlagsConfig(fs *flag.FlagSet, args []string) (*configFlags, error) {
	config := &configFlags{}

	argC := fs.String("c", "", "Path to JSON config file")
	argConfig := fs.String("config", "", "Path to JSON config file")
	argCryptoKey := fs.String("crypto-key", "", "Public crypto key path")
	argK := fs.String("k", "", "Hash key")
	argA := fs.String("a", "", "HTTP server endpoint")
	argP := fs.Int64("p", 0, "Poll interval")
	argR := fs.Int64("r", 0, "Report interval")
	argL := fs.Int64("l", 0, "Rate limit")
	argG := fs.Bool("g", false, "gRPC enabled")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse argument %w", err)
	}

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
	if argG != nil {
		config.grpcEnabled = *argG
		config.grpcEnabledIsValue = true
	}

	return config, nil
}

// getFlagsConfigFromOS получает значения флагов из аргументов запуска приложения в ОС.
func getFlagsConfigFromOS() (*configFlags, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	config, err := getFlagsConfig(fs, os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("get flag config %w", err)
	}
	return config, nil
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
	if conf.grpcEnabledIsValue {
		c.GrpcEnabled = conf.grpcEnabled
	}
}
