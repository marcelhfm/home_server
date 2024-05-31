package http

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/marcelhfm/home_server/views/components"
)

func generateChart(data []TimeseriesData) (string, error) {
	lineChart := charts.NewLine()

	var co2 []opts.LineData
	var humidity []opts.LineData
	var temp []opts.LineData
	var timestamps []string
	seen := make(map[string]bool)

	for _, el := range data {
		if !seen[el.Timestamp] {
			seen[el.Timestamp] = true
			timestamps = append(timestamps, el.Timestamp)
		}

		switch el.Metric {
		case "co2":
			co2 = append(co2, opts.LineData{Value: el.Value})
			break
		case "humidty":
			humidity = append(humidity, opts.LineData{Value: el.Value})
			break
		case "temperature":
			temp = append(temp, opts.LineData{Value: el.Value})
			break
		default:
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
		return "", err
	}

	return buf.String(), nil
}

type TimeseriesData struct {
	Metric    string
	Value     int
	Timestamp string
}

func getTimeseriesData(db *sql.DB, dsId string) ([]TimeseriesData, error) {
	fmt.Println("DataPaneHandler: Fetching timeseries data for ds", dsId)

	timeseriesQuery := fmt.Sprintf("SELECT metric, value, timestamp FROM timeseries WHERE datasource_id = %s ORDER BY timestamp desc LIMIT 1000", dsId)

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

		chart, err := generateChart(data)
		if err != nil {
			fmt.Println("DataPaneHandler: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = components.DatasourceDataPane(chart).Render(r.Context(), w)

		if err != nil {
			fmt.Println("DataPaneHandler: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
