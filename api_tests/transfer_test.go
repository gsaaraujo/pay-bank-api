package apitests_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/daos"
	testhelpers "github.com/gsaaraujo/pay-bank-api/internal/test_helpers"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/stretchr/testify/suite"
)

type TransferSuite struct {
	suite.Suite
	customerDAO     daos.CustomerDAO
	accountDAO      daos.AccountDAO
	transactionDAO  daos.TransactionDAO
	testEnvironment *testhelpers.TestEnvironment
}

func (tr *TransferSuite) SetupSuite() {
	tr.testEnvironment = testhelpers.NewTestEnvironment()
	tr.testEnvironment.Start()
	tr.customerDAO = daos.NewCustomerDAO(tr.testEnvironment.PgxPool())
	tr.accountDAO = daos.NewAccountDAO(tr.testEnvironment.PgxPool())
	tr.transactionDAO = daos.NewTransactionDAO(tr.testEnvironment.PgxPool())
}

func (tr *TransferSuite) SetupTest() {
	tr.customerDAO.DeleteAll()
	tr.accountDAO.DeleteAll()
	tr.transactionDAO.DeleteAll()
}

func (tr *TransferSuite) Test1() {
	tr.Run("when transferring, then returns 204 and credits the receiver and debits the sender", func() {
		tr.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Name:      "John Doe",
			Email:     "john.doe@gmail.com",
			Password:  "$2a$10$asLIHej6kxd3Fsdc76QHieBugwCGvsYJeLiZmP1K7/t1GbIbUy.pK",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		tr.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Name:      "Richard Smith",
			Email:     "richard.smith@gmail.com",
			Password:  "$2a$10$1dS5NaFw0pZgGA.SvQ5awOm5jr36Z5pE2wl51mHHIQTz5fO9wwBTC",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		tr.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			CustomerId: uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Balance:    12500,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})
		tr.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"),
			CustomerId: uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Balance:    3200,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})

		request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(`
			{
				"customerReceiverId": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
				"amount": 2500
			}
		`)))
		accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+accessToken)
		request.Header.Add("Idempotency-Key", "2108b394-b875-40cf-9ee6-1d8bd6fb1ec5")

		response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

		body := utils.GetOrThrow(io.ReadAll(response.Body))
		tr.Equal(204, response.StatusCode)
		tr.Equal("", string(body))

		accountSender := tr.accountDAO.FindOneById(uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"))
		tr.Require().NotNil(accountSender)
		tr.Require().True(utils.IsValidUUID(accountSender.Id.String()))
		tr.Require().Equal("f59207c8-e837-4159-b67d-78c716510747", accountSender.CustomerId.String())
		tr.Require().Equal(int64(10000), accountSender.Balance)
		tr.Require().WithinDuration(time.Now().UTC(), accountSender.UpdatedAt, 5*time.Second)
		tr.Require().WithinDuration(time.Now().UTC(), accountSender.CreatedAt, 5*time.Second)

		accountReceiver := tr.accountDAO.FindOneById(uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"))
		tr.Require().NotNil(accountReceiver)
		tr.Require().True(utils.IsValidUUID(accountReceiver.Id.String()))
		tr.Require().Equal("a06f5c45-f824-4cb1-a666-805035ae2ae1", accountReceiver.CustomerId.String())
		tr.Require().Equal(int64(5700), accountReceiver.Balance)
		tr.Require().WithinDuration(time.Now().UTC(), accountReceiver.UpdatedAt, 5*time.Second)
		tr.Require().WithinDuration(time.Now().UTC(), accountReceiver.CreatedAt, 5*time.Second)

		transactionSchema := tr.transactionDAO.FindOneByAccountSenderIdAndAccountReceiverId(uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"))
		tr.Require().NotNil(transactionSchema)
		tr.Require().True(utils.IsValidUUID(transactionSchema.Id.String()))
		tr.Require().Equal("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d", transactionSchema.AccountSenderId.String())
		tr.Require().Equal("c7333b68-6f2a-46db-89c8-fd833fd3546d", transactionSchema.AccountReceiverId.String())
		tr.Require().Equal("2108b394-b875-40cf-9ee6-1d8bd6fb1ec5", transactionSchema.IdempotencyKey)
		tr.Require().Equal(int64(2500), transactionSchema.Amount)
		tr.Require().WithinDuration(time.Now().UTC(), transactionSchema.UpdatedAt, 5*time.Second)
		tr.Require().WithinDuration(time.Now().UTC(), transactionSchema.CreatedAt, 5*time.Second)
	})
}

