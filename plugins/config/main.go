package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/logger"
)

type PluginConfig struct {
	ServerName   string `yaml:"server_name"`
	InternalKey  string `yaml:"internal_key"`
	AgentManager string `yaml:"agent_manager"`
	Backend      string `yaml:"backend"`
	Logstash     string `yaml:"logstash"`
	CertsFolder  string `yaml:"certs_folder"`
}

func main() {
	pCfg, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	db, e := connect(pCfg.InternalKey)
	if e != nil {
		os.Exit(1)
	}

	helpers.Logger().Info("connected to database")


}

// connect to postgres database
func connect(password string) (*sql.DB, *logger.Error) {
	// Replace the connection details with your own
	connStr := fmt.Sprintf("user=postgres password=%s dbname=utmstack host=postgres port=5432 sslmode=disable", password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, helpers.Logger().ErrorF("failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, helpers.Logger().ErrorF("failed to ping database: %v", err)
	}

	return db, nil
}

func getFilters(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM filters")
	if err != nil {
		helpers.Logger().ErrorF("failed to get filters: %v", err)
		return
	}
	
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			helpers.Logger().ErrorF("failed to scan row: %v", err)
			return
		}

		fmt.Printf("id: %d, name: %s\n", id, name)
	}
}