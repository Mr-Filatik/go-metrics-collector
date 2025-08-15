package interceptor

import (
	"context"
	"errors"
	"strings"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HashingInterceptor добавляет проверку заголовка hash.
//
// Параметры:
//   - ctx: контекст запроса;
//   - req: запрос;
//   - info: информация о сервере;
//   - handler: следующий обработчик.
func (c *Conveyor) HashingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Получаем заголовок "hashsha256" из метаданных.
	hash, ok := getStringFromContextMetadata(ctx, strings.ToLower(common.HeaderHashSHA256))
	if !ok {
		c.log.Error("Get string from context metadata error", errors.New("hashsha256 not exist"))
		return nil, status.Errorf(codes.PermissionDenied, "hashsha256 not exist")
	}

	// Получаем содержимое запроса.
	body, err := getBodyFromRequest(req)
	if err != nil {
		c.log.Error("Get request body error", err)
		return nil, status.Errorf(codes.InvalidArgument, "get request body error")
	}

	// Рассчитываем хэш на основе содержимого запроса и ключа хэширования.
	calculatedHash, hashErr := common.HashBytesToString(body, c.hashKey)
	if hashErr != nil {
		c.log.Error("Create hash error", hashErr)
		return nil, status.Errorf(codes.Internal, "create hash error")
	}

	if !common.HashValidateStrings(calculatedHash, hash) {
		c.log.Error("Hashes not equals", errors.New("hashes not equals"))
		return nil, status.Errorf(codes.PermissionDenied, "hashes not equals")
	}

	return handler(ctx, req)
}
