// Пакет storage предоставляет конкретную реализацию хранилища
// для хранения данных в файловой системе.
package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/Mr-Filatik/go-metrics-collector/internal/entity"
	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
)

// Костанты для работы с файловой системой.
const (
	filePermission os.FileMode = 0o600 // разрешения для работы с файлом
)

// FileStorage реализация хранилища для файловой системы.
type FileStorage struct {
	log      logger.Logger // логгер
	filePath string        // путь до файла
}

func New(filePath string, log logger.Logger) *FileStorage {
	return &FileStorage{
		filePath: filePath,
		log:      log,
	}
}

func (s *FileStorage) LoadData() ([]entity.Metrics, error) {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return make([]entity.Metrics, 0), errors.New("file does not exist")
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return make([]entity.Metrics, 0), errors.New("failed to read metrics from file")
	}

	var metrics []entity.Metrics
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return make([]entity.Metrics, 0), errors.New("failed to deserialize metrics")
	}

	return metrics, nil
}

func (s *FileStorage) SaveData(data []entity.Metrics) error {
	fd, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.New("failed to serialize metrics")
	}

	err = os.WriteFile(s.filePath, fd, filePermission)
	if err != nil {
		return errors.New("failed to write metrics to file")
	}

	return nil
}
