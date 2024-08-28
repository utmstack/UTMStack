package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/logger"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
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
	Id     int
	Name   string
	Filter string
}

type Tenant helpers.Tenant

type Asset helpers.Asset
type Rule struct {
	Id          int64          `yaml:"id"`
	DataTypes   []string       `yaml:"data_types"`
	Name        string         `yaml:"name"`
	Impact      plugins.Impact `yaml:"impact"`
	Category    string         `yaml:"category"`
	Technique   string         `yaml:"technique"`
	References  []string       `yaml:"references"`
	Description string         `yaml:"description"`
	Where       helpers.Where  `yaml:"where"`
}

func (t *Tenant) FromVar(disabledRules []int64, assets []Asset) {
	t.Id = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	t.Name = "Default"
	t.DisabledRules = disabledRules
	t.Assets = make([]helpers.Asset, len(assets))

	for i, asset := range assets {
		t.Assets[i] = helpers.Asset(asset)
	}
}

func (a *Asset) FromVar(name interface{}, hostnames interface{}, ips interface{}, confidentiality, integrity, availability interface{}) {
	a.Name = helpers.CastString(name)

	hostnamesStr := helpers.CastString(hostnames)
	hostnamesStr = strings.ReplaceAll(hostnamesStr, "[", "")
	hostnamesStr = strings.ReplaceAll(hostnamesStr, "]", "")
	hostnamesStr = strings.ReplaceAll(hostnamesStr, ",", "")
	hostnamesStr = strings.ReplaceAll(hostnamesStr, "\"", "")

	for _, hostname := range strings.Fields(hostnamesStr) {
		a.Hostnames = append(a.Hostnames, helpers.CastString(hostname))
	}

	ipsStr := helpers.CastString(ips)
	ipsStr = strings.ReplaceAll(ipsStr, "[", "")
	ipsStr = strings.ReplaceAll(ipsStr, "]", "")
	ipsStr = strings.ReplaceAll(ipsStr, ",", "")
	ipsStr = strings.ReplaceAll(ipsStr, "\"", "")

	for _, ip := range strings.Fields(ipsStr) {
		a.IPs = append(a.IPs, helpers.CastString(ip))
	}

	a.Confidentiality = int32(helpers.CastInt64(confidentiality))
	a.Integrity = int32(helpers.CastInt64(integrity))
	a.Availability = int32(helpers.CastInt64(availability))
}

func (r *Rule) FromVar(id int64, ruleName interface{}, confidentiality interface{}, integrity interface{},
	availability interface{}, category interface{}, technique interface{}, description interface{},
	references interface{}, where interface{}, dataTypes interface{}) {

	referencesStr := helpers.CastString(references)
	referencesStr = strings.ReplaceAll(referencesStr, "[", "")
	referencesStr = strings.ReplaceAll(referencesStr, "]", "")
	referencesStr = strings.ReplaceAll(referencesStr, ",", "")
	referencesStr = strings.ReplaceAll(referencesStr, "\"", "")
	referencesList := strings.Fields(referencesStr)

	dataTypesStr := helpers.CastString(dataTypes)
	dataTypesStr = strings.ReplaceAll(dataTypesStr, "[", "")
	dataTypesStr = strings.ReplaceAll(dataTypesStr, "]", "")
	dataTypesStr = strings.ReplaceAll(dataTypesStr, ",", "")
	dataTypesStr = strings.ReplaceAll(dataTypesStr, "\"", "")
	dataTypesList := strings.Fields(dataTypesStr)

	r.Id = id
	r.DataTypes = make([]string, len(dataTypesList))
	r.Name = helpers.CastString(ruleName)
	r.Impact.Confidentiality = int32(helpers.CastInt64(confidentiality))
	r.Impact.Integrity = int32(helpers.CastInt64(integrity))
	r.Impact.Availability = int32(helpers.CastInt64(availability))
	r.Category = helpers.CastString(category)
	r.Technique = helpers.CastString(technique)
	r.References = make([]string, len(referencesList))
	r.Description = helpers.CastString(description)

	for i, dataType := range dataTypesList {
		r.DataTypes[i] = helpers.CastString(dataType)
	}

	for i, reference := range referencesList {
		r.References[i] = helpers.CastString(reference)
	}

	w := helpers.CastString(where)

	whereObj := helpers.Where{
		Expression: gjson.Get(w, "ruleExpression").String(),
	}

	for _, variable := range gjson.Get(w, "ruleVariables").Array() {
		whereObj.Variables = append(whereObj.Variables, helpers.Variable{
			Get:    gjson.Get(variable.String(), "get").String(),
			As:     gjson.Get(variable.String(), "as").String(),
			OfType: gjson.Get(variable.String(), "of_type").String(),
		})
	}

	r.Where = whereObj
}

