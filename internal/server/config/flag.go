package config

import (
	"flag"
	"fmt"
	"os"
)

// configFlags - структура, содержащая основные флаги приложения.
type configFlags struct {
	configPath           string // путь до JSON конфига
	connString           string // строка подключения к базе данных
	cryptoKeyPath        string // путь до публичного ключа
	hashKey              string // ключ хэширования
	serverAddress        string // адрес сервера
	storagePath          string // путь до файла хранилища (относительный)
	trustedSubnet        string // разрешённые подсети
	storeInterval        int64  // интервал сохранения данных в хранилище (в секундах)
	restore              bool   // флаг, указывающий загружать ли данные из хранилища при старте приложения
	configPathIsValue    bool
	connStringIsValue    bool
	cryptoKeyPathIsValue bool
	hashKeyIsValue       bool
	serverAddressIsValue bool
	storagePathIsValue   bool
	trustedSubnetIsValue bool
	storeIntervalIsValue bool
	restoreIsValue       bool
}

// getFlagsConfig получает конфиг из указанных аргументов.
func getFlagsConfig(fs *flag.FlagSet, args []string) (*configFlags, error) {
	config := &configFlags{}

	argC := fs.String("c", "", "Path to JSON config file")
	argConfig := fs.String("config", "", "Path to JSON config file")
	argD := fs.String("d", "", "Database connection string")
	argCryptoKey := fs.String("crypto-key", "", "Public crypto key path")
	argK := fs.String("k", "", "Hash key")
	argA := fs.String("a", "", "HTTP server endpoint")
	argF := fs.String("f", "", "Path to file")
	argT := fs.String("t", "", "Trusted subnet")
	argI := fs.Int64("i", 0, "Interval in seconds to save data")
	argR := fs.Bool("r", false, "Loading data when the application starts")

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
	if argD != nil && *argD != "" {
		config.connString = *argD
		config.connStringIsValue = true
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
	if argF != nil && *argF != "" {
		config.storagePath = *argF
		config.storagePathIsValue = true
	}
	if argT != nil && *argT != "" {
		config.trustedSubnet = *argT
		config.trustedSubnetIsValue = true
	}
	if argI != nil && *argI != 0 {
		config.storeInterval = *argI
		config.storeIntervalIsValue = true
	}
	if argR != nil && *argR {
		config.restore = *argR
		config.restoreIsValue = true
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
//
//nolint:dupl // Логика похожа на overrideConfigFromEnvs, но для другого типа.
func (c *Config) overrideConfigFromFlags(conf *configFlags) {
	if conf == nil {
		return
	}

	if conf.connStringIsValue {
		c.ConnectionString = conf.connString
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
	if conf.storagePathIsValue {
		c.FileStoragePath = conf.storagePath
	}
	if conf.trustedSubnetIsValue {
		c.TrustedSubnet = conf.trustedSubnet
	}
	if conf.storeIntervalIsValue {
		c.StoreInterval = conf.storeInterval
	}
	if conf.restoreIsValue {
		c.Restore = conf.restore
	}
}
