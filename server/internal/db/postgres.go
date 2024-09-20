package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/marcelhfm/home_server/pkg/config"
	"github.com/marcelhfm/home_server/pkg/log"
)

func Init_pq() *sql.DB {
	host := config.GetenvStr("PQ_HOST")
	port := config.GetenvInt("PQ_PORT")
	user := config.GetenvStr("PQ_USER")
	password := config.GetenvStr("PQ_PASSWORD")
	dbname := config.GetenvStr("PQ_DBNAME")

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
