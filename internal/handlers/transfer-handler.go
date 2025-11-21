package handlers

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/usecases"
	webhttp "github.com/gsaaraujo/pay-bank-api/internal/web-http"
	"github.com/labstack/echo/v4"
)

type TransferHandlerInput struct {
	CustomerReceiverId any `validate:"required,uuid4"`
	Amount             any `validate:"required,integer,positive"`
}

type TransferHandler struct {
	jsonBodyValidator webhttp.JSONBodyValidator
	transferUsecase   usecases.TransferUsecase
}

func NewTransferHandler(jsonBodyValidator webhttp.JSONBodyValidator, transferUsecase usecases.TransferUsecase) TransferHandler {
	return TransferHandler{jsonBodyValidator, transferUsecase}
}

func (t *TransferHandler) Handle(c echo.Context) error {
	var input TransferHandlerInput

	if err := c.Bind(&input); err != nil {
		return c.NoContent(415)
	}

	if messages := t.jsonBodyValidator.Validate(input); len(messages) > 0 {
		return c.JSON(400, map[string]any{"message": messages})
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")

	if idempotencyKey == "" {
		return c.JSON(400, map[string]any{"message": "idempotency-key header is required"})
	}

	if err := uuid.Validate(idempotencyKey); err != nil {
		return c.JSON(400, map[string]any{"message": "idempotency-key header must be uuidv4"})
	}

	token := c.Get("customer").(*jwt.Token)
	claims := token.Claims.(*usecases.JwtAccessTokenClaims)

	err := t.transferUsecase.Execute(usecases.TransferUsecaseInput{
		SenderCustomerId:   uuid.MustParse(claims.Subject),
		ReceiverCustomerId: uuid.MustParse(input.CustomerReceiverId.(string)),
		IdempotencyKey:     uuid.MustParse(idempotencyKey),
		Amount:             int64(input.Amount.(float64)),
	})

	if err != nil {
		switch err.Error() {
		case "you cannot transfer to yourself":
			return c.JSON(409, map[string]any{"message": err.Error()})
		case "the amount to be transferred cannot be zero":
			return c.JSON(409, map[string]any{"message": err.Error()})
		case "the sender does not have enough balance to make the transfer":
			return c.JSON(409, map[string]any{"message": err.Error()})
		default:
			return c.JSON(500, map[string]any{"message": "Internal Server Error"})
		}
	}

	return c.NoContent(204)
}
