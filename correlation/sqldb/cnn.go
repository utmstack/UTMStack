package sqldb

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/utmstack/UTMStack/correlation/utils"
)

var db *sql.DB
var err error

func Connect() {
	cnf := utils.GetConfig()
	log.Printf("Connecting to Postgres server: %s using port: %v", cnf.Postgres.Server, cnf.Postgres.Port)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable",
		cnf.Postgres.Server,
		cnf.Postgres.User,
		cnf.Postgres.Password,
		cnf.Postgres.Database,
		cnf.Postgres.Port,
	)
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not connect to Postgres: %v", err)
	}

	ping()
}

func ping() {
	if err := db.Ping(); err != nil {
		log.Fatalf("Could not reconnect to Postgres: %v", err)
	}
}
