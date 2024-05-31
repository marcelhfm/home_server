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

	"github.com/marcelhfm/home_server/internal/db"
	"github.com/marcelhfm/home_server/pkg/types"
)

// TODO: Store this information in the db
var picow_value_descr = [...]string{"datasourceId", "co2", "temperature", "humidity", "display_status"}

var connections = make(map[int]net.Conn)
var mut sync.Mutex

func StartTCPServer(db *sql.DB, commandChannel <-chan types.CommandRequest, commandResponseChannel chan<- types.CommandResponse) {
	ln, err := net.Listen("tcp", ":5001")
	if err != nil {
		fmt.Println("tcp: Error listening:", err.Error())
		os.Exit(1)
	}

	defer ln.Close()

	fmt.Println("TCP server listening on port 5001")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("tcp: Error accepting: ", err.Error())
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
			fmt.Printf("tcp: Sending command: %d to datasource: %d\n", command.Command, command.DatasourceId)
			err := sendCommand(conn, command.Command)
			commandResponseChannel <- types.CommandResponse{
				Id:           command.Id,
				Command:      command.Command,
				DatasourceId: command.DatasourceId,
				Error:        err,
			}
		} else {
			fmt.Printf("tcp: No connection found for datasource: %d\n", command.DatasourceId)
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
	defer conn.Close()

	reader := bufio.NewReader(conn)
	isFirstMessage := true
	var datasourceId int

	for {
		message, err := reader.ReadString('\n')

		if err != nil {
			fmt.Printf("tcp: Error on connection for datasource %d: %s\n", datasourceId, err.Error())
			mut.Lock()
			if current, ok := connections[datasourceId]; ok && current == conn {
				delete(connections, datasourceId)
				fmt.Printf("tcp: Removed connection for datasource %d due to error\n", datasourceId)
			}
			mut.Unlock()
			conn.Close()
			return
		}

		fmt.Printf("tcp: Received message: %s", message)
		values := parseCsv(message)

		if isFirstMessage {
			datasourceId = values[0]
			mut.Lock()
			exitingConn, exists := connections[datasourceId]
			if exists {
				fmt.Printf("tcp: Connection updated for datasource %d\n", datasourceId)
				if exitingConn != conn {
					exitingConn.Close()
				}
			} else {
				fmt.Printf("tcp: New Connection added for datasource %d\n", datasourceId)
			}
			connections[datasourceId] = conn
			mut.Unlock()
			isFirstMessage = false
		}

		for i := 1; i < len(values); i++ {
			value := values[i]

			db.IngestIotData(pg_db, datasourceId, picow_value_descr[i], value)

		}
		fmt.Println("tcp: Successfully inserted message")
	}
}

func parseCsv(message string) []int {
	messageTrimmed := strings.TrimSuffix(message, "\n")
	values := strings.Split(messageTrimmed, ",")
	intValues := make([]int, len(values))

	for i, value := range values {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("tcp: Error parsing int form string '%s': %s", value, err)
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
		fmt.Println("tcp: Error writing command to connection:", err)
		return err
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("tcp: Error flushing writer:", err)
		return err
	}

	return nil
}
