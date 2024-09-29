package testsetup

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var Container testcontainers.Container
var Db *sql.DB

func StartPostgresContainer(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "timescale/timescaledb:latest-pg14",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "admin",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	mappedPort, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	host := "localhost"
	port := mappedPort.Port()
	user := "admin"
	password := "password"
	dbname := "testdb"

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS timeseries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  datasource_id INT NOT NULL,
  metric VARCHAR NOT NULL,
  value JSONB NOT NULL,
  timestamp TIMESTAMPTZ NOT NULL
);`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS datasources (
  id INT PRIMARY KEY,
  name TEXT NOT NULL,
  status TEXT NOT NULL
);`)

	db.Exec(`INSERT INTO datasources (id, name, status) VALUES (1, 'pico_w', 'DISCONNECTED')`)

	require.NoError(t, err)

	Db = db
	Container = pgContainer
}

func TearDownTests(t *testing.T) {
	t.Helper()
	require.NoError(t, Container.Terminate(context.Background()))
	require.NoError(t, Db.Close())
}
