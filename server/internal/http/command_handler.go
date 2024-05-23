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

var errorAlert string = `<div id="alert" class="transitio-opacity ease-in duration-150 max-w-md ml-4 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative" role="alert">%s</div>`
var successAlert string = `<div id="alert" class="transition-opacity ease-in duration-150 max-w-md ml-4 bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded relative" role="alert">%s</div>`

func SendCommandHandler(commandChannel chan<- types.CommandRequest, commandResponseChannel <-chan types.CommandResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Println("http: Received request with wrong method")
			w.Write([]byte(fmt.Sprintf(errorAlert, "Something went wrong :( (Wront http method)")))
			return
		}

		id := r.PathValue("id")

		datasourceId, err := strconv.Atoi(id)
		if err != nil {
			w.Write([]byte(fmt.Sprintf(errorAlert, "Something went wrong :( (Invalid id)")))
			return
		}

		cmd := r.PathValue("cmd")

		commandCode, ok := commandMap[cmd]
		if !ok {
			w.Write([]byte(fmt.Sprintf(errorAlert, "Something went wrong :( (Invalid command)")))
			return
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
			w.Write([]byte(fmt.Sprintf(errorAlert, fmt.Sprintf("Internal server error (%v)", response.Error))))
		} else {
			w.Write([]byte(fmt.Sprintf(errorAlert, "Command sent successfully!")))
		}
	}
}
