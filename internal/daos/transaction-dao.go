package daos

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionSchema struct {
	Id                uuid.UUID
	AccountSenderId   uuid.UUID
	AccountReceiverId uuid.UUID
	IdempotencyKey    string
	Amount            int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TransactionDAO struct {
	pgxPool *pgxpool.Pool
}

func NewTransactionDAO(pgxPool *pgxpool.Pool) TransactionDAO {
	return TransactionDAO{pgxPool}
}

func (t *TransactionDAO) Create(transactionSchema TransactionSchema) {
	_ = utils.GetOrThrow(t.pgxPool.Exec(context.Background(),
		"INSERT INTO transactions (id, account_sender_id, account_receiver_id, idempotency_key, amount, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		transactionSchema.Id, transactionSchema.AccountSenderId, transactionSchema.AccountReceiverId, transactionSchema.IdempotencyKey, transactionSchema.Amount,
		transactionSchema.UpdatedAt, transactionSchema.CreatedAt))
}

func (c *TransactionDAO) FindAllByAccountSenderIdAndAccountReceiverId(accountSenderId uuid.UUID, accountReceiverId uuid.UUID) []TransactionSchema {
	rows := utils.GetOrThrow(c.pgxPool.Query(context.Background(),
		`SELECT id, account_sender_id, account_receiver_id, idempotency_key, amount, created_at, updated_at FROM transactions 
	WHERE account_sender_id = $1 AND account_receiver_id = $2`, accountSenderId, accountReceiverId))

	transactionsSchema := []TransactionSchema{}

	for rows.Next() {
		var item TransactionSchema
		utils.ThrowOnError(rows.Scan(&item.Id, &item.AccountSenderId, &item.AccountReceiverId, &item.IdempotencyKey, &item.Amount, &item.UpdatedAt, &item.CreatedAt))
		transactionsSchema = append(transactionsSchema, item)
	}

	return transactionsSchema
}

func (c *TransactionDAO) FindOneByIdempotencyKey(idempotencyKey uuid.UUID) *TransactionSchema {
	var transactionSchema TransactionSchema

	err := c.pgxPool.QueryRow(context.Background(),
		`SELECT id, account_sender_id, account_receiver_id, idempotency_key, amount, created_at, updated_at FROM transactions WHERE idempotency_key = $1`, idempotencyKey).
		Scan(&transactionSchema.Id, &transactionSchema.AccountSenderId, &transactionSchema.AccountReceiverId, &transactionSchema.IdempotencyKey, &transactionSchema.Amount,
			&transactionSchema.UpdatedAt, &transactionSchema.CreatedAt)

	if err != nil && err == pgx.ErrNoRows {
		return nil
	}

	if err != nil {
		panic(err)
	}

	return &transactionSchema
}

func (c *TransactionDAO) FindOneByAccountSenderIdAndAccountReceiverId(accountSenderId uuid.UUID, accountReceiverId uuid.UUID) *TransactionSchema {
	var transactionSchema TransactionSchema

	err := c.pgxPool.QueryRow(context.Background(),
		`SELECT id, account_sender_id, account_receiver_id, idempotency_key, amount, created_at, updated_at FROM transactions 
		WHERE account_sender_id = $1 AND account_receiver_id = $2`, accountSenderId, accountReceiverId).
		Scan(&transactionSchema.Id, &transactionSchema.AccountSenderId, &transactionSchema.AccountReceiverId, &transactionSchema.IdempotencyKey, &transactionSchema.Amount,
			&transactionSchema.UpdatedAt, &transactionSchema.CreatedAt)

	if err != nil && err == pgx.ErrNoRows {
		return nil
	}

	if err != nil {
		panic(err)
	}

	return &transactionSchema
}

func (c *TransactionDAO) DeleteAll() {
	_ = utils.GetOrThrow(c.pgxPool.Exec(context.Background(), "TRUNCATE TABLE transactions CASCADE"))
}
