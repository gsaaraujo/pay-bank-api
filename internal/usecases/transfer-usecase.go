package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/daos"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransferUsecaseInput struct {
	SenderCustomerId   uuid.UUID
	ReceiverCustomerId uuid.UUID
	IdempotencyKey     uuid.UUID
	Amount             int64
}

type TransferUsecase struct {
	pgxPool        *pgxpool.Pool
	accountDAO     daos.AccountDAO
	transactionDAO daos.TransactionDAO
}

func NewTransferUsecase(pgxPool *pgxpool.Pool, accountDAO daos.AccountDAO, transactionDAO daos.TransactionDAO) TransferUsecase {
	return TransferUsecase{pgxPool, accountDAO, transactionDAO}
}

func (t *TransferUsecase) Execute(input TransferUsecaseInput) error {
	if input.SenderCustomerId == input.ReceiverCustomerId {
		return errors.New("you cannot transfer to yourself")
	}

	if input.Amount == 0 {
		return errors.New("the amount to be transferred cannot be zero")
	}

	transaction := t.transactionDAO.FindOneByIdempotencyKey(input.IdempotencyKey)

	if transaction != nil {
		return nil
	}

	senderAccount := t.accountDAO.FindOneByCustomerId(input.SenderCustomerId)
	receiverAccount := t.accountDAO.FindOneByCustomerId(input.ReceiverCustomerId)

	if senderAccount == nil {
		panic("sender account was not found")
	}

	if receiverAccount == nil {
		panic("receiver account was not found")
	}

	if senderAccount.Balance < input.Amount {
		return errors.New("the sender does not have enough balance to make the transfer")
	}

	tx := utils.GetOrThrow(t.pgxPool.Begin(context.TODO()))
	defer func() {
		_ = tx.Rollback(context.TODO())
	}()

	_ = utils.GetOrThrow(tx.Exec(context.TODO(), "UPDATE accounts SET balance = balance - $1 WHERE id = $2", input.Amount, senderAccount.Id))
	_ = utils.GetOrThrow(tx.Exec(context.TODO(), "UPDATE accounts SET balance = balance + $1 WHERE id = $2", input.Amount, receiverAccount.Id))
	_ = utils.GetOrThrow(tx.Exec(context.TODO(),
		"INSERT INTO transactions (id, account_sender_id, account_receiver_id, idempotency_key, amount, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		uuid.New(), senderAccount.Id, receiverAccount.Id, input.IdempotencyKey, input.Amount, time.Now().UTC(), time.Now().UTC()))

	utils.ThrowOnError(tx.Commit(context.TODO()))
	return nil
}
