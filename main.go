package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/dchest/uniuri"
	"github.com/pbnjay/memory"
)

const (
	master = 0
	probe  = 1
)

var containersImages [11]string = [11]string{
	"opendistro:1.11.0",
	"openvas:11",
	"postgres:13",
	"logstash:7.9.3",
	"opendistro-kibana:1.11.0",
	"rsyslog:8.36.0",
	"wazuh:3.11.1",
	"scanner:1.0.0",
	"nginx:1.19.5",
	"panel:7.0.0",
	"datasources:7.0.0",
}

func main() {
	remove := flag.Bool("remove", false, "Remove application's docker containers")
	user := flag.String("db-user", "", "User name that will be used for database connections")
	pass := flag.String("db-pass", "", "Password for the database user. Please use a secure password")
	fqdn := flag.String("fqdn", "", "Full qualified domain name, example: utmmaster.utmstack.com")
	customerName := flag.String("customer-name", "", "Your name, example: John Doe")
	customerEmail := flag.String("customer-email", "", "A valid email address to send important notifications about the system health. Example: john@doe.com")
	datadir := flag.String("datadir", "", "Data directory")
	flag.Parse()

	if *remove {
		uninstall()
	} else {
		if *user == "" || *pass == "" || *datadir == "" || *fqdn == "" || *customerName == "" || *customerEmail == "" {
			log.Fatal("ERROR: Missing arguments")
		}
		install(*user, *pass, *datadir, *fqdn, *customerName, *customerEmail)
	}
}

func uninstall() {
	check(runCmd("docker", "stack", "rm", "utmstack"))

	// remove images
	for _, image := range containersImages {
		check(runCmd("docker", "rmi", "utmstack.azurecr.io/"+image))
	}

	// logout from registry
	runCmd("docker", "logout", "utmstack.azurecr.io")
}

func install(user, pass, datadir, fqdn, customerName, customerEmail string) {
	serverName, err := os.Hostname()
	check(err)
	secret := uniuri.NewLen(10)
	esData := filepath.Join(datadir, "elasticsearch", "data")
	esBackups := filepath.Join(datadir, "elasticsearch", "backups")
	nginxCert := filepath.Join(datadir, "nginx", "cert")

	// create data folders
	os.MkdirAll(esData, 0777)
	os.MkdirAll(esBackups, 0777)
	os.MkdirAll(nginxCert, 0777)

	// setup docker
	if runCmd("docker", "version") != nil {
		installDocker()
	}
	runCmd("docker", "swarm", "init")

	// login to registry
	runCmd("docker", "login", "-u", "client", "-p", "4xYkVIAH8kdAH7mP/9BBhbb2ByzLGm4F", "utmstack.azurecr.io")

	// pull images from registry
	for _, image := range containersImages {
		check(runCmd("docker", "pull", "utmstack.azurecr.io/"+image))
	}

	// generate composer file and deploy
	f, err := os.Create(composerFile)
	check(err)
	defer f.Close()
	f.WriteString(composerTemplate)
	env := []string{
		"SERVER_TYPE=aio",
		"SERVER_NAME=" + serverName,
		"DB_USER=" + user,
		"DB_PASS=" + pass,
		"CLIENT_DOMAIN=" + fqdn,
		"CLIENT_NAME=" + customerName,
		"CLIENT_MAIL=" + customerEmail,
		"CLIENT_SECRET=" + secret,
		fmt.Sprint("ES_MEM=", (memory.TotalMemory()/uint64(math.Pow(1024, 3))-4)/2),
		"ES_DATA=" + esData,
		"ES_BACKUPS=" + esBackups,
		"NGINX_CERT=" + nginxCert,
	}
	check(runEnvCmd(env, "docker", "stack", "deploy", "--compose-file", composerFile, stackName))

	// configure elastic
	initializeElastic(secret)
}
