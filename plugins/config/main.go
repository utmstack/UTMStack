package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	_ "github.com/lib/pq"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

type Filter struct {
	Id     int
	Name   string
	Filter string
}

type Tenant plugins.Tenant

type Asset plugins.Asset

type Rule struct {
	Id          int64          `yaml:"id"`
	DataTypes   []string       `yaml:"dataTypes"`
	Name        string         `yaml:"name"`
	Impact      plugins.Impact `yaml:"impact"`
	Category    string         `yaml:"category"`
	Technique   string         `yaml:"technique"`
	References  []string       `yaml:"references"`
	Description string         `yaml:"description"`
	Where       plugins.Where  `yaml:"where"`
}

func (t *Tenant) FromVar(disabledRules []int64, assets []Asset) {
	t.Id = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	t.Name = "Default"
	t.DisabledRules = disabledRules
	t.Assets = make([]*plugins.Asset, 0, len(assets))

	for _, asset := range assets {
		sdkAsset := plugins.Asset(asset)
		t.Assets = append(t.Assets, &sdkAsset)
	}
}

func (a *Asset) FromVar(name any, hostnames any, ips any, confidentiality, integrity, availability any) {
	a.Name = utils.CastString(name)

	hostnamesStr := utils.CastString(hostnames)
	hostnamesStr = strings.ReplaceAll(hostnamesStr, "[", "")
	hostnamesStr = strings.ReplaceAll(hostnamesStr, "]", "")
	hostnamesStr = strings.ReplaceAll(hostnamesStr, ",", "")
	hostnamesStr = strings.ReplaceAll(hostnamesStr, "\"", "")

	for _, hostname := range strings.Fields(hostnamesStr) {
		a.Hostnames = append(a.Hostnames, utils.CastString(hostname))
	}

	ipsStr := utils.CastString(ips)
	ipsStr = strings.ReplaceAll(ipsStr, "[", "")
	ipsStr = strings.ReplaceAll(ipsStr, "]", "")
	ipsStr = strings.ReplaceAll(ipsStr, ",", "")
	ipsStr = strings.ReplaceAll(ipsStr, "\"", "")

	for _, ip := range strings.Fields(ipsStr) {
		a.Ips = append(a.Ips, utils.CastString(ip))
	}

	a.Confidentiality = int32(utils.CastInt64(confidentiality))
	a.Integrity = int32(utils.CastInt64(integrity))
	a.Availability = int32(utils.CastInt64(availability))
}

func (r *Rule) FromVar(id int64, ruleName any, confidentiality any, integrity any,
	availability any, category any, technique any, description any,
	references any, where any, dataTypes any) {

	referencesStr := utils.CastString(references)
	referencesStr = strings.ReplaceAll(referencesStr, "[", "")
	referencesStr = strings.ReplaceAll(referencesStr, "]", "")
	referencesStr = strings.ReplaceAll(referencesStr, ",", "")
	referencesStr = strings.ReplaceAll(referencesStr, "\"", "")
	referencesList := strings.Fields(referencesStr)

	dataTypesStr := utils.CastString(dataTypes)
	dataTypesStr = strings.ReplaceAll(dataTypesStr, "[", "")
	dataTypesStr = strings.ReplaceAll(dataTypesStr, "]", "")
	dataTypesStr = strings.ReplaceAll(dataTypesStr, ",", "")
	dataTypesStr = strings.ReplaceAll(dataTypesStr, "\"", "")
	dataTypesList := strings.Fields(dataTypesStr)

	r.Id = id
	r.DataTypes = make([]string, len(dataTypesList))
	r.Name = utils.CastString(ruleName)
	r.Impact.Confidentiality = int32(utils.CastInt64(confidentiality))
	r.Impact.Integrity = int32(utils.CastInt64(integrity))
	r.Impact.Availability = int32(utils.CastInt64(availability))
	r.Category = utils.CastString(category)
	r.Technique = utils.CastString(technique)
	r.References = make([]string, len(referencesList))
	r.Description = utils.CastString(description)

	for i, dataType := range dataTypesList {
		r.DataTypes[i] = utils.CastString(dataType)
	}

	for i, reference := range referencesList {
		r.References[i] = utils.CastString(reference)
	}

	w := utils.CastString(where)

	whereObj := plugins.Where{
		Expression: gjson.Get(w, "ruleExpression").String(),
	}

	for _, variable := range gjson.Get(w, "ruleVariables").Array() {
		whereObj.Variables = append(whereObj.Variables, &plugins.Variable{
			Get:    gjson.Get(variable.String(), "get").String(),
			As:     gjson.Get(variable.String(), "as").String(),
			OfType: gjson.Get(variable.String(), "ofType").String(),
		})
	}

	r.Where = whereObj
}

