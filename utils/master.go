package utils

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/quantfall/secp"
)

func InstallMaster(mode, datadir, pass, tag string, lite bool) error {
	p := secp.New()
	p.AllowedCapitalLetters = secp.AllEnglishCapitalLetters()
	p.AllowedLowerLetters = secp.AllEnglishLowerLetters()
	p.AllowedNumbers = secp.AllNumbers()
	p.AllowedSpecialCharacters = secp.SymbolsNonReservedByYAML()
	p.MinCapitalLetters = 3
	p.MinLowerLetters = 5
	p.MinNumbers = 5
	p.MinSpecialCharacters = 3

	com, perr := p.IsCompliant(pass)
	
	if perr != nil {
		return fmt.Errorf("Your password contains an invalid character. Please use only English letters, numbers and the following special characters: . , = + $ € / § _ ; <")
	}

	if !com {
		return fmt.Errorf("Your password does not meet the minimum security requirements. Please use at least 5 lower case letters, 3 capital case letters, 5 numbers and 3 special characters")
	}
	
	if lite {
		if err := CheckCPU(4); err != nil {
			return err
		}
		if err := CheckMem(5); err != nil {
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

	esData := MakeDir(0777, datadir, "opendistro", "data")
	esBackups := MakeDir(0777, datadir, "opendistro", "backups")
	cert := MakeDir(0777, datadir, "cert")
	logstashPipeline := MakeDir(0777, datadir, "logstash", "pipeline")
	datasourcesDir := MakeDir(0777, datadir, "datasources")
	rules := MakeDir(0777, datadir, "rules")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	mainIface, err := GetMainIface(mode)
	if err != nil {
		return err
	}

	mainIP, err := GetMainIP()
	if err != nil {
		return err
	}

	m := sigar.Mem{}
	m.Get()
	em := m.Total / 1024 / 1024 / 1024 / 4
	lm := m.Total / 1024 / 1024 / 1024 / 4

	var updates uint32

	if strings.Contains(tag, "testing") {
		updates = 60
	} else {
		updates = 3600
	}

	var c = Config{
		ServerType:       "aio",
		Lite:             lite,
		ServerName:       serverName,
		DBHost:           mainIP,
		DBPass:           pass,
		LogstashPipeline: logstashPipeline,
		ESMem:            em,
		LSMem:            lm,
		Updates:          updates,
		ESData:           esData,
		ESBackups:        esBackups,
		Cert:             cert,
		Datasources:      datasourcesDir,
		ScannerIface:     mainIface,
		ScannerIP:        mainIP,
		Correlation:      "http://correlation:8080/v1/newlog",
		Rules:            rules,
		Tag:              tag,
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

	if err := InitDocker(mode, c, true, tag, lite); err != nil {
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

	// Install OpenVPN Master
	if err := InstallOpenVPNMaster(mode); err != nil {
		return err
	}

	baseURL := "https://127.0.0.1/"

	for intent := 0; intent <= 10; intent++ {
		time.Sleep(1 * time.Minute)

		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := &http.Client{Transport: transCfg}

		_, err := client.Get(baseURL + "/api/ping")

		if err != nil && intent <= 9 {
			continue
		} else if err == nil {
			break
		}

		return err
	}

	if err := ConfigureFirewall(mode, c); err != nil {
		return err
	}

	return nil
}
