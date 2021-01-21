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
	user := flag.String("user", "", "DB username")
	pass := flag.String("pass", "", "DB password")
	datadir := flag.String("datadir", "", "Data directory")
	// TODO: request needed client data
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
	check(runCmd("docker", "stack", "rm", "utmstack"))
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
		aptGet("update")
		aptGet("install", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common")

		env := []string{"DEBIAN_FRONTEND=noninteractive"}
		runEnvCmd(env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -")
		runEnvCmd(env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`)

		aptGet("update")
		aptGet("install", "docker-ce", "docker-ce-cli", "containerd.io")
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
