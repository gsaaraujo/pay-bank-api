package apitests_test

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/daos"
	testhelpers "github.com/gsaaraujo/pay-bank-api/internal/test_helpers"
	"github.com/gsaaraujo/pay-bank-api/internal/usecases"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/stretchr/testify/suite"
)

type LoginSuite struct {
	suite.Suite
	customerDAO     daos.CustomerDAO
	testEnvironment *testhelpers.TestEnvironment
}

func (l *LoginSuite) SetupSuite() {
	l.testEnvironment = testhelpers.NewTestEnvironment()
	l.testEnvironment.Start()
	l.customerDAO = daos.NewCustomerDAO(l.testEnvironment.PgxPool())
}

func (l *LoginSuite) SetupTest() {
	l.customerDAO.DeleteAll()
}

func (l *LoginSuite) Test1() {
	l.Run("given that the customer is already signed up, when logging in, then returns 200", func() {
		l.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Name:      "John Doe",
			Email:     "john.doe@gmail.com",
			Password:  "$2a$10$asLIHej6kxd3Fsdc76QHieBugwCGvsYJeLiZmP1K7/t1GbIbUy.pK",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})

		response := utils.GetOrThrow(l.testEnvironment.Client().Post(l.testEnvironment.BaseUrl()+"/v1/login", "application/json", strings.NewReader(`
			{
				"email": "john.doe@gmail.com",
				"password": "123456"
			}
		`)))

		l.Equal(200, response.StatusCode)
		body := utils.ParseJSONBody[map[string]map[string]any](response.Body)
		customerId := body["data"]["customerId"].(string)
		accessToken := body["data"]["accessToken"].(string)
		l.True(utils.IsValidUUID(customerId))
		l.NotEmpty(accessToken)

		token := utils.GetOrThrow(jwt.ParseWithClaims(accessToken, &usecases.JwtAccessTokenClaims{}, func(token *jwt.Token) (any, error) {
			return []byte("81c4a8d5b2554de4ba736e93255ba633"), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()})))

		claims := token.Claims.(*usecases.JwtAccessTokenClaims)
		l.Require().Equal("f59207c8-e837-4159-b67d-78c716510747", claims.Subject)
		l.Require().Equal([]string{"customer"}, claims.Roles)
		l.Require().WithinDuration(time.Now().UTC(), claims.IssuedAt.Time, 5*time.Second)
		l.Require().WithinDuration(time.Now().UTC().Add(30*time.Minute), claims.ExpiresAt.Time, 5*time.Second)
	})
}

func (l *LoginSuite) Test2() {
	l.Run("given that the customer is already signed up, when logging in and password is incorrect, then returns 409", func() {
		l.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Name:      "John Doe",
			Email:     "john.doe@gmail.com",
			Password:  "$2a$10$asLIHej6kxd3Fsdc76QHieBugwCGvsYJeLiZmP1K7/t1GbIbUy.pK",
			CreatedAt: time.Now().UTC(),
		})

		response := utils.GetOrThrow(l.testEnvironment.Client().Post(l.testEnvironment.BaseUrl()+"/v1/login", "application/json", strings.NewReader(`
			{
				"email": "john.doe@gmail.com",
				"password": "abc123"
			}
		`)))

		body := utils.GetOrThrow(io.ReadAll(response.Body))

		l.Equal(409, response.StatusCode)
		l.JSONEq(`
			{
				"message": "email or password is incorrect"
			}
		`, string(body))
	})
}

func (l *LoginSuite) Test3() {
	l.Run("given that the customer is not signed up, when logging in, then returns 409", func() {
		response := utils.GetOrThrow(l.testEnvironment.Client().Post(l.testEnvironment.BaseUrl()+"/v1/login", "application/json", strings.NewReader(`
			{
				"email": "john.doe@gmail.com",
				"password": "123456"
			}
		`)))

		body := utils.GetOrThrow(io.ReadAll(response.Body))

		l.Equal(409, response.StatusCode)
		l.JSONEq(`
			{
				"message": "email or password is incorrect"
			}
		`, string(body))
	})
}

func (l *LoginSuite) Test4() {
	l.Run("when logging in and email address is invalid, then returns 409", func() {
		response := utils.GetOrThrow(l.testEnvironment.Client().Post(l.testEnvironment.BaseUrl()+"/v1/login", "application/json", strings.NewReader(`
			{
				"email": "john",
				"password": "123456"
			}
		`)))

		body := utils.GetOrThrow(io.ReadAll(response.Body))

		l.Equal(409, response.StatusCode)
		l.JSONEq(`
			{
				"message": "email address is invalid"
			}
		`, string(body))
	})
}

func (l *LoginSuite) Test5() {
	l.Run("when logging in and body is invalid, then returns 400", func() {
		templates := []map[string]string{
			{
				"body": `{}`,
				"error": `[
					"email is required",
					"password is required"
				]`,
			},
			{
				"body": `{
					"email": null,
					"password": null
				}`,
				"error": `[
					"email is required",
					"password is required"
				]`,
			},
			{
				"body": `{
					"email": "",
					"password": ""
				}`,
				"error": `[
					"email must not be empty",
					"password must not be empty"
				]`,
			},
			{
				"body": `{
					"email": " ",
					"password": " "
				}`,
				"error": `[
					"email must not be empty",
					"password must not be empty"
				]`,
			},
			{
				"body": `{
					"email": 1,
					"password": 1
				}`,
				"error": `[
					"email must be string",
					"password must be string"
				]`,
			},
			{
				"body": `{
					"email": 1.5,
					"password": 1.5
				}`,
				"error": `[
					"email must be string",
					"password must be string"
				]`,
			},
			{
				"body": `{
					"email": -1,
					"password": -1
				}`,
				"error": `[
					"email must be string",
					"password must be string"
				]`,
			},
			{
				"body": `{
					"email": true,
					"password": true
				}`,
				"error": `[
					"email must be string",
					"password must be string"
				]`,
			},
			{
				"body": `{
					"email": {},
					"password": {}
				}`,
				"error": `[
					"email must be string",
					"password must be string"
				]`,
			},
			{
				"body": `{
					"email": [],
					"password": []
				}`,
				"error": `[
					"email must be string",
					"password must be string"
				]`,
			},
		}

		for _, template := range templates {
			response := utils.GetOrThrow(l.testEnvironment.Client().Post(l.testEnvironment.BaseUrl()+"/v1/login", "application/json",
				strings.NewReader(template["body"])))

			body := utils.GetOrThrow(io.ReadAll(response.Body))

			l.Equal(400, response.StatusCode)
			l.JSONEq(fmt.Sprintf(`
				{
					"message": %s
				}
			`, template["error"]), string(body))
		}
	})
}

func TestLogin(t *testing.T) {
	suite.Run(t, new(LoginSuite))
}
