package http

import (
	"net/http"

	"github.com/marcelhfm/home_server/views"
)

func DatasourceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		views.Datasource(r.PathValue("id")).Render(r.Context(), w)
	}
}
