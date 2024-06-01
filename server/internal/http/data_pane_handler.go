package http

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/marcelhfm/home_server/views/components"
)

func generateChart(data []TimeseriesData) (string, int, int, string, error) {
	lineChart := charts.NewLine()

	var co2 []opts.LineData
	var humidity []opts.LineData
	var temp []opts.LineData
	var timestamps []string
	var lastCo2 int
	var display_status int
	seen := make(map[string]bool)

	for i := len(data) - 1; i >= 0; i-- {
		el := data[i]
		if !seen[el.Timestamp] {
			seen[el.Timestamp] = true

			loc, err := time.LoadLocation("Europe/Berlin")
			if err != nil {
				return "", 0, 0, "", err
			}

			t, err := time.Parse(time.RFC3339, el.Timestamp)
			if err != nil {
				return "", 0, 0, "", err
			}

			localTime := t.In(loc)
			formattedTime := localTime.Format("15:04")

			timestamps = append(timestamps, formattedTime)
		}

		switch el.Metric {
		case "co2":
			co2 = append(co2, opts.LineData{Value: el.Value})
			break
		case "humidity":
			humidity = append(humidity, opts.LineData{Value: el.Value})
			break
		case "temperature":
			temp = append(temp, opts.LineData{Value: el.Value})
			break
		default:
		}
	}

	// Find last display_status and co2
	co2found := false
	displayFound := false
	for _, el := range data {
		if el.Metric == "co2" && !co2found {
			lastCo2 = el.Value
			co2found = true
		}

		if el.Metric == "display_status" && !displayFound {
			display_status = el.Value
			displayFound = true
		}

		if displayFound && co2found {
			break
		}
	}

	lineChart.SetXAxis(timestamps).
		AddSeries("Co2", co2).
		AddSeries("Temperature", temp).
		AddSeries("Humidity", humidity).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true, ShowSymbol: true}), charts.WithSeriesAnimation(false))

	var buf bytes.Buffer
	err := lineChart.Render(&buf)

	if err != nil {
		return "", 0, 0, "", err
	}

	return buf.String(), lastCo2, display_status, timestamps[len(timestamps)-1], nil
}

type TimeseriesData struct {
	Metric    string
	Value     int
	Timestamp string
}

func getTimeseriesData(db *sql.DB, dsId string) ([]TimeseriesData, error) {
	fmt.Println("DataPaneHandler: Fetching timeseries data for ds", dsId)

	timeseriesQuery := fmt.Sprintf("SELECT metric, value, timestamp FROM timeseries WHERE datasource_id = %s AND timestamp >=NOW() - INTERVAL '30 minutes' ORDER BY timestamp desc", dsId)

	rows, err := db.Query(timeseriesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []TimeseriesData

	for rows.Next() {
		var metric string
		var value int
		var ts string

		if err := rows.Scan(&metric, &value, &ts); err != nil {
			return nil, err
		}

		res = append(res, TimeseriesData{Metric: metric, Value: value, Timestamp: ts})
	}

	return res, nil
}

func DataPaneHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dsId := r.PathValue("id")
		data, err := getTimeseriesData(db, dsId)
		if err != nil {
			fmt.Println("DataPaneHandler: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		chart, lastCo2, display_status, last_seen, err := generateChart(data)
		if err != nil {
			fmt.Println("DataPaneHandler: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = components.DatasourceDataPane(chart, lastCo2, display_status, last_seen).Render(r.Context(), w)

		if err != nil {
			fmt.Println("DataPaneHandler: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
