package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/utmstack/UTMStack/shared/fs"
)

type UserEntry struct {
	Kind      string `json:"kind"`       // syslog|netflow|file|http|https
	UDP       string `json:"udp"`        // default UDP port (empty if not applicable)
	TCP       string `json:"tcp"`        // default TCP port (empty if not applicable)
	CreatedAt string `json:"created_at"` // RFC3339
}

type Catalog struct {
	UserDataTypes map[string]UserEntry `json:"user_datatypes"`
}

type DataTypeEntry struct {
	Name   string
	Kind   string // syslog|netflow|file|http|https
	UDP    string
	TCP    string
	Origin string
}

func IsBuiltin(name string) bool {
	dt := DataType(name)
	_, inProto := ProtoPorts[dt]
	_, inFile := FilePaths[dt]
	return inProto || inFile
}

func ResolveDataType(name string) (DataTypeEntry, bool) {
	dt := DataType(name)

	if pp, ok := ProtoPorts[dt]; ok {
		kind := "syslog"
		if dt == DataTypeNetflow {
			kind = "netflow"
		}
		return DataTypeEntry{
			Name:   name,
			Kind:   kind,
			UDP:    pp.UDP,
			TCP:    pp.TCP,
			Origin: "built-in",
		}, true
	}

	if _, ok := FilePaths[dt]; ok {
		return DataTypeEntry{
			Name:   name,
			Kind:   "file",
			Origin: "built-in",
		}, true
	}

	cat, err := readCatalog()
	if err != nil {
		return DataTypeEntry{}, false
	}

	if u, ok := cat.UserDataTypes[name]; ok {
		return DataTypeEntry{
			Name:   name,
			Kind:   u.Kind,
			UDP:    u.UDP,
			TCP:    u.TCP,
			Origin: "user",
		}, true
	}

	return DataTypeEntry{}, false
}

func AddUserDataType(name, kind, udp, tcp string) error {
	if IsBuiltin(name) {
		return fmt.Errorf("data type %q already exists as a built-in type", name)
	}

	lock, err := acquireLock()
	if err != nil {
		return err
	}
	defer releaseLock(lock)

	cat, err := readCatalog()
	if err != nil {
		return fmt.Errorf("error reading catalog: %w", err)
	}

	if cat.UserDataTypes == nil {
		cat.UserDataTypes = make(map[string]UserEntry)
	}

	cat.UserDataTypes[name] = UserEntry{
		Kind:      kind,
		UDP:       udp,
		TCP:       tcp,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	return writeCatalog(cat)
}

func RemoveUserDataType(name string) error {
	if IsBuiltin(name) {
		return fmt.Errorf("data type %q is a built-in type and cannot be removed", name)
	}

	collectorCfgPath := CollectorFileName
	if fs.Exists(collectorCfgPath) {
		raw, err := os.ReadFile(collectorCfgPath)
		if err == nil {
			var cfg struct {
				Integrations     map[string]interface{} `json:"integrations"`
				FileIntegrations map[string]interface{} `json:"file_integrations"`
			}
			if json.Unmarshal(raw, &cfg) == nil {
				if _, used := cfg.Integrations[name]; used {
					return fmt.Errorf("data type %q is still in use — disable the integration first", name)
				}
				if _, used := cfg.FileIntegrations[name]; used {
					return fmt.Errorf("data type %q is still in use — disable the integration first", name)
				}
			}
		}
	}

	lock, err := acquireLock()
	if err != nil {
		return err
	}
	defer releaseLock(lock)

	cat, err := readCatalog()
	if err != nil {
		return fmt.Errorf("error reading catalog: %w", err)
	}

	if _, exists := cat.UserDataTypes[name]; !exists {
		return fmt.Errorf("user data type %q not found", name)
	}

	delete(cat.UserDataTypes, name)
	return writeCatalog(cat)
}

func ListAllDataTypes() []DataTypeEntry {
	entries := make([]DataTypeEntry, 0)

	protoNames := make([]string, 0, len(ProtoPorts))
	for dt := range ProtoPorts {
		protoNames = append(protoNames, string(dt))
	}
	sort.Strings(protoNames)
	for _, name := range protoNames {
		pp := ProtoPorts[DataType(name)]
		kind := "syslog"
		if DataType(name) == DataTypeNetflow {
			kind = "netflow"
		}
		entries = append(entries, DataTypeEntry{
			Name: name, Kind: kind, UDP: pp.UDP, TCP: pp.TCP, Origin: "built-in",
		})
	}

	fileNames := make([]string, 0, len(FilePaths))
	for dt := range FilePaths {
		fileNames = append(fileNames, string(dt))
	}
	sort.Strings(fileNames)
	for _, name := range fileNames {
		entries = append(entries, DataTypeEntry{
			Name: name, Kind: "file", Origin: "built-in",
		})
	}

	cat, err := readCatalog()
	if err != nil {
		return entries
	}
	userNames := make([]string, 0, len(cat.UserDataTypes))
	for name := range cat.UserDataTypes {
		userNames = append(userNames, name)
	}
	sort.Strings(userNames)
	for _, name := range userNames {
		u := cat.UserDataTypes[name]
		entries = append(entries, DataTypeEntry{
			Name: name, Kind: u.Kind, UDP: u.UDP, TCP: u.TCP, Origin: "user",
		})
	}

	return entries
}

func readCatalog() (Catalog, error) {
	cat := Catalog{UserDataTypes: make(map[string]UserEntry)}
	catalogPath := CatalogFileName
	if !fs.Exists(catalogPath) {
		return cat, nil
	}
	data, err := os.ReadFile(catalogPath)
	if err != nil {
		return cat, fmt.Errorf("error reading catalog: %w", err)
	}
	if err := json.Unmarshal(data, &cat); err != nil {
		return cat, fmt.Errorf("error parsing catalog: %w", err)
	}
	if cat.UserDataTypes == nil {
		cat.UserDataTypes = make(map[string]UserEntry)
	}
	return cat, nil
}

func writeCatalog(cat Catalog) error {
	catalogPath := CatalogFileName
	tmpPath := catalogPath + ".tmp"
	data, err := json.MarshalIndent(cat, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling catalog: %w", err)
	}
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("error writing catalog: %w", err)
	}
	if err := os.Rename(tmpPath, catalogPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error saving catalog: %w", err)
	}
	return nil
}

func acquireLock() (*os.File, error) {
	lockPath := CatalogLockFile
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return nil, fmt.Errorf("catalog locked by another process: %w", err)
	}
	return f, nil
}

func releaseLock(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}
