package testhelpers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
)

func TestGenerateAccessToken(customerId uuid.UUID) string {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   customerId.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(30 * time.Minute)),
	})

	acessTokenSigned := utils.GetOrThrow(accessToken.SignedString([]byte("81c4a8d5b2554de4ba736e93255ba633")))
	return acessTokenSigned
}
