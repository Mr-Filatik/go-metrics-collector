// Пакет config предоставляет функционал загрузки конфигурации из флагов командной строки и переменных окружения.
// Конфигурация включает такие параметры как: адрес сервера, интервал опроса, интервал отправки и т.п.
package config

import (
	"flag"
	"os"
	"strconv"
)

// Костанты - значения по умолчанию.
const (
	defaultServerAddress  string = "localhost:8080" // адрес сервера
	defaultHashKey        string = ""               // ключ хэширования (отсутствует)
	defaultPollInterval   int64  = 2                // интервал опроса (в секундах)
	defaultReportInterval int64  = 10               // интервал отправки данных (в секундах)
	defaultRateLimit      int64  = 1                // лимит запросов для агента
	defaultCryptoKeyPath  string = ""               // путь до публичного ключа
)

// Config - структура, содержащая основные параметры приложения.
type Config struct {
	ServerAddress  string // адрес сервера
	HashKey        string // ключ хэширования
	CryptoKeyPath  string // путь до публичного ключа
	PollInterval   int64  // интервал опроса (в секундах)
	ReportInterval int64  // интервал отправки данных (в секундах)
	RateLimit      int64  // лимит запросов для агента
}

// Initialize создаёт и иницализирует объект *Config.
// Значения присваиваются в следующем порядке (переприсваивают):
//   - значения по умолчания;
//   - значения из флагов командной строки;
//   - значения из переменных окружения.
func Initialize() *Config {
	config := Config{
		ServerAddress:  "http://" + defaultServerAddress,
		HashKey:        defaultHashKey,
		CryptoKeyPath:  defaultCryptoKeyPath,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		RateLimit:      defaultRateLimit,
	}

	config.getFlags()
	config.getEnvironments()

	return &config
}

func (c *Config) getFlags() {
	argEndpValue := flag.String("a", defaultServerAddress, "HTTP server endpoint")
	argKeyValue := flag.String("k", defaultHashKey, "Hash key")
	argCryptoValue := flag.String("crypto-key", defaultCryptoKeyPath, "Public crypto key path")
	argRepValue := flag.Int64("r", defaultReportInterval, "Report interval")
	argPollValue := flag.Int64("p", defaultPollInterval, "Poll interval")
	argLimitValue := flag.Int64("l", defaultPollInterval, "Rate limit")

	flag.Parse()

	if argEndpValue != nil && *argEndpValue != "" {
		c.ServerAddress = "http://" + *argEndpValue
	}
	if argKeyValue != nil && *argKeyValue != "" {
		c.HashKey = *argKeyValue
	}
	if argCryptoValue != nil && *argCryptoValue != "" {
		c.CryptoKeyPath = *argCryptoValue
	}
	if argRepValue != nil && *argRepValue != 0 {
		c.ReportInterval = *argRepValue
	}
	if argPollValue != nil && *argPollValue != 0 {
		c.PollInterval = *argPollValue
	}
	if argLimitValue != nil && *argLimitValue != 0 {
		c.RateLimit = *argLimitValue
	}
}

func (c *Config) getEnvironments() {
	envEndpValue, ok := os.LookupEnv("ADDRESS")
	if ok && envEndpValue != "" {
		c.ServerAddress = "http://" + envEndpValue
	}

	envKeyValue, ok := os.LookupEnv("KEY")
	if ok && envKeyValue != "" {
		c.HashKey = envKeyValue
	}

	envCryptoValue, ok := os.LookupEnv("CRYPTO_KEY")
	if ok && envCryptoValue != "" {
		c.CryptoKeyPath = envCryptoValue
	}

	envRepValue, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok && envRepValue != "" {
		if val, err := strconv.ParseInt(envRepValue, 10, 64); err == nil {
			c.ReportInterval = val
		}
	}

	envPollValue, ok := os.LookupEnv("POLL_INTERVAL")
	if ok && envPollValue != "" {
		if val, err := strconv.ParseInt(envPollValue, 10, 64); err == nil {
			c.PollInterval = val
		}
	}

	envLimitValue, ok := os.LookupEnv("RATE_LIMIT")
	if ok && envLimitValue != "" {
		if val, err := strconv.ParseInt(envLimitValue, 10, 64); err == nil {
			c.RateLimit = val
		}
	}
}
