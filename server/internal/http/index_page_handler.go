package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
	_ "time/tzdata"

	l "github.com/marcelhfm/home_server/pkg/log"
	"github.com/marcelhfm/home_server/pkg/types"
	"github.com/marcelhfm/home_server/views"
)

func getDatasources(db *sql.DB) ([]types.Datasource, error) {
	getDatasourcesQuery := "SELECT id, name, status FROM datasources"

	rows, err := db.Query(getDatasourcesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasourceArry []types.Datasource

	for rows.Next() {
		var id int
		var name string
		var status string
		err = rows.Scan(&id, &name, &status)
		if err != nil {
			return nil, err
		}

		datasourceArry = append(datasourceArry, types.Datasource{Id: id, Name: name, Status: status})
	}

	return datasourceArry, nil
}

func getLastSeen(db *sql.DB, datasources []types.Datasource) ([]types.Datasource, error) {
	if len(datasources) == 0 {
		return nil, fmt.Errorf("datasources list empty")
	}

	placeholder := make([]string, len(datasources))
	params := make([]interface{}, len(datasources))
	for i, ds := range datasources {
		placeholder[i] = fmt.Sprintf("$%d", i+1)
		params[i] = ds.Id
	}

	getLastSeenQuery := fmt.Sprintf(`SELECT datasource_id, MAX(timestamp) AS latest_timestamp	FROM timeseries	WHERE	datasource_id IN (%s)	GROUP BY datasource_id;`, strings.Join(placeholder, ", "))

	rows, err := db.Query(getLastSeenQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	foundDatasources := make(map[int]time.Time)

	for rows.Next() {
		var id int
		var last_seen time.Time
		if err := rows.Scan(&id, &last_seen); err != nil {
			return nil, err
		}

		foundDatasources[id] = last_seen
	}

	l.Log.Debug().Msgf("Found datasources: %v for query: %s and params %v", foundDatasources, getLastSeenQuery, params)

	if err := rows.Err(); err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return nil, err
	}

	for i := range datasources {
		ds := datasources[i]
		last_seen, found := foundDatasources[ds.Id]

		if found {
			localTime := last_seen.In(loc)
			formattedTime := localTime.Format("15:04")
			datasources[i].Last_seen = formattedTime
		}
	}

	return datasources, nil
}

func IndexPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasources, err := getDatasources(db)
		if err != nil {
			l.Log.Error().Msgf("Error while fetching datasources: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		datasourcesLastSeen, err := getLastSeen(db, datasources)

		if err != nil {
			l.Log.Error().Msgf("Error getting last seen: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = views.Index(datasourcesLastSeen).Render(r.Context(), w)

		if err != nil {
			l.Log.Error().Msgf("Error rendering: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
