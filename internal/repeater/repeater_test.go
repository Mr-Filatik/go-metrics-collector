package repeater

import (
	"errors"
	"testing"
	"time"

	logger "github.com/Mr-Filatik/go-metrics-collector/internal/logger/zap/sugar"
)

// fakeAction возвращает действие, которое будет возвращать заданный результат и ошибки
func fakeAction[T any, R any](result R, err error) func(T) (R, error) {
	return func(t T) (R, error) {
		return result, err
	}
}

func TestNewCreatesRepeaterWithDefaults(t *testing.T) {
	type args struct{}

	tests := []struct {
		name string
	}{
		{"default initialization"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New[args, int](nil)
			if len(r.delays) != 3 || r.delays[0] != 1 || r.delays[1] != 3 || r.delays[2] != 5 {
				t.Errorf("ожидаемые delays = [1, 3, 5], получено %v", r.delays)
			}
			if r.condition == nil {
				t.Error("ожидается default condition")
			}
			if r.action != nil {
				t.Error("ожидается action == nil по умолчанию")
			}
		})
	}
}

func TestSetFuncSetsAction(t *testing.T) {
	r := New[string, int](nil).SetFunc(func(s string) (int, error) {
		return 42, nil
	})

	if r.action == nil {
		t.FailNow()
	}
}

func TestSetConditionSetsCustomCondition(t *testing.T) {
	customCond := func(err error) bool {
		return err != nil
	}

	r := New[string, int](nil).SetCondition(customCond)

	if r.condition == nil {
		t.Error("condition не установлен")
	}
	if r.condition(errors.New("error")) != true {
		t.Error("ожидается custom condition")
	}
}

func TestRunActionNotSet(t *testing.T) {
	log := logger.New(logger.LevelDebug)
	r := New[string, int](log)

	result, err := r.Run("data")

	if err != ErrActionNotSet {
		t.Errorf("ожидаемая ошибка %v, получено %v", ErrActionNotSet, err)
	}
	if result != 0 {
		t.Errorf("ожидаемое значение по умолчанию 0, получено %v", result)
	}
}

func TestRunSuccessOnFirstAttempt(t *testing.T) {
	log := logger.New(logger.LevelDebug)
	r := New[string, int](log).
		SetFunc(fakeAction[string](42, nil))

	result, err := r.Run("input")

	if err != nil {
		t.Errorf("не ожидалось ошибок, получено %v", err)
	}
	if result != 42 {
		t.Errorf("ожидаемый результат 42, получено %d", result)
	}
}

func TestRunRetrySuccessAfterOneFail(t *testing.T) {
	log := logger.New(logger.LevelDebug)
	failOnce := func(data string) (int, error) {
		if data == "first" {
			return 0, errors.New("fail once")
		}
		return 42, nil
	}

	r := New[string, int](log).
		SetFunc(func(s string) (int, error) {
			return failOnce(s)
		}).
		SetCondition(func(err error) bool {
			return err == nil
		})

	// Первая попытка: возвращаем ошибку
	firstResult, firstErr := r.Run("first")
	if firstErr == nil || firstResult != 0 {
		t.Errorf("первый запуск должен вернуть ошибку")
	}

	// Вторая попытка: успех
	secondResult, secondErr := r.Run("second")
	if secondErr != nil {
		t.Errorf("второй запуск должен быть успешным, получено %v", secondErr)
	}
	if secondResult != 42 {
		t.Errorf("ожидаемый результат 42, получено %d", secondResult)
	}
}

func TestRunExceedsMaxAttempts_ReturnsErrAttemptsOver(t *testing.T) {
	log := logger.New(logger.LevelDebug)
	failingAction := func(s string) (int, error) {
		return 0, errors.New("always fails")
	}

	r := New[string, int](log).
		SetFunc(failingAction)

	result, err := r.Run("input")

	if err == nil || !errors.Is(err, ErrAttemptsOver) {
		t.Errorf("ожидаемая ошибка %v, получено %v", ErrAttemptsOver, err)
	}
	if result != 0 {
		t.Errorf("ожидаемое значение 0, получено %d", result)
	}
}

func TestRunUsesCustomDelays(t *testing.T) {
	delayCounter := 0
	startTime := time.Now()

	r := New[string, int](logger.New(logger.LevelDebug)).SetFunc(func(s string) (int, error) {
		if delayCounter == 0 {
			delayCounter++
			return 0, errors.New("try again")
		}
		return 1, nil
	})

	r.delays = []int{1, 2, 3}
	r.SetCondition(func(err error) bool {
		return err == nil
	})

	result, err := r.Run("data")

	if err != nil {
		t.Errorf("ожидаемый результат без ошибок, получено %v", err)
	}
	if result != 1 {
		t.Errorf("ожидаемый результат 1, получено %d", result)
	}

	elapsed := time.Since(startTime)
	if elapsed < 1*time.Second {
		t.Errorf("ожидаемая задержка перед повтором")
	}
}
