package testhelpers

import (
	"context"
	"fmt"

	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type WiremockContainer struct {
	url string
}

func NewWiremockContainer() WiremockContainer {
	ctx := context.Background()

	localStackContainer := utils.GetOrThrow(testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "wiremock/wiremock:3.13.1",
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor:   wait.ForListeningPort("8080"),
		},
	}))

	host := utils.GetOrThrow(localStackContainer.Host(ctx))
	port := utils.GetOrThrow(localStackContainer.MappedPort(ctx, "8080/tcp"))

	return WiremockContainer{
		url: fmt.Sprintf("http://%s:%s", host, port.Port()),
	}
}

func (p *WiremockContainer) Url() string {
	return p.url
}
