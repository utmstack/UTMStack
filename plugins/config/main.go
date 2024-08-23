package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/logger"
)

type PluginConfig struct {
	RulesFolder   string        `yaml:"rules_folder"`
	GeoIPFolder   string        `yaml:"geoip_folder"`
	Elasticsearch string        `yaml:"elasticsearch"`
	PostgreSQL    PostgreConfig `yaml:"postgresql"`
	ServerName    string        `yaml:"server_name"`
	InternalKey   string        `yaml:"internal_key"`
	AgentManager  string        `yaml:"agent_manager"`
	Backend       string        `yaml:"backend"`
	Logstash      string        `yaml:"logstash"`
	CertsFolder   string        `yaml:"certs_folder"`
}

type PostgreConfig struct {
	Server   string `yaml:"server"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Filter struct {
	ID            int
	Name          string
	Filter        string
	GroupID       int
	SystemOwned   bool
	ModuleName    string
	IsActive      bool
	FilterVersion string
	DataTypeID    int
	UpdatedAt     string
}

func (f *Filter) FromVar(id int, name string, filter string, groupID int,
	systemOwned bool, moduleName string, isActive bool, filterVersion string,
	dataTypeID int, updatedAt string) {
	f.ID = id
	f.Name = name
	f.Filter = filter
	f.GroupID = groupID
	f.SystemOwned = systemOwned
	f.ModuleName = moduleName
	f.IsActive = isActive
	f.FilterVersion = filterVersion
	f.DataTypeID = dataTypeID
	f.UpdatedAt = updatedAt
}

func main() {
	pCfg, e := helpers.PluginCfg[PluginConfig]("com.utmstack")
	if e != nil {
		os.Exit(1)
	}

	gCfg := helpers.GetCfg()

	db, e := connect(pCfg.PostgreSQL.Password)
	if e != nil {
		os.Exit(1)
	}

	helpers.Logger().Info("connected to database")

	for {
		filters, e := getFilters(db)
		if e != nil {
			os.Exit(1)
		}

		e = cleanUpFiles(gCfg, filters)
		if e != nil {
			os.Exit(1)
		}

		e = writeFilters(gCfg, filters)
		if e != nil {
			os.Exit(1)
		}
	}
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

func getFilters(db *sql.DB) ([]Filter, *logger.Error) {
	rows, err := db.Query("SELECT * FROM filters")
	if err != nil {
		return nil, helpers.Logger().ErrorF("failed to get filters: %v", err)
	}

	defer rows.Close()

	filters := make([]Filter, 0, 10)

	for rows.Next() {
		var (
			id            int
			name          string
			body          string
			groupId       int
			systemOwned   bool
			moduleName    string
			isActive      bool
			filterVersion string
			dataTypeId    int
			updatedAt     string
		)

		err = rows.Scan(&id, &body, &name, &groupId, &systemOwned,
			&moduleName, &isActive, &filterVersion, &dataTypeId, &updatedAt)
		if err != nil {
			return nil, helpers.Logger().ErrorF("failed to scan row: %v", err)
		}

		filter := Filter{}
		filter.FromVar(id, name, body, groupId, systemOwned, moduleName, isActive, filterVersion, dataTypeId, updatedAt)
		filters = append(filters, filter)
	}

	return filters, nil
}

func listFiles(folder string) ([]string, *logger.Error) {
	var files []string

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		if logger.Is(err, "not found", "not exists") {
			return []string{}, nil
		}

		return nil, helpers.Logger().ErrorF("failed to list files: %v", err)
	}

	return files, nil
}

func cleanUpFiles(gCfg *helpers.Config, filters []Filter) *logger.Error {
	files, e := listFiles(filepath.Join(gCfg.Env.Workdir, "pipeline", "filters"))
	if e != nil {
		os.Exit(1)
	}

	for _, file := range files {
		var keep bool
		for _, filter := range filters {
			if file == filepath.Join(gCfg.Env.Workdir, "pipeline", "filters", fmt.Sprintf("%s.yaml", filter.Name)) {
				keep = true
				break
			}
		}

		if !keep {
			err := os.Remove(file)
			if err != nil {
				return helpers.Logger().ErrorF("failed to remove file: %v", err)
			}
		}
	}

	return nil
}

func writeFilters(pCfg *helpers.Config, filters []Filter) *logger.Error {
	for _, filter := range filters {
		file, err := os.Create(filepath.Join(pCfg.Env.Workdir, "pipeline", "filters", fmt.Sprintf("%s.yaml", filter.Name)))
		if err != nil {
			return helpers.Logger().ErrorF("failed to create file: %v", err)
		}

		_, err = file.WriteString(filter.Filter)
		if err != nil {
			return helpers.Logger().ErrorF("failed to write to file: %v", err)
		}

		err = file.Close()
		if err != nil {
			return helpers.Logger().ErrorF("failed to close file: %v", err)
		}
	}

	return nil
}
