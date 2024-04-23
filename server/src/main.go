package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var picow_value_descr = [...]string{"datasourceId", "co2", "temperature", "humidity"}

func init_pq() *sql.DB {
	host, err := GetenvStr("PQ_HOST")
	if err != nil {
		log.Fatal(err)
	}

	port, err := GetenvInt("PQ_PORT")
	if err != nil {
		log.Fatal(err)
	}

	user, err := GetenvStr("PQ_USER")
	if err != nil {
		log.Fatal(err)
	}

	password, err := GetenvStr("PQ_PASSWORD")
	if err != nil {
		log.Fatal(err)
	}

	dbname, err := GetenvStr("PQ_DBNAME")
	if err != nil {
		log.Fatal(err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to db")

	return db
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %s\n", err)
	}

	db := init_pq()
	defer db.Close()

	ln, err := net.Listen("tcp", ":5001")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer ln.Close()

	fmt.Println("TCP server listening on port 5001")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Connected to client: ", conn.RemoteAddr().String())

		go handleRequest(conn, db)
	}
}

func handleRequest(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	sqlStatement := `INSERT INTO timeseries (datasource_id, metric, value, timestamp) VALUES ($1, $2, $3, now())`

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			break
		}
		fmt.Printf("Received: %s", message)
		values := parseCsv(message)

		datasourceId := values[0]

		for i := 1; i < len(values); i++ {
			value := values[i]

			_, err = db.Exec(sqlStatement, datasourceId, picow_value_descr[i], value)

			if err != nil {
				fmt.Printf("An error occured while trying to insert into database: %s", err)
			}
		}
		fmt.Println("Successfully inserted message")
	}
}

func parseCsv(message string) []int {
	messageTrimmed := strings.TrimSuffix(message, "\n")
	values := strings.Split(messageTrimmed, ",")
	intValues := make([]int, len(values))

	for i, value := range values {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("Error parsing int form string '%s': %s", value, err)
			continue
		}
		intValues[i] = intValue
	}

	return intValues
}
