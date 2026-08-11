package repository

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/utmstack/utmstack/backend/modules/storage/connectors"
	"github.com/utmstack/utmstack/backend/modules/storage/domain"
)

const configFileName = "utmstack-tiering.xml"

type xmlConfigRepository struct {
	dir string
	mu  sync.Mutex
}

func NewConfigRepository(dir string) connectors.ConfigRepository {
	return &xmlConfigRepository{dir: dir}
}

func (r *xmlConfigRepository) path() string { return filepath.Join(r.dir, configFileName) }

type clickhouseConfig struct {
	XMLName xml.Name             `xml:"clickhouse"`
	Storage storageConfiguration `xml:"storage_configuration"`
}

type storageConfiguration struct {
	Disks    disks    `xml:"disks"`
	Policies policies `xml:"policies"`
}

type disks struct {
	Cold coldDisk `xml:"cold_disk"`
}

type coldDisk struct {
	Type            string `xml:"type"`
	Endpoint        string `xml:"endpoint"`
	AccessKeyID     string `xml:"access_key_id"`
	SecretAccessKey string `xml:"secret_access_key"`
	MetadataPath    string `xml:"metadata_path"`
	CacheEnabled    bool   `xml:"cache_enabled"`
	CacheMaxSize    int64  `xml:"data_cache_max_size"`
}

type policies struct {
	HotCold hotCold `xml:"hot_cold"`
}

type hotCold struct {
	Volumes    volumes `xml:"volumes"`
	MoveFactor float64 `xml:"move_factor"`
}

type volumes struct {
	Default hotVolume  `xml:"default"`
	Cold    coldVolume `xml:"cold"`
}

type hotVolume struct {
	Disk string `xml:"disk"`
}

type coldVolume struct {
	Disk string `xml:"disk"`
}

func (r *xmlConfigRepository) Read() (domain.Tiering, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.path())
	if errors.Is(err, os.ErrNotExist) {
		return domain.Tiering{}, nil
	}
	if err != nil {
		return domain.Tiering{}, err
	}

	var cfg clickhouseConfig
	if err := xml.Unmarshal(data, &cfg); err != nil {
		return domain.Tiering{}, err
	}
	endpoint := cfg.Storage.Disks.Cold.Endpoint
	if endpoint == "" {
		return domain.Tiering{}, nil
	}
	return domain.Tiering{Configured: true, Endpoint: endpoint, Policy: domain.PolicyName}, nil
}

func (r *xmlConfigRepository) Write(o domain.ObjectStore) (func() error, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	previous, err := os.ReadFile(r.path())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	undo := func() error {
		r.mu.Lock()
		defer r.mu.Unlock()
		if previous == nil {
			return os.Remove(r.path())
		}
		return os.WriteFile(r.path(), previous, 0o644)
	}

	o = o.Normalized()
	cfg := clickhouseConfig{
		Storage: storageConfiguration{
			Disks: disks{Cold: coldDisk{
				Type:            "s3",
				Endpoint:        o.Endpoint,
				AccessKeyID:     o.AccessKey,
				SecretAccessKey: o.SecretKey,
				MetadataPath:    "/var/lib/clickhouse/disks/" + domain.ColdDisk + "/",
				CacheEnabled:    true,
				CacheMaxSize:    o.CacheBytes,
			}},
			Policies: policies{HotCold: hotCold{
				Volumes: volumes{
					Default: hotVolume{Disk: "default"},
					Cold:    coldVolume{Disk: domain.ColdDisk},
				},
				MoveFactor: domain.MoveFactor,
			}},
		},
	}

	body, err := xml.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return nil, err
	}
	body = append([]byte(xml.Header), body...)
	body = append(body, '\n')

	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return nil, err
	}

	tmp, err := os.CreateTemp(r.dir, configFileName+".*.tmp")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return nil, err
	}
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp.Name(), r.path()); err != nil {
		return nil, err
	}
	return undo, nil
}