func (f *Filter) FromVar(id int, name any, filter any) {
	f.Id = id
	f.Name = utils.CastString(name)
	f.Filter = utils.CastString(filter)
}

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	for {
		gCfg := plugins.GetCfg()
		err := createFolderStructure(gCfg)
		if err != nil {
			_ = catcher.Error("failed to create folder structure", err, map[string]any{})
			os.Exit(1)
		}

		db, err := connect()
		if err != nil {
			_ = catcher.Error("failed to connect to database", err, map[string]any{})
			os.Exit(1)
		}

		filters, err := getFilters(db)
		if err != nil {
			_ = catcher.Error("failed to get filters", err, map[string]any{})
			os.Exit(1)
		}

		assets, err := getAssets(db)
		if err != nil {
			_ = catcher.Error("failed to get assets", err, map[string]any{})
			os.Exit(1)
		}

		rules, err := getRules(db)
		if err != nil {
			_ = catcher.Error("failed to get rules", err, map[string]any{})
			os.Exit(1)
		}

		patterns, err := getPatterns(db)
		if err != nil {
			_ = catcher.Error("failed to get patterns", err, map[string]any{})
			os.Exit(1)
		}

		_ = db.Close()

		tenant := Tenant{}
		tenant.FromVar([]int64{}, assets)

		err = cleanUpFilters(gCfg, filters)
		if err != nil {
			_ = catcher.Error("failed to clean up filters", err, map[string]any{})
			os.Exit(1)
		}

		err = writeFilters(gCfg, filters)
		if err != nil {
			_ = catcher.Error("failed to write filters", err, map[string]any{})
			os.Exit(1)
		}

		err = cleanUpRules(gCfg, rules)
		if err != nil {
			_ = catcher.Error("failed to clean up rules", err, map[string]any{})
			os.Exit(1)
		}

		err = writeRules(gCfg, rules)
		if err != nil {
			_ = catcher.Error("failed to write rules", err, map[string]any{})
			os.Exit(1)
		}

		err = writeTenant(gCfg, tenant)
		if err != nil {
			_ = catcher.Error("failed to write tenant", err, map[string]any{})
			os.Exit(1)
		}

		err = writePatterns(gCfg, patterns)
		if err != nil {
			_ = catcher.Error("failed to write patterns", err, map[string]any{})
			os.Exit(1)
		}

		time.Sleep(5 * time.Minute)
	}
}

// connect to postgres database
func connect() (*sql.DB, error) {
	pCfg := plugins.PluginCfg("com.utmstack", false)
	password := pCfg.Get("postgresql.password").String()
	server := pCfg.Get("postgresql.server").String()
	port := pCfg.Get("postgresql.port").Int()
	database := pCfg.Get("postgresql.database").String()
	user := pCfg.Get("postgresql.user").String()

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable", user, password,
		database, server, port)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, catcher.Error("failed to open database connection", err, map[string]any{"connStr": connStr})
	}

	err = db.Ping()
	if err != nil {
		return nil, catcher.Error("failed to ping database", err, map[string]any{})
	}

	return db, nil
}

