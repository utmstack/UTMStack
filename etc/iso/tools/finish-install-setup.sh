#!/bin/bash
echo *************************
echo ****  Finish Setup   ****
echo *************************
echo 'Enter the hostname for this system: '
read NEW_HOSTNAME
hostnamectl set-hostname \${NEW_HOSTNAME}
echo
echo 'Enter the timezone for this system: '
echo 'America/Los_Angeles America/Denver America/Chicago America/New_York'
read NEW_TIMEZONE
timedatectl set-timezone \${NEW_TIMEZONE}
echo *************************
echo
echo *************************
echo 'Restarting to finish ...'
shutdown -r 3
