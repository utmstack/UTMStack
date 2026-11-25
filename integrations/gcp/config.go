package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"os"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/integrations/shared/modulecfg"
	"golang.org/x/crypto/pbkdf2"
	"gopkg.in/yaml.v3"
)

const (
	processName = "plugin_com.utmstack.gcp"
)

type ConfigurationSection struct {
	ModuleActive bool
	ModuleGroups []*ModuleGroup
}

type ModuleGroup struct {
	Id                        int32
	GroupName                 string
	TenantId                  string
	ModuleGroupConfigurations []*Configuration
}

// Key identifies a group across tenants and across reloads. GroupName is unique
// only inside its own tenant, and Id is the position in the file — it shifts as
// soon as any tenant adds or removes a group, which would restart collectors
// that did not change and hand one tenant's pagination cursor to another.
func (g *ModuleGroup) Key() string { return g.TenantId + "/" + g.GroupName }

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
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		_ = catcher.Error("ENCRYPTION_KEY is not set; the module configuration cannot be read", nil, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	err := modulecfg.Watch(context.Background(), os.Getenv("NATS_URL"), "gcp", NewCipher(encKey),
		func(data []byte) {
			sec := parseConfig(data, encKey)
			if sec == nil {
				return
			}

			mu.Lock()
			cnf = sec
			mu.Unlock()
			push(sec)
		},
		func(err error) {
			_ = catcher.Error("module configuration could not be read", err, map[string]any{"process": processName})
		})
	if err != nil {
		_ = catcher.Error("module configuration watch failed", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

// pollFallback is used if fsnotify setup fails — polls every 30s instead.

func push(sec *ConfigurationSection) {
	select {
	case configUpdateChan <- sec:
	default:
	}
}

// The backend writes one section per platform tenant, each holding that
// tenant's configured connector instances. A group name is unique inside its
// tenant only — two customers may each call theirs "production" — so nothing
// downstream may identify a group by name alone.
type configGroupYAML struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Config      map[string]string `yaml:"config"`
}

type tenantYAML struct {
	ID     string            `yaml:"id"`
	Groups []configGroupYAML `yaml:"groups"`
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

func parseConfig(data []byte, encKey string) *ConfigurationSection {
	if data == nil {
		// No stored configuration means the module is disabled, not an error.
		return &ConfigurationSection{ModuleActive: false}
	}

	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		_ = catcher.Error("failed to parse config file", err, map[string]any{"process": processName})
		return nil
	}
	tenants := pf.Plugins[pluginKey].Tenants

	sec := &ConfigurationSection{}
	var id int32
	for _, tc := range tenants {
		for _, g := range tc.Groups {
			id++
			grp := &ModuleGroup{
				Id:        id,
				GroupName: g.Name,
				TenantId:  tc.ID,
			}
			for k, v := range g.Config {
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
	}
	// A tenant section with no groups must not read as configured.
	sec.ModuleActive = len(sec.ModuleGroups) > 0
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
