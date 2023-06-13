package utils

import (
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
)

func InstallMaster(mode, datadir, pass, tag string) error {
	if err := CheckCPU(4); err != nil {
		return err
	}

	if err := CheckMem(8); err != nil {
		return err
	}

	esData := MakeDir(0777, datadir, "opendistro", "data")
	esBackups := MakeDir(0777, datadir, "opendistro", "backups")
	cert := MakeDir(0777, datadir, "cert")
	logstashPipeline := MakeDir(0777, datadir, "logstash", "pipeline")
	datasourcesDir := MakeDir(0777, datadir, "datasources")
	rules := MakeDir(0777, datadir, "rules")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	mainIP, err := GetMainIP()
	if err != nil {
		return err
	}

	m := sigar.Mem{}
	m.Get()
	em := m.Total / 1024 / 1024 / 1024 / 4
	lm := m.Total / 1024 / 1024 / 1024 / 4

	var updates uint32

	if strings.Contains(tag, "testing") {
		updates = 60
	} else {
		updates = 3600
	}

	var c = Config{
		ServerType:       "aio",
		ServerName:       serverName,
		DBHost:           mainIP,
		DBPass:           pass,
		LogstashPipeline: logstashPipeline,
		ESMem:            em,
		LSMem:            lm,
		Updates:          updates,
		ESData:           esData,
		ESBackups:        esBackups,
		Cert:             cert,
		Datasources:      datasourcesDir,
		Correlation:      "http://correlation:8080/v1/newlog",
		Rules:            rules,
		Tag:              tag,
	}

	// Generate auto-signed cert and key
	if err := generateCerts(cert); err != nil {
		return err
	}

	if err := InitDocker(mode, c, true, tag, mainIP); err != nil {
		return err
	}

	baseURL := "https://127.0.0.1/"

	for intent := 0; intent <= 10; intent++ {
		time.Sleep(1 * time.Minute)

		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{Transport: transCfg}

		_, err := client.Get(baseURL + "/api/ping")

		if err != nil && intent <= 9 {
			continue
		} else if err == nil {
			break
		}

		return err
	}

	return nil
}