func (tr *TransferSuite) Test2() {
	tr.Run("when transferring and there's concurrency, then returns 204 and credits the receiver and debits the sender", func() {
		tr.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Name:      "John Doe",
			Email:     "john.doe@gmail.com",
			Password:  "$2a$10$asLIHej6kxd3Fsdc76QHieBugwCGvsYJeLiZmP1K7/t1GbIbUy.pK",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		tr.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Name:      "Richard Smith",
			Email:     "richard.smith@gmail.com",
			Password:  "$2a$10$1dS5NaFw0pZgGA.SvQ5awOm5jr36Z5pE2wl51mHHIQTz5fO9wwBTC",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		tr.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			CustomerId: uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Balance:    10,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})
		tr.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"),
			CustomerId: uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Balance:    0,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})

		var wg sync.WaitGroup

		for range 10 {
			wg.Go(func() {
				request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(`
					{
						"customerReceiverId": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
						"amount": 1
					}
				`)))
				accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
				request.Header.Add("Content-Type", "application/json")
				request.Header.Add("Idempotency-Key", uuid.New().String())
				request.Header.Add("Authorization", "Bearer "+accessToken)

				response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

				body := utils.GetOrThrow(io.ReadAll(response.Body))
				tr.Equal(204, response.StatusCode)
				tr.Equal("", string(body))
			})
		}

		wg.Wait()

		accountSender := tr.accountDAO.FindOneById(uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"))
		tr.Require().NotNil(accountSender)
		tr.Require().True(utils.IsValidUUID(accountSender.Id.String()))
		tr.Require().Equal("f59207c8-e837-4159-b67d-78c716510747", accountSender.CustomerId.String())
		tr.Require().Equal(int64(0), accountSender.Balance)
		tr.Require().WithinDuration(time.Now().UTC(), accountSender.UpdatedAt, 5*time.Second)
		tr.Require().WithinDuration(time.Now().UTC(), accountSender.CreatedAt, 5*time.Second)

		accountReceiver := tr.accountDAO.FindOneById(uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"))
		tr.Require().NotNil(accountReceiver)
		tr.Require().True(utils.IsValidUUID(accountReceiver.Id.String()))
		tr.Require().Equal("a06f5c45-f824-4cb1-a666-805035ae2ae1", accountReceiver.CustomerId.String())
		tr.Require().Equal(int64(10), accountReceiver.Balance)
		tr.Require().WithinDuration(time.Now().UTC(), accountReceiver.UpdatedAt, 5*time.Second)
		tr.Require().WithinDuration(time.Now().UTC(), accountReceiver.CreatedAt, 5*time.Second)

		transactionSchema := tr.transactionDAO.FindAllByAccountSenderIdAndAccountReceiverId(uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"))
		tr.Require().Equal(10, len(transactionSchema))
	})
}

