```markdown
# UTMStack / Server Autoinstall ISO

## Introduction

With the release of Ubuntu 20.04, the server autoinstall method changed. Previously, the Debian pre-seed method 
was used. The new autoinstall method uses a "user-data" file, similar to what is done with cloud-init. 
The Ubuntu installer, ubiquity, was modified for this and became subiquity (server ubiquity).

The autoinstall "user-data" YAML file is a superset of the cloud-init user-data file and contains directives 
for the install tool curtin.

## Step 0: Pre-requisites

To build the autoinstall ISO on an Ubuntu 22.04.4 system, you will need the following packages:

- **7z**: `sudo apt install p7zip` for unpacking the source ISO (including mbr and efi partition images)
- **wget**: `sudo apt install wget` to download a fresh daily build of the 22.04 service ISO
- **xorriso**: `sudo apt install xorriso` for building the modified ISO

Two of the biggest sources of trouble when creating the user-data file for an autoinstall ISO are:

- Syntax mistakes in user-data.
- Misconfigured YAML.

## Step 1: Set up the build environment, create a directory to work in and get a fresh copy of the server ISO:

mkdir -p ISO/source-files
cd ISO
wget https://cdimage.ubuntu.com/ubuntu-server/jammy/daily-live/current/jammy-live-server-amd64.iso
cd ISO/source-files

## Step 2: Unpack files and partition images from the Ubuntu 22.04 live server ISO

The Ubuntu 22.04 server ISO layout differs from the 20.04 ISO. 20.04 used a single partition on the ISO but 22.04 
has separate gpt partitions for mbr, efi, and the install root image, 7zip is very useful for unpacking the ISO 
since it will create image files for the mbr and efi partitions for you:

7z -y x jammy-live-server-amd64.iso -osource-files

In the source-files directory, you will see the ISO files plus a directory named '[BOOT]'. This directory contains 
the files `1-Boot-NoEmul.img` and `2-Boot-NoEmul.img` which are, respectively, the mbr (master boot record) and 
efi (UEFI) partition images from the ISO. These will be used when we create the modified ISO. There is no reason to
leave the raw image files on the new ISO, so move them out of the way and give the directory a better name:

mv '[BOOT]' ../BOOT

## Step 3: Edit the ISO grub.cfg file

Copy the `grub.cfg` to `source-files/boot/grub/grub.cfg`:

...add the directory for the user-data and meta-data files

mkdir -p ISO/source-files/server

Note; you can create other directories to contain alternative user-data file configurations and add extra grub menu 
entries pointing to those directories. That way you could have multiple install configurations on the same ISO and 
select the appropriate one from the boot menu during install.

## Step 4: Create and add your custom autoinstall user-data files

This is where you will need to read the documentation for the user-data syntax and format. I will provide a sample 
file to get you started.

`Note; the meta-data file is just an empty file that cloud-init expects to be present (it would be populated with data 
needed when using cloud services)`

touch source-files/server/meta-data && source-files/server/user-data

## Step 5: Generate a new Ubuntu 22.04 server autoinstall ISO

The following command is helpful when trying to setup the arguments for building an ISO. It will give flags and data 
to closely reproduce the source base install ISO.

./iso-build.sh

The partition images that 7z extracted for us are being added back to the recreated ISO.

