//go:build integration

package testsupport

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Redis is a running Redis container plus its dialable address.
type Redis struct {
	Addr      string
	container testcontainers.Container
}

// StartRedis boots a disposable Redis container (the asynq broker) and returns
// its host:port address.
func StartRedis(ctx context.Context) (*Redis, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("start redis: %w", err)
	}
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}
	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}
	return &Redis{Addr: fmt.Sprintf("%s:%s", host, port.Port()), container: container}, nil
}

// Close terminates the container.
func (r *Redis) Close(ctx context.Context) {
	if r.container != nil {
		_ = r.container.Terminate(ctx)
	}
}
