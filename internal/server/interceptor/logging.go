package interceptor

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Mr-Filatik/go-metrics-collector/internal/common"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor добавляет логирование в gRPC-сервер.
//
// Параметры:
//   - ctx: контекст запроса;
//   - req: запрос;
//   - info: информация о сервере;
//   - handler: следующий обработчик.
func (c *Conveyor) LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	startTime := time.Now().UTC()

	// Получаем заголовок "x-request-id" из метаданных.
	requestID, ok := getStringFromContextMetadata(ctx, strings.ToLower(common.HeaderXRequestID))
	if !ok {
		requestID = uuid.New().String()
		err := setStringToContextMetadata(ctx, strings.ToLower(common.HeaderXRequestID), requestID)
		if err != nil {
			c.log.Error("Set string to context metadata error", errors.New(strings.ToLower(common.HeaderXRequestID)))
			return nil, status.Errorf(codes.Internal, strings.ToLower(common.HeaderXRequestID))
		}
	}

	// Проверяем что заголовок "x-request-id" из метаданных корректный, является uuid.
	_, err := uuid.Parse(requestID)
	if err != nil {
		c.log.Error("Parse x-request-id error", err)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	resp, err := handler(ctx, req)

	statusErr := errors.New("200 OK")
	if err != nil {
		statusErr = err
	}

	body, berr := getBodyFromRequest(req)
	if berr != nil {
		c.log.Error("Get body error", berr)
		return nil, status.Errorf(codes.InvalidArgument, berr.Error())
	}

	c.log.Info(
		"gRPC-Call",
		"call_id", requestID,
		"call_method", info.FullMethod,
		"call_time", startTime.String(),
		"call_duration", time.Since(startTime),
		"status", statusErr.Error(),
		"content_lenght", len(body),
	)

	return resp, err
}
