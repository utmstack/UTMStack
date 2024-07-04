package sqldb

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/utmstack/UTMStack/plugins/events/utils"
)

var db *sql.DB
var err error

func init() {
	cnf, e := helpers.PluginCfg[utils.Config]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

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
