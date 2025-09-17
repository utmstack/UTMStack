package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
	k8syaml "sigs.k8s.io/yaml"
)

type Filter struct {
	Id     int
	Name   string
	Filter string
}

type Tenant plugins.Tenant

type Asset plugins.Asset

type Rule struct {
	Id            int64           `yaml:"id"`
	DataTypes     []string        `yaml:"dataTypes"`
	Name          string          `yaml:"name"`
	Impact        *plugins.Impact `yaml:"impact"`
	Category      string          `yaml:"category"`
	Technique     string          `yaml:"technique"`
	Adversary     string          `yaml:"adversary"`
	References    []string        `yaml:"references"`
	Description   string          `yaml:"description"`
	Where         string          `yaml:"where"`
	AfterEvents   []SearchRequest `yaml:"afterEvents,omitempty"`
	DeduplicateBy []string        `yaml:"deduplicateBy,omitempty"`
}

type SearchRequest struct {
	IndexPattern string          `yaml:"indexPattern"`
	With         []Expression    `yaml:"with"`
	Or           []SearchRequest `yaml:"or,omitempty"`
	Within       string          `yaml:"within"`
	Count        int64           `yaml:"count"`
}

type SearchRequestBackend struct {
	IndexPattern string                 `yaml:"indexPattern"`
	With         []ExpressionBackend    `yaml:"with"`
	Or           []SearchRequestBackend `yaml:"or,omitempty"`
	Within       string                 `yaml:"within"`
	Count        int64                  `yaml:"count"`
}

type Expression struct {
	Field    string      `yaml:"field"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
}

type ExpressionBackend struct {
	Field    string      `yaml:"field"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
}

func (b *ExpressionBackend) ToExpression() Expression {
	return Expression{
		Field:    b.Field,
		Operator: b.Operator,
		Value:    b.Value,
	}
}

func (b *SearchRequestBackend) ToSearchRequest() SearchRequest {
	// Convert With field: convert each ExpressionBackend to Expression
	with := make([]Expression, 0, len(b.With))
	for _, expr := range b.With {
		with = append(with, expr.ToExpression())
	}

	// Convert Or field: recursively convert each SearchRequestBackend to SearchRequest
	or := make([]SearchRequest, 0, len(b.Or))
	for _, req := range b.Or {
		or = append(or, req.ToSearchRequest())
	}

	return SearchRequest{
		IndexPattern: b.IndexPattern,
		With:         with,
		Or:           or,
		Within:       b.Within,
		Count:        b.Count,
	}
}

func (t *Tenant) FromVar(disabledRules []uint64, assets []Asset) {
	t.Id = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	t.Name = "Default"
	t.DisabledRules = disabledRules
	t.Assets = make([]*plugins.Asset, 0, len(assets))

	for _, asset := range assets {
		sdkAsset := plugins.Asset(asset)
		t.Assets = append(t.Assets, &sdkAsset)
	}
}

func (a *Asset) FromVar(name any, hostnames any, ipAddresses any, confidentiality, integrity, availability any) {
	var hostnamesList []string

	if hostnames != nil {
		hostnamesStr := utils.CastString(hostnames)
		err := json.Unmarshal([]byte(hostnamesStr), &hostnamesList)
		if err != nil {
			_ = catcher.Error("failed to unmarshal hostnames list", err, map[string]any{})
			return
		}
	}

	var ipAddressesList []string

	if ipAddresses != nil {
		ipAddressesStr := utils.CastString(ipAddresses)
		err := json.Unmarshal([]byte(ipAddressesStr), &ipAddressesList)
		if err != nil {
			_ = catcher.Error("failed to unmarshal ip addresses list", err, map[string]any{})
			return
		}
	}

	a.Name = utils.CastString(name)
	a.Confidentiality = castUint32(confidentiality)
	a.Integrity = castUint32(integrity)
	a.Availability = castUint32(availability)
	a.Hostnames = hostnamesList
	a.Ips = ipAddressesList
}

func castUint32(value interface{}) uint32 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int64:
		return uint32(v)
	case float64:
		return uint32(v)
	case string:
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			_ = catcher.Error("failed to cast string to int64", err, map[string]any{"value": v})
			return 0
		}
		return uint32(val)
	default:
		return 0
	}
}

