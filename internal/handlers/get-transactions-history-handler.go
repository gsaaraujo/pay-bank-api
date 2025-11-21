package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type customer struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type account struct {
	Id uuid.UUID `json:"id"`
}

type transaction struct {
	Id               uuid.UUID `json:"id"`
	CustomerSender   customer  `json:"customerSender"`
	CustomerReceiver customer  `json:"customerReceiver"`
	AccountSender    account   `json:"accountSender"`
	AccountReceiver  account   `json:"accountReceiver"`
	Amount           int64     `json:"amount"`
}

type GetTransactionsHistoryHandler struct {
	pgxPool *pgxpool.Pool
}

func NewGetTransactionsHistoryHandler(pgxPool *pgxpool.Pool) GetTransactionsHistoryHandler {
	return GetTransactionsHistoryHandler{pgxPool}
}

func (g *GetTransactionsHistoryHandler) Handle(c echo.Context) error {
	rows := utils.GetOrThrow(g.pgxPool.Query(context.TODO(), `
		SELECT
			t.id AS transaction_id,
			cs.id AS customer_sender_id,
			cs.name AS customer_sender_name,
			cr.id AS customer_receiver_id,
			cr.name AS customer_receiver_name,
			asnd.id AS account_sender_id,
			arec.id AS account_receiver_id,
			t.amount
		FROM transactions t
		JOIN accounts as asnd
			ON t.account_sender_id = asnd.id
		JOIN customers cs
			ON asnd.customer_id = cs.id
		JOIN accounts as arec
			ON t.account_receiver_id = arec.id
		JOIN customers cr
			ON arec.customer_id = cr.id;
	`))

	transactions := []transaction{}

	for rows.Next() {
		item := transaction{}
		utils.ThrowOnError(rows.Scan(&item.Id, &item.CustomerSender.Id, &item.CustomerSender.Name, &item.CustomerReceiver.Id, &item.CustomerReceiver.Name,
			&item.AccountSender.Id, &item.AccountReceiver.Id, &item.Amount))
		transactions = append(transactions, item)
	}

	return c.JSON(200, map[string]any{
		"data": transactions,
	})
}
