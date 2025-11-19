package daos

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerSchema struct {
	Id        uuid.UUID
	Name      string
	Email     string
	Password  string
	UpdatedAt time.Time
	CreatedAt time.Time
}

type CustomerDAO struct {
	pgxPool *pgxpool.Pool
}

func NewCustomerDAO(pgxPool *pgxpool.Pool) CustomerDAO {
	return CustomerDAO{pgxPool}
}

func (p *CustomerDAO) Create(customerSchema CustomerSchema) {
	_ = utils.GetOrThrow(p.pgxPool.Exec(context.Background(),
		"INSERT INTO customers (id, name, email, password, updated_at, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		customerSchema.Id, customerSchema.Name, customerSchema.Email, customerSchema.Password, customerSchema.UpdatedAt, customerSchema.CreatedAt))
}

func (c *CustomerDAO) FindOneByEmail(email string) *CustomerSchema {
	var customerSchema CustomerSchema

	err := c.pgxPool.QueryRow(context.Background(),
		"SELECT id, name, email, password, updated_at, created_at FROM customers WHERE email = $1", email).
		Scan(&customerSchema.Id, &customerSchema.Name, &customerSchema.Email, &customerSchema.Password, &customerSchema.UpdatedAt, &customerSchema.CreatedAt)

	if err != nil && err == pgx.ErrNoRows {
		return nil
	}

	if err != nil {
		panic(err)
	}

	return &customerSchema
}

func (c *CustomerDAO) DeleteAll() {
	_ = utils.GetOrThrow(c.pgxPool.Exec(context.Background(), "TRUNCATE TABLE customers CASCADE"))
}
