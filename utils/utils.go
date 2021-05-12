package utils

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dchest/uniuri"
	"github.com/levigross/grequests"
	"github.com/pbnjay/memory"

	_ "github.com/lib/pq" //Import PostgreSQL driver
)

var containersImages = [9]string{
	"opendistro:testing",
	"openvas:testing",
	"postgres:testing",
	"logstash:testing",
	"nginx:testing",
	"panel:testing",
	"datasources:testing",
	"zapier:testing",
	"correlation:testing",
}

func runEnvCmd(mode string, env []string, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Env = append(os.Environ(), env...)
	if mode == "cli" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func runCmd(mode, command string, arg ...string) error {
	return runEnvCmd(mode, []string{}, command, arg...)
}

func mkdirs(mode os.FileMode, arg ...string) string {
	path := ""
	for _, folder := range arg {
		path = filepath.Join(path, folder)
		os.MkdirAll(path, mode)
		os.Chmod(path, mode)
	}
	return path
}

// Check Check if error is not nil or exit with error code
func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Uninstall Uninstall UTMStack
func Uninstall(mode string) error {
	err := runCmd(mode, "docker", "stack", "rm", "utmstack")
	if err != nil {
		return errors.New(`failed to remove "utmstack" docker stack`)
	}

	// sleep while docker is removing the containers
	time.Sleep(120 * time.Second)

	// uninstall scanner
	if err := runCmd(mode, "systemctl", "stop", "utm_scanner"); err != nil {
		return err
	}

	if err := runCmd(mode, "rm", "/etc/systemd/system/utm_scanner.service"); err != nil {
		return err
	}

	if err := runCmd(mode, "systemctl", "daemon-reload"); err != nil {
		return err
	}

	if err := runCmd(mode, "rm", "-R", "/opt/scanner/"); err != nil {
		return err
	}

	// uninstall suricata
	if err := runCmd(mode, "apt", "remove", "-y", "--purge", "suricata"); err != nil {
		return err
	}

	if err := runCmd(mode, "rm", "-R", "/var/log/suricata"); err != nil {
		return err
	}

	// remove images
	for _, image := range containersImages {
		image = "utmstack.azurecr.io/" + image
		if err := runCmd(mode, "docker", "rmi", image); err != nil {
			return errors.New("failed to remove docker image: " + image)
		}
	}

	// logout from registry
	if err := runCmd(mode, "docker", "logout", "utm_scanner.service"); err != nil {
		return err
	}

	// leave swarm
	if err := runCmd(mode, "docker", "swarm", "leave", "--force"); err != nil {
		return err
	}

	return nil
}

// InstallProbe Install UTMStack probe
func InstallProbe(mode, datadir, pass, host string) error {
	logstashPipeline := mkdirs(0777, datadir, "logstash", "pipeline")
	datasourcesDir := mkdirs(0777, datadir, "datasources")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	mainIP := getMainIP()
	mainIface := getMainIface(mode)

	env := []string{
		"SERVER_TYPE=probe",
		"SERVER_NAME=" + serverName,
		"DB_HOST=" + host,
		"DB_PASS=" + pass,
		"LOGSTASH_PIPELINE=" + logstashPipeline,
		"UTMSTACK_DATASOURCES=" + datasourcesDir,
		"SCANNER_IFACE=" + mainIface,
		"SCANNER_IP=" + mainIP,
		"CORRElATION_URL=http://" + host + ":9090",
	}

	if err := installScanner(mode); err != nil {
		return err
	}

	if err := installSuricata(mode, mainIface); err != nil {
		return err
	}

	if err := initDocker(mode, baseTemplate, env); err != nil {
		return err
	}

	return nil
}

// InstallMaster Install UTMStack Master
func InstallMaster(mode, datadir, pass, fqdn, customerName, customerEmail string) error {
	esData := mkdirs(0777, datadir, "elasticsearch", "data")
	esBackups := mkdirs(0777, datadir, "elasticsearch", "backups")
	_ = mkdirs(0777, datadir, "elasticsearch", "backups", "incremental")
	nginxCert := mkdirs(0777, datadir, "nginx", "cert")
	logstashPipeline := mkdirs(0777, datadir, "logstash", "pipeline")
	datasourcesDir := mkdirs(0777, datadir, "datasources")
	rules := mkdirs(0777, datadir, "rules")

	serverName, err := os.Hostname()
	if err != nil {
		return err
	}

	secret := uniuri.NewLenChars(10, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))

	mainIP := getMainIP()
	mainIface := getMainIface(mode)

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
		"CORRELATION_URL=http://correlation:8080",
		"UTMSTACK_RULES=" + rules,
	}

	if err := installScanner(mode); err != nil {
		return err
	}

	if err := installSuricata(mode, mainIface); err != nil {
		return err
	}

	if err := initDocker(mode, masterTemplate, env); err != nil {
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
	if err := initializePostgres(pass, customerName, fqdn, secret, customerEmail); err != nil {
		return err
	}

	return nil
}

func getMainIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func getMainIface(mode string) string {
	if err := runCmd(mode, "apt", "update"); err != nil {
		log.Fatal(err)
	}

	if err := runCmd(mode, "apt", "install", "-y", "wget", "net-tools"); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("/bin/sh", "-c", "route | grep '^default' | grep -o '[^ ]*$'")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	s := buf.String()
	return s[:len(s)-1]
}

func initDocker(mode, composerTemplate string, env []string) error {
	if err := installDocker(mode); err != nil {
		return err
	}

	if err := runCmd(mode, "docker", "swarm", "init"); err != nil {
		return err
	}

	// login to registry
	if err := runCmd(mode, "docker", "login", "-u", "client", "-p", "4xYkVIAH8kdAH7mP/9BBhbb2ByzLGm4F", "utmstack.azurecr.io"); err != nil {
		return err
	}

	// pull images from registry
	for _, image := range containersImages {
		image = "utmstack.azurecr.io/" + image
		if err := runCmd(mode, "docker", "pull", image); err != nil {
			return errors.New("Failed to pull docker image: " + image)
		}
	}

	// generate composer file and deploy
	f, err := os.Create(composerFile)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(composerTemplate)

	for i := 1; i <= 3; i++ {
		err := runEnvCmd(mode, env, "docker", "stack", "deploy", "--compose-file", composerFile, stackName)
		if err == nil {
			break
		} else if i == 3 {
			return errors.New("failed to deploy stack")
		}
	}

	return nil
}

func installScanner(mode string) error {
	if err := runCmd(mode, "apt", "install", "-y", "python3", "python3-pip", "wget", "zip", "unzip"); err != nil {
		return err
	}

	if err := runCmd(mode, "wget", "-O", "scanner.zip", "https://updates.utmstack.com/assets/scanner.zip"); err != nil {
		return err
	}

	if err := runCmd(mode, "unzip", "scanner.zip"); err != nil {
		return err
	}

	if err := runCmd(mode, "mv", "UTMStackHostScanner/", "/opt/scanner"); err != nil {
		return err
	}

	if err := runCmd(mode, "mv", "/opt/scanner/utm_scanner.service", "/etc/systemd/system/"); err != nil {
		return err
	}

	if err := runCmd(mode, "chmod", "777", "/opt/scanner/run.sh"); err != nil {
		return err
	}

	if err := runCmd(mode, "pip3", "install", "-r", "/opt/scanner/requirements.txt"); err != nil {
		return err
	}

	if err := runCmd(mode, "systemctl", "enable", "utm_scanner"); err != nil {
		return err
	}

	if err := runCmd(mode, "systemctl", "start", "utm_scanner"); err != nil {
		return err
	}

	return nil
}

