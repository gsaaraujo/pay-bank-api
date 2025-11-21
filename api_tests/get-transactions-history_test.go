package apitests_test

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/daos"
	testhelpers "github.com/gsaaraujo/pay-bank-api/internal/test_helpers"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/stretchr/testify/suite"
)

type GetTransactionsHistorySuite struct {
	suite.Suite
	customerDAO     daos.CustomerDAO
	accountDAO      daos.AccountDAO
	transactionDAO  daos.TransactionDAO
	testEnvironment *testhelpers.TestEnvironment
}

func (g *GetTransactionsHistorySuite) SetupSuite() {
	g.testEnvironment = testhelpers.NewTestEnvironment()
	g.testEnvironment.Start()
	g.customerDAO = daos.NewCustomerDAO(g.testEnvironment.PgxPool())
	g.accountDAO = daos.NewAccountDAO(g.testEnvironment.PgxPool())
	g.transactionDAO = daos.NewTransactionDAO(g.testEnvironment.PgxPool())
}

func (g *GetTransactionsHistorySuite) SetupTest() {
	g.customerDAO.DeleteAll()
	g.accountDAO.DeleteAll()
	g.transactionDAO.DeleteAll()
}

func (g *GetTransactionsHistorySuite) Test1() {
	g.Run("given that there's transactions, when getting transaction history, then returns 200", func() {
		g.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Name:      "John Doe",
			Email:     "john.doe@gmail.com",
			Password:  "$2a$10$asLIHej6kxd3Fsdc76QHieBugwCGvsYJeLiZmP1K7/t1GbIbUy.pK",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		g.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Name:      "Richard Smith",
			Email:     "richard.smith@gmail.com",
			Password:  "$2a$10$1dS5NaFw0pZgGA.SvQ5awOm5jr36Z5pE2wl51mHHIQTz5fO9wwBTC",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		g.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			CustomerId: uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Balance:    12500,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})
		g.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"),
			CustomerId: uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Balance:    3200,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})
		g.transactionDAO.Create(daos.TransactionSchema{
			Id:                uuid.MustParse("7e7fc500-0699-4e21-895c-dc8908da9329"),
			AccountSenderId:   uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			AccountReceiverId: uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"),
			IdempotencyKey:    "2108b394-b875-40cf-9ee6-1d8bd6fb1ec5",
			Amount:            4900,
			UpdatedAt:         time.Now().UTC(),
			CreatedAt:         time.Now().UTC(),
		})
		g.transactionDAO.Create(daos.TransactionSchema{
			Id:                uuid.MustParse("661d6052-ba0b-4d53-80b4-0e0b1e78623e"),
			AccountSenderId:   uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			AccountReceiverId: uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"),
			IdempotencyKey:    "03066c51-ce8d-420a-a21b-b905e1b37b2a",
			Amount:            78594,
			UpdatedAt:         time.Now().UTC(),
			CreatedAt:         time.Now().UTC(),
		})
		g.transactionDAO.Create(daos.TransactionSchema{
			Id:                uuid.MustParse("b648c932-becb-48ca-89e1-3fda8677e7dd"),
			AccountSenderId:   uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			AccountReceiverId: uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"),
			IdempotencyKey:    "d7e7b27b-2ec7-4215-8f7d-31f8fa73e662",
			Amount:            2539,
			UpdatedAt:         time.Now().UTC(),
			CreatedAt:         time.Now().UTC(),
		})

		request := utils.GetOrThrow(http.NewRequest("GET", g.testEnvironment.BaseUrl()+"/v1/transactions-history", nil))
		accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+accessToken)
		request.Header.Add("Idempotency-Key", "2108b394-b875-40cf-9ee6-1d8bd6fb1ec5")

		response := utils.GetOrThrow(g.testEnvironment.Client().Do(request))

		body := utils.GetOrThrow(io.ReadAll(response.Body))
		g.Equal(200, response.StatusCode)
		g.JSONEq(`
		{
			"data": [
				{
					"id": "7e7fc500-0699-4e21-895c-dc8908da9329",
					"customerSender": {
						"id": "f59207c8-e837-4159-b67d-78c716510747",
						"name": "John Doe"
					},
					"customerReceiver": {
						"id": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
						"name": "Richard Smith"
					},
					"accountSender": {
						"id": "2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"
					},
					"accountReceiver": {
						"id": "c7333b68-6f2a-46db-89c8-fd833fd3546d"
					},
					"amount": 4900
				},
				{
					"id": "661d6052-ba0b-4d53-80b4-0e0b1e78623e",
					"customerSender": {
						"id": "f59207c8-e837-4159-b67d-78c716510747",
						"name": "John Doe"
					},
					"customerReceiver": {
						"id": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
						"name": "Richard Smith"
					},
					"accountSender": {
						"id": "2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"
					},
					"accountReceiver": {
						"id": "c7333b68-6f2a-46db-89c8-fd833fd3546d"
					},
					"amount": 78594
				},
				{
					"id": "b648c932-becb-48ca-89e1-3fda8677e7dd",
					"customerSender": {
						"id": "f59207c8-e837-4159-b67d-78c716510747",
						"name": "John Doe"
					},
					"customerReceiver": {
						"id": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
						"name": "Richard Smith"
					},
					"accountSender": {
						"id": "2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"
					},
					"accountReceiver": {
						"id": "c7333b68-6f2a-46db-89c8-fd833fd3546d"
					},
					"amount": 2539
				}
			]
		}
	`, string(body))
	})
}

func TestGetTransactionsHistorySuite(t *testing.T) {
	suite.Run(t, new(GetTransactionsHistorySuite))
}
