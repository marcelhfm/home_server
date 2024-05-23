package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/marcelhfm/home_server/pkg/types"
)

var commandMap = map[string]int{
	"command_co2_display_off": 1,
	"command_co2_display_on":  2,
}

func SendCommandHandler(commandChannel chan<- types.CommandRequest, commandResponseChannel <-chan types.CommandResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Println("http: Received request with wrong method")
			http.Error(w, "Method is not supported", http.StatusNotFound)
			return
		}

		id := r.PathValue("id")

		datasourceId, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "ID is not int", http.StatusBadRequest)
			return
		}

		cmd := r.PathValue("cmd")

		commandCode, ok := commandMap[cmd]
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
			fmt.Fprintf(w, "Command %s sent to datasource %d successfully.", cmd, datasourceId)
		}
	}
}
