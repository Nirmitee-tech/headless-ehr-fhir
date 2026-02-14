package integration

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// startWithTestcontainers spins up a postgres:16-alpine container using the Docker CLI
// and returns the connection string and a cleanup function.
func startWithTestcontainers(ctx context.Context) (string, func(), error) {
	// Find a free port
	port, err := getFreePort()
	if err != nil {
		return "", nil, fmt.Errorf("find free port: %w", err)
	}

	containerName := fmt.Sprintf("ehr-integration-test-%d", port)

	// Remove any existing container with the same name
	exec.CommandContext(ctx, "docker", "rm", "-f", containerName).Run()

	// Start postgres:16-alpine container
	cmd := exec.CommandContext(ctx, "docker", "run",
		"--name", containerName,
		"-d",
		"-p", fmt.Sprintf("%d:5432", port),
		"-e", "POSTGRES_USER=testuser",
		"-e", "POSTGRES_PASSWORD=testpass",
		"-e", "POSTGRES_DB=ehrtest",
		"postgres:16-alpine",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", nil, fmt.Errorf("docker run: %w\noutput: %s", err, string(output))
	}
	containerID := strings.TrimSpace(string(output))

	cleanup := func() {
		exec.Command("docker", "rm", "-f", containerID).Run()
	}

	// Wait for postgres to be ready
	connStr := fmt.Sprintf("postgres://testuser:testpass@localhost:%d/ehrtest?sslmode=disable", port)
	if err := waitForPostgres(ctx, connStr, 30*time.Second); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("wait for postgres: %w", err)
	}

	return connStr, cleanup, nil
}

// getFreePort returns a free TCP port on localhost.
func getFreePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// waitForPostgres waits until postgres accepts connections and responds to queries.
func waitForPostgres(ctx context.Context, connStr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try to connect using pgxpool
		connCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		pool, err := pgxpool.New(connCtx, connStr)
		if err == nil {
			err = pool.Ping(connCtx)
			pool.Close()
			cancel()
			if err == nil {
				return nil
			}
		} else {
			cancel()
		}

		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("postgres not ready after %v", timeout)
}
