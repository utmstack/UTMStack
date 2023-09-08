# UTMStack Installer

## Recommendations

### Operating System

This installation guide provides instructions to perform the UTMStack installation on Ubuntu 22.04 LTS.

### MINIMUM RECOMMENDED RESOURCES
- For each 100 GB of live logs: 4 Cores, 8 GB RAM, 256 GB Disk Space

## Installation steps

### Preparing for installation

- Update packages list: `sudo apt update`
- Install WGET: `sudo apt install wget`
- Download the latest version of the installer by typing `wget http://github.com/AtlasInsideCorp/UTMStackInstaller/releases/latest/download/installer`
- Change to root user: `sudo su`
- Set execution permissions with `chmod +x installer`

### Running installation

- Execute the installer without parameters: `./installer`

Once UTMStack is installed, use admin as the user and the password generated during the installation for the default user to login. You can found the password and other generated configurations in /root/utmstack.yml
Note: Use HTTPS in front of your server name or IP to access the login page.

### Required ports
- 22/TCP Secure Shell (We recommend to create a firewall rule to allow it only from admin workstation)
- 80/TCP UTMStack Web-based Graphical User Interface Redirector (We recommend to create a firewall rule to allow it only from admin and security analyst workstations)
- 443/TCP UTMStack Web-based Graphical User Interface (We recommend to create a firewall rule to allow it only from admin and security analyst workstations)
- 9090/TCP Cockpit Web-based Graphical Interface for Servers (We recommend to create a firewall rule to allow it only from admin workstation)
- Others ports will be required during the configuration of UTMStack's integrations in order to receive logs. (Please follow the security recommendations given on the integration guide if exists)