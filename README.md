# UTMStack Installer

## Recommendations

### Operating System
This installation guide provides instructions to perform the UTMStack installation on Ubuntu 20.04 LTS.
### Resources for Master
| :exclamation:  Minimum Resources Requeriments
|-----------------------------------------|

- MINIMUM REQUERED for non-prod environments: 6 Cores, 8 GB of RAM, 20GB Disk Space (Minimum requered)
- For 100 GB of live logs: 8 Cores, 12 GB RAM, 280GB Disk Space
- For 1000 GB of live logs: 16 Cores, 64 GB RAM, 2080GB Disk Space
- For 10000 GB of data and 1000GB of live logs: 16 Cores, 64 GB RAM, 10080 Disk Space
- For 10000GB of live logs: A cluster of 10 servers with 16 Cores, 64 GB RAM, 2080GB Disk Space

### Resources for Probe or Proxy server
- 50 GB Disk Space for system.
- 4 Cores.
- 8 GB of RAM.
(The master can act as probe if a proxy is not used)

## Requirements
### Firewall rules
- Open the following input ports on the master for access from the probe/proxy.

  50051/TCP and 5044/TCP (Used to send logs)

  5432/TCP and 9200/TCP (Used to data access. This ports must be closed from any other source, for security, only probes can reach this ports)

- Open the following input ports on the probe/proxy for access from the master.

  9390/TCP and 8888/TCP (Used to connect with the vulnerabilities scanner)

  5000/TCP and 8000/TCP (Used to connect with the assets discovery service)

- Open the following input ports for internet access to the master:

  9999/TCP (Used to conntect Zapier to UTMStack)
  
  1194/TCP (Used to connect probe/proxy over the internet using VPN)
  
- Open the following ports from agentless devices (firewalls, hypervisors, etc) to master or probe/proxy:

  2055/UDP (Used to send Netflow packets)
  
  514/UDP (Used to send syslog logs)
  
  514/TCP (Used to send syslog logs)
  
- Open the following ports on the master for agents communication with master or probe/proxy:
  
  5044/TCP (Used to send logs)

  1514-1516/TCP (Used for HIDS agent communications)
  
  1514-1516/UDP (Used for HIDS agent communications)
  
  55000/TCP (Used for HIDS management API)
  
  23949/TCP (Used for connect to the probe API)
  
- Open port 443 for accessing the UTMStack Web console.

## Installation steps

### Preparing for installation
- Update packages list: apt update
- Install WGET and NET-TOOLS: apt install wget net-tools
- Download the latest version from https://github.com/UTMStack/installer/releases (You can use `# wget [URL]` to download the installer directly to the server)
- Set execution permissions with `# chmod +x installer`

### Install using Terminal User Interface
- Execute the installer without parameters: `# ./installer`

### Install using the parameterized mode
You can replace the markups of the next examples by real values in order to use the parameterized mode to install UTMStack Master or Probe.
- Master:
`# ./installer master --datadir "/example/dir" --db-pass "ExAmPlEpaSsWoRd" --fqdn "server.example.domain" --customer-name "Your Business" --customer-email "your@email.com"`

- Probe:
`# ./installer probe --datadir "/example/dir" --db-pass "Master's DB password" --host "Master's IP or FQDN"`

Once a UTMStack master server is installed, use admin admin as the default first time login user and password.
Note: Use HTTPS in front of your server name or IP to access the login page.

| :exclamation: Demo Environment
|-----------------------------------------|

To see a fully operating UTMStack environment access our demo at: https://utmstack.com/demo

Watch this short 10 minutes installation video if you still have questions.

[![Alt text](https://img.youtube.com/vi/dM9dC9HNXUs/0.jpg)](https://youtu.be/dM9dC9HNXUs)

Why UTMStack

[![Alt text](https://img.youtube.com/vi/wv87dj15G5k/0.jpg)](https://youtu.be/wv87dj15G5k)

