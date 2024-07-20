package main

import (
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/marcelhfm/home_server/internal/cron"
	"github.com/marcelhfm/home_server/internal/db"
	"github.com/marcelhfm/home_server/internal/http"
	"github.com/marcelhfm/home_server/internal/mqtt"
	"github.com/marcelhfm/home_server/internal/tcp"
	"github.com/marcelhfm/home_server/internal/udp"
	"github.com/marcelhfm/home_server/pkg/log"
	"github.com/marcelhfm/home_server/pkg/types"
)

func main() {
	l.Log.Info().Msg("Hello from home server :)")

	err := godotenv.Load()
	if err != nil {
		l.Log.Error().Msgf("Error loading .env file %s", err)
	}

	db := db.Init_pq()
	defer db.Close()

	commandChannel := make(chan types.CommandRequest)
	commandResponseChannel := make(chan types.CommandResponse)

	// crons
	s, err := gocron.NewScheduler()
	if err != nil {
		l.Log.Error().Msgf("Error starting cron scheduler: %v", err)
		return
	}

	j, err := s.NewJob(
		gocron.DurationJob(1*time.Hour),
		gocron.NewTask(
			cron.CleanupLogs,
			db,
		),
	)
	if err != nil {
		l.Log.Error().Msgf("Error creating cron job (cleanup logs): %v", err)
		return
	}

	k, err := s.NewJob(
		gocron.DurationJob(1*time.Hour),
		gocron.NewTask(
			cron.CleanupTimeseries,
			db,
		),
	)
	if err != nil {
		l.Log.Error().Msgf("MAIN: Error creating cron job (cleanup timeseries): %v", err)
		return
	}

	l.Log.Info().Msgf("MAIN: Created CRON jobs: CleanupLogs (%s), CleanupTimeseries (%s)", j.ID(), k.ID())
	s.Start()

	// protocols
	go tcp.StartTCPServer(db, commandChannel, commandResponseChannel)
	go udp.StartLogServer(db)
	go mqtt.StartMqttListener(db)
	http.StartHttpServer(db, commandChannel, commandResponseChannel)
}
