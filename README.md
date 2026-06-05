> [!IMPORTANT]
> ## ⚠️ UPGRADE NOTIFICATION and UTMStack v10 End-of-Life Notice
>
> **UTMStack v10 will be sunset on December 5, 2026 (6 months from now).**
>
> As of today, **v10 is no longer receiving any new features** and will only receive critical fixes until the sunset date. After December 5, 2026, v10 will no longer be supported.
>
> 👉 **All users must upgrade to v11 as soon as possible** using the upgrade package. Full guide here: [https://utmstack.com](https://utmstack.com)


---

# UTMStack

<p align="center">
  <a href="https://utmstack.com">
    <img src="https://utmstack.com/wp-content/uploads/2023/02/utmstack-logo-favicon.png?v=2" alt="UTMStack" width="150px">
  </a>
</p>

[![Contributors](https://img.shields.io/github/contributors-anon/utmstack/utmstack)](https://github.com/utmstack/UTMStack/graphs/contributors)
[![Release](https://img.shields.io/github/release/utmstack/utmstack)](https://github.com/utmstack/UTMStack/releases/)
[![Issues](https://img.shields.io/github/issues-raw/utmstack/utmstack)](https://github.com/utmstack/UTMStack/issues)
[![Commit Activity](https://img.shields.io/github/commit-activity/m/utmstack/utmstack)](https://github.com/utmstack/UTMStack/commits/main)
[![License](https://img.shields.io/github/license/utmstack/utmstack?color=blue)](https://github.com/utmstack/UTMStack/blob/master/LICENSE)
[![Discord](https://img.shields.io/discord/1154016563775672400.svg?logo=discord)](https://discord.gg/ZznvZ8xcHh)

#### Enterprise-ready SIEM and XDR powered by Real-Time correlation and Threat Intelligence

## Introduction

Welcome to the UTMStack open-source project! UTMStack is a unified threat management platform that merges SIEM (Security Information and Event Management) and XDR (Extended Detection and Response) technologies. Our unique approach allows real-time correlation of log data, threat intelligence, and malware activity patterns from multiple sources, enabling the identification and halting of complex threats that use stealthy techniques. Visit an [online demo here.](https://utmstack.com/demo)

<table>
  <tr>
    <td width="50%" valign="top">
      <a href="https://utmstack.com">
        <img src="https://utmstack.com/wp-content/uploads/2023/07/dashboard-two.gif?v=2" alt="UTMStack dashboard" width="100%">
      </a>
    </td>
    <td width="50%" valign="top">
      <a href="https://utmstack.com">
        <img src="https://utmstack.com/wp-content/uploads/2023/07/dashboard-one.gif?v=2" alt="UTMStack dashboard" width="100%">
      </a>
    </td>
  </tr>
</table>

## Features

- Log Management and Correlation
- Threat Detection and Response
- Threat Intelligence
- Alert Investigation
- File Classification
- SOC AI-Powered Analysis
- Security Compliance

## Why UTMStack?

UTMStack stands out in threat prevention by surpassing the boundaries of traditional systems. Our software platform can swiftly analyze log data to identify and halt threats at their source in real-time, even if the threat was not directly detected on the server itself. This seamless integration of SIEM and XDR capabilities sets UTMStack apart from competitors, providing organizations with an effective, holistic cybersecurity suite that enhances threat detection, response, and remediation across clients’ valuable digital infrastructure. Correlation happens before data ingestion, reducing workload and improving response times.

## Getting Started

To get started with UTMStack, visit our demo at utmstack.com/demo. You can also watch our videos to learn more about our platform:

- [Official Documentation](https://docs.utmstack.com)
- [Advanced Persistent Threats](https://www.youtube.com/watch?v=Rqbl65cJMuA)
- [UTMStack Features Overview](https://www.youtube.com/watch?v=lKkydWFiu4Y)

## Contributing

We welcome contributions from the community! Whether you're a developer, a security expert, or just someone interested in cybersecurity, your contributions can help make UTMStack even better. Check out our [Contributing Guide](CONTRIBUTING.md) for more information on how you can contribute to this project.

## Security

UTMStack code is reviewed daily for vulnerable dependencies. Penetration testing is performed on the system yearly and after every major release. All data in transit between agents and UTMStack servers is encrypted using TLS. UTMStack services are isolated by containers and microservices with strong authentication. Connections to the UTMStack server are authenticated with a +24 characters unique key. User credentials are encrypted in the database and protected by fail2ban mechanisms and 2FA.

## License

UTMStack is open-source software licensed under the AGPL version 3. For more information, see the [LICENSE](LICENSE) file.

## Contact

If you have any questions or suggestions, feel free to open an issue or submit a pull request. We're always happy to hear from our community!

Join us in making UTMStack the best it can be!

# Installation

## Recommendations

### Operating System

This installation guide provides instructions to perform the UTMStack installation on Ubuntu 22.04 LTS.

### SYSTEM RESOURCES
Assumptions: 60 data sources (devices) generate approximately 100 GB of monthly data.

Definitions:
- Hot log storage: not archived data that can be accessed for analysis at any time.
- Cold log storage: archived data that should be restored before accessing it.
- Data source: any individual source of logs, for example, devices, agents, SaaS integrations.

Required resources for one month of hot log storage.
- For 50 data sources (120 GB) of hot log storage you will need 4 Cores, 16 GB RAM, 150 GB Disk Space
- For 120 data sources (250 GB) of hot log storage you will need 8 Cores, 16 GB RAM, 250 GB Disk Space
- For 240 data sources (500 GB) of hot log storage you will need 16 Cores, 32 GB RAM, 500 GB Disk Space
- For 500 data sources (1000 GB) of hot log storage you will need 32 Cores, 64 GB RAM, 1000 GB Disk Space
- You may combine these tiers to allocate resources based on the number of devices and desired hot log storage retention

IMPORTANT: Going above 500 data sources/devices requires adding secondary nodes for horizontal scaling.

## Installation steps
The installation can be performed using an installer file or an [ISO image](https://utmstack.com/install). The instructions below are only for the installer file option; please skip them if you use the ISO image instead.
> **_NOTE:_** The default Ubuntu Server credentials are; "user: utmstack", "password: utmstack"

### Preparing for installation

- Update packages list: `sudo apt update`
- Install WGET: `sudo apt install wget`
- Download the latest version of the installer by typing `wget http://github.com/utmstack/UTMStack/releases/latest/download/installer`
- Change to root user: `sudo su`
- Set execution permissions with `chmod +x installer`

### Running installation

- Execute the installer without parameters: `./installer`

Once UTMStack is installed, use admin as the user and the password generated during the installation for the default user to login. You can find the password and other generated configurations in /root/utmstack.yml
Note: Use HTTPS in front of your server name or IP to access the login page.

### Required ports
- 22/TCP Secure Shell (We recommend creating a firewall rule to allow it only from admins workstations)
- 80/TCP UTMStack Web-based Graphical User Interface Redirector (We recommend creating a firewall rule to allow it only from admin and security analyst workstations)
- 443/TCP UTMStack Web-based Graphical User Interface (We recommend creating a firewall rule to allow it only from admin and security analyst workstations)
- 9090/TCP Cockpit Web-based Graphical Interface for Servers (We recommend creating a firewall rule to allow it only from admin workstation)
- Others ports will be required during the configuration of UTMStack's integrations to receive logs. (Please follow the security recommendations given on the integration guide if exists)

# FAQ
- Is this based on Grafana, Kibana, or a similar reporting tool?
  Answer: It is not. UTMStack has been built from the ground up to be a simple and intuitive SIEM/XDR.
- Does UTMStack use ELK for log correlation?
  Answer: It does not. UTMStack correlation engine was built from scratch to analyze data before ingestion and maximize real-time correlation.
- What is the difference between the Open Source and Enterprise versions?
  The enterprise version includes features that would typically benefit enterprises and MSPs. For example, support, faster correlation, frequent threat intelligence updates, and Artificial Intelligence.
