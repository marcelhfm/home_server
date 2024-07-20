package cron

import (
	"database/sql"

	"github.com/marcelhfm/home_server/pkg/log"
)

func CleanupLogs(db *sql.DB) {
	l.Log.Info().Msg("CRON: Deleting old logs...")
	deleteLogsQuery := "DELETE FROM logs WHERE timestamp < NOW() - INTERVAL '7 days'"

	result, err := db.Exec(deleteLogsQuery)
	if err != nil {
		l.Log.Error().Msgf("CRON: Error deleting logs from db: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		l.Log.Error().Msgf("CRON: Error getting rows affected: %v", err)
	}

	l.Log.Info().Msgf("CRON: Deleted %d rows from logs table", rowsAffected)
}
