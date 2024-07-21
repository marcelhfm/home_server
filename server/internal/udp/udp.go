package udp

import (
	"database/sql"
	"net"
	"strconv"
	"time"

	l "github.com/marcelhfm/home_server/pkg/log"
)

const (
	UDP_PORT = 12345
	BUF_SIZE = 1024
)

func ingestLogs(db *sql.DB, datasourceId int, message string, ts string) {
	sqlStatement := `INSERT INTO logs (datasource_id, message, timestamp) VALUES ($1, $2, $3)`

	_, err := db.Exec(sqlStatement, datasourceId, message, ts)

	if err != nil {
		l.Log.Error().Msgf("db: An error occured while trying to insert into database: %s", err)
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
		l.Log.Error().Msgf("Error starting UDP server: %v", err)
	}
	defer conn.Close()

	l.Log.Info().Msgf("UDP server listening on %s:%d", addr.IP.String(), addr.Port)

	buffer := make([]byte, BUF_SIZE)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			l.Log.Error().Msgf("Error reading from UDP: %v", err)
			continue
		}
		message := string(buffer[:n])
		l.Log.Debug().Msgf("UDP: Received message %s", message)
		dsId, message, err := parseMessage(message)
		if err != nil {
			l.Log.Error().Msgf("UDP: Error parsing message. err: %s", err)
			continue
		}

		currTimestamp := time.Now().Format(time.RFC3339)
		ingestLogs(db, dsId, message, currTimestamp)
	}
}
