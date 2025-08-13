// Пакет config предоставляет функционал загрузки конфигурации из флагов командной строки и переменных окружения.
// Конфигурация включает такие параметры как: адрес сервера, настройки хранилища, путь до хранилища и т.п.
package config

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
	defaultTrustedSubnet    string = "" // разрешённые подсети
)

// Config - структура, содержащая основные параметры приложения.
type Config struct {
	ServerAddress    string // Адрес сервера
	HashKey          string // Ключ хэширования
	CryptoKeyPath    string // Путь до приватного ключа
	FileStoragePath  string // Путь до файла хранилища (относительный)
	ConnectionString string // Строка подключения к базе данных
	TrustedSubnet    string // Разрешённые подсети
	StoreInterval    int64  // Интервал сохранения данных в хранилище (в секундах)
	Restore          bool   // Флаг, указывающий загружать ли данные из хранилища при старте приложения
}

// Initialize создаёт и иницализирует объект *Config.
// Значения присваиваются в следующем порядке (переприсваивают):
//   - значения по умолчания;
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

	return config
}

func createAndOverrideConfig(fileConf *configJSONs, flagsConf *configFlags, envsConf *configEnvs) *Config {
	config := &Config{
		ServerAddress:    defaultServerAddress,
		HashKey:          defaultHashKey,
		CryptoKeyPath:    defaultCryptoKeyPath,
		StoreInterval:    defaultStoreInterval,
		FileStoragePath:  defaultFileStoragePath,
		TrustedSubnet:    defaultTrustedSubnet,
		ConnectionString: defaultConnectionString,
		Restore:          defaultRestore,
	}

	config.overrideConfigFromJSONs(fileConf)
	config.overrideConfigFromFlags(flagsConf)
	config.overrideConfigFromEnvs(envsConf)

	return config
}
