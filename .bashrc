if [ ! -f "/done.txt" ]; then
    sudo chmod 777 /usr/local/bin/utmstack
    sudo -u root utmstack && sudo -u root touch /done.txt
fi