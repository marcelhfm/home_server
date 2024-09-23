package cron

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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

type Bouncy struct {
	lastNotif         time.Time
	threshholdReached bool
}

var (
	statefulBouncies = make(map[string]*Bouncy)
	debounce         = 10 * time.Minute
	mu               sync.Mutex
)

const DEBOUNCE_DEFAULT = 10 * time.Minute

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

func shouldSendNotification(notificationKey string) bool {
	mu.Lock()
	defer mu.Unlock()

	bouncy, exists := statefulBouncies[notificationKey]

	if !exists || (time.Since(bouncy.lastNotif) > DEBOUNCE_DEFAULT && bouncy.threshholdReached) {
		if !exists {
			bouncy = &Bouncy{}
			statefulBouncies[notificationKey] = bouncy
		}

		bouncy.lastNotif = time.Now()
		bouncy.threshholdReached = false

		l.Log.Info().Msgf("I should send a notification for %s", notificationKey)
		return true
	}

	l.Log.Info().Msgf("I should NOT send a notification for %s", notificationKey)
	return false
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
  WHERE timestamp >= NOW() - INTERVAL '8 hours'
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
		notificationKey := fmt.Sprintf("%s_%s", metricsData.DatasourceId, metricsData.Metric)
		bouncy, exists := statefulBouncies[notificationKey]
		if !exists {
			bouncy = &Bouncy{threshholdReached: true, lastNotif: time.Now().Add(-DEBOUNCE_DEFAULT)} // inital threshold is reached and last notif is before default debounce
			statefulBouncies[notificationKey] = bouncy
		}

		if metricsData.Metric == "co2" && metricsData.Value >= 1200 {
			message := fmt.Sprintf("Öffne ein Fenster! Aktuelle ppm: %d", metricsData.Value)

			if shouldSendNotification(notificationKey) {
				sendPush("Schlechte Luft", message)
			}

			if metricsData.Value < 1000 {
				mu.Lock()
				bouncy.threshholdReached = true
				mu.Unlock()
			}
		}

		if metricsData.Metric == "moisture" && metricsData.Value <= 4000 {
			message := fmt.Sprintf("Pflanze mit id %s. Letzter Messwert: %d%%", metricsData.DatasourceId, metricsData.Value/100)

			if shouldSendNotification(notificationKey) {
				sendPush("Eine Pflanze benötigt Wasser", message)
			}

			if metricsData.Value > 5500 {
				mu.Lock()
				bouncy.threshholdReached = true
				mu.Unlock()
			}
		}
	}
}
