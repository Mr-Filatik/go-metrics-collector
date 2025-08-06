package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// configJSONs - структура, содержащая основные настройки в JSON для приложения.
type configJSONs struct {
	CryptoKeyPath         string `json:"crypto_key"`
	ServerAddress         string `json:"server_address"`
	PollInterval          int64  `json:"poll_interval"`
	ReportInterval        int64  `json:"report_interval"`
	cryptoKeyPathIsValue  bool   `json:"-"`
	serverAddressIsValue  bool   `json:"-"`
	pollIntervalIsValue   bool   `json:"-"`
	reportIntervalIsValue bool   `json:"-"`
}

// initializeJSONs получает значения из JSON файла.
func initializeJSONs(path string) (*configJSONs, error) {
	config := &configJSONs{}

	if path == "" {
		return nil, errors.New("path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %w", err)
	}

	var c configJSONs
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unmarshal data from file %w", err)
	}

	if c.CryptoKeyPath != "" {
		config.cryptoKeyPathIsValue = true
	}
	if c.ServerAddress != "" {
		config.cryptoKeyPathIsValue = true
	}
	if c.PollInterval != 0 {
		config.pollIntervalIsValue = true
	}
	if c.ReportInterval != 0 {
		config.reportIntervalIsValue = true
	}

	return config, nil
}

// overrideConfig переопределяет основной конфиг новыми значениями.
func (c *Config) overrideConfigFromJSONs(conf *configJSONs) {
	if conf == nil {
		return
	}

	if conf.cryptoKeyPathIsValue {
		c.CryptoKeyPath = conf.CryptoKeyPath
	}
	if conf.serverAddressIsValue {
		c.ServerAddress = conf.ServerAddress
	}
	if conf.pollIntervalIsValue {
		c.PollInterval = conf.PollInterval
	}
	if conf.reportIntervalIsValue {
		c.ReportInterval = conf.ReportInterval
	}
}
