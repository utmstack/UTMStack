package sqldb

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/correlation/utils"
)

var db *sql.DB
var err error

func Connect() {
	cnf := utils.GetConfig()
	catcher.Info("Connecting to Postgres server", map[string]any{"server": cnf.Postgres.Server, "port": cnf.Postgres.Port})

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable",
		cnf.Postgres.Server,
		cnf.Postgres.User,
		cnf.Postgres.Password,
		cnf.Postgres.Database,
		cnf.Postgres.Port,
	)
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		catcher.Error("Could not connect to Postgres", err, nil)
		os.Exit(1)
	}

	ping()
}

func ping() {
	if err := db.Ping(); err != nil {
		catcher.Error("Could not reconnect to Postgres", err, nil)
		os.Exit(1)
	}
}
