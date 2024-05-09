package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/marcelhfm/home_server/pkg/types"
)

func SendCommandHandler(commandChannel chan<- types.CommandRequest, commandResponseChannel <-chan types.CommandResponse) http.HandlerFunc {
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

		var c types.CommandBody

		err = json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			http.Error(w, "Body is in wrong format", http.StatusBadRequest)
		}

		commandCode, ok := commandMap[c.Command]
		if !ok {
			http.Error(w, "Invaild command", http.StatusBadRequest)
		}

		rq := types.CommandRequest{
			Id:           uuid.New(),
			Command:      commandCode,
			DatasourceId: datasourceId,
		}

		log.Printf("Send command %d to datasource with id: %d", rq.Command, rq.DatasourceId)
		commandChannel <- rq

		response := <-commandResponseChannel

		if response.Error != nil {
			http.Error(w, response.Error.Error(), http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "Command %s sent to datasource %d successfully.", c.Command, datasourceId)
		}
	}
}
