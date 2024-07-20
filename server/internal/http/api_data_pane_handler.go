package http

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"time"
	_ "time/tzdata"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"

	"github.com/marcelhfm/home_server/pkg/types"
	"github.com/marcelhfm/home_server/views/components"
)

type GenerateChartReturn struct {
	Chart         string
	LastCo2       int
	DisplayStatus int
	LastMoisture  int
	LastSeen      string
	Error         error
}

func generateChart(data []TimeseriesData, dsType string) GenerateChartReturn {
	lineChart := charts.NewLine()

	var co2 []opts.LineData
	var humidity []opts.LineData
	var moisture []opts.LineData
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
				return GenerateChartReturn{Error: err}
			}

			t, err := time.Parse(time.RFC3339, el.Timestamp)
			if err != nil {
				return GenerateChartReturn{Error: err}
			}

			localTime := t.In(loc)
			formattedTime := localTime.Format("15:04")

			timestamps = append(timestamps, formattedTime)
		}

		if dsType == "IRRIGATION" {
			switch el.Metric {
			case "moisture":
				moisture = append(moisture, opts.LineData{Value: el.Value / 100})
				break
			default:
				break
			}
		} else {
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
	}

	// Find last display_status and co2
	co2found := false
	displayFound := false
	moistureFound := false
	var lastMoisture int
	for _, el := range data {
		if el.Metric == "co2" && !co2found {
			lastCo2 = el.Value
			co2found = true
		}

		if el.Metric == "display_status" && !displayFound {
			display_status = el.Value
			displayFound = true
		}

		if el.Metric == "moisture" && !moistureFound {
			lastMoisture = el.Value
			moistureFound = true
		}

		if displayFound && co2found && moistureFound {
			break
		}
	}

	if dsType == "IRRIGATION" {
		lineChart.SetXAxis(timestamps).
			AddSeries("Moisture", moisture).
			SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true, ShowSymbol: true}), charts.WithSeriesAnimation(false))
	} else {
		lineChart.SetXAxis(timestamps).
			AddSeries("Co2", co2).
			AddSeries("Temperature", temp).
			AddSeries("Humidity", humidity).
			SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true, ShowSymbol: true}), charts.WithSeriesAnimation(false))
	}

	var buf bytes.Buffer
	err := lineChart.Render(&buf)

	if err != nil {
		return GenerateChartReturn{Error: err}
	}

	return GenerateChartReturn{
		Chart:         buf.String(),
		LastCo2:       lastCo2,
		LastMoisture:  lastMoisture,
		DisplayStatus: display_status,
		LastSeen:      timestamps[len(timestamps)-1],
		Error:         nil,
	}
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

func getDsStatus(db *sql.DB, dsId string) (string, string, error) {
	statusQuery := fmt.Sprintf("SELECT status, type FROM datasources WHERE id = %s LIMIT 1", dsId)

	var status string
	var dsType string
	err := db.QueryRow(statusQuery).Scan(&status, &dsType)

	if err != nil {
		return "", "", err
	}

	return status, dsType, nil
}

func ApiDataPaneHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dsId := r.PathValue("id")
		dsName := r.URL.Query().Get("name")

		status, dsType, err := getDsStatus(db, dsId)
		if err != nil {
			fmt.Println("DataPaneHandler: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		data, err := getTimeseriesData(db, dsId)
		if err != nil {
			fmt.Println("DataPaneHandler: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if len(data) == 0 {
			fmt.Printf("DataPaneHandler: datasource %s has no timeseries data\n", dsId)
			components.DatasourceDataPane(types.DsDataPaneProps{DsId: dsId, DsName: dsName}).Render(r.Context(), w)
			return
		}

		chartStruct := generateChart(data, dsType)
		if err != nil {
			fmt.Println("DataPaneHandler: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = components.DatasourceDataPane(types.DsDataPaneProps{DsId: dsId,
			DsName:        dsName,
			Chart:         chartStruct.Chart,
			LastSeen:      chartStruct.LastSeen,
			Moisture:      chartStruct.LastMoisture,
			DisplayStatus: chartStruct.DisplayStatus,
			Status:        status,
			Data:          true,
			DsType:        dsType,
			Co2:           chartStruct.LastCo2}).Render(r.Context(), w)

		if err != nil {
			fmt.Println("DataPaneHandler: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
