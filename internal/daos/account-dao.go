package daos

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountSchema struct {
	Id         uuid.UUID
	CustomerId uuid.UUID
	Balance    int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type AccountDAO struct {
	pgxPool *pgxpool.Pool
}

func NewAccountDAO(pgxPool *pgxpool.Pool) AccountDAO {
	return AccountDAO{pgxPool}
}

func (p *AccountDAO) Create(accountSchema AccountSchema) {
	_ = utils.GetOrThrow(p.pgxPool.Exec(context.Background(),
		"INSERT INTO accounts (id, customer_id, balance, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		accountSchema.Id, accountSchema.CustomerId, accountSchema.Balance, accountSchema.UpdatedAt, accountSchema.CreatedAt))
}

func (c *AccountDAO) FindOneByCustomerId(customerId uuid.UUID) *AccountSchema {
	var accountSchema AccountSchema

	err := c.pgxPool.QueryRow(context.Background(),
		"SELECT id, customer_id, balance, created_at, updated_at FROM accounts WHERE customer_id = $1", customerId).
		Scan(&accountSchema.Id, &accountSchema.CustomerId, &accountSchema.Balance, &accountSchema.CreatedAt, &accountSchema.UpdatedAt)

	if err != nil && err == pgx.ErrNoRows {
		return nil
	}

	if err != nil {
		panic(err)
	}

	return &accountSchema
}

func (c *AccountDAO) DeleteAll() {
	_ = utils.GetOrThrow(c.pgxPool.Exec(context.Background(), "TRUNCATE TABLE accounts CASCADE"))
}
