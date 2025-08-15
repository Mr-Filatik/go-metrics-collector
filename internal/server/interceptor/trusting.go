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

// TrustingInterceptor добавляет ограничения доступа для неразрешённых подсетей в gRPC-сервер.
//
// Параметры:
//   - ctx: контекст запроса;
//   - req: запрос;
//   - info: информация о сервере;
//   - handler: следующий обработчик.
func (c *Conveyor) TrustingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Пропуск проверки, если разрешённые подсети не указаны.
	if c.trustedSubnet == "" {
		return handler(ctx, req)
	}

	// Получение заголовка "x-real-ip" из метаданных.
	realIP, ok := getStringFromContextMetadata(ctx, strings.ToLower(common.HeaderXRealIP))
	if !ok {
		c.log.Error("Get string from context metadata error", errors.New("x-real-ip not exist"))
		return nil, status.Errorf(codes.PermissionDenied, "x-real-ip not exist")
	}

	// Проверяем значение заголовка "x-real-ip" с разрешёнными адресами.
	if realIP != c.trustedSubnet {
		msg := strings.Join([]string{"subnet", realIP, "not trusted"}, " ")
		c.log.Error("Subnet not trusted", errors.New(msg))
		return nil, status.Errorf(codes.PermissionDenied, msg)
	}

	return handler(ctx, req)
}
