package utils

import (
	"os"
	"strings"

	sigar "github.com/cloudfoundry/gosigar"
)

func InstallProbe(mode, datadir, pass, host, tag string, lite bool) error {
	if lite {
		if err := CheckCPU(2); err != nil {
			return err
		}
		if err := CheckMem(3); err != nil {
			return err
		}
	} else {
		if err := CheckCPU(4); err != nil {
			return err
		}
		if err := CheckMem(5); err != nil {
			return err
		}
	}

	logstashPipeline := MakeDir(0777, datadir, "logstash", "pipeline")
	datasourcesDir := MakeDir(0777, datadir, "datasources")
	cert := MakeDir(0777, datadir, "cert")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	if err := InstallOpenVPNClient(mode, host); err != nil {
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
		Lite:             lite,
		ServerName:       serverName,
		DBHost:           "10.21.199.1",
		DBPass:           pass,
		LogstashPipeline: logstashPipeline,
		LSMem:            lm,
		Updates:          updates,
		Cert:             cert,
		Datasources:      datasourcesDir,
		ScannerIface:     mainIface,
		ScannerIP:        tunIP,
		Correlation:      "http://10.21.199.1:9090/v1/newlog",
		Tag:              tag,
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

	if err := InitDocker(mode, c, false, tag, lite); err != nil {
		return err
	}

	if err := ConfigureFirewall(mode); err != nil {
		return err
	}

	return nil
}
