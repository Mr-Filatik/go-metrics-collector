package config

import (
	"os"
	"strconv"
)

// configEnvs - структура, содержащая основные переменные окружения для приложения.
type configEnvs struct {
	configPath           string // путь до JSON конфига
	connString           string // строка подключения к базе данных
	cryptoKeyPath        string // путь до публичного ключа
	hashKey              string // ключ хэширования
	serverAddress        string // адрес сервера
	storagePath          string // путь до файла хранилища (относительный)
	storeInterval        int64  // интервал сохранения данных в хранилище (в секундах)
	restore              bool   // флаг, указывающий загружать ли данные из хранилища при старте приложения
	configPathIsValue    bool
	connStringIsValue    bool
	cryptoKeyPathIsValue bool
	hashKeyIsValue       bool
	serverAddressIsValue bool
	storagePathIsValue   bool
	storeIntervalIsValue bool
	restoreIsValue       bool
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

	envDatabaseDSN, ok := getenv("DATABASE_DSN")
	if ok && envDatabaseDSN != "" {
		config.connString = envDatabaseDSN
		config.connStringIsValue = true
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

	envFileStoragePath, ok := getenv("FILE_STORAGE_PATH")
	if ok && envFileStoragePath != "" {
		config.storagePath = envFileStoragePath
		config.storagePathIsValue = true
	}

	envReportInterval, ok := getenv("STORE_INTERVAL")
	if ok && envReportInterval != "" {
		if val, err := strconv.ParseInt(envReportInterval, 10, 64); err == nil {
			config.storeInterval = val
			config.storeIntervalIsValue = true
		}
	}

	envRestore, ok := getenv("RESTORE")
	if ok && envRestore != "" {
		if val, err := strconv.ParseBool(envRestore); err == nil {
			config.restore = val
			config.restoreIsValue = true
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
//
//nolint:dupl // Логика похожа на overrideConfigFromFlags, но для другого типа.
func (c *Config) overrideConfigFromEnvs(conf *configEnvs) {
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
	if conf.storeIntervalIsValue {
		c.StoreInterval = conf.storeInterval
	}
	if conf.restoreIsValue {
		c.Restore = conf.restore
	}
}
