package main

import (
	"flag"
	"log"
	"math"
	"os"
	"text/template"

	"github.com/dchest/uniuri"
	"github.com/levigross/grequests"
	"github.com/pbnjay/memory"
)

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
		User:    user,
		Pass:    pass,
		DataDir: datadir,
		FQDN: fqdn,
		CustomerName: customerName,
		CustomerEmail: customerEmail,
	}
	var err error
	args.ServerName, err = os.Hostname()
	check(err)
	args.Secret = uniuri.NewLen(10)
	args.EsMem = (memory.TotalMemory()/uint64(math.Pow(1024, 3)) - 4) / 2

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

	// wait for elastic to be ready
	for {
		_, err := grequests.Get("http://localhost:9200/_cluster/healt", &grequests.RequestOptions{
			Params: map[string]string{
				"wait_for_status": "yellow",
				"timeout":         "50s",
			},
		})
		if err == nil {
			break
		}
	}
	// TODO: configure elastic
}
