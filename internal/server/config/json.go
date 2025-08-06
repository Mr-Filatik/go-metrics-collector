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
	ConnString           string `json:"database_dsn,omitempty"`
	CryptoKeyPath        string `json:"crypto_key,omitempty"`
	ServerAddress        string `json:"address,omitempty"`
	StoragePath          string `json:"store_file,omitempty"`
	StoreInterval        int64  `json:"store_interval,omitempty"`
	Restore              bool   `json:"restore,omitempty"`
	connStringIsValue    bool   `json:"-"`
	cryptoKeyPathIsValue bool   `json:"-"`
	serverAddressIsValue bool   `json:"-"`
	storagePathIsValue   bool   `json:"-"`
	storeIntervalIsValue bool   `json:"-"`
	restoreIsValue       bool   `json:"-"`
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

	if c.ConnString != "" {
		config.ConnString = c.ConnString
		config.connStringIsValue = true
	}
	if c.CryptoKeyPath != "" {
		config.CryptoKeyPath = c.CryptoKeyPath
		config.cryptoKeyPathIsValue = true
	}
	if c.ServerAddress != "" {
		config.ServerAddress = c.ServerAddress
		config.serverAddressIsValue = true
	}
	if c.StoragePath != "" {
		config.StoragePath = c.StoragePath
		config.storagePathIsValue = true
	}
	if c.StoreInterval != 0 {
		config.StoreInterval = c.StoreInterval
		config.storeIntervalIsValue = true
	}
	if !c.Restore {
		config.Restore = c.Restore
		config.restoreIsValue = true
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

	if conf.connStringIsValue {
		c.ConnectionString = conf.ConnString
	}
	if conf.cryptoKeyPathIsValue {
		c.CryptoKeyPath = conf.CryptoKeyPath
	}
	if conf.serverAddressIsValue {
		c.ServerAddress = conf.ServerAddress
	}
	if conf.storagePathIsValue {
		c.FileStoragePath = conf.StoragePath
	}
	if conf.storeIntervalIsValue {
		c.StoreInterval = conf.StoreInterval
	}
	if conf.restoreIsValue {
		c.Restore = conf.Restore
	}
}
