# UTMStack Installer

![Go](https://github.com/UTMStack/installer/workflows/Go/badge.svg)
[![CodeQL](https://github.com/UTMStack/installer/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/UTMStack/installer/actions/workflows/codeql-analysis.yml)
[![Quality Gate Status](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=alert_status)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Bugs](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=bugs)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Vulnerabilities](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=vulnerabilities)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Security Rating](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=security_rating)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Reliability Rating](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=reliability_rating)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Maintainability Rating](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=sqale_rating)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Lines of Code](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=ncloc)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Code Smells](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=code_smells)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Duplicated Lines (%)](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=duplicated_lines_density)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)
[![Technical Debt](https://qube.atlasinside.com/api/project_badges/measure?project=utmstack_installer&metric=sqale_index)](https://qube.atlasinside.com/dashboard?id=utmstack_installer)


## Recommendations

### Operating System
This installation guide provides instructions to perform the UTMStack installation on Linux 18.04 LTS operating system.

### Resources for Master
- For 100 GB of live logs: 4 Cores, 16 GB RAM, 280GB Disk Space
- For 1000 GB of live logs: 16 Cores, 64 GB RAM, 2080GB Disk Space
- For 10000 GB of data and 1000GB of live logs: 16 Cores, 64 GB RAM, 10080 Disk Space
- For 10000GB of live logs: A cluster of 10 servers with 16 Cores, 64 GB RAM, 2080GB Disk Space

### Resources for Probe or Proxy server
- 50 GB Disk Space for system.
- 1 Core per 10 agents or devices sending logs.
- 2 GB of RAM per 10 agents or devices sending logs.
- 30 GB of disk space per 10 agents or devices sending logs.
(The master can act as probe if a proxy is not used)

## Requirements
### Firewall rules
- Open the next input ports from probes to master:

  50051/TCP and 5044/TCP (Used to send logs)

  5432/TCP and 9200/TCP (Used to data access. This ports must be closed from any other source, for security, only probes can reach this ports)

- Open the next input ports from master to probes:

  9390/TCP and 8888/TCP (Used to connect with the vulnerabilities scanner)

  5000/TCP and 8000/TCP (Used to connect with the assets discovery service)

## Installation steps

### Preparing for installation
- Download the latest version from https://github.com/UTMStack/installer/releases (You can use `# wget [URL]` to download the installer directly to the server)
- Set execution permissions with `# chmod +x installer`

### Install using Terminal User Interface
- Execute the installer without parameters: `# ./installer`

### Install using the parameterized mode
You can replace the markups of the next examples by real values in order to use the parameterized mode to install UTMStack Master or Probe.
- Master:
`# ./installer master --datadir "[/example/dir]" --db-pass "[ExAmPlEpaSsWoRd]" --fqdn "[server.example.domain]" --customer-name "[Your Business]" --customer-email "[your@email.com]"`

- Probe:
`# ./installer probe --datadir "[/example/dir]" --db-pass "[Master's DB password]" --host "[Master's IP or FQDN]"`

Once a UTMStack master server is installed, use admin admin as the default first time login user and password.
