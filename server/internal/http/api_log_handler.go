package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	_ "time/tzdata"

	"github.com/marcelhfm/home_server/pkg/types"
	"github.com/marcelhfm/home_server/views/components"
)

type LogData struct {
	Message   string
	Timestamp string
}

func getLogs(db *sql.DB, dsId string, timerange string) ([]LogData, error) {
	fmt.Println("ApiLogHandler: Fetching logs for ds", dsId)

	var logQuery string

	if timerange == "0" {
		logQuery = fmt.Sprintf("SELECT message, timestamp FROM logs WHERE datasource_id = %s ORDER BY timestamp desc LIMIT 100", dsId)
	} else {
		logQuery = fmt.Sprintf("SELECT message, timestamp FROM logs WHERE datasource_id = %s AND timestamp >=NOW() - INTERVAL '%s' ORDER BY timestamp desc LIMIT 100", dsId, timerange)
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

	fmt.Printf("ApiLogHandler: Got %d log messages for range %s and ds %s", len(res), timerange, dsId)
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

func getColorFromCode(str string) string {
	colorCodes := map[string]string{
		"\x1b[0;30m": "text-black-600",
		"\x1b[0;31m": "text-red-600",
		"\x1b[0;32m": "text-green-600",
		"\x1b[0;33m": "text-yellow-600",
		"\x1b[0;34m": "text-blue-600",
		"\x1b[0;35m": "text-magenta-600",
		"\x1b[0;36m": "text-cyan-600",
		"\x1b[0;37m": "text-white-600",
	}

	for code, color := range colorCodes {
		if strings.Contains(str, code) {
			return color
		}
	}
	return "unknown"
}

func removeAnsiCodes(str string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(str, "")
}

func formatMessages(messages []LogData) []types.FormattedLogs {
	var results []types.FormattedLogs
	for _, message := range messages {
		color := getColorFromCode(message.Message)

		results = append(results, types.FormattedLogs{Message: fmt.Sprintf("[%s] %s", message.Timestamp, removeAnsiCodes(message.Message)), Color: color})
	}

	return results
}

func ApiLogHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dsId := r.PathValue("id")
		timeRange := r.URL.Query().Get("timerange")
		fmt.Printf("ApiLogHandler: Called for ds %s and range: %s\n", dsId, timeRange)

		formattedTimeRange := timeRangeToPqInterval(timeRange)

		logs, err := getLogs(db, dsId, formattedTimeRange)

		if err != nil {
			fmt.Println("ApiLogHandler: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		formattedLogs := formatMessages(logs)

		if len(logs) == 0 {
			err = components.LogComponent(dsId, formattedLogs, false).Render(r.Context(), w)
		} else {
			err = components.LogComponent(dsId, formattedLogs, true).Render(r.Context(), w)
		}

		if err != nil {
			fmt.Println("ApiLogHandler: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
