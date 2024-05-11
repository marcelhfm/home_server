package http

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/marcelhfm/home_server/pkg/types"
)

var commandMap = map[string]int{
	"command_co2_display_off": 1,
	"command_co2_display_on":  2,
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func StartHttpServer(db *sql.DB, commandChannel chan<- types.CommandRequest, commandResponseChannel <-chan types.CommandResponse) {
	router := http.NewServeMux()

	router.HandleFunc("GET /", IndexHandler(db))
	router.HandleFunc("POST /api/datasource/command/", SendCommandHandler(commandChannel, commandResponseChannel))
	fmt.Println("Http Server listening on port 8080")

	loggedRouter := LoggerMiddleware(router)
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}
