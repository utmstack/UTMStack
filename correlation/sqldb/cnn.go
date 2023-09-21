package sqldb

import (
	"database/sql"
	"fmt"

	"github.com/utmstack/UTMStack/correlation/utils"
	_ "github.com/lib/pq"
	"github.com/quantfall/holmes"
)

var db *sql.DB
var err error
var h = holmes.New(utils.GetConfig().ErrorLevel, "SQLDB")

func Connect() {
	cnf := utils.GetConfig()
	h.Info("Connecting to PostgreSQL server: %s using port: %v", cnf.PostgreSQL.Server, cnf.PostgreSQL.Port)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable",
		cnf.PostgreSQL.Server,
		cnf.PostgreSQL.User,
		cnf.PostgreSQL.Password,
		cnf.PostgreSQL.Database,
		cnf.PostgreSQL.Port,
	)
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		h.FatalError("Could not connect to PostgreSQL: %v", err)
	}

	ping()
}

func ping() {
	if err := db.Ping(); err != nil {
		h.FatalError("Could not reconnect to PostgreSQL: %v", err)
	}
}