func installSuricata(mode string, iface string) error {
	if err := runCmd(mode, "add-apt-repository", "-y", "ppa:oisf/suricata-stable"); err != nil {
		return err
	}

	env := []string{
		"DEBIAN_FRONTEND=noninteractive",
	}

	if err := runEnvCmd(mode, env, "apt", "install", "-y", "suricata"); err != nil {
		return err
	}

	if err := runCmd(mode, "wget", "-O", "/etc/suricata/suricata.yaml", "https://updates.utmstack.com/assets/suricata.yaml"); err != nil {
		return err
	}

	sed := fmt.Sprintf("sed -i \"s/SCANNER_IFACE/%s/g\" /etc/suricata/suricata.yaml", iface)

	if err := runCmd(mode, "/bin/sh", "-c", sed); err != nil {
		return err
	}

	if err := runCmd(mode, "systemctl", "restart", "suricata"); err != nil {
		return err
	}

	if err := runCmd(mode, "suricata-update"); err != nil {
		return err
	}

	return nil
}

func generateCerts(nginxCert string) error {
	cert := `-----BEGIN CERTIFICATE-----
MIIFJjCCBA6gAwIBAgISA7ylpw0Ob1YkGwHhx3lwj3gwMA0GCSqGSIb3DQEBCwUA
MDIxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MQswCQYDVQQD
EwJSMzAeFw0yMTAxMDExNDMwMTRaFw0yMTA0MDExNDMwMTRaMBwxGjAYBgNVBAMT
EXNpZW0udXRtc3RhY2suY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEAwOCCCS37KDMRpsW36wBt80hz4G5HCHnQZhvkpPE3eSGTAHYu1H4ZoPVpe/47
qf1qxW3+Y4dzICn4KpeBSiDhC4L4vU/hEmjrhtpQgy+7Acxqk3WmlJ9i/fSsTiwA
ahyIGdyR1FDW1WzpT5nCF2v+xVQM2vW3HOCa2P4M0Edc7T8ki7ZBuonc3ViMCc8K
yuXQ37auVVU6znH1Nm/MhimSFGvf6ZKVRm/ANPtMnnrw6J2IZ4YtPhzfBJ1fWNYg
5FqI/+iZ/HE/gxYxPowb9RyEFi7y8OWoD+j+8c1PFlfq9nGcEOtw/LWsnwvG2nnS
afQ89BuSxW5xf3Er3P4KqV6/jQIDAQABo4ICSjCCAkYwDgYDVR0PAQH/BAQDAgWg
MB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMB0G
A1UdDgQWBBStPJIHEpzk/cLXMUjzToJEWcxFFzAfBgNVHSMEGDAWgBQULrMXt1hW
y65QCUDmH6+dixTCxjBVBggrBgEFBQcBAQRJMEcwIQYIKwYBBQUHMAGGFWh0dHA6
Ly9yMy5vLmxlbmNyLm9yZzAiBggrBgEFBQcwAoYWaHR0cDovL3IzLmkubGVuY3Iu
b3JnLzAcBgNVHREEFTATghFzaWVtLnV0bXN0YWNrLmNvbTBMBgNVHSAERTBDMAgG
BmeBDAECATA3BgsrBgEEAYLfEwEBATAoMCYGCCsGAQUFBwIBFhpodHRwOi8vY3Bz
LmxldHNlbmNyeXB0Lm9yZzCCAQIGCisGAQQB1nkCBAIEgfMEgfAA7gB1AG9Tdqwx
8DEZ2JkApFEV/3cVHBHZAsEAKQaNsgiaN9kTAAABdr6SGvEAAAQDAEYwRAIgL/0c
Q+x3pzOGj8V6ri9SuAq6D3isp2DH2jXaxtVzz5MCIGbiBtVeZVd2X1IBCokHepnX
YeDQsFGAQFytPIhLxllLAHUA9lyUL9F3MCIUVBgIMJRWjuNNExkzv98MLyALzE7x
ZOMAAAF2vpIa2gAABAMARjBEAiA6zpHXvk7WW+EFAcNDwsxUxERrZESge4REj+TN
6hL97wIgXYEsqAFAzvd6/zT2LujrjtYiLXYsnY6lhQQHjb4FE/kwDQYJKoZIhvcN
AQELBQADggEBACPPScgPwbItxm0c7IADO4BZAdAofcZXUNmgdTG7a/N1+c5wbEat
D3Q+2tIwaoo10NLWodncIdazL/zFxCjxlw3V8ms2/Hq4Y+eCelsgtkinll7cMiUO
1g0YwSRhfx1bZG8VLmyp2Mt8tP1t3fo//pKz8///mdfN31FQjP7n3w3EM+eWQNtc
ZkZKOE10irEjZwHs1np49iqlq0y1hw7eOdVylZSBmnKu+EGTLlpq4xKHIwNpi6B8
+Fz4r/BzbcvH64B++rFxjp8qFcHvNAUCMDCASSX5DTn0lOxr6m5wuHUmQ47kwdBZ
5GbDRHdkYFV3FoteCj+ztlIx1iVY/tJRYUQ=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEZTCCA02gAwIBAgIQQAF1BIMUpMghjISpDBbN3zANBgkqhkiG9w0BAQsFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTIwMTAwNzE5MjE0MFoXDTIxMDkyOTE5MjE0MFow
MjELMAkGA1UEBhMCVVMxFjAUBgNVBAoTDUxldCdzIEVuY3J5cHQxCzAJBgNVBAMT
AlIzMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuwIVKMz2oJTTDxLs
jVWSw/iC8ZmmekKIp10mqrUrucVMsa+Oa/l1yKPXD0eUFFU1V4yeqKI5GfWCPEKp
Tm71O8Mu243AsFzzWTjn7c9p8FoLG77AlCQlh/o3cbMT5xys4Zvv2+Q7RVJFlqnB
U840yFLuta7tj95gcOKlVKu2bQ6XpUA0ayvTvGbrZjR8+muLj1cpmfgwF126cm/7
gcWt0oZYPRfH5wm78Sv3htzB2nFd1EbjzK0lwYi8YGd1ZrPxGPeiXOZT/zqItkel
/xMY6pgJdz+dU/nPAeX1pnAXFK9jpP+Zs5Od3FOnBv5IhR2haa4ldbsTzFID9e1R
oYvbFQIDAQABo4IBaDCCAWQwEgYDVR0TAQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8E
BAMCAYYwSwYIKwYBBQUHAQEEPzA9MDsGCCsGAQUFBzAChi9odHRwOi8vYXBwcy5p
ZGVudHJ1c3QuY29tL3Jvb3RzL2RzdHJvb3RjYXgzLnA3YzAfBgNVHSMEGDAWgBTE
p7Gkeyxx+tvhS5B1/8QVYIWJEDBUBgNVHSAETTBLMAgGBmeBDAECATA/BgsrBgEE
AYLfEwEBATAwMC4GCCsGAQUFBwIBFiJodHRwOi8vY3BzLnJvb3QteDEubGV0c2Vu
Y3J5cHQub3JnMDwGA1UdHwQ1MDMwMaAvoC2GK2h0dHA6Ly9jcmwuaWRlbnRydXN0
LmNvbS9EU1RST09UQ0FYM0NSTC5jcmwwHQYDVR0OBBYEFBQusxe3WFbLrlAJQOYf
r52LFMLGMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjANBgkqhkiG9w0B
AQsFAAOCAQEA2UzgyfWEiDcx27sT4rP8i2tiEmxYt0l+PAK3qB8oYevO4C5z70kH
ejWEHx2taPDY/laBL21/WKZuNTYQHHPD5b1tXgHXbnL7KqC401dk5VvCadTQsvd8
S8MXjohyc9z9/G2948kLjmE6Flh9dDYrVYA9x2O+hEPGOaEOa1eePynBgPayvUfL
qjBstzLhWVQLGAkXXmNs+5ZnPBxzDJOLxhF2JIbeQAcH5H0tZrUlo5ZYyOqA7s9p
O5b85o3AM/OJ+CktFBQtfvBhcJVd9wvlwPsk+uyOy2HI7mNxKKgsBTt375teA2Tw
UdHkhVNcsAKX1H7GNNLOEADksd86wuoXvg==
-----END CERTIFICATE-----
`
	key := `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDA4IIJLfsoMxGm
xbfrAG3zSHPgbkcIedBmG+Sk8Td5IZMAdi7Ufhmg9Wl7/jup/WrFbf5jh3MgKfgq
l4FKIOELgvi9T+ESaOuG2lCDL7sBzGqTdaaUn2L99KxOLABqHIgZ3JHUUNbVbOlP
mcIXa/7FVAza9bcc4JrY/gzQR1ztPySLtkG6idzdWIwJzwrK5dDftq5VVTrOcfU2
b8yGKZIUa9/pkpVGb8A0+0yeevDonYhnhi0+HN8EnV9Y1iDkWoj/6Jn8cT+DFjE+
jBv1HIQWLvLw5agP6P7xzU8WV+r2cZwQ63D8tayfC8baedJp9Dz0G5LFbnF/cSvc
/gqpXr+NAgMBAAECggEAWx82EAQrLhjCdBnhBCNVzqQiKpuu73AtZrAy20IixRV5
j7RF95oDnamTxkKcCXlyIggPMRJi74Uke2rMaCsUZw3fWgENAleTPkiR1QiNvxHG
IBhrNLgAWp5ncR8Uqw0Dt8QfGUF/3aDwsOyhZ9Nbr/o/gIqpkfkC7xVHFHdCjcp9
StvzhNeZIq1U/NYULrtnwkOnvV2f2dUEblruP6SzLksr4ZZXuSSC5RmX/U53GP0z
zOJj5pR2AJzjYjE4NuFxn4xtgj4eYSoGK0evyij01HFFHxjLQtLbeGo19WIXmJjA
0BwzJm7FuwbrhPzd49O66isRATXoA95/DAmQzha+OQKBgQD9DHQZ/DAvM1Xyw60v
LyPdj2+KZFPeU9P/bNzf+ZiJduKyRZWrFpnGpvyuSY3lje39HrUDJfpozO34H4BW
mrwe6uHfkDgAAvDaHzpt0mZ69XTCpGEjTtFQBXR+CIHJoAv8o11KewJNoOScAw/l
ElyT8S+ZCb6ZKWOGzRClNa7stwKBgQDDIGU1k/VRs/HVZh8SQSh42QWwdZU4PFVy
ITjcnqqCag4c6n9PXF3N846f/pXtYNW0HLJEmBcc7boea/sq18CCHiulLzlTWaVS
LwejTjZRHaJYxud8aBN7CbwZG0N6MsqMnnHtzFBEXu9m73zSKsgr/0A4Acbz67QX
tGSu06a52wKBgHkvr6KKLiFMuoqqv5PrRYfkG4zxg2DkUJDw986j4DNlJiguPwFS
r459hmGJhFU9ZY5lWFcLpyLtkcHUhEf1jsZXwpionskSn3o2nmrd6opUZviYdJTO
OFvUYPfC5zVCWrtBGXqD8pRuy00UAla4NnH7fcoS6p67PZjfOGuGjCF9AoGAWkFf
zzqTHKmpUNYdxSnSeKOZ2BdrYEm4FER9sr7Ji+1WfdWR8bl9wkfITwVJgDVsZBVp
+ASJnF3x2ySDVzvY1dbyxUNktsMejzclx0nkIf0dHQdUB910NVM5aDuOKLXZrtWT
STVaY2WuQuS/zc7wLDmzELTxu93ovZY5hAxucEUCgYBp08momp1VsxY+UVwRt/in
on1VquDXH0oinGHn9tMGNrvTb0M7u1UxXldP3KiIP35QZbumZK/kiFod3p6fJ+OU
nzvOGfUJga8KRGJAAenaKpxCw4S9RASrDoilCtlWDM4dBneB4daj4NoT0WNkSmCY
/kRDA9tgjNCiPbGaoIwTiw==
-----END PRIVATE KEY-----
`

	crtFile, err := os.Create(filepath.Join(nginxCert, "utm.crt"))
	if err != nil {
		return err
	}
	defer crtFile.Close()
	crtFile.WriteString(cert)

	keyFile, err := os.Create(filepath.Join(nginxCert, "utm.key"))
	if err != nil {
		return err
	}
	defer keyFile.Close()
	keyFile.WriteString(key)

	return nil
}

