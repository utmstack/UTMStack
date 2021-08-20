package utils

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/dchest/uniuri"
	"github.com/pbnjay/memory"
)

func InstallMaster(mode, datadir, pass, branch string) error {
	esData := MakeDir(0777, datadir, "elasticsearch", "data")
	esBackups := MakeDir(0777, datadir, "elasticsearch", "backups")
	_ = MakeDir(0777, datadir, "elasticsearch", "backups", "incremental")
	nginxCert := MakeDir(0777, datadir, "nginx", "cert")
	logstashPipeline := MakeDir(0777, datadir, "logstash", "pipeline")
	datasourcesDir := MakeDir(0777, datadir, "datasources")
	rules := MakeDir(0777, datadir, "rules")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	secret := uniuri.NewLenChars(10, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))

	mainIP := GetMainIP()
	mainIface := GetMainIface(mode)

	env := []string{
		"SERVER_TYPE=aio",
		"SERVER_NAME=" + serverName,
		"DB_PASS=" + pass,
		"CLIENT_SECRET=" + secret,
		fmt.Sprint("ES_MEM=", (memory.TotalMemory()/uint64(math.Pow(1024, 3))-4)/2),
		"ES_DATA=" + esData,
		"ES_BACKUPS=" + esBackups,
		"NGINX_CERT=" + nginxCert,
		"LOGSTASH_PIPELINE=" + logstashPipeline,
		"UTMSTACK_DATASOURCES=" + datasourcesDir,
		"SCANNER_IFACE=" + mainIface,
		"SCANNER_IP=" + mainIP,
		"CORRELATION_URL=http://correlation:8080/v1/newlog",
		"UTMSTACK_RULES=" + rules,
		"BRANCH=" + branch,
	}

	if err := InstallScanner(mode); err != nil {
		return err
	}

	if err := InstallSuricata(mode, mainIface); err != nil {
		return err
	}

	if err := InitDocker(mode, masterTemplate, env, true, branch); err != nil {
		return err
	}

	// Generate nginx auto-signed cert and key
	if err := generateCerts(nginxCert); err != nil {
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

	time.Sleep(7 * time.Minute)

	return nil
}
