#!/bin/bash
cd /home/utmstack/
rm installer
wget http://github.com/utmstack/UTMStack/releases/latest/download/installer
chmod +x installer
./installer