func initializeElastic() error {
	// wait for elastic to be ready
	baseURL := "http://127.0.0.1:9200/"
	for {
		time.Sleep(50 * time.Second)

		_, err := grequests.Get(baseURL+"_cluster/healt", &grequests.RequestOptions{
			Params: map[string]string{
				"wait_for_status": "green",
				"timeout":         "50s",
			},
		})

		if err == nil {
			break
		}
	}

	_, err := grequests.Put(baseURL+"_snapshot/utm_geoip", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"type": "fs",
			"settings": map[string]interface{}{
				"location": "utm-geoip",
			},
		},
	})
	if err != nil {
		return err
	}

	// restore geoip snapshot
	_, err = grequests.Post(baseURL+"_snapshot/utm_geoip/utm-geoip/_restore?wait_for_completion=false", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"indices":              "utm-*",
			"ignore_unavailable":   true,
			"include_global_state": false,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func initializePostgres(dbPassword string, clientName string, clientDomain string,
	clientPrefix string, clientMail string) error {
	// Connecting to PostgreSQL
	psqlconn := fmt.Sprintf("host=localhost port=5432 user=postgres password=%s sslmode=disable",
		dbPassword)
	srv, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return err
	}

	// Close connection when finish
	defer srv.Close()

	// Check connection status
	err = srv.Ping()
	if err != nil {
		return err
	}

	// Crating utmstack
	_, err = srv.Exec("CREATE DATABASE utmstack")
	if err != nil {
		return err
	}

	// Connecting to utmstack
	psqlconn = fmt.Sprintf("host=localhost port=5432 user=postgres password=%s sslmode=disable database=utmstack",
		dbPassword)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return err
	}

	// Close connection when finish
	defer db.Close()

	// Check connection status
	err = db.Ping()
	if err != nil {
		return err
	}

	// Creating utm_client
	_, err = db.Exec(`CREATE TABLE public.utm_client (		
	id serial NOT NULL,
	client_name varchar(100) NULL,
	client_domain varchar(100) NULL,
	client_prefix varchar(10) NULL,
	client_mail varchar(100) NULL,
	client_user varchar(50) NULL,
	client_pass varchar(50) NULL,
	client_licence_creation timestamp(0) NULL,
	client_licence_expire timestamp(0) NULL,
	client_licence_id varchar(100) NULL,
	client_licence_verified bool NOT NULL,
	CONSTRAINT utm_client_pkey PRIMARY KEY (id)
	);`)
	if err != nil {
		return err
	}

	// Insert client data
	_, err = db.Exec(`INSERT INTO public.utm_client (
	client_name, client_domain, client_prefix, 
	client_mail, client_user, client_pass, client_licence_verified
	) VALUES ($1, $2, $3, $4, 'admin', $5, false);`,
		clientName, clientDomain, clientPrefix, clientMail, dbPassword)
	if err != nil {
		return err
	}

	return nil
}

func installDocker(mode string) error {
	// set map_max_count size to 262144 and disable IPv6
	sysctl := []string{
		"vm.max_map_count=262144",
		"net.ipv6.conf.all.disable_ipv6=1",
		"net.ipv6.conf.default.disable_ipv6=1",
	}
	f, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, config := range sysctl {
		if err := runCmd(mode, "sysctl", "-w", config); err != nil {
			return errors.New("failed to set sysctl config")
		}
		f.WriteString(config + "\n")
	}

	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	if err := runEnvCmd(mode, env, "apt", "update"); err != nil {
		return err
	}
	if err := runEnvCmd(mode, env, "apt", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg-agent", "software-properties-common"); err != nil {
		return err
	}
	if err := runEnvCmd(mode, env, "sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -"); err != nil {
		return err
	}
	if err := runEnvCmd(mode, env, "sh", "-c", `add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"`); err != nil {
		return err
	}
	if err := runEnvCmd(mode, env, "apt", "update"); err != nil {
		return err
	}
	if err := runEnvCmd(mode, env, "apt", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io"); err != nil {
		return err
	}

	time.Sleep(20 * time.Second)

	return nil
}
