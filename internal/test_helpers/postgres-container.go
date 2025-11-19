package testhelpers

import (
	"context"
	"fmt"
	"time"

	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	url string
}

func NewPostgresContainer() PostgresContainer {
	ctx := context.Background()

	postgresContainer := utils.GetOrThrow(testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:17.2-alpine3.21",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10 * time.Second),
			Env: map[string]string{
				"POSTGRES_DB":       "postgres",
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
			},
		},
	}))

	host := utils.GetOrThrow(postgresContainer.Host(ctx))
	port := utils.GetOrThrow(postgresContainer.MappedPort(ctx, "5432/tcp"))

	return PostgresContainer{
		url: fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres", host, port.Port()),
	}
}

func (p *PostgresContainer) Url() string {
	return p.url
}