func (r *Rule) FromVar(id int64, dataTypes []string, ruleName any, confidentiality any, integrity any,
	availability any, category any, technique any, description any,
	references any, where any, adversary any, deduplicate any, after any) {

	var referencesList []string

	if references != nil {
		referencesStr := utils.CastString(references)
		err := json.Unmarshal([]byte(referencesStr), &referencesList)
		if err != nil {
			_ = catcher.Error("failed to unmarshal references list", err, map[string]any{})
			return
		}
	}

	var deduplicateList []string

	if deduplicate != nil {
		deduplicateStr := utils.CastString(deduplicate)
		err := json.Unmarshal([]byte(deduplicateStr), &deduplicateList)
		if err != nil {
			_ = catcher.Error("failed to unmarshal deduplicate list", err, map[string]any{})
			return
		}
	}

	var afterObj []SearchRequest

	if after != nil {
		var afterBackendObj []SearchRequestBackend
		afterStr := utils.CastString(after)
		err := json.Unmarshal([]byte(afterStr), &afterBackendObj)
		if err != nil {
			_ = catcher.Error("failed to unmarshal after list", err, map[string]any{})
			return
		}

		// Convert each SearchRequestBackend to SearchRequest
		for _, req := range afterBackendObj {
			afterObj = append(afterObj, req.ToSearchRequest())
		}
	}

	r.Impact = new(plugins.Impact)
	r.Id = id
	r.DataTypes = dataTypes
	r.Name = utils.CastString(ruleName)
	r.Impact.Confidentiality = castUint32(confidentiality)
	r.Impact.Integrity = castUint32(integrity)
	r.Impact.Availability = castUint32(availability)
	r.Category = utils.CastString(category)
	r.Technique = utils.CastString(technique)
	r.References = make([]string, len(referencesList))
	r.Description = utils.CastString(description)
	r.Adversary = utils.CastString(adversary)
	r.DeduplicateBy = deduplicateList
	r.AfterEvents = afterObj
	r.References = referencesList
	r.Where = utils.CastString(where)
}

func (f *Filter) FromVar(id int, name any, filter any) {
	f.Id = id
	f.Name = utils.CastString(name)
	f.Filter = utils.CastString(filter)
}

