// Пакет config предоставляет функционал загрузки конфигурации из флагов командной строки и переменных окружения.
// Конфигурация включает такие параметры как: адрес сервера, настройки хранилища, путь до хранилища и т.п.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Костанты - значения по умолчанию.
const (
	defaultServerAddress   string = "localhost:8080"          // aдрес сервера
	defaultHashKey         string = ""                        // ключ хэширования (отсутствует)
	defaultStoreInterval   int64  = 300                       // интервал сохранения данных в хранилище (в секундах)
	defaultFileStoragePath string = "../../temp_metrics.json" // путь до файла хранилища (относительный)
	// Флаг, указывающий загружать ли данные из хранилища при старте приложения.
	defaultRestore          bool   = false
	defaultConnectionString string = "" // строка подключения к базе данных
	defaultCryptoKeyPath    string = "" // путь до приватного ключа
)

// Config - структура, содержащая основные параметры приложения.
type Config struct {
	ServerAddress    string // адрес сервера
	HashKey          string // ключ хэширования
	CryptoKeyPath    string // путь до приватного ключа
	FileStoragePath  string // интервал сохранения данных в хранилище (в секундах)
	ConnectionString string // путь до файла хранилища (относительный)
	StoreInterval    int64  // флаг, указывающий загружать ли данные из хранилища при старте приложения
	Restore          bool   // строка подключения к базе данных
}

// Initialize создаёт и иницализирует объект *Config.
// Значения присваиваются в следующем порядке (переприсваивают):
//   - значения по умолчания;
//   - значения из флагов командной строки;
//   - значения из переменных окружения.
func Initialize() *Config {
	config := Config{
		ServerAddress:    defaultServerAddress,
		HashKey:          defaultHashKey,
		CryptoKeyPath:    defaultCryptoKeyPath,
		StoreInterval:    defaultStoreInterval,
		FileStoragePath:  defaultFileStoragePath,
		Restore:          defaultRestore,
		ConnectionString: defaultConnectionString,
	}

	config.getFlags()
	config.getEnvironments()

	return &config
}

func (c *Config) getFlags() {
	argEndpValue := flag.String("a", defaultServerAddress, "HTTP server endpoint")
	argKeyValue := flag.String("k", defaultHashKey, "Hash key")
	argCryptoValue := flag.String("crypto-key", defaultCryptoKeyPath, "Public crypto key path")
	argIntervalValue := flag.Int64("i", defaultStoreInterval, "Interval in seconds to save data")
	argFileValue := flag.String("f", defaultFileStoragePath, "Path to file")
	argRestoreValue := flag.Bool("r", defaultRestore, "Loading data when the application starts")
	argConnStr := flag.String("d", defaultConnectionString, "Database connection string")

	flag.Parse()

	if argEndpValue != nil && *argEndpValue != "" {
		c.ServerAddress = *argEndpValue
	}
	if argKeyValue != nil && *argKeyValue != "" {
		c.HashKey = *argKeyValue
	}
	if argCryptoValue != nil && *argCryptoValue != "" {
		c.CryptoKeyPath = *argCryptoValue
	}
	if argIntervalValue != nil && *argIntervalValue >= 0 {
		c.StoreInterval = *argIntervalValue
	}
	if argFileValue != nil && *argFileValue != "" {
		c.FileStoragePath = *argFileValue
	}
	if argRestoreValue != nil {
		c.Restore = *argRestoreValue
	}
	if argConnStr != nil && *argConnStr != "" {
		c.ConnectionString = *argConnStr
	}
}

func (c *Config) getEnvironments() {
	envEndpValue, ok := os.LookupEnv("ADDRESS")
	if ok && envEndpValue != "" {
		c.ServerAddress = envEndpValue
	}

	envKeyValue, ok := os.LookupEnv("KEY")
	if ok && envKeyValue != "" {
		c.HashKey = envKeyValue
	}

	envCryptoValue, ok := os.LookupEnv("CRYPTO_KEY")
	if ok && envCryptoValue != "" {
		c.CryptoKeyPath = envCryptoValue
	}

	envStoreValue, ok := os.LookupEnv("STORE_INTERVAL")
	if ok && envStoreValue != "" {
		if val, err := strconv.ParseInt(envStoreValue, 10, 64); err == nil {
			c.StoreInterval = val
		}
	}

	envFileValue, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok && envFileValue != "" {
		c.FileStoragePath = envFileValue
	}

	envRestoreValue, ok := os.LookupEnv("RESTORE")
	if ok && envRestoreValue != "" {
		if val, err := strconv.ParseBool(envRestoreValue); err == nil {
			c.Restore = val
		}
	}

	envConnStrValue, ok := os.LookupEnv("DATABASE_DSN")
	if ok && envConnStrValue != "" {
		c.ConnectionString = envConnStrValue
	}
}
