package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"golang.org/x/crypto/pbkdf2"
	"gopkg.in/yaml.v3"
)

const (
	pluginFile         = "system_plugins_gcp.yaml"
	processName        = "plugin_com.utmstack.gcp"
	pipelineDirDefault = "/workdir/pipeline"
)

type ConfigurationSection struct {
	ModuleActive bool
	ModuleGroups []*ModuleGroup
}

type ModuleGroup struct {
	Id                        int32
	GroupName                 string
	ModuleGroupConfigurations []*Configuration
}

type Configuration struct {
	ConfKey   string
	ConfValue string
}

var (
	cnf              *ConfigurationSection
	mu               sync.Mutex
	configUpdateChan chan *ConfigurationSection
)

func init() {
	configUpdateChan = make(chan *ConfigurationSection, 1)
}

// GetConfig returns the current configuration (nil-safe).
func GetConfig() *ConfigurationSection {
	mu.Lock()
	defer mu.Unlock()
	if cnf == nil {
		return &ConfigurationSection{}
	}
	return cnf
}

// GetConfigUpdateChannel returns the channel that receives config updates.
func GetConfigUpdateChannel() <-chan *ConfigurationSection {
	return configUpdateChan
}

// StartConfigurationSystem starts the file watcher. Blocks until configured,
// then runs indefinitely.
func StartConfigurationSystem() {
	pipelineDir := pipelineDirDefault
	var encKey string
	for {
		cfg := plugins.PluginCfg("com.utmstack")
		if cfg.Exists() {
			if d := cfg.Get("pipelineDir").String(); d != "" {
				pipelineDir = d
			}
			encKey = cfg.Get("encryptionKey").String()
			break
		}
		_ = catcher.Error("plugin configuration not found", nil, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
	}

	filePath := filepath.Join(pipelineDir, pluginFile)

	// Initial load.
	if sec := readConfig(filePath, encKey); sec != nil {
		mu.Lock()
		cnf = sec
		mu.Unlock()
		push(sec)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_ = catcher.Error("failed to create file watcher", err, map[string]any{"process": processName})
		pollFallback(filePath, encKey)
		return
	}
	defer watcher.Close()

	// Watch the directory so we catch atomic write (rename) events.
	if err := watcher.Add(pipelineDir); err != nil {
		_ = catcher.Error("failed to watch pipeline dir", err, map[string]any{"process": processName})
		pollFallback(filePath, encKey)
		return
	}

	catcher.Info(fmt.Sprintf("watching %s for config changes", filePath), map[string]any{"process": processName})

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Name != filePath {
				continue
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) ||
				event.Has(fsnotify.Rename) || event.Has(fsnotify.Remove) {
				if sec := readConfig(filePath, encKey); sec != nil {
					mu.Lock()
					cnf = sec
					mu.Unlock()
					push(sec)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			_ = catcher.Error("watcher error", err, map[string]any{"process": processName})
		}
	}
}

// pollFallback is used if fsnotify setup fails — polls every 30s instead.
func pollFallback(filePath, encKey string) {
	catcher.Warn("falling back to 30s polling", map[string]any{"process": processName})
	for range time.Tick(30 * time.Second) {
		if sec := readConfig(filePath, encKey); sec != nil {
			mu.Lock()
			cnf = sec
			mu.Unlock()
			push(sec)
		}
	}
}

func push(sec *ConfigurationSection) {
	select {
	case configUpdateChan <- sec:
	default:
	}
}

// tenantYAML matches the flat YAML format the backend writes.
type tenantYAML struct {
	Name   string            `yaml:"name"`
	Config map[string]string `yaml:",inline"`
}

type pluginsFile struct {
	Plugins map[string]struct {
		Tenants []tenantYAML `yaml:"tenants"`
	} `yaml:"plugins"`
}

const pluginKey = "gcp"

// sensitiveKeys lists the configuration keys whose values are stored encrypted.
var sensitiveKeys = map[string]bool{
	"jsonKey": true,
}

func readConfig(path, encKey string) *ConfigurationSection {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File removed → module disabled / no configuration. Report an empty,
			// inactive section so the module is treated as disabled and all work stops.
			return &ConfigurationSection{ModuleActive: false}
		}
		_ = catcher.Error("failed to read config file", err, map[string]any{"process": processName, "file": path})
		return nil
	}

	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		_ = catcher.Error("failed to parse config file", err, map[string]any{"process": processName, "file": path})
		return nil
	}
	tenants := pf.Plugins[pluginKey].Tenants

	sec := &ConfigurationSection{
		ModuleActive: len(tenants) > 0,
	}
	for i, t := range tenants {
		grp := &ModuleGroup{
			Id:        int32(i + 1),
			GroupName: t.Name,
		}
		for k, v := range t.Config {
			grp.ModuleGroupConfigurations = append(grp.ModuleGroupConfigurations,
				&Configuration{ConfKey: k, ConfValue: v})
		}
		if encKey != "" {
			for _, conf := range grp.ModuleGroupConfigurations {
				if sensitiveKeys[conf.ConfKey] {
					dec, err := NewCipher(encKey).Decrypt(conf.ConfValue)
					if err == nil {
						conf.ConfValue = dec
					}
				}
			}
		}
		sec.ModuleGroups = append(sec.ModuleGroups, grp)
	}
	return sec
}

const (
	iterationCount = 65536
	keyLength      = 16
)

type Cipher struct {
	key []byte
}

func NewCipher(key string) *Cipher {
	return &Cipher{key: []byte(key)}
}

func (c *Cipher) setKey() (cipher.Block, []byte, error) {
	h := sha1.New()
	h.Write(c.key)
	salt := h.Sum(nil)
	keyEnc := pbkdf2.Key(c.key, salt, iterationCount, keyLength, sha1.New)
	block, err := aes.NewCipher(keyEnc)
	if err != nil {
		return nil, nil, err
	}
	return block, salt[:keyLength], nil
}

func (c *Cipher) Decrypt(crypt string) (string, error) {
	if crypt == "" {
		return "", nil
	}
	encryptedData, err := base64.StdEncoding.DecodeString(crypt)
	if err != nil {
		return crypt, nil // not base64 → already plaintext
	}
	blk, iv, err := c.setKey()
	if err != nil {
		return crypt, err
	}
	if len(encryptedData)%aes.BlockSize != 0 {
		return crypt, nil // not a valid CBC block → already plaintext
	}
	dec := cipher.NewCBCDecrypter(blk, iv)
	decrypted := make([]byte, len(encryptedData))
	dec.CryptBlocks(decrypted, encryptedData)
	return string(pkcs5Trim(decrypted)), nil
}

func pkcs5Trim(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return data
	}
	return data[:len(data)-padding]
}
