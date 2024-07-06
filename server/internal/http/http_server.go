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

	router.HandleFunc("GET /", IndexPageHandler(db))
	router.HandleFunc("GET /ds/{id}", DatasourcePageHandler(db))
	router.HandleFunc("GET /ds/{id}/logs", LogPageHandler())
	router.HandleFunc("POST /api/ds/{id}/cmd/{cmd}", ApiCommandHandler(commandChannel, commandResponseChannel))
	router.HandleFunc("GET /api/ds/{id}/display_button", ApiDisplayButtonHandler(db))
	router.HandleFunc("GET /api/ds/{id}/data_pane", ApiDataPaneHandler(db))
	router.HandleFunc("GET /api/ds/{id}/logs", ApiLogHandler(db))
	fmt.Println("Http Server listening on port 8080")

	loggedRouter := LoggerMiddleware(router)
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}
