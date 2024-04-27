package main

import (
	"fmt"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/marcelhfm/home_server/internal/db"
	"github.com/marcelhfm/home_server/internal/http"
	"github.com/marcelhfm/home_server/internal/tcp"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %s\n", err)
	}

	db := db.Init_pq()
	defer db.Close()

	commandChannel := make(chan http.CommandRequest)

	go tcp.StartTCPServer(db, commandChannel)
	http.StartHttpServer(commandChannel)
}