func getPatterns(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT pattern_id, pattern_definition FROM utm_regex_pattern")
	if err != nil {
		return nil, fmt.Errorf("failed to get patterns: %v", err)
	}

	defer func() { _ = rows.Close() }()

	patterns := make(map[string]string, 10)

	for rows.Next() {
		var name string
		var pattern string

		err = rows.Scan(&name, &pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		patterns[name] = pattern
	}

	return patterns, nil
}

func getFilters(db *sql.DB) ([]Filter, error) {
	rows, err := db.Query("SELECT id, filter_name, logstash_filter FROM utm_logstash_filter WHERE is_active = true")
	if err != nil {
		return nil, fmt.Errorf("failed to get filters: %v", err)
	}

	defer func() { _ = rows.Close() }()

	filters := make([]Filter, 0, 10)

	for rows.Next() {
		var (
			id   int
			name any
			body any
		)

		err = rows.Scan(&id, &name, &body)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
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

func getAssets(db *sql.DB) ([]Asset, error) {
	rows, err := db.Query("id,asset_name,asset_hostname_list_def,asset_ip_list_def,asset_confidentiality,asset_integrity,asset_availability,last_update FROM utm_tenant_config")
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %v", err)
	}

	defer func() { _ = rows.Close() }()

	assets := make([]Asset, 0, 10)

	for rows.Next() {
		var (
			id              int
			name            any
			hostnames       any
			ips             any
			confidentiality any
			integrity       any
			availability    any
			lastUpdate      any
		)

		err = rows.Scan(&id, &name, &hostnames, &ips, &confidentiality,
			&integrity, &availability, &lastUpdate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		asset := Asset{}

		asset.FromVar(name, hostnames, ips, confidentiality, integrity, availability)

		assets = append(assets, asset)
	}

	return assets, nil
}

func getRules(db *sql.DB) ([]Rule, error) {
	rows, err := db.Query("SELECT id,rule_name,rule_confidentiality,rule_integrity,rule_availability,rule_category,rule_technique,rule_description,rule_references_def,rule_definition_def FROM utm_correlation_rules WHERE rule_active = true")
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %v", err)
	}

	defer func() { _ = rows.Close() }()

	rules := make([]Rule, 0, 10)

	for rows.Next() {
		var (
			id              int64
			ruleName        any
			confidentiality any
			integrity       any
			availability    any
			category        any
			technique       any
			description     any
			references      any
			where           any
		)

		err = rows.Scan(&id, &ruleName, &confidentiality, &integrity, &availability,
			&category, &technique, &description, &references, &where)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		rule := Rule{}

		dataTypes, err := getRuleDataTypes(db, id)
		if err != nil {
			return nil, err
		}

		rule.FromVar(id, ruleName, confidentiality, integrity, availability, category, technique, description, references, where, dataTypes)

		rules = append(rules, rule)
	}

	return rules, nil
}

func getRuleDataTypes(db *sql.DB, ruleId int64) ([]any, error) {
	rows, err := db.Query("SELECT data_type_id FROM utm_group_rules_data_type WHERE rule_id = $1", ruleId)
	if err != nil {
		return nil, fmt.Errorf("failed to get data types: %v", err)
	}

	defer func() { _ = rows.Close() }()

	dataTypes := make([]any, 0, 10)

	for rows.Next() {
		var (
			dataTypeId int64
			dataType   any
		)

		err = rows.Scan(&dataTypeId)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		row := db.QueryRow("SELECT data_type FROM utm_data_types WHERE id = $1", dataTypeId)

		err := row.Scan(&dataType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		dataTypes = append(dataTypes, dataType)
	}

	return dataTypes, nil
}

func listFiles(folder string) ([]string, error) {
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
		return nil, fmt.Errorf("failed to list files: %v", err)
	}

	return files, nil
}

func cleanUpFilters(gCfg *plugins.Config, filters []Filter) error {
	filePath, err := utils.MkdirJoin(plugins.WorkDir, "pipeline", "filters")
	if err != nil {
		_ = catcher.Error("cannot create filters directory", err, nil)
		os.Exit(1)
	}

	files, e := listFiles(string(filePath))
	if e != nil {
		os.Exit(1)
	}

	for _, file := range files {
		var keep bool
		for _, filter := range filters {
			filePath := filePath.FileJoin(fmt.Sprintf("%d.yaml", filter.Id))
			if file == filePath {
				keep = true
				break
			}
		}

		if !keep {
			err := os.Remove(file)
			if err != nil {
				return fmt.Errorf("failed to remove file: %v", err)
			}
		}
	}

	return nil
}

func cleanUpRules(gCfg *plugins.Config, rules []Rule) error {
	filePath, err := utils.MkdirJoin(plugins.WorkDir, "rules", "utmstack")
	if err != nil {
		_ = catcher.Error("cannot create filters directory", err, nil)
		os.Exit(1)
	}

	files, err := listFiles(string(filePath))
	if err != nil {
		_ = catcher.Error("failed to list files", err, map[string]any{})
		os.Exit(1)
	}

	for _, file := range files {
		var keep bool
		for _, rule := range rules {
			filePath := filePath.FileJoin(fmt.Sprintf("%d.yaml", rule.Id))
			if file == filePath {
				keep = true
				break
			}
		}

		if !keep {
			err := os.Remove(file)
			if err != nil {
				return fmt.Errorf("failed to remove file: %v", err)
			}
		}
	}

	return nil
}

func writeFilters(pCfg *plugins.Config, filters []Filter) error {
	for _, filter := range filters {
		filePath, err := utils.MkdirJoin(plugins.WorkDir, "pipeline", "filters")
		if err != nil {
			_ = catcher.Error("cannot create filters directory", err, nil)
			os.Exit(1)
		}

		file, err := os.Create(filePath.FileJoin(fmt.Sprintf("%d.yaml", filter.Id)))
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}

		_, err = file.WriteString(filter.Filter)
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("failed to close file: %v", err)
		}
	}

	return nil
}

