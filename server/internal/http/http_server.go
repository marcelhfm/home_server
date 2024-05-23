package http

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/marcelhfm/home_server/pkg/types"
)

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
	router.HandleFunc("GET /ds/{id}", DatasourceHandler())
	router.HandleFunc("POST /api/ds/{id}/cmd/{cmd}", SendCommandHandler(commandChannel, commandResponseChannel))
	fmt.Println("Http Server listening on port 8080")

	loggedRouter := LoggerMiddleware(router)
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}
