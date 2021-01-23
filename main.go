package main

import (
	"flag"
	"log"
	"math"
	"os"
	"path/filepath"
	"text/template"

	"github.com/dchest/uniuri"
	"github.com/pbnjay/memory"
)

type TemplateArgs struct {
	ServerName string
	User       string
	Pass       string
	// master specific:
	FQDN          string
	CustomerName  string
	CustomerEmail string
	Secret        string
	EsMem         uint64
	EsData        string
	EsBackups     string
	NginxCert     string
}

const (
	master = 0
	probe  = 1
)

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
}

func install(user, pass, datadir, fqdn, customerName, customerEmail string) {
	args := TemplateArgs{
		User:          user,
		Pass:          pass,
		FQDN:          fqdn,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Secret:        uniuri.NewLen(10),
		EsMem:         (memory.TotalMemory()/uint64(math.Pow(1024, 3)) - 4) / 2,
		EsData:        filepath.Join("elasticsearch", "data"),
		EsBackups:     filepath.Join("elasticsearch", "backups"),
		NginxCert:     filepath.Join("nginx", "cert"),
	}
	var err error
	args.ServerName, err = os.Hostname()
	check(err)

	// setup docker
	if runCmd("docker", "version") != nil {
		installDocker()
	}
	runCmd("docker", "swarm", "init")

	// generate composer file
	tmplName := "utmstack.yml"
	tmpl := template.Must(template.New(tmplName).Parse(master_template))
	f, err := os.Create(tmplName)
	check(err)
	defer f.Close()
	tmpl.Execute(f, args)
	// deploy
	check(runCmd("docker", "stack", "deploy", "--compose-file", tmplName, "utmstack"))

	// configure elastic
	initializeElastic(args.Secret)
}
