# OpenVPN deployment guide

## Setup OpenVPN Server on Ubuntu 20.04

* First, run the apt command to apply security updates:

`sudo apt update`

`sudo apt upgrade`

* Find and note down your external IP address with command:

`dig +short myip.opendns.com @resolver1.opendns.com`

or

`curl ifconfig.me`

### A note about IP address assigned to your server

Most cloud and bare-metal servers have two types of IP address provided by the ISP:

Public static IP address directly assigned to your box and routed from the Internet. For example, Linode, Digital Ocean, and others give you direct public IP address.
Private static IP address directly attached to your server and your server is behind NAT with public IP address. For example, AWS EC2/Lightsail give you this kind of NAT public IP address.
The script will automatically detect your networking setup. All you have to do is provide a correct IP address when asked for it.

* Download and run openvpn-install.sh script:

`wget https://git.io/vpn -O openvpn-ubuntu-install.sh`

`chmod -v +x openvpn-ubuntu-install.sh`

`sudo ./openvpn-ubuntu-install.sh`

`sudo systemctl restart openvpn-server@server.service`

## Setup OpenVPN client on Ubuntu 20.04

* First, run the apt command to apply security updates:

`sudo apt update`

`sudo apt upgrade`

* Install OpenVPN with command:

`sudo apt install openvpn`

* Upload the client config to the probe to dir /etc/openvpn/ and replace the extension from .ovpn to .conf

* Restart OpenVPN with command:

`sudo systemctl restart openvpn`
