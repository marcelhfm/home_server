package cron

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/marcelhfm/home_server/pkg/config"
	"github.com/marcelhfm/home_server/pkg/log"
)

type MetricsData struct {
	DatasourceId string
	Timestamp    time.Time
	Metric       string
	Value        int
}

func sendPush(title string, message string) {
	pushUser := config.GetenvStr("PUSHOVER_USER")
	pushToken := config.GetenvStr("PUSHOVER_TOKEN")

	type Body struct {
		Token   string `json:"token"`
		User    string `json:"user"`
		Message string `json:"message"`
		Title   string `json:"title"`
	}

	body := Body{
		Token:   pushToken,
		User:    pushUser,
		Title:   title,
		Message: message,
	}

	marshalled, err := json.Marshal(body)
	if err != nil {
		l.Log.Error().Msgf("Error sending push. Can't marshall request body: %v", err)
		return
	}

	req, err := http.NewRequest("POST", "https://api.pushover.net/1/messages.json", bytes.NewReader(marshalled))
	if err != nil {
		l.Log.Error().Msgf("Error sending push. Error creating http request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		l.Log.Error().Msgf("Error sending push. Error sending http request: %v", err)
		return
	}

	if res.StatusCode > 299 {
		l.Log.Error().Msgf("Error sending push. Received status code: %d", res.StatusCode)
	} else {
		l.Log.Info().Msgf("Successfully send push. Status code: %d", res.StatusCode)
	}
}

func Notifications(db *sql.DB) {
	l.Log.Info().Msg("Notifications: Checking timeseries data...")
	lastMetricsQuery := `
  SELECT DISTINCT ON (datasource_id, metric) 
      datasource_id,
      metric,
      value,
      timestamp
  FROM timeseries
  ORDER BY datasource_id, metric, timestamp DESC;`

	rows, err := db.Query(lastMetricsQuery)
	if err != nil {
		l.Log.Error().Msgf("Notifications: Error executing lastMetricsQuery: %v", err)
		return
	}
	defer rows.Close()

	var res []MetricsData

	for rows.Next() {
		var datasourceId string
		var ts time.Time
		var metric string
		var value int

		if err := rows.Scan(&datasourceId, &metric, &value, &ts); err != nil {
			l.Log.Error().Msgf("Notifications: Error getting rows lastMetricsQuery: %v", err)
			return
		}

		res = append(res, MetricsData{DatasourceId: datasourceId, Timestamp: ts, Metric: metric, Value: value})
	}

	for _, metricsData := range res {
		if metricsData.Metric == "co2" && metricsData.Value >= 1200 {
			message := fmt.Sprintf("Öffne ein Fenster! Aktuelle ppm: %d", metricsData.Value)
			sendPush("Schlechte Luft", message)
		}
		if metricsData.Metric == "moisture" && metricsData.Value <= 4000 {
			message := fmt.Sprintf("Pflanze mit id %s. Letzter Messwert: %d%%", metricsData.DatasourceId, metricsData.Value/100)
			sendPush("Eine Pflanze benötigt Wasser", message)
		}
	}
}