func (tr *TransferSuite) Test3() {
	tr.Run("when transferring more than once with the same idempotency key, then returns 204 and credits the receiver and debits the sender only once", func() {
		tr.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Name:      "John Doe",
			Email:     "john.doe@gmail.com",
			Password:  "$2a$10$asLIHej6kxd3Fsdc76QHieBugwCGvsYJeLiZmP1K7/t1GbIbUy.pK",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		tr.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Name:      "Richard Smith",
			Email:     "richard.smith@gmail.com",
			Password:  "$2a$10$1dS5NaFw0pZgGA.SvQ5awOm5jr36Z5pE2wl51mHHIQTz5fO9wwBTC",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		tr.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			CustomerId: uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Balance:    12500,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})
		tr.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"),
			CustomerId: uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Balance:    3200,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})

		for range 4 {
			request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(`
				{
					"customerReceiverId": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
					"amount": 2500
				}
			`)))
			accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
			request.Header.Add("Content-Type", "application/json")
			request.Header.Add("Idempotency-Key", "2108b394-b875-40cf-9ee6-1d8bd6fb1ec5")
			request.Header.Add("Authorization", "Bearer "+accessToken)

			response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

			body := utils.GetOrThrow(io.ReadAll(response.Body))
			tr.Equal(204, response.StatusCode)
			tr.Equal("", string(body))
		}

		accountSender := tr.accountDAO.FindOneById(uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"))
		tr.Require().NotNil(accountSender)
		tr.Require().True(utils.IsValidUUID(accountSender.Id.String()))
		tr.Require().Equal("f59207c8-e837-4159-b67d-78c716510747", accountSender.CustomerId.String())
		tr.Require().Equal(int64(10000), accountSender.Balance)
		tr.Require().WithinDuration(time.Now().UTC(), accountSender.UpdatedAt, 5*time.Second)
		tr.Require().WithinDuration(time.Now().UTC(), accountSender.CreatedAt, 5*time.Second)

		accountReceiver := tr.accountDAO.FindOneById(uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"))
		tr.Require().NotNil(accountReceiver)
		tr.Require().True(utils.IsValidUUID(accountReceiver.Id.String()))
		tr.Require().Equal("a06f5c45-f824-4cb1-a666-805035ae2ae1", accountReceiver.CustomerId.String())
		tr.Require().Equal(int64(5700), accountReceiver.Balance)
		tr.Require().WithinDuration(time.Now().UTC(), accountReceiver.UpdatedAt, 5*time.Second)
		tr.Require().WithinDuration(time.Now().UTC(), accountReceiver.CreatedAt, 5*time.Second)

		transactionSchema := tr.transactionDAO.FindAllByAccountSenderIdAndAccountReceiverId(uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"))
		tr.Require().Equal(1, len(transactionSchema))
		tr.Require().True(utils.IsValidUUID(transactionSchema[0].Id.String()))
		tr.Require().Equal("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d", transactionSchema[0].AccountSenderId.String())
		tr.Require().Equal("c7333b68-6f2a-46db-89c8-fd833fd3546d", transactionSchema[0].AccountReceiverId.String())
		tr.Require().Equal("2108b394-b875-40cf-9ee6-1d8bd6fb1ec5", transactionSchema[0].IdempotencyKey)
		tr.Require().Equal(int64(2500), transactionSchema[0].Amount)
		tr.Require().WithinDuration(time.Now().UTC(), transactionSchema[0].UpdatedAt, 5*time.Second)
		tr.Require().WithinDuration(time.Now().UTC(), transactionSchema[0].CreatedAt, 5*time.Second)
	})
}

func (tr *TransferSuite) Test4() {
	tr.Run("when transferring to yourself, then returns 409", func() {
		request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(`
			{
				"customerReceiverId": "f59207c8-e837-4159-b67d-78c716510747",
				"amount": 2500
			}
		`)))
		accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+accessToken)
		request.Header.Add("Idempotency-Key", "2108b394-b875-40cf-9ee6-1d8bd6fb1ec5")

		response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

		body := utils.GetOrThrow(io.ReadAll(response.Body))
		tr.Equal(409, response.StatusCode)
		tr.JSONEq(`
			{
				"message": "you cannot transfer to yourself"
			}
		`, string(body))
	})
}

func (tr *TransferSuite) Test5() {
	tr.Run("when transferring and amount is zero, then returns 409", func() {
		request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(`
			{
				"customerReceiverId": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
				"amount": 0
			}
		`)))
		accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+accessToken)
		request.Header.Add("Idempotency-Key", "2108b394-b875-40cf-9ee6-1d8bd6fb1ec5")

		response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

		body := utils.GetOrThrow(io.ReadAll(response.Body))
		tr.Equal(409, response.StatusCode)
		tr.JSONEq(`
			{
				"message": "the amount to be transferred cannot be zero"
			}
		`, string(body))
	})
}

func (tr *TransferSuite) Test6() {
	tr.Run("when transferring and idempotency key header is missing, then returns 400", func() {
		request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(`
			{
				"customerReceiverId": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
				"amount": 0
			}
		`)))
		accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+accessToken)

		response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

		body := utils.GetOrThrow(io.ReadAll(response.Body))
		tr.Equal(400, response.StatusCode)
		tr.JSONEq(`
			{
				"message": "idempotency-key header is required"
			}
		`, string(body))
	})
}

func (tr *TransferSuite) Test7() {
	tr.Run("when transferring and idempotency key header is missing, then returns 400", func() {
		request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(`
			{
				"customerReceiverId": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
				"amount": 0
			}
		`)))
		accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Idempotency-Key", "abc")
		request.Header.Add("Authorization", "Bearer "+accessToken)

		response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

		body := utils.GetOrThrow(io.ReadAll(response.Body))
		tr.Equal(400, response.StatusCode)
		tr.JSONEq(`
			{
				"message": "idempotency-key header must be uuidv4"
			}
		`, string(body))
	})
}

