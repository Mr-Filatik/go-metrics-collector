package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// configJSONs - структура, содержащая основные настройки в JSON для приложения.
type configJSONs struct {
	CryptoKeyPath         string `json:"crypto_key,omitempty"`
	ServerAddress         string `json:"server_address,omitempty"`
	PollInterval          int64  `json:"poll_interval,omitempty"`
	ReportInterval        int64  `json:"report_interval,omitempty"`
	cryptoKeyPathIsValue  bool   `json:"-"`
	serverAddressIsValue  bool   `json:"-"`
	pollIntervalIsValue   bool   `json:"-"`
	reportIntervalIsValue bool   `json:"-"`
}

// getJSONConfig получает конфиг из универсального io.Reader.
func getJSONConfig(r io.Reader) (*configJSONs, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read error %w", err)
	}

	var c configJSONs
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unmarshal data from file %w", err)
	}

	config := &configJSONs{}

	if c.CryptoKeyPath != "" {
		config.CryptoKeyPath = c.CryptoKeyPath
		config.cryptoKeyPathIsValue = true
	}
	if c.ServerAddress != "" {
		config.ServerAddress = c.ServerAddress
		config.serverAddressIsValue = true
	}
	if c.PollInterval != 0 {
		config.PollInterval = c.PollInterval
		config.pollIntervalIsValue = true
	}
	if c.ReportInterval != 0 {
		config.ReportInterval = c.ReportInterval
		config.reportIntervalIsValue = true
	}

	return config, nil
}

// getJSONConfigFromFile получает конфиг из файла с форматом хранения JSON.
func getJSONConfigFromFile(path string) (*configJSONs, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return getJSONConfig(file)
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
