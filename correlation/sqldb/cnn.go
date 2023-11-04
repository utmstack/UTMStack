package sqldb

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/utmstack/UTMStack/correlation/utils"
	_ "github.com/lib/pq"
)

var db *sql.DB
var err error

func Connect() {
	cnf := utils.GetConfig()
	log.Printf("Connecting to PostgreSQL server: %s using port: %v", cnf.PostgreSQL.Server, cnf.PostgreSQL.Port)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable",
		cnf.PostgreSQL.Server,
		cnf.PostgreSQL.User,
		cnf.PostgreSQL.Password,
		cnf.PostgreSQL.Database,
		cnf.PostgreSQL.Port,
	)
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not connect to PostgreSQL: %v", err)
	}

	ping()
}

func ping() {
	if err := db.Ping(); err != nil {
		log.Fatalf("Could not reconnect to PostgreSQL: %v", err)
	}
}

