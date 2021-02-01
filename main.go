package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/dchest/uniuri"
	"github.com/pbnjay/memory"
)

const (
	master = 0
	probe  = 1
)

var containersImages [10]string = [10]string{
	"opendistro:1.11.0",
	"openvas:11",
	"postgres:13",
	"logstash:7.9.3",
	"rsyslog:8.36.0",
	"wazuh:3.11.1",
	"scanner:1.0.0",
	"nginx:1.19.5",
	"panel:7.0.0",
	"datasources:7.0.0",
}

func main() {
	parser := argparse.NewParser("", "UTMStack installer")
	removeCmd := parser.NewCommand("remove", "Uninstall UTMStack")

	masterCmd := parser.NewCommand("master", "Install master server")
	masterDataDir := masterCmd.String("", "datadir", &argparse.Options{Required: true, Help: "Data directory"})
	masterPass := masterCmd.String("", "db-pass", &argparse.Options{Required: true, Help: "Password to access the database. Please use a secure password"})
	fqdn := masterCmd.String("", "fqdn", &argparse.Options{Required: true, Help: "Full qualified domain name, example: utmmaster.utmstack.com"})
	customerName := masterCmd.String("", "customer-name", &argparse.Options{Required: true, Help: "Your name, example: John Doe"})
	customerEmail := masterCmd.String("", "customer-email", &argparse.Options{Required: true, Help: "A valid email address to send important notifications about the system health. Example: john@doe.com"})

	probeCmd := parser.NewCommand("probe", "Install probe")
	probeDataDir := probeCmd.String("", "datadir", &argparse.Options{Required: true, Help: "Data directory"})
	probePass := probeCmd.String("", "db-pass", &argparse.Options{Required: true, Help: "Password to access the database"})
	host := probeCmd.String("", "host", &argparse.Options{Required: true, Help: "Master server address"})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Print(parser.Usage(err))
		return
	}

	if removeCmd.Happened() {
		uninstall()
	} else if masterCmd.Happened() {
		installMaster(*masterDataDir, *masterPass, *fqdn, *customerName, *customerEmail)
	} else if probeCmd.Happened() {
		installProbe(*probeDataDir, *probePass, *host)
	}
}

func uninstall() {
	check(runCmd("docker", "stack", "rm", "utmstack"))
	
	// sleep while docker is removing the containers
	time.Sleep(50 * time.Second)
	
	// remove images
	for _, image := range containersImages {
		check(runCmd("docker", "rmi", "utmstack.azurecr.io/"+image))
	}

	// logout from registry
	runCmd("docker", "logout", "utmstack.azurecr.io")
}

func installProbe(datadir, pass, host string) {
	logstashPipeline := mkdirs(0777, datadir, "logstash", "pipeline")
	logsDir := mkdirs(0777, datadir, "logs")

	serverName, err := os.Hostname()
	check(err)
	env := []string{
		"SERVER_TYPE=probe",
		"SERVER_NAME=" + serverName,
		"DB_HOST=" + host,
		"DB_PASS=" + pass,
		"LOGSTASH_PIPELINE=" + logstashPipeline,
		"UTMSTACK_LOGSDIR=" + logsDir,
	}

	initDocker(baseTemplate, env)
}

func installMaster(datadir, pass, fqdn, customerName, customerEmail string) {
	esData := mkdirs(0777, datadir, "elasticsearch", "data")
	esBackups := mkdirs(0777, datadir, "elasticsearch", "backups")
	nginxCert := mkdirs(0777, datadir, "nginx", "cert")
	logstashPipeline := mkdirs(0777, datadir, "logstash", "pipeline")
	logsDir := mkdirs(0777, datadir, "logs")

	serverName, err := os.Hostname()
	check(err)
	secret := uniuri.NewLenChars(10, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
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
		"UTMSTACK_LOGSDIR=" + logsDir,
	}

	initDocker(masterTemplate, env)

	// Generate nginx auto-signed cert and key
	generateCerts(nginxCert)

	// configure elastic
	initializeElastic(secret)

	// Initialize PostgreSQL Database
	initializePostgres(pass, customerName, fqdn, secret, customerEmail)
}
