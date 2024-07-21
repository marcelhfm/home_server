package http

import (
	"database/sql"
	"fmt"
	"net/http"

	l "github.com/marcelhfm/home_server/pkg/log"
)

func getDisplayStatus(db *sql.DB, dsId string) (bool, error) {
	l.Log.Info().Msgf("Fetching display_status for ds %s", dsId)
	getTimeseriesQuery := fmt.Sprintf("SELECT * FROM timeseries WHERE datasource_id=%s AND metric='display_status' order by timestamp desc LIMIT 1", dsId)

	rows, err := db.Query(getTimeseriesQuery)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var dsId int
		var metric string
		var value int
		var timestamp string
		err = rows.Scan(&id, &dsId, &metric, &value, &timestamp)

		if err != nil {
			return false, err
		}

		return value == 1, nil
	}

	return false, nil
}

func ApiDisplayButtonHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			l.Log.Error().Msg("Received req with wrong method")
			http.Error(w, "Wrong method", http.StatusBadRequest)
			return
		}

		dsId := r.PathValue("id")

		var turnOffButton string = fmt.Sprintf(`<div hx-target="#alert" hx-swap="outerHTML" hx-post="/api/ds/%s/cmd/command_co2_display_off" class="ml-2 bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded">Turn Off</div>`, dsId)
		var turnOnButton string = fmt.Sprintf(`<div hx-target="#alert" hx-swap="outerHTML" hx-post="/api/ds/%s/cmd/command_co2_display_on" class="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded">Turn On</div>`, dsId)

		display_status, err := getDisplayStatus(db, dsId)

		if err != nil {
			l.Log.Error().Msgf("Error fetching display_status: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if display_status {
			w.Write([]byte(turnOffButton))
		} else {
			w.Write([]byte(turnOnButton))
		}
	}
}
