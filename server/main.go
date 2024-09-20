package main

import (
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
	cron.StartCrons(db)

	// servers
	go tcp.StartTCPServer(db, commandChannel, commandResponseChannel)
	go udp.StartLogServer(db)
	go mqtt.StartMqttListener(db)
	http.StartHttpServer(db, commandChannel, commandResponseChannel)
}
