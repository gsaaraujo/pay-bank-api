package middlewares

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEchoRecoverMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1024,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.LogAttrs(context.Background(), slog.LevelError, err.Error(),
				slog.String("stack_trace", string(stack)),
			)

			return err
		},
	})
}
