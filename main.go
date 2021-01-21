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
	// serverType := flag.Int("type", master, "Server type: 0(master) or 1(probe)")
	// host := flag.String("host", "", "Master host name")
	user := flag.String("user", "", "DB username")
	pass := flag.String("pass", "", "DB password")
	datadir := flag.String("datadir", "", "Data directory")
	flag.Parse()

	if *remove {
		uninstall()
	} else {
		if *user == "" || *pass == "" || *datadir == "" {
			log.Fatal("ERROR: Missing arguments")
		}
		install(*user, *pass, *datadir)
	}
}

func uninstall() {
	checkCmd("docker", "stack", "rm", "utmstack")
}

func install(user, pass string, datadir string) {
	args := TemplateArgs{
		User:    user,
		Pass:    pass,
		DataDir: datadir,
	}
	var err error
	args.ServerName, err = os.Hostname()
	check(err)
	args.Secret = uniuri.New()
	args.EsMem = (memory.TotalMemory()/uint64(math.Pow(1024, 3)) - 4) / 2

	// setup docker
	if runCmd("docker", "version") != nil {
		// TODO: install docker
	}
	runCmd("docker", "swarm", "init")

	// generate composer file
	tmpl := template.Must(template.New("utmstack.yml").Parse(master_template))
	f, err := os.Create("utmstack.yml")
	check(err)
	defer f.Close()
	tmpl.Execute(f, args)
	// deploy
	checkCmd("docker", "stack", "deploy", "--compose-file", "utmstack.yml", "utmstack")

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
