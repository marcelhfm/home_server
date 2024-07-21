package tcp

import (
	"bufio"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/marcelhfm/home_server/internal/db"
	l "github.com/marcelhfm/home_server/pkg/log"
	"github.com/marcelhfm/home_server/pkg/types"
)

var picow_value_descr = [...]string{"datasourceId", "co2", "temperature", "humidity", "display_status"}

var connections = make(map[int]net.Conn)
var mut sync.Mutex

func StartTCPServer(db *sql.DB, commandChannel <-chan types.CommandRequest, commandResponseChannel chan<- types.CommandResponse) {
	ln, err := net.Listen("tcp", ":5001")
	if err != nil {
		l.Log.Error().Msgf("tcp: Error listening: %v", err.Error())
		os.Exit(1)
	}

	defer ln.Close()

	l.Log.Info().Msg("TCP server listening on port 5001")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				l.Log.Error().Msgf("tcp: Error accepting: %v", err.Error())
				os.Exit(1)
			}
			fmt.Println("tcp: Connected to client: ", conn.RemoteAddr().String())

			go handleConnection(conn, db)
		}
	}()

	for command := range commandChannel {
		mut.Lock()
		conn, ok := connections[command.DatasourceId]
		mut.Unlock()

		if ok {
			l.Log.Info().Msgf("tcp: Sending command: %d to datasource: %d", command.Command, command.DatasourceId)
			err := sendCommand(conn, command.Command)
			commandResponseChannel <- types.CommandResponse{
				Id:           command.Id,
				Command:      command.Command,
				DatasourceId: command.DatasourceId,
				Error:        err,
			}
		} else {
			l.Log.Warn().Msgf("tcp: No connection found for datasource: %d", command.DatasourceId)
			commandResponseChannel <- types.CommandResponse{
				Id:           command.Id,
				Command:      command.Command,
				DatasourceId: command.DatasourceId,
				Error:        fmt.Errorf("Connection does not exist"),
			}
		}
	}
}

func handleConnection(conn net.Conn, pg_db *sql.DB) {
	var datasourceId int // Declare datasourceId at the start of the function

	defer func() {
		mut.Lock()
		if current, ok := connections[datasourceId]; ok && current == conn {
			delete(connections, datasourceId)
			updateDatasourceStatus(pg_db, datasourceId, "DISCONNECTED")
			l.Log.Info().Msgf("tcp: Removed connection and updated status to DISCONNECTED for datasource %d", datasourceId)
		}
		mut.Unlock()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	isFirstMessage := true

	for {
		conn.SetReadDeadline(time.Now().Add(1 * time.Minute)) // Set a read deadline to detect timeouts

		message, err := reader.ReadString('\n')

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				l.Log.Error().Msgf("tcp: Timeout on connection for datasource %d: %s", datasourceId, err.Error())
			} else {
				l.Log.Error().Msgf("tcp: Error on connection for datasource %d: %s", datasourceId, err.Error())
			}
			return
		}

		fmt.Printf("tcp: Received message: %s", message)
		values := parseCsv(message)

		if isFirstMessage {
			datasourceId = values[0]
			mut.Lock()
			existingConn, exists := connections[datasourceId]
			if exists {
				l.Log.Info().Msgf("tcp: Connection updated for datasource %d", datasourceId)
				if existingConn != conn {
					existingConn.Close()
				}
			} else {
				l.Log.Info().Msgf("tcp: New Connection added for datasource %d", datasourceId)
			}
			connections[datasourceId] = conn
			updateDatasourceStatus(pg_db, datasourceId, "CONNECTED")
			mut.Unlock()
			isFirstMessage = false
		}

		currTimestamp := time.Now().Format(time.RFC3339)

		for i := 1; i < len(values); i++ {
			value := values[i]
			db.IngestIotData(pg_db, datasourceId, picow_value_descr[i], value, currTimestamp)
		}
		l.Log.Debug().Msg("tcp: Successfully inserted message")
	}
}

func parseCsv(message string) []int {
	messageTrimmed := strings.TrimSuffix(message, "\n")
	values := strings.Split(messageTrimmed, ",")
	intValues := make([]int, len(values))

	for i, value := range values {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			l.Log.Error().Msgf("tcp: Error parsing int form string '%s': %s", value, err)
			continue
		}
		intValues[i] = intValue
	}

	return intValues
}

func sendCommand(conn net.Conn, command int) error {
	writer := bufio.NewWriter(conn)

	message := fmt.Sprintf("%d\n", command)

	_, err := writer.WriteString(message)
	if err != nil {
		l.Log.Error().Msgf("tcp: Error writing command to connection: %v", err)
		return err
	}

	err = writer.Flush()
	if err != nil {
		l.Log.Error().Msgf("tcp: Error flushing writer: %v", err)
		return err
	}

	return nil
}

func updateDatasourceStatus(db *sql.DB, datasourceId int, status string) {
	query := `UPDATE datasources SET status = $1 WHERE id = $2`
	_, err := db.Exec(query, status, datasourceId)
	if err != nil {
		l.Log.Error().Msgf("tcp: Error updating status for datasource %d to %s: %s", datasourceId, status, err.Error())
	}
}
