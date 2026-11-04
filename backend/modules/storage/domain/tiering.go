package domain

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	PolicyName              = "hot_cold"
	HotVolume               = "default"
	ColdVolume              = "cold"
	ColdDisk                = "cold_disk"
	MoveFactor              = 0.15
	DefaultCacheBytes int64 = 10 << 30 // 10 GiB

	// LocalMount is where the installer mounts the directory that backs a local
	// cold disk, so a folder or an NFS share is whatever the host put there.
	LocalMount = "/mnt/utmstack-cold/"
)

type ColdKind string

const (
	KindS3    ColdKind = "s3"
	KindAzure ColdKind = "azure"
	KindLocal ColdKind = "local"
)

// ColdTarget is where the cold volume lives. For s3 the Endpoint is the bucket
// URL; for azure it is the storage account URL and Container names the
// container; for local nothing is asked, because the installer decides the
// mount. With UseRole the s3 keys are left to the server's own environment
// (an IAM role), so none are stored.
type ColdTarget struct {
	Kind       ColdKind
	Endpoint   string
	Container  string
	AccessKey  string
	SecretKey  string
	UseRole    bool
	CacheBytes int64
}

type Tiering struct {
	Configured bool
	Ready      bool
	Kind       ColdKind
	Endpoint   string
	Policy     string
}

// ColdDiskInfo is what the server reports about the disk behind the cold volume.
type ColdDiskInfo struct {
	Present   bool
	ReadOnly  bool
	FreeBytes uint64
}

// Location identifies where records live.
func (o ColdTarget) Location() string {
	if o.Kind == KindAzure {
		return o.Endpoint + "/" + o.Container
	}
	return o.Endpoint
}

func (o ColdTarget) Normalized() ColdTarget {
	if o.Kind == "" {
		o.Kind = KindS3
	}
	o.Endpoint = strings.TrimSpace(o.Endpoint)
	o.Container = strings.TrimSpace(o.Container)
	o.AccessKey = strings.TrimSpace(o.AccessKey)
	o.SecretKey = strings.TrimSpace(o.SecretKey)

	switch o.Kind {
	case KindLocal:
		return ColdTarget{Kind: KindLocal, Endpoint: LocalMount}
	case KindAzure:
		o.Endpoint = strings.TrimRight(o.Endpoint, "/")
		o.UseRole = false
	default:
		if o.Endpoint != "" && !strings.HasSuffix(o.Endpoint, "/") {
			o.Endpoint += "/"
		}
		o.Container = ""
		if o.UseRole {
			o.AccessKey, o.SecretKey = "", ""
		}
	}
	if o.CacheBytes == 0 {
		o.CacheBytes = DefaultCacheBytes
	}
	return o
}

func (o ColdTarget) Validate() error {
	o = o.Normalized()

	switch o.Kind {
	case KindLocal:
		return nil
	case KindS3, KindAzure:
	default:
		return fmt.Errorf("%w: %s", ErrKindUnknown, o.Kind)
	}

	if o.Endpoint == "" {
		return ErrEndpointRequired
	}
	u, err := url.Parse(o.Endpoint)
	if err != nil || !u.IsAbs() || u.Host == "" {
		return fmt.Errorf("%w: %s", ErrEndpointNotURL, o.Endpoint)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: %s", ErrEndpointNotURL, u.Scheme)
	}
	if o.CacheBytes < 0 {
		return ErrCacheNegative
	}

	if o.Kind == KindAzure {
		if o.Container == "" {
			return ErrContainerRequired
		}
		if o.AccessKey == "" || o.SecretKey == "" {
			return ErrAccountNeeded
		}
		return nil
	}

	if strings.Trim(u.Path, "/") == "" {
		return ErrEndpointNoBucket
	}
	if !o.UseRole && (o.AccessKey == "" || o.SecretKey == "") {
		return ErrCredentialsNeeded
	}
	return nil
}