func (tr *TransferSuite) Test8() {
	tr.Run("given that the sender has not enough balance, when transferring an amount higher than it's balance, then returns 409", func() {
		tr.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Name:      "John Doe",
			Email:     "john.doe@gmail.com",
			Password:  "$2a$10$asLIHej6kxd3Fsdc76QHieBugwCGvsYJeLiZmP1K7/t1GbIbUy.pK",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		tr.customerDAO.Create(daos.CustomerSchema{
			Id:        uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Name:      "Richard Smith",
			Email:     "richard.smith@gmail.com",
			Password:  "$2a$10$1dS5NaFw0pZgGA.SvQ5awOm5jr36Z5pE2wl51mHHIQTz5fO9wwBTC",
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		})
		tr.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("2a351ae8-cd0b-41c0-b28b-570f8dd5fb4d"),
			CustomerId: uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"),
			Balance:    12500,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})
		tr.accountDAO.Create(daos.AccountSchema{
			Id:         uuid.MustParse("c7333b68-6f2a-46db-89c8-fd833fd3546d"),
			CustomerId: uuid.MustParse("a06f5c45-f824-4cb1-a666-805035ae2ae1"),
			Balance:    3200,
			UpdatedAt:  time.Now().UTC(),
			CreatedAt:  time.Now().UTC(),
		})

		request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(`
			{
				"customerReceiverId": "a06f5c45-f824-4cb1-a666-805035ae2ae1",
				"amount": 80000
			}
		`)))
		accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+accessToken)
		request.Header.Add("Idempotency-Key", "2108b394-b875-40cf-9ee6-1d8bd6fb1ec5")

		response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

		body := utils.GetOrThrow(io.ReadAll(response.Body))
		tr.Equal(409, response.StatusCode)
		tr.JSONEq(`
			{
				"message": "the sender does not have enough balance to make the transfer"
			}
		`, string(body))
	})
}

func (tr *TransferSuite) Test9() {
	tr.Run("when transferring and body is invalid, then returns 400", func() {
		templates := []map[string]string{
			{
				"body": `{}`,
				"error": `[
					"customerReceiverId is required",
					"amount is required"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": null,
					"amount": null
				}`,
				"error": `[
					"customerReceiverId is required",
					"amount is required"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": "",
					"amount": ""
				}`,
				"error": `[
					"customerReceiverId must be uuidv4",
					"amount must be integer"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": " ",
					"amount": " "
				}`,
				"error": `[
					"customerReceiverId must be uuidv4",
					"amount must be integer"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": 1,
					"amount": 1
				}`,
				"error": `[
					"customerReceiverId must be uuidv4"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": 1.5,
					"amount": 1.5
				}`,
				"error": `[
					"customerReceiverId must be uuidv4",
					"amount must be integer"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": -1,
					"amount": -1
				}`,
				"error": `[
					"customerReceiverId must be uuidv4",
					"amount must be positive"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": true,
					"amount": false
				}`,
				"error": `[
					"customerReceiverId must be uuidv4",
					"amount must be integer"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": {},
					"amount": {}
				}`,
				"error": `[
					"customerReceiverId must be uuidv4",
					"amount must be integer"
				]`,
			},
			{
				"body": `{
					"customerReceiverId": [],
					"amount": []
				}`,
				"error": `[
					"customerReceiverId must be uuidv4",
					"amount must be integer"
				]`,
			},
		}

		for _, template := range templates {
			request := utils.GetOrThrow(http.NewRequest("POST", tr.testEnvironment.BaseUrl()+"/v1/transfer", strings.NewReader(template["body"])))
			accessToken := testhelpers.TestGenerateAccessToken(uuid.MustParse("f59207c8-e837-4159-b67d-78c716510747"))
			request.Header.Add("Content-Type", "application/json")
			request.Header.Add("Authorization", "Bearer "+accessToken)
			request.Header.Add("Idempotency-Key", "2108b394-b875-40cf-9ee6-1d8bd6fb1ec5")

			response := utils.GetOrThrow(tr.testEnvironment.Client().Do(request))

			body := utils.GetOrThrow(io.ReadAll(response.Body))
			tr.Equal(400, response.StatusCode)
			tr.JSONEq(fmt.Sprintf(`
				{
					"message": %s
				}
			`, template["error"]), string(body))
		}
	})
}

func TestTransfer(t *testing.T) {
	suite.Run(t, new(TransferSuite))
}
