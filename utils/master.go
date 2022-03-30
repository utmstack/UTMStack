package utils

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
)

func InstallMaster(mode, datadir, pass, tag string, lite bool) error {
	if lite {
		if err := CheckCPU(4); err != nil {
			return err
		}
		if err := CheckMem(5); err != nil {
			return err
		}
	} else {
		if err := CheckCPU(4); err != nil {
			return err
		}
		if err := CheckMem(7); err != nil {
			return err
		}
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

	mainIface, err := GetMainIface(mode)
	if err != nil {
		return err
	}

	m := sigar.Mem{}
	m.Get()
	em := m.Total / 1024 / 1024 / 1024 / 3
	lm := m.Total / 1024 / 1024 / 1024 / 3

	var updates uint32

	if strings.Contains(tag, "testing") {
		updates = 60
	} else {
		updates = 3600
	}

	env := []string{
		"SERVER_TYPE=aio",
		"LITE=" + strconv.FormatBool(lite),
		"SERVER_NAME=" + serverName,
		"DB_HOST=10.21.199.1", // is always 10.21.199.1, deprecated
		"DB_PASS=" + pass,
		"LOGSTASH_PIPELINE=" + logstashPipeline,
		fmt.Sprint("ES_MEM=", em),
		fmt.Sprint("LS_MEM=", lm),
		fmt.Sprint("UPDATES=", updates),
		"ES_DATA=" + esData,
		"ES_BACKUPS=" + esBackups,
		"CERT=" + cert,
		"UTMSTACK_DATASOURCES=" + datasourcesDir,
		"SCANNER_IFACE=" + mainIface,
		"SCANNER_IP=10.21.199.1",
		"CORRELATION_URL=http://10.21.199.1:9090/v1/newlog", // is always the same, deprecated
		"UTMSTACK_RULES=" + rules,
		"TAG=" + tag,
	}

	if !lite {
		if err := InstallSuricata(mode, mainIface); err != nil {
			return err
		}
	}

	// Generate auto-signed cert and key
	if err := generateCerts(cert); err != nil {
		return err
	}

	if err := InitDocker(mode, env, true, tag, lite); err != nil {
		return err
	}

	// configure elastic
	if err := initializeElastic(); err != nil {
		return err
	}

	// Initialize PostgreSQL Database
	if err := initializePostgres(pass); err != nil {
		return err
	}

	// Install OpenVPN Master
	if err := InstallOpenVPNMaster(mode); err != nil {
		return err
	}

	baseURL := "https://127.0.0.1/"

	for intent := 0; intent <= 5; intent++ {
		time.Sleep(2 * time.Minute)

		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{Transport: transCfg}

		_, err := client.Get(baseURL + "/utmstack/api/ping")

		if err == nil {
			break
		}
	}

	if err := ConfigureFirewall(mode); err != nil {
		return err
	}

	return nil
}
