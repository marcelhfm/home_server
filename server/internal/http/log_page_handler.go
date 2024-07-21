package http

import (
	"net/http"

	l "github.com/marcelhfm/home_server/pkg/log"
	"github.com/marcelhfm/home_server/views"
)

func LogPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dsId := r.PathValue("id")
		dsName := r.URL.Query().Get("name")
		err := views.LogPage(dsId, dsName).Render(r.Context(), w)

		if err != nil {
			l.Log.Error().Msgf("Error in DataPaneHandler: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
