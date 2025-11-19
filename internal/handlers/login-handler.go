package handlers

import (
	"github.com/gsaaraujo/pay-bank-api/internal/usecases"
	webhttp "github.com/gsaaraujo/pay-bank-api/internal/web-http"
	"github.com/labstack/echo/v4"
)

type LoginHandlerInput struct {
	Email    any `validate:"required,string,notEmpty"`
	Password any `validate:"required,string,notEmpty"`
}

type LoginHandler struct {
	jsonBodyValidator webhttp.JSONBodyValidator
	LoginUsecase      usecases.LoginUsecase
}

func NewLoginHandler(JSONBodyValidator webhttp.JSONBodyValidator, LoginUsecase usecases.LoginUsecase) LoginHandler {
	return LoginHandler{JSONBodyValidator, LoginUsecase}
}

func (l *LoginHandler) Handle(c echo.Context) error {
	var input LoginHandlerInput

	if err := c.Bind(&input); err != nil {
		return c.NoContent(415)
	}

	if messages := l.jsonBodyValidator.Validate(input); len(messages) > 0 {
		return c.JSON(400, map[string]any{"message": messages})
	}

	loginUsecaseOutput, err := l.LoginUsecase.Execute(usecases.LoginUsecaseInput{
		Email:    input.Email.(string),
		Password: input.Password.(string),
	})

	if err != nil {
		switch err.Error() {
		case "email address is invalid":
			return c.JSON(409, map[string]any{"message": err.Error()})
		case "email or password is incorrect":
			return c.JSON(409, map[string]any{"message": err.Error()})
		default:
			return c.JSON(500, map[string]any{"message": "Internal Server Error"})
		}
	}

	return c.JSON(200, map[string]any{
		"data": map[string]any{
			"customerId":  loginUsecaseOutput.CustomerId,
			"accessToken": loginUsecaseOutput.AccessToken,
		},
	})
}
