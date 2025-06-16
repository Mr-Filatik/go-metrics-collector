package repeater

import (
	"errors"
	"fmt"

	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
)

// ExampleRepeater_Run — пример с повтором при ошибке.
func ExampleRepeater_Run() {
	log := logger.New(logger.LevelDebug)
	attempts := 0
	maxAttempts := 3

	r := New[int, string](log).
		SetCondition(func(err error) bool {
			return err == nil // повторять, пока не будет nil
		}).
		SetFunc(func(n int) (string, error) {
			attempts++
			if attempts < maxAttempts {
				return "", errors.New("временная ошибка")
			}
			return "успех", nil
		})

	result, err := r.Run(42)

	fmt.Printf("Результат: %s, Ошибка: %v\n", result, err)
	// Output:
	// Результат: успех, Ошибка: <nil>
}