func writeTenant(pCfg *plugins.Config, tenant Tenant) error {
	filePath, err := utils.MkdirJoin(plugins.WorkDir, "pipeline")
	if err != nil {
		_ = catcher.Error("cannot create pipeline directory", err, nil)
		os.Exit(1)
	}

	file, err := os.Create(filePath.FileJoin("tenants.yaml"))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	sdkTenant := plugins.Tenant(tenant)

	tenants := plugins.Config{
		Tenants: []*plugins.Tenant{&sdkTenant},
	}

	bTenants, err := yaml.Marshal(tenants)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant: %v", err)
	}

	_, err = file.Write(bTenants)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file: %v", err)
	}

	return nil
}

func writeRules(pCfg *plugins.Config, rules []Rule) error {
	for _, rule := range rules {
		filePath, err := utils.MkdirJoin(plugins.WorkDir, "rules", "utmstack")
		if err != nil {
			_ = catcher.Error("cannot create rules directory", err, nil)
			os.Exit(1)
		}

		file, err := os.Create(filePath.FileJoin(fmt.Sprintf("%d.yaml", rule.Id)))
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}

		bRule, err := yaml.Marshal([]Rule{rule})
		if err != nil {
			return fmt.Errorf("failed to marshal rule: %v", err)
		}

		_, err = file.Write(bRule)
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}

		err = file.Close()
		if err != nil {
			return fmt.Errorf("failed to close file: %v", err)
		}
	}

	return nil
}

func writePatterns(pCfg *plugins.Config, patterns map[string]string) error {
	filePath, err := utils.MkdirJoin(plugins.WorkDir, "pipeline")
	if err != nil {
		_ = catcher.Error("cannot create pipeline directory", err, nil)
		os.Exit(1)
	}
	file, err := os.Create(filePath.FileJoin("patterns.yaml"))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	config := plugins.Config{
		Patterns: patterns,
	}

	bPatterns, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal patterns: %v", err)
	}

	_, err = file.Write(bPatterns)
	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file: %v", err)
	}

	return nil
}

func createFolderStructure(gCfg *plugins.Config) error {
	folders := []string{
		filepath.Join(plugins.WorkDir, "rules"),
		filepath.Join(plugins.WorkDir, "rules", "utmstack"),
		filepath.Join(plugins.WorkDir, "pipeline"),
		filepath.Join(plugins.WorkDir, "pipeline", "filters"),
	}

	for _, folder := range folders {
		if _, err := os.Stat(folder); err == nil {
			continue
		}

		err := os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return catcher.Error("failed to create folder structure", err, map[string]any{
				"folder": folder,
			})
		}
	}

	return nil
}
