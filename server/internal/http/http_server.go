package http

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/marcelhfm/home_server/pkg/types"
	"github.com/marcelhfm/home_server/views"
)

var commandMap = map[string]int{
	"command_co2_display_off": 1,
	"command_co2_display_on":  2,
}

func getDatasources(db *sql.DB) ([]types.Datasource, error) {
	getDatasourcesQuery := "SELECT * FROM datasources"

	rows, err := db.Query(getDatasourcesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasourceArry []types.Datasource

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}

		datasourceArry = append(datasourceArry, types.Datasource{Id: id, Name: name})
	}

	return datasourceArry, nil
}

func IndexHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasources, err := getDatasources(db)
		if err != nil {
			fmt.Println("Error fetching datasources", err)
		}

		views.Index(datasources).Render(r.Context(), w)
	}
}

func StartHttpServer(db *sql.DB, commandChannel chan<- types.CommandRequest, commandResponseChannel <-chan types.CommandResponse) {
	http.HandleFunc("GET /", IndexHandler(db))
	http.HandleFunc("POST /api/datasource/command/", SendCommandHandler(commandChannel, commandResponseChannel))
	fmt.Println("Http Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
