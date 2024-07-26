package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	l "github.com/marcelhfm/home_server/pkg/log"
)

// Get all datasource

type Datasource struct {
	Id     int
	Name   string
	Status string
	Type   string
}

func getAllDatasources(db *sql.DB) ([]Datasource, error) {
	l.Log.Debug().Msgf("Fetching datasources")
	getAllDatasourceQuery := "SELECT * FROM datasources"

	var datasources []Datasource
	rows, err := db.Query(getAllDatasourceQuery)
	if err != nil {
		return datasources, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var status string
		var dsType string
		err = rows.Scan(&id, &name, &status, &dsType)

		if err != nil {
			return datasources, err
		}

		datasources = append(datasources, Datasource{Id: id, Name: name, Status: status, Type: dsType})
	}

	return datasources, nil
}

func ApiHomeBridgeGetDatasources(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		datasources, err := getAllDatasources(db)
		if err != nil {
			l.Log.Error().Msgf("Error getting datasources. err: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(datasources)
	}
}

// Get last metric

type Metric struct {
	Value     int
	Timestamp string
}

func getMetric(db *sql.DB, dsId string, metric string) (*Metric, error) {
	l.Log.Debug().Msgf("Fetching metric")
	timeseriesQuery := fmt.Sprintf("SELECT value, timestamp FROM timeseries WHERE datasource_id = %s AND metric = '%s' ORDER BY timestamp desc LIMIT 1", dsId, metric)
	l.Log.Debug().Msgf("query: %s", timeseriesQuery)

	var value int
	var timestamp string

	err := db.QueryRow(timeseriesQuery).Scan(&value, &timestamp)
	if err != nil {
		return nil, err
	}

	metricObj := Metric{Value: value, Timestamp: timestamp}

	return &metricObj, nil
}

func ApiHomeBridgeGetMetric(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dsId := r.PathValue("id")
		metric := r.PathValue("metric")

		metricObj, err := getMetric(db, dsId, metric)
		if err != nil {
			l.Log.Error().Msgf("Error getting metric. err: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(*metricObj)
	}
}