func (f *Filter) FromVar(id int, name interface{}, filter interface{}) {
	f.Id = id
	f.Name = helpers.CastString(name)
	f.Filter = helpers.CastString(filter)
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

	e = createFolderStructure(gCfg)
	if e != nil {
		os.Exit(1)
	}

	helpers.Logger().Info("created folder structure")

	for {
		filters, e := getFilters(db)
		if e != nil {
			os.Exit(1)
		}

		assets, e := getAssets(db)
		if e != nil {
			os.Exit(1)
		}

		rules, e := getRules(db)
		if e != nil {
			os.Exit(1)
		}

		tenant := Tenant{}
		tenant.FromVar([]int64{}, assets)

		e = cleanUpFilters(gCfg, filters)
		if e != nil {
			os.Exit(1)
		}

		e = writeFilters(gCfg, filters)
		if e != nil {
			os.Exit(1)
		}

		e = cleanUpRules(gCfg, rules)
		if e != nil {
			os.Exit(1)
		}

		e = writeRules(gCfg, rules)
		if e != nil {
			os.Exit(1)
		}

		e = writeTenant(gCfg, tenant)
		if e != nil {
			os.Exit(1)
		}

		time.Sleep(5 * time.Minute)
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
	rows, err := db.Query("SELECT id, filter_name, logstash_filter FROM utm_logstash_filter WHERE is_active = true")
	if err != nil {
		return nil, helpers.Logger().ErrorF("failed to get filters: %v", err)
	}

	defer rows.Close()

	filters := make([]Filter, 0, 10)

	for rows.Next() {
		var (
			id   int
			name interface{}
			body interface{}
		)

		err = rows.Scan(&id, &name, &body)
		if err != nil {
			return nil, helpers.Logger().ErrorF("failed to scan row: %v", err)
		}

		filter := Filter{}
		filter.FromVar(
			id,
			name,
			body,
		)
		filters = append(filters, filter)
	}

	return filters, nil
}

func getAssets(db *sql.DB) ([]Asset, *logger.Error) {
	rows, err := db.Query("SELECT * FROM utm_tenant_config")
	if err != nil {
		return nil, helpers.Logger().ErrorF("failed to get assets: %v", err)
	}

	defer rows.Close()

	assets := make([]Asset, 0, 10)

	for rows.Next() {
		var (
			id              int
			name            interface{}
			hostnames       interface{}
			ips             interface{}
			confidentiality interface{}
			integrity       interface{}
			availability    interface{}
			lastUpdate      interface{}
		)

		err = rows.Scan(&id, &name, &hostnames, &ips, &confidentiality,
			&integrity, &availability, &lastUpdate)
		if err != nil {
			return nil, helpers.Logger().ErrorF("failed to scan row: %v", err)
		}

		asset := Asset{}

		asset.FromVar(name, hostnames, ips, confidentiality, integrity, availability)

		assets = append(assets, asset)
	}

	return assets, nil
}

func getRules(db *sql.DB) ([]Rule, *logger.Error) {
	rows, err := db.Query("SELECT id,rule_name,rule_confidentiality,rule_integrity,rule_availability,rule_category,rule_technique,rule_description,rule_references_def,rule_definition_def FROM utm_correlation_rules WHERE rule_active = true")
	if err != nil {
		return nil, helpers.Logger().ErrorF("failed to get rules: %v", err)
	}

	defer rows.Close()

	rules := make([]Rule, 0, 10)

	for rows.Next() {
		var (
			id              int64
			ruleName        interface{}
			confidentiality interface{}
			integrity       interface{}
			availability    interface{}
			category        interface{}
			technique       interface{}
			description     interface{}
			references      interface{}
			where           interface{}
		)

		err = rows.Scan(&id, &ruleName, &confidentiality, &integrity, &availability,
			&category, &technique, &description, &references, &where)
		if err != nil {
			return nil, helpers.Logger().ErrorF("failed to scan row: %v", err)
		}

		rule := Rule{}

		dataTypes, e := getRuleDataTypes(db)
		if e != nil {
			return nil, e
		}

		rule.FromVar(id, ruleName, confidentiality, integrity, availability, category, technique, description, references, where, dataTypes)

		rules = append(rules, rule)
	}

	return rules, nil
}

func getRuleDataTypes(db *sql.DB) ([]interface{}, *logger.Error) {
	rows, err := db.Query("SELECT data_type_id FROM utm_group_rules_data_type")
	if err != nil {
		return nil, helpers.Logger().ErrorF("failed to get data types: %v", err)
	}

	defer rows.Close()

	dataTypes := make([]interface{}, 0, 10)

	for rows.Next() {
		var (
			dataTypeId int64
			dataType   interface{}
		)

		err = rows.Scan(&dataTypeId)
		if err != nil {
			return nil, helpers.Logger().ErrorF("failed to scan row: %v", err)
		}

		row := db.QueryRow("SELECT data_type FROM utm_data_types WHERE id = $1", dataTypeId)

		err := row.Scan(&dataType)
		if err != nil {
			return nil, helpers.Logger().ErrorF("failed to scan row: %v", err)
		}

		dataTypes = append(dataTypes, dataType)
	}

	return dataTypes, nil
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
		return nil, helpers.Logger().ErrorF("failed to list files: %v", err)
	}

	return files, nil
}

