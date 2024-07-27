package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	_ "time/tzdata"

	l "github.com/marcelhfm/home_server/pkg/log"
	"github.com/marcelhfm/home_server/pkg/types"
	"github.com/marcelhfm/home_server/views/components"
)

type LogData struct {
	Message   string
	Timestamp string
}

func getLogs(db *sql.DB, dsId string, timerange string, level string) ([]LogData, error) {
	l.Log.Info().Msgf("Fetching logs for ds %s", dsId)

	var where_level string = ""

	if level == "info" {
		where_level = "AND message LIKE '%[0;32m%'"
	}
	if level == "debug" {
		where_level = "AND message LIKE '%[0;34m%'"
	}
	if level == "error" {
		where_level = "AND message LIKE '%[0;31m%'"
	}
	if level == "warning" {
		where_level = "AND message LIKE '%[0;33m%'"
	}

	var logQuery string

	if timerange == "0" {
		logQuery = fmt.Sprintf("SELECT message, timestamp FROM logs WHERE datasource_id = %s %s ORDER BY timestamp desc LIMIT 1000", dsId, where_level)
	} else {
		logQuery = fmt.Sprintf("SELECT message, timestamp FROM logs WHERE datasource_id = %s AND timestamp >=NOW() - INTERVAL '%s' %s ORDER BY timestamp desc LIMIT 1000", dsId, timerange, where_level)
	}

	rows, err := db.Query(logQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []LogData

	for rows.Next() {
		var message string
		var ts string

		if err := rows.Scan(&message, &ts); err != nil {
			return nil, err
		}

		res = append(res, LogData{Message: message, Timestamp: ts})
	}

	l.Log.Info().Msgf("Got %d log messages for range %s and ds %s", len(res), timerange, dsId)
	return res, nil
}

func timeRangeToPqInterval(timerange string) string {
	switch timerange {
	case "0":
		return "0"
	case "0.5":
		return "30 Minutes"
	case "1":
		return "1 hour"
	case "24":
		return "24 hours"
	default:
		return "0"
	}
}

func getColorFromCode(str string) (string, string) {
	colorCodes := map[string][2]string{
		"\x1b[0;31m": {"text-red-600", "ERROR"},
		"\x1b[0;32m": {"text-green-600", "INFO"},
		"\x1b[0;33m": {"text-yellow-600", "WARN"},
		"\x1b[0;34m": {"text-blue-600", "DEBUG"},
	}

	for code, colorLevel := range colorCodes {
		if strings.Contains(str, code) {
			return colorLevel[0], colorLevel[1]
		}
	}
	return "unknown", "UNKNOWN"
}

func removeAnsiCodes(str string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(str, "")
}

func formatMessages(messages []LogData) []types.FormattedLogs {
	var results []types.FormattedLogs
	for _, message := range messages {
		color, level := getColorFromCode(message.Message)

		results = append(results, types.FormattedLogs{Message: removeAnsiCodes(message.Message), Color: color, Timestamp: message.Timestamp, Level: level})
	}

	return results
}

func ApiLogHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dsId := r.PathValue("id")
		timeRange := r.URL.Query().Get("timerange")
		level := r.URL.Query().Get("loglevel")
		l.Log.Info().Msgf("ApiLogHandler: Called for ds %s, level %s and range: %s", dsId, level, timeRange)

		formattedTimeRange := timeRangeToPqInterval(timeRange)

		logs, err := getLogs(db, dsId, formattedTimeRange, level)

		if err != nil {
			l.Log.Error().Msgf("Error in ApiLogHandler: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		formattedLogs := formatMessages(logs)

		if len(logs) == 0 {
			err = components.LogComponent(dsId, formattedLogs, false).Render(r.Context(), w)
		} else {
			err = components.LogComponent(dsId, formattedLogs, true).Render(r.Context(), w)
		}

		if err != nil {
			l.Log.Error().Msgf("Error in ApiLogHandler: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
