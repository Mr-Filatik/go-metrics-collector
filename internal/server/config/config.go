// Пакет config предоставляет функционал загрузки конфигурации из флагов командной строки и переменных окружения.
// Конфигурация включает такие параметры как: адрес сервера, настройки хранилища, путь до хранилища и т.п.
package config

import (
	"encoding/json"
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
	// Адрес сервера
	ServerAddress string `json:"server_address"`
	// Ключ хэширования
	HashKey string `json:"-"`
	// Путь до приватного ключа
	CryptoKeyPath string `json:"crypto_key"`
	// Путь до файла хранилища (относительный)
	FileStoragePath string `json:"store_file"`
	// Строка подключения к базе данных
	ConnectionString string `json:"database_dsn"`
	// Интервал сохранения данных в хранилище (в секундах)
	StoreInterval int64 `json:"store_interval"`
	// Флаг, указывающий загружать ли данные из хранилища при старте приложения
	Restore bool `json:"restore"`
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

	path := config.getConfigPathFromFlag()
	if newPath := config.getConfigPathFromEnvironment(); newPath != "" {
		path = newPath
	}

	if path != "" {
		config.loadFromJSON(path)
	}

	config.getFlags()
	config.getEnvironments()

	return &config
}

func (c *Config) loadFromJSON(path string) {
	if path == "" {
		return // файл не указан
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var fileConfig Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return
	}

	if fileConfig.ServerAddress != "" {
		c.ServerAddress = fileConfig.ServerAddress
	}
	if fileConfig.CryptoKeyPath != "" {
		c.CryptoKeyPath = fileConfig.CryptoKeyPath
	}
	if fileConfig.ConnectionString != "" {
		c.ConnectionString = fileConfig.ConnectionString
	}
	if fileConfig.FileStoragePath != "" {
		c.FileStoragePath = fileConfig.FileStoragePath
	}
	if fileConfig.StoreInterval != 0 {
		c.StoreInterval = fileConfig.StoreInterval
	}
	c.Restore = fileConfig.Restore
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

func (c *Config) getConfigPathFromFlag() string {
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c", "-config":
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}

	return ""
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

func (c *Config) getConfigPathFromEnvironment() string {
	envConfigValue, ok := os.LookupEnv("CONFIG")
	if ok && envConfigValue != "" {
		return envConfigValue
	}
	return ""
}
