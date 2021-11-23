package utils

import (
	"os"

)

func InstallProbe(mode, datadir, pass, host, tag string) error {
	if err := CheckCPU(4); err != nil {
		return err
	}
	if err := CheckMem(3); err != nil {
		return err
	}

	logstashPipeline := MakeDir(0777, datadir, "logstash", "pipeline")
	datasourcesDir := MakeDir(0777, datadir, "datasources")

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

	env := []string{
		"SERVER_TYPE=probe",
		"SERVER_NAME=" + serverName,
		"DB_HOST=10.21.199.1",
		"DB_PASS=" + pass,
		"LOGSTASH_PIPELINE=" + logstashPipeline,
		"UTMSTACK_DATASOURCES=" + datasourcesDir,
		"SCANNER_IFACE=" + mainIface,
		"SCANNER_IP=" + mainIP,
		"CORRELATION_URL=http://10.21.199.1:9090/v1/newlog",
		"TAG=" + tag,
	}

	if err := InstallScanner(mode); err != nil {
		return err
	}

	if err := InstallSuricata(mode, mainIface); err != nil {
		return err
	}

	if err := InstallOpenVPNClient(mode, host); err != nil {
		return err
	}

	if err := InitDocker(mode, probeTemplate, env, false, tag); err != nil {
		return err
	}

	return nil
}
