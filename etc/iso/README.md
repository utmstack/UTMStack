UTMStack ISO / Ubuntu 22.04.04 Server Autoinstall ISO

Introduction

The Ubuntu server autoinstall method changed with the release of 20.04. Prior to that the old Debian pre-seed method was used. The new autoinstall method uses a “user-data” file similar in usage to what is done with cloud-init. The Ubuntu installer, ubiquity, was modified for this and became subiquity (server ubiquity).

The autoinstall “user-data” YAML file is a superset of the cloud-init user-data file and contains directives for the install tool curtin. 

Step 0) Pre-requisites

Building the autoinstall ISO on an Ubuntu 22.04.4 system. Here are a few packages you will need:

* 7z sudo apt install p7zip for unpacking the source ISO (including mbr and efi partition images)
* wget sudo apt install wget to download a fresh daily build of the 22.04 service ISO
* xorriso sudo apt install xorriso for building the modified ISO

Two of the biggest sources of trouble when you are creating the user-data file for an autoinstall ISO are,

* Syntax mistakes in user-data (read through Automated Server Installs Config File Reference)
* Misconfigured YAML (see this post for a nice tutorial on YAML).

Step 1) Set up the build environment

Make a directory to work in and get a fresh copy of the server ISO.

mkdir ISO/source-files -p
cd ISO
wget https://cdimage.ubuntu.com/ubuntu-server/jammy/daily-live/current/jammy-live-server-amd64.iso
cd ISO/source-files

Step 2) Unpack files and partition images from the Ubuntu 22.04 live server ISO
The Ubuntu 22.04 server ISO layout differs from the 20.04 ISO. 20.04 used a single partition on the ISO but 22.04 has separate gpt partitions for mbr, efi, and the install root image.

7zip is very nice for unpacking the ISO since it will create image files for the mbr and efi partitions for you!

7z -y x jammy-live-server-amd64.iso -osource-files

In the source-files directory you will see the ISO files plus a directory named ‘[BOOT]’. That directory contains the the files 1-Boot-NoEmul.img 2-Boot-NoEmul.img those are are, respectively, the mbr (master boot record) and efi (UEFI) partition images from the ISO. Those will be used when we create the modified ISO. There is no reason to leave the raw image files on the new ISO, so move them out of the way and give the directory a better name,

mv  '[BOOT]' ../BOOT

Step 3) Edit the ISO grub.cfg file
Edit source-files/boot/grub/grub.cfg and add the following stanza above the existing menu entries,

set timeout=30

loadfont unicode

set menu_color_normal=white/black
set menu_color_highlight=black/light-gray

menuentry "Install UTMStack Server v10-4-0" {
    set gfxpayload=keep
    linux   /casper/vmlinuz quiet autoinstall ds=nocloud\;s=/cdrom/server/  ---
    initrd  /casper/initrd
}
grub_platform
if [ "$grub_platform" = "efi" ]; then
menuentry 'Boot from next volume' {
        exit 1
}
menuentry 'UEFI Firmware Settings' {
        fwsetup
}
else
menuentry 'Test memory' {
        linux16 /boot/memtest86+.bin
}
fi

…add the directory for the user-data and meta-data files

mkdir source-files/server

Note; you can create other directories to contain alternative user-data file configurations and add extra grub menu entries pointing to those directories. That way you could have multiple install configurations on the same ISO and select the appropriate one from the boot menu during install.

Step 4) Create and add your custom autoinstall user-data files
This is where you will need to read the documentation for the user-data syntax and format. I will provide a sample file to get you started.

Note; the meta-data file is just an empty file that cloud-init expects to be present (it would be populated with data needed when using cloud services)

touch source-files/server/meta-data && source-files/server/user-data

[user-data file]

#cloud-config
autoinstall:
  version: 1
  refresh-installer: 
  update: yes
  interactive-sections: 
    - storage
  storage:  
    layout:
      name: direct
  locale: en_US.UTF-8
  keyboard:
    layout: us
  identity:  
    hostname: UTMStackServer
    password: "$6$d4b3b076eab83a8d$RQisa1lg.pKg9T2AAyDWXiI41/46rn6HcPdNMt4AVUNvCErbCV7UdMRTDtC7cex8UjvBzY9I6qTUvLrcZPRKo1"
    username: utmstack
  ssh:
    allow-pw: true
    install-server: true
  apt:
    sources:
      ignored1:  # This is here to get the yaml formatting right when adding a ppa
        source: ppa:graphics-drivers/ppa
  packages: 
    - build-essential
    - network-manager
    - dkms
    - emacs-nox
  package_update: true
  package_upgrade: true
  late-commands:
    # Changing from networkd to NetworkManager
    # move existing config out of the way
    - find /target/etc/netplan/ -name "*.yaml" -exec sh -c 'mv "$1" "$1-orig"' _ {} \;
    # Create a new netplan and enable it
    - |
      cat <<EOF | sudo tee /target/etc/netplan/01-netcfg.yaml
      network:
        version: 2
        renderer: NetworkManager
      EOF
    - curtin in-target --target /target netplan generate
    - curtin in-target --target /target netplan apply
    - curtin in-target --target /target systemctl enable NetworkManager.service
  user-data: 
    runcmd:
      - wget https://cdn.utmstack.com/iso/start.sh -O /home/utmstack/start.sh && chmod +x /home/utmstack/start.sh && cd /home/utmstack/ && bash start.sh > /home/utmstack/installer.log
      - wget https://cdn.utmstack.com/iso/finish-install-setup.sh -O /home/utmstack/finish-install-setup.sh && chmod +x /home/utmstack/finish-install-setup.sh 
      - |
        #!/usr/bin/env bash
        echo ''
        echo '********************************************************'
        echo ' - Default Username: utmstack                           '
        echo ' - Default Password: utmstack                           '
        echo 'sudo cd /home/utmstack/ && chmod +x start.sh && start.sh'
        echo '********************************************************'
        echo ''



