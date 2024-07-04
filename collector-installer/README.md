# Collector-Installer

The Collector-Installer is a command-line tool designed for the installation, running, and uninstallation of collectors. It provides a streamlined process for managing collectors, ensuring they are correctly installed and operational within your environment.

### Features
- Install Collectors: Easily install collectors by specifying the collector type, IP address, and a unique connection key.
- Run Collectors: Start the collector process to begin data collection immediately after installation or at any given time.
- Uninstall Collectors: Remove collectors cleanly from your system when they are no longer needed.

### Requirements
Access to the target system's IP address and open ports required by the collector (default: 9000 for the Agent Manager and 50051 for the Log Auth Proxy).

### Supported Collectors
Currently, the Collector-Installer supports the following collector:
- UTMStack AS400 Collector 
