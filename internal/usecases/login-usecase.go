package usecases

import (
	"errors"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/daos"
	"github.com/gsaaraujo/pay-bank-api/internal/gateways"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type JwtAccessTokenClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type LoginUsecaseInput struct {
	Email    string
	Password string
}

type LoginUsecaseOutput struct {
	CustomerId  uuid.UUID
	AccessToken string
}

type LoginUsecase struct {
	customerDAO       daos.CustomerDAO
	awsSecretsGateway gateways.AwsSecretsGateway
}

func NewLoginUsecase(customerDAO daos.CustomerDAO, awsSecretsGateway gateways.AwsSecretsGateway) LoginUsecase {
	return LoginUsecase{customerDAO, awsSecretsGateway}
}

func (l *LoginUsecase) Execute(input LoginUsecaseInput) (LoginUsecaseOutput, error) {
	_, err := mail.ParseAddress(input.Email)
	if err != nil {
		return LoginUsecaseOutput{}, errors.New("email address is invalid")
	}

	customerSchema := l.customerDAO.FindOneByEmail(input.Email)

	if customerSchema == nil {
		return LoginUsecaseOutput{}, errors.New("email or password is incorrect")
	}

	err = bcrypt.CompareHashAndPassword([]byte(customerSchema.Password), []byte(input.Password))
	if err != nil {
		return LoginUsecaseOutput{}, errors.New("email or password is incorrect")
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, JwtAccessTokenClaims{
		Roles: []string{"customer"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   customerSchema.Id.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
		},
	})

	accessTokenSigningKey := l.awsSecretsGateway.Get("ACCESS_TOKEN_SIGNING_KEY").(string)
	acessTokenSigned := utils.GetOrThrow(accessToken.SignedString([]byte(accessTokenSigningKey)))

	return LoginUsecaseOutput{
		CustomerId:  customerSchema.Id,
		AccessToken: acessTokenSigned,
	}, nil
}
