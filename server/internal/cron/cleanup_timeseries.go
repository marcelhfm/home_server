package cron

import (
	"database/sql"

	"github.com/marcelhfm/home_server/pkg/log"
)

func CleanupTimeseries(db *sql.DB) {
	l.Log.Info().Msg("CRON: Deleting old timeseries data...")
	deleteLogsQuery := "DELETE FROM timeseries WHERE timestamp < NOW() - INTERVAL '14 days'"

	result, err := db.Exec(deleteLogsQuery)
	if err != nil {
		l.Log.Error().Msgf("CRON: Error deleting timeseries from db: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.Error().Msgf("CRON: Error getting rows affected: %v", err)
	}

	l.Log.Info().Msgf("CRON: Deleted %d rows from timeseries table", rowsAffected)
}
