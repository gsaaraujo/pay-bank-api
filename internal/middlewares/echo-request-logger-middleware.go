package middlewares

import (
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gsaaraujo/pay-bank-api/internal/usecases"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEchoRequestLoggerMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRequestID: true,
		LogRemoteIP:  true,
		LogHost:      true,
		LogMethod:    true,
		LogURI:       true,
		LogUserAgent: true,
		LogStatus:    true,
		LogError:     true,
		LogLatency:   true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log := logger.With(
				slog.String("request_id", v.RequestID),
				slog.String("remote_ip", v.RemoteIP),
				slog.String("host", v.Host),
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.String("user_agent", v.UserAgent),
				slog.Int("response_status_code", v.Status),
				slog.Float64("latency_ms", float64(v.Latency.Microseconds())/1000),
			)

			if c.Get("customer") != nil {
				token := c.Get("customer").(*jwt.Token)
				claims := token.Claims.(*usecases.JwtAccessTokenClaims)
				customerId := claims.Subject

				log = log.With(slog.String("customer_id", customerId))
			}

			if v.Error != nil {
				log = log.With(slog.String("error", v.Error.Error()))
			}

			log.Info("http request")
			return v.Error
		},
	})
}
