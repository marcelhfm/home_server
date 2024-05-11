package http

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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

func getLastSeen(db *sql.DB, datasources []types.Datasource) ([]types.DatasourceLastSeen, error) {
	if len(datasources) == 0 {
		return nil, fmt.Errorf("datasources list empty")
	}

	placeholder := make([]string, len(datasources))
	params := make([]interface{}, len(datasources))
	for i, ds := range datasources {
		placeholder[i] = fmt.Sprintf("$%d", i+1)
		params[i] = ds.Id
	}

	getLastSeenQuery := fmt.Sprintf("SELECT timeseries.datasource_id, timestamp FROM timeseries WHERE timeseries.datasource_id IN (%s) ORDER BY timestamp DESC LIMIT 1", strings.Join(placeholder, ", "))

	rows, err := db.Query(getLastSeenQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.DatasourceLastSeen
	foundDatasources := make(map[int]time.Time)

	for rows.Next() {
		var id int
		var last_seen time.Time
		if err := rows.Scan(&id, &last_seen); err != nil {
			return nil, err
		}

		foundDatasources[id] = last_seen
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, ds := range datasources {
		last_seen, found := foundDatasources[ds.Id]

		if !found {
			results = append(results, types.DatasourceLastSeen{Datasource: ds, Last_seen: nil})
		} else {
			results = append(results, types.DatasourceLastSeen{Datasource: ds, Last_seen: &last_seen})
		}
	}

	return results, nil
}

func IndexHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasources, err := getDatasources(db)
		if err != nil {
			fmt.Println("Error fetching datasources", err)
		}

		datasourcesLastSeen, err := getLastSeen(db, datasources)

		if err != nil {
			fmt.Println("Error getting last seen", err)
		}

		views.Index(datasourcesLastSeen).Render(r.Context(), w)
	}
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
