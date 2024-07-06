package http

import (
	"database/sql"
	"net/http"

	"github.com/marcelhfm/home_server/views"
)

type TimeseriesResponse struct {
	Id            string
	Datasource_id int
	Metric        string
	Value         int
	Timestamp     string
}

func DatasourcePageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dsId := r.PathValue("id")
		dsName := r.URL.Query().Get("name")

		views.Datasource(dsId, dsName).Render(r.Context(), w)
	}
}
