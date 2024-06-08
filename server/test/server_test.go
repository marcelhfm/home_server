package test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/marcelhfm/home_server/internal/tcp"
	"github.com/marcelhfm/home_server/pkg/types"
)

func startPostgresContainer(t *testing.T) (testcontainers.Container, *sql.DB) {
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

	db.Exec(`CREATE TABLE IF NOT EXISTS timeseries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  datasource_id INT NOT NULL,
  metric VARCHAR NOT NULL,
  value JSONB NOT NULL,
  timestamp TIMESTAMPTZ NOT NULL
);`)

	require.NoError(t, err)

	return pgContainer, db
}

type TimeseriesData struct {
	Metric    string
	Value     int
	Timestamp string
}

func TestTcpServer(t *testing.T) {
	pgContainer, db := startPostgresContainer(t)
	defer pgContainer.Terminate(context.Background())
	defer db.Close()

	commandChannel := make(chan types.CommandRequest)
	commandResponseChannel := make(chan types.CommandResponse)

	go tcp.StartTCPServer(db, commandChannel, commandResponseChannel)

	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "localhost:5001")
	require.NoError(t, err)
	defer conn.Close()

	fmt.Fprintf(conn, "1,400,22,60,1\n")

	time.Sleep(1 * time.Second)

	id := uuid.New()

	commandChannel <- types.CommandRequest{
		Id:           id,
		Command:      100,
		DatasourceId: 1,
	}

	select {
	case res := <-commandResponseChannel:
		assert.Equal(t, id, res.Id)
		assert.Equal(t, 100, res.Command)
		assert.Equal(t, 1, res.DatasourceId)
		assert.NoError(t, res.Error)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for command res")
	}

	query := fmt.Sprintf("SELECT metric, value, timestamp FROM timeseries WHERE datasource_id = %s AND timestamp >=NOW() - INTERVAL '30 minutes' ORDER BY timestamp desc", "1")
	rows, err := db.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	var res []TimeseriesData

	for rows.Next() {
		var metric string
		var value int
		var ts string

		err := rows.Scan(&metric, &value, &ts)
		require.NoError(t, err)
		res = append(res, TimeseriesData{Metric: metric, Value: value, Timestamp: ts})
	}

	for _, entry := range res {
		if entry.Metric == "temperature" {
			assert.Equal(t, 22, entry.Value)
		}
		if entry.Metric == "co2" {
			assert.Equal(t, 400, entry.Value)
		}
		if entry.Metric == "humidiy" {
			assert.Equal(t, 60, entry.Value)
		}
		if entry.Metric == "display_status" {
			assert.Equal(t, 1, entry.Value)
		}
	}
}
