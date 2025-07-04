// Пакет repeater предоставляет реализацию сущности для повторения действий при ошибках, временных сбоях.
package repeater

import (
	"errors"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
)

var (
	ErrActionNotSet = errors.New("action not set")
	ErrAttemptsOver = errors.New("attempts are over")
)

// Repeater позвляет повторить действие несколько раз, если при его выполнении не выполнилось условие.
type Repeater[Tin any, Tout any] struct {
	log       logger.Logger           // логгер
	action    func(Tin) (Tout, error) // основное действие
	condition func(error) bool        // условие для выхода из повторителя
	delays    []int                   // задержки перед повторением
	current   int                     // текущая попытка
}

// New создаёт и инициализирует новый объект *Repeater[Tin, Tout].
//
// Параметры:
//   - log: логгер
func New[Tin any, Tout any](log logger.Logger) *Repeater[Tin, Tout] {
	return &Repeater[Tin, Tout]{
		current:   0,
		delays:    []int{1, 3, 5},
		log:       log,
		condition: defaultCondition,
		action:    nil,
	}
}

// SetFunc устанавливает основное действие для повторения.
//
// Параметры:
//   - f: функция
func (r *Repeater[Tin, Tout]) SetFunc(f func(Tin) (Tout, error)) *Repeater[Tin, Tout] {
	r.action = f
	return r
}

// SetCondition устанавливает условие для повторения.
//
// Параметры:
//   - c: функция условие
func (r *Repeater[Tin, Tout]) SetCondition(c func(error) bool) *Repeater[Tin, Tout] {
	r.condition = c
	return r
}

// Run запускает повторитель.
//
// Параметры:
//   - data: данные
func (r *Repeater[Tin, Tout]) Run(data Tin) (Tout, error) {
	if r.action == nil {
		var zero Tout
		r.log.Error("Action not set", ErrActionNotSet)
		return zero, ErrActionNotSet
	}

	result, err := r.action(data)
	if r.condition(err) {
		return result, nil
	}
	for r.current = 0; r.current < len(r.delays); r.current++ {
		time.Sleep(time.Duration(r.delays[r.current]) * time.Second)
		r.log.Info("Repeater retry", "attempt", r.current+1)
		result, err = r.action(data)
		if r.condition(err) {
			return result, nil
		}
	}
	r.log.Error("Repeater retry attempts are over", ErrAttemptsOver)
	return result, ErrAttemptsOver
}

func defaultCondition(err error) bool {
	return err == nil
}
