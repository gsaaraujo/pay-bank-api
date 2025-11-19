package usecases

import (
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/gsaaraujo/pay-bank-api/internal/daos"
	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type SignUpUsecaseInput struct {
	Name     string
	Email    string
	Password string
}

type SignUpUsecase struct {
	pgxPool     *pgxpool.Pool
	customerDAO daos.CustomerDAO
}

func NewSignUpUsecase(pgxPool *pgxpool.Pool, customerDAO daos.CustomerDAO) SignUpUsecase {
	return SignUpUsecase{pgxPool, customerDAO}
}

func (s SignUpUsecase) Execute(input SignUpUsecaseInput) error {
	if len(input.Name) < 2 {
		return errors.New("name must be at least 2 characters")
	}

	_, err := mail.ParseAddress(input.Email)
	if err != nil {
		return errors.New("email address is invalid")
	}

	if len(input.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	customerSchema := s.customerDAO.FindOneByEmail(input.Email)

	if customerSchema != nil {
		return errors.New("this email address has already been taken by someone")
	}

	hashedPassword := utils.GetOrThrow(bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost))

	s.customerDAO.Create(daos.CustomerSchema{
		Id:        uuid.New(),
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hashedPassword),
		UpdatedAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	})

	return nil
}
