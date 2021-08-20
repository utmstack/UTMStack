package utils

import "os"

func InstallProbe(mode, datadir, pass, host string) error {
	logstashPipeline := MakeDir(0777, datadir, "logstash", "pipeline")
	datasourcesDir := MakeDir(0777, datadir, "datasources")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	mainIP := GetMainIP()
	mainIface := GetMainIface(mode)

	env := []string{
		"SERVER_TYPE=probe",
		"SERVER_NAME=" + serverName,
		"DB_HOST=" + host,
		"DB_PASS=" + pass,
		"LOGSTASH_PIPELINE=" + logstashPipeline,
		"UTMSTACK_DATASOURCES=" + datasourcesDir,
		"SCANNER_IFACE=" + mainIface,
		"SCANNER_IP=" + mainIP,
		"CORRELATION_URL=http://" + host + ":9090/v1/newlog",
	}

	if err := InstallScanner(mode); err != nil {
		return err
	}

	if err := InstallSuricata(mode, mainIface); err != nil {
		return err
	}

	if err := InitDocker(mode, probeTemplate, env, false); err != nil {
		return err
	}

	return nil
}
