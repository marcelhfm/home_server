package udp

import (
	"database/sql"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	UDP_PORT = 12345
	BUF_SIZE = 1024
)

func ingestLogs(db *sql.DB, datasourceId int, message string, ts string) {

	sqlStatement := `INSERT INTO logs (datasource_id, message, timestamp) VALUES ($1, $2, $3)`

	_, err := db.Exec(sqlStatement, datasourceId, message, ts)

	if err != nil {
		log.Printf("db: An error occured while trying to insert into database: %s", err)
	}
}

func parseMessage(message string) (int, string, error) {
	slices := strings.Split(message, ";")

	dsId, err := strconv.Atoi(slices[0])
	if err != nil {
		return -1, "", err
	}

	return dsId, slices[1], err
}

func StartLogServer(db *sql.DB) {
	addr := net.UDPAddr{
		Port: UDP_PORT,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Error starting UDP server: %v\n", err)
	}
	defer conn.Close()

	log.Printf("UDP server listening on %s:%d\n", addr.IP.String(), addr.Port)

	buffer := make([]byte, BUF_SIZE)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v\n", err)
			continue
		}
		message := string(buffer[:n])
		dsId, message, err := parseMessage(message)
		if err != nil {
			log.Printf("UDP: Error parsing message. err: %s", err)
			continue
		}

		currTimestamp := time.Now().Format(time.RFC3339)
		ingestLogs(db, dsId, message, currTimestamp)
	}
}
