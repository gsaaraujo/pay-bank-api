package testhelpers

import (
	"context"
	"fmt"
	"time"

	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type LocalstackContainer struct {
	url string
}

func NewLocalstackContainer() LocalstackContainer {
	ctx := context.Background()

	localStackContainer := utils.GetOrThrow(testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "localstack/localstack:4.2.0",
			ExposedPorts: []string{"4566/tcp"},
			WaitingFor:   wait.ForLog("Ready.").WithStartupTimeout(10 * time.Second),
			Env: map[string]string{
				"SERVICES": "secretsmanager,sqs",
			},
		},
	}))

	host := utils.GetOrThrow(localStackContainer.Host(ctx))
	port := utils.GetOrThrow(localStackContainer.MappedPort(ctx, "4566/tcp"))

	return LocalstackContainer{
		url: fmt.Sprintf("http://%s:%s", host, port),
	}
}

func (p *LocalstackContainer) Url() string {
	return p.url
}
