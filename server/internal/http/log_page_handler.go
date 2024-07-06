package http

import (
	"fmt"
	"net/http"

	"github.com/marcelhfm/home_server/views"
)

func LogPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dsId := r.PathValue("id")
		dsName := r.URL.Query().Get("name")
		err := views.LogPage(dsId, dsName).Render(r.Context(), w)

		if err != nil {
			fmt.Println("DataPaneHandler: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
