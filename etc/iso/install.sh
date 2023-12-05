#!/bin/bash

if [ ! -f "/done.txt" ]; then
    sudo wget -O /usr/local/bin/utmstack https://github.com/utmstack/UTMStack/releases/latest/download/installer
    sudo chmod 777 /usr/local/bin/utmstack
    sudo -u root utmstack && sudo -u root touch /done.txt
fi