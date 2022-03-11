package utils

import (
	"fmt"
	"os"
	"strconv"

	sigar "github.com/cloudfoundry/gosigar"
)

func InstallProbe(mode, datadir, pass, host, tag string, lite bool) error {
	if lite {
		if err := CheckCPU(4); err != nil {
			return err
		}
		if err := CheckMem(4); err != nil {
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

	logstashPipeline := MakeDir(0777, datadir, "logstash", "pipeline")
	datasourcesDir := MakeDir(0777, datadir, "datasources")
	cert := MakeDir(0777, datadir, "cert")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	mainIP, err := GetMainIP()
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

	if tag == "testing" {
		updates = 60
	} else {
		updates = 3600
	}

	env := []string{
		"SERVER_TYPE=probe",
		"LITE=" + strconv.FormatBool(lite),
		"SERVER_NAME=" + serverName,
		"DB_HOST=10.21.199.1",
		"DB_PASS=" + pass,
		"LOGSTASH_PIPELINE=" + logstashPipeline,
		fmt.Sprint("LS_MEM=", lm),
		fmt.Sprint("UPDATES=", updates),
		"CERT=" + cert,
		"UTMSTACK_DATASOURCES=" + datasourcesDir,
		"SCANNER_IFACE=" + mainIface,
		"SCANNER_IP=" + mainIP,
		"CORRELATION_URL=http://10.21.199.1:9090/v1/newlog",
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

	if err := InitDocker(mode, env, false, tag, lite, mainIP); err != nil {
		return err
	}

	if err := InstallOpenVPNClient(mode, host); err != nil {
		return err
	}

	if err := ConfigureFirewall(mode); err != nil {
		return err
	}

	return nil
}
