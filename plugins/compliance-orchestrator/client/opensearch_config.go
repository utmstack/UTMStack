package client

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/threatwinds/go-sdk/plugins"
)

type OpenSearchConfig struct {
	Host string
	Port string
	User string
	Pass string
	URL  string
}

func LoadOpenSearchConfig() OpenSearchConfig {
	cfg := plugins.PluginCfg("org.opensearch", false).Get("opensearch")

	host := cfg.Get("host").String()
	port := cfg.Get("port").String()
	user := cfg.Get("user").String()
	pass := cfg.Get("password").String()

	return OpenSearchConfig{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		URL:  fmt.Sprintf("https://%s:%s", host, port),
	}
}

func NewOpenSearchHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
