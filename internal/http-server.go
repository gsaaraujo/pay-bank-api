package internal

import (
	"context"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gsaaraujo/pay-bank-api/internal/daos"
	"github.com/gsaaraujo/pay-bank-api/internal/gateways"
	"github.com/gsaaraujo/pay-bank-api/internal/handlers"
	"github.com/gsaaraujo/pay-bank-api/internal/middlewares"
	"github.com/gsaaraujo/pay-bank-api/internal/usecases"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	webhttp "github.com/gsaaraujo/pay-bank-api/internal/web-http"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type HttpServer struct {
	echo   *echo.Echo
	logger *slog.Logger
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		echo:   echo.New(),
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (h *HttpServer) Ready() {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error(r.(string), "stack_trace", string(debug.Stack()))
			os.Exit(1)
		}
	}()

	h.echo.HidePort = true
	h.echo.HideBanner = true
	h.echo.Use(middleware.RequestID())
	h.echo.Use(middlewares.NewEchoRequestLoggerMiddleware(h.logger))
	h.echo.Use(middlewares.NewEchoRecoverMiddleware(h.logger))

	defaultConfig := utils.GetOrThrow(config.LoadDefaultConfig(context.TODO()))
	secretsClient := secretsmanager.NewFromConfig(defaultConfig)
	jsonBodyValidator := utils.GetOrThrow(webhttp.NewJSONBodyValidator())

	awsSecretsGateway := gateways.NewAwsSecretsGateway(secretsClient)
	postgresUrl := awsSecretsGateway.Get("POSTGRES_URL").(string)
	accessTokenSigningKey := awsSecretsGateway.Get("ACCESS_TOKEN_SIGNING_KEY").(string)

	pgxPool := utils.GetOrThrow(pgxpool.New(context.Background(), postgresUrl))

	customerDAO := daos.NewCustomerDAO(pgxPool)
	accountDAO := daos.NewAccountDAO(pgxPool)
	transactionDAO := daos.NewTransactionDAO(pgxPool)

	loginUsecase := usecases.NewLoginUsecase(customerDAO, awsSecretsGateway)
	signUpUsecase := usecases.NewSignUpUsecase(pgxPool, customerDAO)
	transferUsecase := usecases.NewTransferUsecase(pgxPool, accountDAO, transactionDAO)

	loginHandler := handlers.NewLoginHandler(jsonBodyValidator, loginUsecase)
	signUpHandler := handlers.NewSignUpHandler(jsonBodyValidator, signUpUsecase)
	transferHandler := handlers.NewTransferHandler(jsonBodyValidator, transferUsecase)
	getTransactionsHistoryHandler := handlers.NewGetTransactionsHistoryHandler(pgxPool)

	jwtMiddleware := middlewares.NewEchoJWTMiddleware(accessTokenSigningKey)

	h.echo.GET("/health", func(c echo.Context) error {
		return c.NoContent(204)
	})

	v1 := h.echo.Group("/v1")

	v1.POST("/login", loginHandler.Handle)
	v1.POST("/sign-up", signUpHandler.Handle)

	v1.POST("/transfer", transferHandler.Handle, jwtMiddleware)
	v1.GET("/transactions-history", getTransactionsHistoryHandler.Handle, jwtMiddleware)
}

func (h *HttpServer) Start() {
	h.Ready()

	err := h.echo.Start(":3333")
	if err != nil {
		h.logger.Error(err.Error())
		os.Exit(1)
	}
}

func (h *HttpServer) Echo() *echo.Echo {
	return h.echo
}
