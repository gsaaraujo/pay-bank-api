package testhelpers

import (
	"context"
	"fmt"
	"time"

	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisContainer struct {
	url string
}

func NewRedisContainer() RedisContainer {
	ctx := context.Background()

	postgresContainer := utils.GetOrThrow(testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:8.2.1",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections tcp").WithStartupTimeout(10 * time.Second),
		},
	}))

	host := utils.GetOrThrow(postgresContainer.Host(ctx))
	port := utils.GetOrThrow(postgresContainer.MappedPort(ctx, "6379/tcp"))

	return RedisContainer{
		url: fmt.Sprintf("redis://:%s@%s:%s/0", "", host, port.Port()),
	}
}

func (p *RedisContainer) Url() string {
	return p.url
}
