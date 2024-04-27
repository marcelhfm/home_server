package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var commandMap = map[string]int{
	"command_co2_display_off": 1,
	"command_co2_display_on":  2,
}

type CommandBody struct {
	Command string
}

type CommandRequest struct {
	Command      int
	DatasourceId int
}

func StartHttpServer(commandChannel chan<- CommandRequest) {
	http.HandleFunc("POST /api/datasource/command/", sendCommandHandler(commandChannel))
	fmt.Println("Http Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sendCommandHandler(commandChannel chan<- CommandRequest) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Println("http: Received request with wrong method")
			http.Error(w, "Method is not supported", http.StatusNotFound)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/api/datasource/command/")
		if id == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		datasourceId, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "ID is not int", http.StatusBadRequest)
			return
		}

		var c CommandBody

		err = json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			http.Error(w, "Body is in wrong format", http.StatusBadRequest)
		}

		commandCode, ok := commandMap[c.Command]
		if !ok {
			http.Error(w, "Invaild command", http.StatusBadRequest)
		}

		rq := CommandRequest{
			Command:      commandCode,
			DatasourceId: datasourceId,
		}

		log.Printf("Send command %d to datasource with id: %d", rq.Command, rq.DatasourceId)
		commandChannel <- rq

		fmt.Fprintf(w, "Sent data to datasoruce with id: %s", id)
	}
}
