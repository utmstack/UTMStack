# UTMStack Installer

## Recommendations

### Operating System

This installation guide provides instructions to perform the UTMStack installation on Ubuntu 20.04 LTS and 22.04 LTS.

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

| :exclamation: Demo Environment
|-----------------------------------------|

To see a fully operating UTMStack environment access our demo at: <https://utmstack.com/demo>

### Required ports
- 22/TCP Secure Shell (Only required for admin access)
- 80/TCP UTMStack Web-based Graphical User Interface Redirector (Only required for analyst/admin access)
- 443/TCP UTMStack Web-based Graphical User Interface (Only required for analyst/admin access)
- 9090/TCP Cockpit Web-based Graphical Interface for Servers (Only required for admin access)
- Others ports will be required during the configuration of UTMStack's integrations. (Will be listed on the integration guide)

## Why UTMStack

[![Alt text](https://img.youtube.com/vi/wv87dj15G5k/0.jpg)](https://youtu.be/wv87dj15G5k)
