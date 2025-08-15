// Пакет interceptor предоставляет реализации всех interceptors используемых в серверном приложении.
package interceptor

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mr-Filatik/go-metrics-collector/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Conveyor описывает сущность конвеера для регистрации intercepters.
type Conveyor struct {
	log           logger.Logger // логгер
	trustedSubnet string        // разрешённые подсети
	hashKey       string        // ключ хэширования
}

// New создаёт и инициализирует новый экзепляр *Conveyor.
//
// Параметры:
//   - ts: разрешённые подсети;
//   - hashKey: ключ хэширования;
//   - l: логгер.
func New(ts string, hashKey string, l logger.Logger) *Conveyor {
	return &Conveyor{
		log:           l,
		trustedSubnet: ts,
		hashKey:       hashKey,
	}
}

// getBodyFromRequest получение содержимого запроса в формате []byte.
func getBodyFromRequest(req interface{}) ([]byte, error) {
	mes, ok := req.(proto.Message)
	if !ok {
		return nil, errors.New("message is not proto.Message")
	}
	val, err := proto.Marshal(mes)
	if err != nil {
		return nil, fmt.Errorf("marshal proto message error: %w", err)
	}
	return val, nil
}

// getStringFromContextMetadata получение значения из метаданных запроса по ключу.
func getStringFromContextMetadata(ctx context.Context, key string) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		vals := md.Get(key)
		if len(vals) > 0 {
			return vals[0], true
		}
	}
	return "", false
}

// getStringFromContextMetadata установка значения в метаданные запроса по ключу.
func setStringToContextMetadata(ctx context.Context, key string, val string) error {
	header := metadata.Pairs(
		key, val,
	)
	err := grpc.SetHeader(ctx, header)
	if err != nil {
		return fmt.Errorf("set header to context metadata error: %w", err)
	}
	return nil
}
