package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/marcelhfm/home_server/pkg/config"
)

func Init_pq() *sql.DB {
	host, err := config.GetenvStr("PQ_HOST")
	if err != nil {
		log.Fatal(err)
	}

	port, err := config.GetenvInt("PQ_PORT")
	if err != nil {
		log.Fatal(err)
	}

	user, err := config.GetenvStr("PQ_USER")
	if err != nil {
		log.Fatal(err)
	}

	password, err := config.GetenvStr("PQ_PASSWORD")
	if err != nil {
		log.Fatal(err)
	}

	dbname, err := config.GetenvStr("PQ_DBNAME")
	if err != nil {
		log.Fatal(err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("db: Successfully connected to db!")

	return db
}

func IngestIotData(db *sql.DB, datasourceId int, metric string, value int) {

	sqlStatement := `INSERT INTO timeseries (datasource_id, metric, value, timestamp) VALUES ($1, $2, $3, now())`

	_, err := db.Exec(sqlStatement, datasourceId, metric, value)

	if err != nil {
		fmt.Printf("db: An error occured while trying to insert into database: %s", err)
	}
}
