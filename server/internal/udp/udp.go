package udp

import (
	"database/sql"
	"log"
	"net"
	"strconv"
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
	dsIdString := string(message[0])
	var parsedMessage string

	runes := []rune(message)
	if len(runes) > 1 {
		parsedMessage = string(runes[2:])
	}

	dsId, err := strconv.Atoi(dsIdString)
	if err != nil {
		return -1, "", err
	}

	return dsId, parsedMessage, err
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
		log.Printf("UDP: Received message %s\n", message)
		dsId, message, err := parseMessage(message)
		if err != nil {
			log.Printf("UDP: Error parsing message. err: %s\n", err)
			continue
		}

		currTimestamp := time.Now().Format(time.RFC3339)
		ingestLogs(db, dsId, message, currTimestamp)
	}
}
