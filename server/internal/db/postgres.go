package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/marcelhfm/home_server/pkg/config"
	"github.com/marcelhfm/home_server/pkg/log"
)

func Init_pq() *sql.DB {
	host, err := config.GetenvStr("PQ_HOST")
	if err != nil {
		l.Log.Error().Msgf("Error while getting env: %v", err)
	}

	port, err := config.GetenvInt("PQ_PORT")
	if err != nil {
		l.Log.Error().Msgf("Error while getting env: %v", err)
	}

	user, err := config.GetenvStr("PQ_USER")
	if err != nil {
		l.Log.Error().Msgf("Error while getting env: %v", err)
	}

	password, err := config.GetenvStr("PQ_PASSWORD")
	if err != nil {
		l.Log.Error().Msgf("Error while getting env: %v", err)
	}

	dbname, err := config.GetenvStr("PQ_DBNAME")
	if err != nil {
		l.Log.Error().Msgf("Error while getting env: %v", err)
	}

	psqlError := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", psqlError)

	if err != nil {
		l.Log.Error().Msgf("Error while getting env: %v", err)
	}

	err = db.Ping()
	if err != nil {
		l.Log.Error().Msgf("Error while getting env: %v", err)
	}

	l.Log.Info().Msg("Successfully connected to db!")

	return db
}

func IngestIotData(db *sql.DB, datasourceId int, metric string, value int, ts string) {

	sqlStatement := `INSERT INTO timeseries (datasource_id, metric, value, timestamp) VALUES ($1, $2, $3, $4)`

	_, err := db.Exec(sqlStatement, datasourceId, metric, value, ts)

	if err != nil {
		l.Log.Error().Msgf("An error occured while trying to insert into database: %s", err)
	}
}
