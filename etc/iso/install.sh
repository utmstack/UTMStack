#!/bin/bash

if [ -f "/usr/local/bin/utmstack" ]; then
  rm /usr/local/bin/utmstack
fi

wget -O /usr/local/bin/utmstack https://github.com/utmstack/UTMStack/releases/latest/download/installer
chmod 755 /usr/local/bin/utmstack
utmstack