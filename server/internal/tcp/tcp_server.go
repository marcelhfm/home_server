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
	"github.com/marcelhfm/home_server/internal/http"
)

// TODO: Store this information in the db
var picow_value_descr = [...]string{"datasourceId", "co2", "temperature", "humidity"}

var connections = make(map[int]net.Conn)
var mut sync.Mutex

func StartTCPServer(db *sql.DB, commandChannel <-chan http.CommandRequest) {
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
			sendCommand(conn, command.Command)
		} else {
			fmt.Printf("tcp: No connection found for datasource: %d\n", command.DatasourceId)
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
			fmt.Println("tcp: Error reading: ", err.Error())
			break
		}
		fmt.Printf("tcp: Received message: %s", message)
		values := parseCsv(message)

		if isFirstMessage {
			datasourceId = values[0]
			mut.Lock()
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

	mut.Lock()
	delete(connections, datasourceId)
	mut.Unlock()
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

func sendCommand(conn net.Conn, command int) {
	writer := bufio.NewWriter(conn)

	message := fmt.Sprintf("%d\n", command)

	_, err := writer.WriteString(message)
	if err != nil {
		fmt.Println("tcp: Error writing command to connection:", err)
		return
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("tcp: Error flushing writer:", err)
	}
}