func main() {
	for {
		func() {
			db, err := connect()
			if err != nil {
				_ = catcher.Error("failed to connect to database", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}
			defer func() { _ = db.Close() }()

			filters, err := getFilters(db)
			if err != nil {
				_ = catcher.Error("failed to get filters", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			assets, err := getAssets(db)
			if err != nil {
				_ = catcher.Error("failed to get assets", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			rules, err := getRules(db)
			if err != nil {
				_ = catcher.Error("failed to get rules", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			patterns, err := getPatterns(db)
			if err != nil {
				_ = catcher.Error("failed to get patterns", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			tenant := Tenant{}
			tenant.FromVar([]uint64{}, assets)

			// Try to acquire the lock before modifying configuration files
			maxRetries := 5

			for i := 0; i < maxRetries; i++ {
				acquired, err := plugins.AcquireLock()

				if acquired {
					break
				}

				// Lock not acquired, wait and retry
				if i < maxRetries-1 {
					_ = catcher.Error("failed to acquire lock", err, map[string]interface{}{"retry": i + 1, "maxRetries": maxRetries})
					time.Sleep(plugins.RandomDuration(10, 60))
				} else {
					_ = catcher.Error("failed to acquire lock after multiple retries", nil, nil)
					return
				}
			}

			// Make sure to release the lock when done
			defer func() {
				if err := plugins.ReleaseLock(); err != nil {
					_ = catcher.Error("failed to release lock", err, nil)
				}
			}()

			err = cleanUpFilters(filters)
			if err != nil {
				_ = catcher.Error("failed to clean up filters", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			err = writeFilters(filters)
			if err != nil {
				_ = catcher.Error("failed to write filters", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			err = cleanUpRules(rules)
			if err != nil {
				_ = catcher.Error("failed to clean up rules", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			err = writeRules(rules)
			if err != nil {
				_ = catcher.Error("failed to write rules", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			err = writeTenant(tenant)
			if err != nil {
				_ = catcher.Error("failed to write tenant", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}

			err = writePatterns(patterns)
			if err != nil {
				_ = catcher.Error("failed to write patterns", err, map[string]any{})
				// Don't exit, just sleep and retry
				time.Sleep(30 * time.Second)
				return
			}
		}()

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
	rows, err := db.Query("SELECT id,asset_name,asset_hostname_list_def,asset_ip_list_def,asset_confidentiality,asset_integrity,asset_availability,last_update FROM utm_tenant_config")
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
	rows, err := db.Query("SELECT id,rule_name,rule_confidentiality,rule_integrity,rule_availability,rule_category,rule_technique,rule_description,rule_references_def,rule_definition_def,rule_adversary,rule_deduplicate_by_def,rule_after_events_def FROM utm_correlation_rules WHERE rule_active = true")
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
			adversary       any
			deduplicate     any
			after           any
		)

		err = rows.Scan(&id, &ruleName, &confidentiality, &integrity, &availability,
			&category, &technique, &description, &references, &where, &adversary, &deduplicate, &after)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		rule := Rule{}

		dataTypes, err := getRuleDataTypes(db, id)
		if err != nil {
			return nil, err
		}

		rule.FromVar(id, dataTypes, ruleName, confidentiality, integrity, availability, category, technique, description, references, where, adversary, deduplicate, after)

		rules = append(rules, rule)
	}

	return rules, nil
}

func getRuleDataTypes(db *sql.DB, ruleId int64) ([]string, error) {
	rows, err := db.Query("SELECT data_type_id FROM utm_group_rules_data_type WHERE rule_id = $1", ruleId)
	if err != nil {
		return nil, fmt.Errorf("failed to get data types: %v", err)
	}

	defer func() { _ = rows.Close() }()

	dataTypes := make([]string, 0, 10)

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

		dataTypes = append(dataTypes, utils.CastString(dataType))
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

func cleanUpFilters(filters []Filter) error {
	filtersPath, err := utils.MkdirJoin(plugins.WorkDir, "pipeline", "filters")
	if err != nil {
		return catcher.Error("cannot create filters directory", err, nil)
	}

	files, e := listFiles(filtersPath.String())
	if e != nil {
		return catcher.Error("failed to list files", e, nil)
	}

	for _, file := range files {
		var keep bool
		for _, filter := range filters {
			filterFile := filtersPath.FileJoin(fmt.Sprintf("%d.yaml", filter.Id))
			if file == filterFile {
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

func cleanUpRules(rules []Rule) error {
	rulesFolder, err := utils.MkdirJoin(plugins.WorkDir, "rules", "utmstack")
	if err != nil {
		return catcher.Error("cannot create rules directory", err, nil)
	}

	files, err := listFiles(rulesFolder.String())
	if err != nil {
		return catcher.Error("failed to list files", err, map[string]any{})
	}

	for _, file := range files {
		var keep bool
		for _, rule := range rules {
			ruleFile := rulesFolder.FileJoin(fmt.Sprintf("%d.yaml", rule.Id))
			if file == ruleFile {
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

func writeFilters(filters []Filter) error {
	for _, filter := range filters {
		filtersFolder, err := utils.MkdirJoin(plugins.WorkDir, "pipeline", "filters")
		if err != nil {
			return catcher.Error("cannot create filters directory", err, nil)
		}

		file, err := os.Create(filtersFolder.FileJoin(fmt.Sprintf("%d.yaml", filter.Id)))
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

func writeTenant(tenant Tenant) error {
	pipelineFolder, err := utils.MkdirJoin(plugins.WorkDir, "pipeline")
	if err != nil {
		return catcher.Error("cannot create pipeline directory", err, nil)
	}

	file, err := os.Create(pipelineFolder.FileJoin("tenants.yaml"))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	sdkTenant := plugins.Tenant(tenant)

	tenants := plugins.Config{
		Tenants: []*plugins.Tenant{&sdkTenant},
	}

	bTenants, err := k8syaml.Marshal(tenants)
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

func writeRules(rules []Rule) error {
	for _, rule := range rules {
		filePath, err := utils.MkdirJoin(plugins.WorkDir, "rules", "utmstack")
		if err != nil {
			return catcher.Error("cannot create rules directory", err, nil)
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

func writePatterns(patterns map[string]string) error {
	filePath, err := utils.MkdirJoin(plugins.WorkDir, "pipeline")
	if err != nil {
		return catcher.Error("cannot create pipeline directory", err, nil)
	}
	file, err := os.Create(filePath.FileJoin("patterns.yaml"))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	config := plugins.Config{
		Patterns: patterns,
	}

	bPatterns, err := k8syaml.Marshal(config)
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
