package cron

import (
	"database/sql"
	"time"

	"github.com/go-co-op/gocron/v2"
	l "github.com/marcelhfm/home_server/pkg/log"
)

func StartCrons(db *sql.DB) error {
	s, err := gocron.NewScheduler()

	j, err := s.NewJob(
		gocron.DurationJob(1*time.Hour),
		gocron.NewTask(
			CleanupLogs,
			db,
		),
	)
	if err != nil {
		l.Log.Error().Msgf("CRON: Error creating cron job (cleanup logs): %v", err)
		return err
	}

	k, err := s.NewJob(
		gocron.DurationJob(1*time.Hour),
		gocron.NewTask(
			CleanupTimeseries,
			db,
		),
	)
	if err != nil {
		l.Log.Error().Msgf("CRON: Error creating cron job (cleanup timeseries): %v", err)
		return err
	}

	m, err := s.NewJob(
		gocron.DurationJob(30*time.Second),
		gocron.NewTask(
			Notifications,
			db,
		),
	)

	l.Log.Info().Msgf("CRON: Created CRON jobs: CleanupLogs (%s), CleanupTimeseries (%s), Notifications (%s)", j.ID(), k.ID(), m.ID())
	s.Start()

	return nil
}