func cleanUpFilters(gCfg *helpers.Config, filters []Filter) *logger.Error {
	files, e := listFiles(filepath.Join(gCfg.Env.Workdir, "pipeline", "filters"))
	if e != nil {
		os.Exit(1)
	}

	for _, file := range files {
		var keep bool
		for _, filter := range filters {
			if file == filepath.Join(gCfg.Env.Workdir, "pipeline", "filters", fmt.Sprintf("%s.yaml", filter.Id)) {
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

func cleanUpRules(gCfg *helpers.Config, rules []Rule) *logger.Error {
	files, e := listFiles(filepath.Join(gCfg.Env.Workdir, "rules"))
	if e != nil {
		os.Exit(1)
	}

	for _, file := range files {
		var keep bool
		for _, rule := range rules {
			if file == filepath.Join(gCfg.Env.Workdir, "rules", fmt.Sprintf("%s.yaml", rule.Id)) {
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
		file, err := os.Create(filepath.Join(pCfg.Env.Workdir, "pipeline", "filters", fmt.Sprintf("%d.yaml", filter.Id)))
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

func writeTenant(pCfg *helpers.Config, tenant Tenant) *logger.Error {
	file, err := os.Create(filepath.Join(pCfg.Env.Workdir, "pipeline", "tenant.yaml"))
	if err != nil {
		return helpers.Logger().ErrorF("failed to create file: %v", err)
	}

	tenants := helpers.Config{
		Tenants: []helpers.Tenant{helpers.Tenant(tenant)},
	}

	bTenants, err := yaml.Marshal(tenants)
	if err != nil {
		return helpers.Logger().ErrorF("failed to marshal tenant: %v", err)
	}

	_, err = file.Write(bTenants)
	if err != nil {
		return helpers.Logger().ErrorF("failed to write to file: %v", err)
	}

	err = file.Close()
	if err != nil {
		return helpers.Logger().ErrorF("failed to close file: %v", err)
	}

	return nil
}

func writeRules(pCfg *helpers.Config, rules []Rule) *logger.Error {
	for _, rule := range rules {
		file, err := os.Create(filepath.Join(pCfg.Env.Workdir, "rules", fmt.Sprintf("%d.yaml", rule.Id)))
		if err != nil {
			return helpers.Logger().ErrorF("failed to create file: %v", err)
		}

		bRule, err := yaml.Marshal(rule)
		if err != nil {
			return helpers.Logger().ErrorF("failed to marshal rule: %v", err)
		}

		_, err = file.Write(bRule)
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

func createFolderStructure(gCfg *helpers.Config) *logger.Error {
	folders := []string{
		filepath.Join(gCfg.Env.Workdir, "rules"),
		filepath.Join(gCfg.Env.Workdir, "pipeline"),
		filepath.Join(gCfg.Env.Workdir, "pipeline", "filters"),
	}

	for _, folder := range folders {
		if _, err := os.Stat(folder); err == nil {
			continue
		}

		err := os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return helpers.Logger().ErrorF("failed to create folder: %v", err)
		}
	}

	return nil
}
