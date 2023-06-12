package utils

import (
	"fmt"
	"os"
	"strings"

	sigar "github.com/cloudfoundry/gosigar"
)

func InstallProbe(mode, datadir, pass, host, tag string) error {
	if err := CheckCPU(2); err != nil {
		return err
	}

	if err := CheckMem(4); err != nil {
		return err
	}

	logstashPipeline := MakeDir(0777, datadir, "logstash", "pipeline")
	datasourcesDir := MakeDir(0777, datadir, "datasources")
	cert := MakeDir(0777, datadir, "cert")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	tunIP, err := GetIfaceIP("tun0")
	if err != nil {
		return err
	}

	mainIface, err := GetMainIface(mode)
	if err != nil {
		return err
	}

	m := sigar.Mem{}
	m.Get()
	lm := m.Total / 1024 / 1024 / 1024 / 3

	var updates uint32

	if strings.Contains(tag, "testing") {
		updates = 60
	} else {
		updates = 3600
	}

	var c = Config{
		ServerType:       "probe",
		ServerName:       serverName,
		DBHost:           host,
		DBPass:           pass,
		LogstashPipeline: logstashPipeline,
		LSMem:            lm,
		Updates:          updates,
		Cert:             cert,
		Datasources:      datasourcesDir,
		ScannerIface:     mainIface,
		ScannerIP:        tunIP,
		Correlation:      fmt.Sprintf("http://%s:9090/v1/newlog", host),
		Tag:              tag,
		Kind:             "probe",
		Last:             -1,
	}

	// Generate auto-signed cert and key
	if err := generateCerts(cert); err != nil {
		return err
	}

	if err := InitDocker(mode, c, false, tag); err != nil {
		return err
	}

	if err := ConfigureFirewall(mode, c); err != nil {
		return err
	}

	return nil
}
