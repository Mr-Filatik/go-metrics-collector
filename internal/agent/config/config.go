// Пакет config предоставляет функционал загрузки конфигурации из флагов командной строки и переменных окружения.
// Конфигурация включает такие параметры как: адрес сервера, интервал опроса, интервал отправки и т.п.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"strings"
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
	// Aдрес сервера
	ServerAddress string `json:"server_address"`
	// Ключ хэширования
	HashKey string `json:"-"`
	// Путь до публичного ключа
	CryptoKeyPath string `json:"crypto_key"`
	// Интервал опроса (в секундах)
	PollInterval int64 `json:"poll_interval"`
	// Интервал отправки данных (в секундах)
	ReportInterval int64 `json:"report_interval"`
	// Лимит запросов для агента
	RateLimit int64 `json:"-"`
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
		c.ServerAddress = "http://" + stripHTTPPrefix(fileConfig.ServerAddress)
	}
	if fileConfig.CryptoKeyPath != "" {
		c.CryptoKeyPath = fileConfig.CryptoKeyPath
	}
	if fileConfig.PollInterval != 0 {
		c.PollInterval = fileConfig.PollInterval
	}
	if fileConfig.ReportInterval != 0 {
		c.ReportInterval = fileConfig.ReportInterval
	}
}

func stripHTTPPrefix(addr string) string {
	if strings.HasPrefix(addr, "http://") {
		return addr[7:]
	}
	if strings.HasPrefix(addr, "https://") {
		return addr[8:]
	}
	return addr
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

func (c *Config) getConfigPathFromEnvironment() string {
	envConfigValue, ok := os.LookupEnv("CONFIG")
	if ok && envConfigValue != "" {
		return envConfigValue
	}
	return ""
}
