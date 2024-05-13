UTMStack ISO / Ubuntu 22.04.04 Server Autoinstall ISO

Introduction

The Ubuntu server autoinstall method changed with the release of 20.04. Prior to that the old Debian pre-seed method was used. The new autoinstall method uses a “user-data” file similar in usage to what is done with cloud-init. The Ubuntu installer, ubiquity, was modified for this and became subiquity (server ubiquity).

The autoinstall “user-data” YAML file is a superset of the cloud-init user-data file and contains directives for the install tool curtin. 

Step 0) Pre-requisites

Building the autoinstall ISO on an Ubuntu 22.04.04 system. Here are a few packages you will need:

* 7z sudo apt install p7zip for unpacking the source ISO (including mbr and efi partition images)
* wget sudo apt install wget to download a fresh daily build of the 22.04 service ISO
* xorriso sudo apt install xorriso for building the modified ISO

Two of the biggest sources of trouble when you are creating the user-data file for an autoinstall ISO are,

* Syntax mistakes in user-data (read through Automated Server Installs Config File Reference)
* Misconfigured YAML (see this post for a nice tutorial on YAML).

Step 1) Set up the build environment

Make a directory to work in and get a fresh copy of the server ISO.

