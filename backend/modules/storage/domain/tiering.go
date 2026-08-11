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
)

type ObjectStore struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	CacheBytes int64
}

type Tiering struct {
	Configured bool
	Ready      bool
	Endpoint   string
	Policy     string
}

func (o ObjectStore) Normalized() ObjectStore {
	o.Endpoint = strings.TrimSpace(o.Endpoint)
	if o.Endpoint != "" && !strings.HasSuffix(o.Endpoint, "/") {
		o.Endpoint += "/"
	}
	o.AccessKey = strings.TrimSpace(o.AccessKey)
	o.SecretKey = strings.TrimSpace(o.SecretKey)
	if o.CacheBytes == 0 {
		o.CacheBytes = DefaultCacheBytes
	}
	return o
}

func (o ObjectStore) Validate() error {
	o = o.Normalized()

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
	if strings.Trim(u.Path, "/") == "" {
		return ErrEndpointNoBucket
	}
	if o.AccessKey == "" || o.SecretKey == "" {
		return ErrCredentialsNeeded
	}
	if o.CacheBytes < 0 {
		return ErrCacheNegative
	}
	return nil
}
