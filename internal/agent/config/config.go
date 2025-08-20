// Пакет config предоставляет функционал загрузки конфигурации из флагов командной строки и переменных окружения.
// Конфигурация включает такие параметры как: адрес сервера, интервал опроса, интервал отправки и т.п.
package config

// Костанты - значения по умолчанию.
const (
	defaultServerAddress  string = "localhost:8080" // адрес сервера
	defaultHashKey        string = ""               // ключ хэширования (отсутствует)
	defaultPollInterval   int64  = 2                // интервал опроса (в секундах)
	defaultReportInterval int64  = 10               // интервал отправки данных (в секундах)
	defaultRateLimit      int64  = 1                // лимит запросов для агента
	defaultCryptoKeyPath  string = ""               // путь до публичного ключа
	defaultGrpcEnabled    bool   = false            // включать ли поддержку gRPC
)

// Config - структура, содержащая основные параметры приложения.
type Config struct {
	ServerAddress  string // Aдрес сервера
	HashKey        string // Ключ хэширования
	CryptoKeyPath  string // Путь до публичного ключа
	PollInterval   int64  // Интервал опроса (в секундах)
	ReportInterval int64  // Интервал отправки данных (в секундах)
	RateLimit      int64  // Лимит запросов для агента
	GrpcEnabled    bool   // Bключать ли поддержку gRPC
}

// Initialize создаёт и иницализирует объект *Config.
// Значения присваиваются в следующем порядке (переприсваивают):
//   - значения по умолчания;
//   - значения из файла конфигурации;
//   - значения из флагов командной строки;
//   - значения из переменных окружения.
func Initialize() *Config {
	envsConf := getEnvsConfigFromOS()
	flagsConf, _ := getFlagsConfigFromOS()

	var path string
	if flagsConf.configPathIsValue {
		path = flagsConf.configPath
	}
	if envsConf.configPathIsValue {
		path = envsConf.configPath
	}

	fileConf, _ := getJSONConfigFromFile(path) // игнорируем ошибку, т.к. есть дефолтные значения

	config := createAndOverrideConfig(fileConf, flagsConf, envsConf)

	config.ServerAddress = "http://" + stripHTTPPrefix(config.ServerAddress)

	return config
}

func createAndOverrideConfig(fileConf *configJSONs, flagsConf *configEnvsAndFlags, envsConf *configEnvsAndFlags) *Config {
	config := &Config{
		ServerAddress:  defaultServerAddress,
		HashKey:        defaultHashKey,
		CryptoKeyPath:  defaultCryptoKeyPath,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		RateLimit:      defaultRateLimit,
		GrpcEnabled:    defaultGrpcEnabled,
	}

	config.overrideConfigFromJSONs(fileConf)
	config.overrideConfigFromEnvsAndFlags(flagsConf)
	config.overrideConfigFromEnvsAndFlags(envsConf)

	return config
}
