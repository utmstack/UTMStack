# UTMStack Correlation Rules - Detection Gap Analysis

**Date:** February 2026
**Scope:** All 692 correlation rules across 24+ technology directories
**Methodology:** 4 parallel analyst agents reviewed every rule file, mapped MITRE ATT&CK coverage, and identified specific missing detections

---

## Executive Summary

| Domain | Current Rules | Gaps Identified | Critical Gaps |
|--------|:------------:|:---------------:|:-------------:|
| Network Security | 172 | 98 | 12 |
| Cloud & SaaS | 97 | 57 | 10 |
| Endpoint Security | ~220 | ~115 | ~30 |
| Application & Infrastructure | ~203 | ~111 | ~28 |
| **TOTAL** | **~692** | **~381** | **~80** |

**Key Findings:**
- **Persistence is the biggest gap** across Windows and Linux -- registry run keys, scheduled tasks, crontab, SSH authorized_keys, WMI subscriptions all lack detection
- **SSRF and SSTI** have zero detection across all web servers (Apache, Nginx, IIS)
- **Google Cloud Platform** has only 8 rules vs. AWS (24) and Azure (18) -- severely undercovered
- **Firewall policy change detection** is missing from most network devices (FortiGate, Palo Alto, pfSense, SonicWall, Sophos XG)
- **Kafka and NATS** message queue technologies have only 2 rules each -- severely undercovered
- **Log source health monitoring** (detecting when sources stop sending) does not exist

---

## Table of Contents

1. [Network Security](#1-network-security)
   - [Cisco ASA](#cisco-asa-10-rules)
   - [Cisco Switch](#cisco-switch-11-rules)
   - [Cisco Firepower](#cisco-firepower-12-rules)
   - [Cisco Meraki](#cisco-meraki-8-rules)
   - [Fortinet FortiGate](#fortinet-fortigate-11-rules)
   - [Fortinet FortiWeb](#fortinet-fortiweb-13-rules)
   - [Palo Alto](#palo-alto-8-rules)
   - [MikroTik](#mikrotik-17-rules)
   - [pfSense](#pfsense-8-rules)
   - [SonicWall](#sonicwall-19-rules)
   - [Sophos](#sophos-24-rules)
   - [NetFlow](#netflow-14-rules)
   - [NIDS/Suricata](#nidssuricata-17-rules)
2. [Cloud & SaaS](#2-cloud--saas)
   - [AWS](#aws-24-rules)
   - [Azure](#azure-18-rules)
   - [Google Cloud Platform](#google-cloud-platform-8-rules)
   - [Office 365](#office-365-27-rules)
   - [GitHub](#github-20-rules)
3. [Endpoint Security](#3-endpoint-security)
   - [Windows](#windows-28-rules)
   - [macOS](#macos-24-rules)
   - [Linux](#linux-33-rules-debian--rhel)
   - [Antivirus](#antivirus-82-rules)
   - [HIDS/OSSEC](#hidsossec-15-rules)
4. [Application & Infrastructure](#4-application--infrastructure)
   - [Apache](#apache-21-rules)
   - [Nginx](#nginx-7-rules)
   - [IIS](#iis-14-rules)
   - [MySQL](#mysql-9-rules)
   - [PostgreSQL](#postgresql-4-rules)
   - [MongoDB](#mongodb-9-rules)
   - [Redis](#redis-5-rules)
   - [Kafka](#kafka-2-rules)
   - [NATS](#nats-2-rules)
   - [Elasticsearch](#elasticsearch-6-rules)
   - [Kibana](#kibana-9-rules)
   - [Logstash](#logstash-6-rules)
   - [HAProxy](#haproxy-10-rules)
   - [Traefik](#traefik-7-rules)
   - [VMware ESXi](#vmware-esxi-21-rules)
   - [IBM AIX](#ibm-aix-18-rules)
   - [IBM AS/400](#ibm-as400-6-rules)
   - [Auditd](#auditd-18-rules)
   - [System Linux Module](#system-linux-module-11-rules)
   - [Osquery](#osquery-6-rules)
   - [Generic](#generic-9-rules)
   - [JSON Input](#json-input-10-rules)
   - [Syslog](#syslog-23-rules)
5. [Cross-Cutting Gaps](#5-cross-cutting-gaps)
6. [Priority Implementation Roadmap](#6-priority-implementation-roadmap)

---

## 1. Network Security

### Cisco ASA (10 rules)

**Current Coverage:** VPN brute force, privilege escalation, config changes, ACL modifications, botnet C2, IPS signatures, threat detection/DDoS, syslog tampering, geographic impossibility travel, WebVPN anomalies.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Failover/HA manipulation | T1489 | High | No rule to detect failover configuration changes or active/standby state manipulation that could disable redundant firewall protection |
| 2 | NAT rule tampering | T1562.004 | High | No detection for unauthorized NAT rule creation/modification (msg IDs 305005-305012) that could expose internal services |
| 3 | VPN split tunneling changes | T1572 | Medium | No detection for changes to split tunnel policies that could route traffic outside the VPN tunnel |
| 4 | Successful admin login from unusual source | T1078 | High | Existing rule only covers failed VPN attempts; no detection for successful admin logins from new/unusual IPs |
| 5 | Crypto/IKE configuration changes | T1562.001 | Medium | No monitoring of IPsec/IKE crypto map changes that could weaken VPN encryption |
| 6 | Interface shutdown/modification | T1489 | High | No detection for interface state changes (shutdown/no shutdown) that could disrupt network connectivity |
| 7 | Denied connection flood | T1498 | Medium | No rule detecting high volumes of denied connections indicating firewall resource exhaustion attempts |

### Cisco Switch (11 rules)

**Current Coverage:** VLAN hopping, MAC spoofing, ARP poisoning, STP attacks, DHCP snooping, DAI failures, IP source guard, 802.1X failures, MAB attempts, SNMP access, port mirroring.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | CAM table overflow (MAC flooding) | T1498 | High | No detection for CAM table overflow attacks that force switches into hub mode, enabling traffic sniffing |
| 2 | CDP/LLDP reconnaissance | T1018 | Medium | No detection for CDP/LLDP information harvesting that reveals network topology |
| 3 | Private VLAN proxy attacks | T1599 | Medium | No detection for attacks targeting private VLAN isolation boundaries |
| 4 | Switch config change detection | T1562.004 | High | No detection for unauthorized switch configuration changes |
| 5 | TACACS+/RADIUS authentication failure | T1110 | Medium | No detection for management authentication failures using TACACS+ or RADIUS |

### Cisco Firepower (12 rules)

**Current Coverage:** AMP malware, IPS high priority, URL filtering, SSL/TLS bypass, file policy violations, DNS security, email security, encrypted visibility, TID alerts, custom rules, IOC detection, retrospective malware.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Access control policy changes | T1562.004 | **Critical** | No detection for policy deployment changes, rule additions/deletions, or security policy modifications in FMC |
| 2 | Connection event anomalies (C2 ports) | T1571 | High | No detection for allowed connections on non-standard ports that might indicate C2 |
| 3 | User identity anomalies | T1078 | High | No detection for user-to-IP mapping failures or conflicting identity data indicating identity spoofing |
| 4 | Geolocation-based anomalies | T1078 | Medium | No rule correlating connection events with geolocation for traffic from unusual countries |

### Cisco Meraki (8 rules)

**Current Coverage:** Rogue SSID, wireless intrusion, rogue AP, AMP alerts, IDS alerts, security appliance events, API anomalies, dashboard auth failures, org admin changes.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Evil twin AP detection | T1557 | **Critical** | No specific rule for evil twin attacks where rogue AP mimics exact corporate SSID |
| 2 | Client VPN brute force | T1110 | High | No detection for failed client VPN authentication attempts on Meraki MX appliances |
| 3 | Firmware/configuration changes | T1562.001 | High | No detection for unauthorized firmware updates or configuration changes |
| 4 | Content filtering bypass | T1090 | Medium | No detection for attempts to bypass Meraki content filtering |

### Fortinet FortiGate (11 rules)

**Current Coverage:** Admin compromise, AV outbreak, app control violations, DLP exfiltration, DNS filtering, email threats, FortiGuard threat feed, FortiSandbox evasion, ICS/SCADA attacks, IPS critical events, sandbox malicious verdict.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Firewall policy changes | T1562.004 | **Critical** | No detection for FortiGate security policy additions, deletions, or modifications |
| 2 | VPN tunnel failures/anomalies | T1133 | **Critical** | No detection for FortiGate SSL-VPN or IPsec VPN authentication failures (very common attack vector) |
| 3 | HA cluster split-brain | T1489 | High | No detection for HA cluster synchronization failures or split-brain conditions |
| 4 | Admin session anomalies | T1078 | High | No detection for admin logins from unusual IPs or concurrent sessions |
| 5 | Web filtering bypass attempts | T1090 | High | No detection for attempts to bypass FortiGuard web filtering |
| 6 | SSL inspection certificate errors | T1573 | Medium | No detection for SSL deep inspection failures indicating encrypted threat evasion |

### Fortinet FortiWeb (13 rules)

**Current Coverage:** API security, auth bypass, bot detection, cookie security, CSRF, custom signatures, DDoS, file upload, geo-blocking, HTTP header injection, IP reputation, ML anomaly detection, OWASP top 10, rate limiting, session management, web app attacks, XML/JSON attacks, zero-day patterns.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Web shell upload detection | T1505.003 | **Critical** | No FortiWeb-specific rule for detecting web shell upload attempts |
| 2 | SQL injection (specific) | T1190 | High | No dedicated SQL injection rule with specific FortiWeb SQLi signature IDs |
| 3 | Server-side request forgery | T1190 | High | No specific SSRF detection rule |
| 4 | WebSocket attacks | T1071.001 | Medium | No detection for WebSocket-specific attacks |

### Palo Alto (8 rules)

**Current Coverage:** WildFire malware, zero-day exploits, file blocking, IOC threat intel, DoS protection, zone protection, GlobalProtect VPN anomalies, Cortex XDR alerts.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Security policy changes | T1562.004 | **Critical** | No detection for PAN-OS security policy additions, modifications, or deletions |
| 2 | URL filtering blocks | T1071.001 | High | No Palo Alto-specific URL filtering detection for malicious/phishing URL categories |
| 3 | Spyware/vulnerability detections | T1190 | High | No rule for THREAT log subtype=spyware or subtype=vulnerability |
| 4 | DNS Security (PDNS) alerts | T1071.004 | High | No Palo Alto DNS Security subscription alerts (DNS sinkholing, DGA detection, DNS tunneling) |
| 5 | Admin login brute force | T1110 | High | No detection for failed management plane login attempts |
| 6 | User-ID mapping anomalies | T1078 | Medium | No detection for User-ID agent failures or authentication policy bypass |
| 7 | SSL decryption failures | T1573 | Medium | No detection for SSL/TLS decryption failures indicating evasion |

### MikroTik (17 rules)

**Current Coverage:** RouterOS brute force, Winbox exploitation, API access, SSH brute force, telnet access, web interface attacks, config export, backup file access, script execution, scheduler abuse, firewall rule bypass, firmware exploits, DHCP server attacks, DNS cache poisoning, hotspot bypass, VLAN hopping, wireless security breach.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | User account creation/modification | T1136 | **Critical** | No detection for creation of new admin users or modification of existing user privileges |
| 2 | SOCKS proxy enablement | T1090 | **Critical** | No detection for enabling SOCKS proxy (commonly used by Meris botnet for proxying attacks) |
| 3 | DNS redirection/static entries | T1584.002 | High | No detection for creation of static DNS entries that redirect traffic to malicious servers |
| 4 | Port forwarding/NAT rule changes | T1090.001 | High | No detection for unauthorized NAT/port forwarding rules |
| 5 | PPP/VPN server configuration changes | T1133 | High | No detection for enabling PPP, L2TP, PPTP, or SSTP VPN servers |
| 6 | Routing table manipulation | T1565 | High | No detection for static route additions that could redirect traffic |

### pfSense (8 rules)

**Current Coverage:** pfBlockerNG blocks, proxy violations, Squid cache poisoning, DNS resolver cache poisoning, Snort/Suricata IDS, VPN auth failures, RADIUS auth failures, captive portal bypass.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Firewall rule changes | T1562.004 | **Critical** | No detection for pfSense firewall rule additions, modifications, or deletions |
| 2 | Admin login brute force | T1110 | High | No detection for failed WebGUI authentication attempts |
| 3 | Package installation/removal | T1059 | High | No detection for pfSense package manager activity |
| 4 | Certificate changes | T1588.004 | Medium | No detection for CA or server certificate changes |

### SonicWall (19 rules)

**Current Coverage:** App control, botnet detection, capture ATP, content filtering, DDoS, DPI-SSL, email security, gateway AV, geo-IP filtering, anti-spyware, IPS alerts, capture client threats, cloud app security, DNS filtering, encrypted threats, endpoint security, real-time blacklist, SonicWave wireless, zero-day threats.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Firewall policy changes | T1562.004 | **Critical** | No detection for access rule additions/modifications/deletions |
| 2 | Admin authentication failures | T1110 | High | No detection for failed management interface login attempts |
| 3 | VPN tunnel failures/anomalies | T1133 | High | No detection for SSL VPN authentication failures |
| 4 | Log tampering/export | T1562.002 | High | No detection for audit log configuration changes |

### Sophos (24 rules)

**Current Coverage:** Admin anomalies, app control, behavioral analysis, DLP, device control, email protection, endpoint threats, exploit prevention, MTR, ML detections, mobile threats, ransomware, server protection, tamper protection, threat cases, web protection, ZTNA, API security (Central). ATP alerts, email protection, synchronized security, user threat quotient, web protection (XG).

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | XG firewall rule changes | T1562.004 | **Critical** | No detection for firewall rule modifications on Sophos XG |
| 2 | XG VPN authentication failures | T1110 | High | No detection for SSL VPN or IPsec VPN auth failures |
| 3 | XG admin login anomalies | T1078 | High | No detection for management console authentication failures |
| 4 | XG IPS signature matches | T1190 | High | No dedicated IPS alert rule for Sophos XG |

### NetFlow (14 rules)

**Current Coverage:** Application layer attacks, asset enumeration, beaconing behavior, data exfiltration, DDoS patterns, amplification attacks, network reconnaissance, port scanning, service discovery, Tor usage, DNS query anomalies.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Lateral movement via SMB/RDP | T1021 | **Critical** | No rule for detecting internal SMB (445) or RDP (3389) scanning/connection patterns indicating lateral movement |
| 2 | Cryptocurrency mining traffic | T1496 | High | No detection for connections to known mining pool ports (3333, 8333, 14444, 45700) |
| 3 | VPN/tunnel to unusual destinations | T1572 | High | No detection for VPN protocol connections to non-corporate destinations |
| 4 | DNS over HTTPS (DoH) detection | T1071.004 | High | No detection for DoH connections bypassing DNS monitoring |
| 5 | Internal network scanning (east-west) | T1046 | High | Existing scanning rules focus on external; no rule for internal host scanning |
| 6 | ICMP tunnel detection | T1095 | Medium | No ICMP tunneling detection |
| 7 | Long-duration connections | T1571 | Medium | No detection for abnormally long-lived connections indicating persistent C2 |

### NIDS/Suricata (17 rules)

**Current Coverage:** Covert channels, exploit attempts, fragmentation attacks, malware callbacks, port scans, service enumeration, signature matches, threat intel IOCs, C2 traffic, DNS tunneling, ICMP tunneling, known bad actors, lateral movement, tunneling, DDoS patterns, evasion techniques, data exfiltration.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | TLS certificate anomalies | T1573 | High | No detection for self-signed, expired, or mismatched certificates indicating MITM or malicious servers |
| 2 | HTTP request smuggling | T1190 | High | No specific detection for HTTP desync attacks |
| 3 | JA3/JA3S fingerprint matching | T1071.001 | High | No comprehensive JA3 threat intelligence correlation |
| 4 | Malicious file downloads (by MIME type) | T1105 | High | No detection for executable file downloads (PE, ELF, scripts) over HTTP |
| 5 | SSH anomalies | T1021.004 | Medium | No detection for SSH brute force or tunneling at the NIDS level |
| 6 | ARP spoofing detection | T1557.002 | Medium | No NIDS-level ARP spoofing detection |

### Network Security: Cross-Technology Systemic Gaps

| Gap | Priority | Description |
|-----|----------|-------------|
| **Firewall policy change detection** | **CRITICAL** | Only Cisco ASA has config change rules. FortiGate, Palo Alto, pfSense, SonicWall, and Sophos XG ALL lack firewall policy change detection. This is the #1 gap. |
| **VPN authentication brute force** | **HIGH** | Only Cisco ASA, pfSense, and Palo Alto have VPN auth failure rules. FortiGate, SonicWall, Sophos XG, and Meraki lack this. |
| **Admin login monitoring** | **HIGH** | Most firewalls lack management plane authentication monitoring. |
| **HA/Failover manipulation** | **MEDIUM** | No technology has HA state change detection. |
| **Log/audit tampering** | **HIGH** | Only Cisco ASA has syslog modification detection. All other firewalls lack this. |

---

## 2. Cloud & SaaS

### AWS (24 rules)

**Current Coverage:** Root account usage, IAM backdoor, S3 public exposure, EBS snapshot sharing, Lambda privilege escalation, cross-account access, IMDS abuse, Secrets Manager, CloudTrail disable, CloudWatch alarm deletion, security group/network ACL/VPC flow log changes, mass deletion, CloudFormation stack deletion, unusual API patterns, SSO activities, Route53 hijacking, GuardDuty findings, Macie alerts, Config compliance, Control Tower guardrails, cost anomaly, reserved instance mods, support case anomalies, billing alarms, KMS key mods, RDS security group changes.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | EC2 Cryptomining Detection | T1496 | **Critical** | No rule to detect EC2 instance launches with GPU/high-compute types by unusual principals or in unusual regions |
| 2 | Console Login Impossible Travel | T1078.004 | **Critical** | No rule correlating ConsoleLogin events from geographically impossible locations |
| 3 | S3 Bulk Data Exfiltration | T1530 | High | No detection for bulk GetObject operations from unusual IPs or roles |
| 4 | CloudTrail Event Selector Manipulation | T1562.008 | High | PutEventSelectors can selectively exclude data events while leaving trail running |
| 5 | IAM Privilege Escalation Paths | T1098 | High | Missing iam:PassRole abuse, iam:CreatePolicyVersion, sts:AssumeRole chaining, ssm:SendCommand |
| 6 | AWS Organization Modifications | T1098 | High | No detection for CreateAccount, InviteAccountToOrganization, RemoveAccountFromOrganization |
| 7 | RDS/Redshift Snapshot Exfiltration | T1537 | High | Only EBS snapshots covered; RDS snapshot sharing is missing |
| 8 | EC2 Instance Connect / SSM Abuse | T1021 | Medium | No detection for SendSSHPublicKey, StartSession, or SendCommand |
| 9 | ECS/EKS Container Abuse | T1610 | Medium | No rules for suspicious container task definitions, privileged containers, or container escape |
| 10 | AWS Backup Deletion | T1490 | Medium | No rule for DeleteBackupVault, DeleteRecoveryPoint (ransomware precursor) |
| 11 | WAF/Shield Rule Modifications | T1562.001 | Medium | No detection for DeleteWebACL, UpdateWebACL removing rules |
| 12 | API Gateway/Lambda URL Backdoor | T1059 | Low | No detection for CreateFunctionUrlConfig or CreateRestApi creating backdoor endpoints |

### Azure (18 rules)

**Current Coverage:** Privilege escalation, conditional access bypass, subscription ownership, service principal abuse, WAF alerts, container registry vulns, resource group mass mods, Key Vault access, storage account public access, MFA disabled, NSG changes, Sentinel alerts, Defender alerts, VM suspicious activities, ExpressRoute changes, Function Apps security, SQL database firewall, privilege escalation attempts.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Azure AD Impossible Travel / Risky Sign-ins | T1078 | **Critical** | No rule for impossible travel, unfamiliar sign-in properties, or sign-ins from anonymizing services |
| 2 | Managed Identity Abuse | T1078.004 | **Critical** | No detection for tokens acquired from IMDS by unauthorized processes |
| 3 | App Registration Abuse | T1098.001 | High | No rule for new app registrations with high-privilege API permissions (Mail.ReadWrite, Directory.ReadWrite.All) |
| 4 | Disk/Snapshot Exfiltration | T1537 | High | No rule for sharing managed disk snapshots across subscriptions/tenants |
| 5 | Azure AD Password Spray | T1110 | High | No rule for failed sign-ins across multiple usernames from same IP |
| 6 | AKS Security | T1610 | High | No rules for privileged pod creation, container exec, or container escape attempts |
| 7 | Golden SAML / Federation Abuse | T1606.002 | High | No detection for federation domain additions (SolarWinds/NOBELIUM technique) |
| 8 | Automation Runbook Abuse | T1059 | Medium | No detection for runbook creation/modification with high-privilege managed identities |
| 9 | PIM Abuse | T1078 | Medium | No detection for unusual PIM role activation patterns |
| 10 | Diagnostic Settings Tampering | T1562.008 | Medium | No rule for deletion of diagnostic settings (equivalent to disabling CloudTrail) |
| 11 | Azure AD Directory Role Changes (specific) | T1098 | Medium | No specific detection for Global Admin or Privileged Role Admin assignments |

### Google Cloud Platform (8 rules)

**Current Coverage:** IAM policy modifications, organization policy violations, cloud asset inventory, binary authorization, Anthos security, VPC firewall rules, cloud identity sign-ins, cloud load balancing anomalies.

**NOTE: GCP has the most severe coverage deficit with only 8 rules. Should triple in size at minimum.**

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Compute Engine Cryptomining | T1496 | **Critical** | #1 threat to GCP accounts; no rule for high-CPU/GPU instance launches in unusual regions |
| 2 | Cloud Audit Log Disabling | T1562.008 | **Critical** | No rule for modifying audit log sinks, exclusion filters, or disabling Data Access audit logs |
| 3 | Service Account Impersonation | T1550.001 | **Critical** | No detection for GenerateAccessToken, GenerateIdToken, or SignBlob by unusual principals |
| 4 | Cloud Storage Exfiltration | T1530 | High | No rule for bulk downloads, bucket policy changes to public, or signed URL creation |
| 5 | Cloud Function/Cloud Run Abuse | T1059 | High | No rules for serverless compute abuse for persistence/execution |
| 6 | Project Creation/Deletion | T1578 | High | No detection for shadow project creation or project deletion |
| 7 | Secret Manager Access Patterns | T1552 | High | No rule for bulk SecretVersionAccess operations |
| 8 | VPC Network Peering Changes | T1562.004 | Medium | No detection for network peering connections bridging network boundaries |
| 9 | BigQuery Data Exfiltration | T1567 | Medium | No detection for large data exports or table copy to external projects |
| 10 | KMS Key Modifications | T1552 | Medium | No detection for DisableCryptoKeyVersion or DestroyCryptoKeyVersion |
| 11 | Workload Identity Federation Abuse | T1078 | Medium | No detection for external identity binding without service account keys |
| 12 | Custom Role Creation | T1098 | Medium | No detection for roles with overly permissive permissions |

### Office 365 (27 rules)

**Current Coverage:** Conditional access bypass, app consent grant, Exchange admin changes, email forwarding/inbox rules, mass email deletion, Power Automate abuse, Safe Links violations, SharePoint mass downloads, OAuth anomalies, OneDrive mass access, Power Apps exposure, anti-phishing/compliance/Safe Attachments disable, mail flow rule manipulation, Safe Attachment violations, external sharing, audit log tampering, eDiscovery abuse.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Mailbox Delegation Abuse | T1098.002 | **Critical** | No rule for Add-MailboxPermission (FullAccess, SendAs, SendOnBehalf) -- primary BEC technique |
| 2 | Impossible Travel Detection | T1078 | **Critical** | No dedicated impossible travel detection for O365 sign-ins |
| 3 | MFA Fatigue / Push Spam | T1621 | High | No detection for multiple MFA prompts in short succession (Uber/Cisco breach technique) |
| 4 | Teams External User Abuse | T1566 | High | No detection for external user phishing via Teams channels |
| 5 | SharePoint Site Permission Escalation | T1098 | High | No rule for SiteCollectionAdminAdded or granting site admin to external users |
| 6 | Service Principal Credential Addition | T1098.001 | High | No detection for adding credentials to existing app registrations |
| 7 | Mailbox Export (PST) | T1114.002 | Medium | No specific detection for New-MailboxExportRequest |
| 8 | Power BI Export | T1567 | Medium | No rules for Power BI workload at all |
| 9 | Forms/Sway Phishing | T1566 | Medium | No detection for Microsoft Forms or Sway phishing pages |
| 10 | Admin Role Assignment (specific) | T1098 | Medium | No specific detection for Global Admin role assignment |

### GitHub (20 rules)

**Current Coverage:** Repository permissions, large files, sensitive data commits, OAuth apps, SAML SSO, SSH keys, workflow modifications, package registry, releases, mass cloning, repo archive/delete, PATs, webhooks, secrets/secret scanning, GPG keys, code scanning, advanced security events, branch protection bypasses, enterprise member changes, fork permissions, organization creation spikes.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Self-Hosted Runner Compromise | T1195.002 | **Critical** | No detection for runner registration/deregistration or workflow runs targeting self-hosted runners |
| 2 | Workflow Injection via Pull Request | T1195.001 | High | No detection for pull_request_target trigger abuse |
| 3 | Environment Protection Bypass | T1562.001 | High | No detection for environment protection rule changes or required reviewer removals |
| 4 | Repository Visibility Change (Private to Public) | T1537 | High | Making a private repo public instantly exposes all code and secrets history |
| 5 | GitHub App Token Theft | T1528 | High | No detection for GitHub App tokens used from unexpected IPs |
| 6 | Dependabot Config Poisoning | T1195.002 | Medium | No detection for config changes redirecting dependency updates to malicious registries |
| 7 | CODEOWNERS File Modification | T1562.001 | Medium | Removing CODEOWNERS entries bypasses code review requirements |
| 8 | Org Member Role Escalation | T1098 | Medium | No specific rule for member-to-owner escalation |

---

## 3. Endpoint Security

### Windows (28 rules)

**Current Coverage:** LSASS/Mimikatz (T1003.001), SAM dump (T1003.002), NTDS.dit (T1003.003), DCSync (T1003.006), Kerberos/Certificate abuse (T1558), Event Log Clearing (T1070.001), DCShadow (T1207), Process Injection (T1055), AMSI Bypass/Defender Tampering (T1562.001), Token Manipulation (T1134), Named Pipe (T1134.001), SID History (T1134.005), UAC Bypass (T1548.002), PrintNightmare (T1068), DCOM (T1021.003), WinRM (T1021.006), SMBv1 (T1210), AdminSDHolder (T1098), GPO Tampering (T1484.001), PowerShell/Empire (T1059.001), BloodHound (T1087), Network Sniffing (T1040), Keylogging (T1056.001), Volume Shadow Copy Deletion (T1490), ADFS abuse (T1078), RDP Brute Force (T1110.001).

#### Credential Access Gaps
| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | LSASS Memory Protection Bypass | T1003.001 | **Critical** | No rule for procdump, comsvcs.dll MiniDump, or Task Manager LSASS dump (alternatives to Mimikatz) |
| 2 | Credential Dumping via Registry | T1003.004 | **Critical** | No detection for `reg save HKLM\SAM`, `reg save HKLM\SYSTEM`, `reg save HKLM\SECURITY` |
| 3 | Kerberoasting | T1558.003 | **Critical** | No rule for TGS ticket requests (Event ID 4769) with RC4 encryption targeting service accounts |
| 4 | Golden Ticket Detection | T1558.001 | **Critical** | No rule for TGT with anomalous lifetimes or domain admin SID from non-DC sources |
| 5 | AS-REP Roasting | T1558.004 | High | No rule for AS-REP requests targeting accounts without pre-authentication |
| 6 | Silver Ticket Detection | T1558.002 | High | No rule for forged TGS tickets bypassing the KDC |

#### Persistence Gaps (MAJOR GAP AREA)
| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 7 | Registry Run Keys | T1547.001 | **Critical** | No rule for HKLM/HKCU Run, RunOnce, RunServices key modifications -- the most common Windows persistence mechanism |
| 8 | Scheduled Tasks Creation | T1053.005 | **Critical** | No rule for suspicious schtasks.exe usage or Event ID 4698 with suspicious task actions |
| 9 | WMI Event Subscriptions | T1546.003 | **Critical** | No rule for WMI permanent event subscriptions (fileless persistence) |
| 10 | Windows Service Creation | T1543.003 | **Critical** | No rule for suspicious service creation via sc.exe or Event ID 7045 |
| 11 | DLL Search Order Hijacking | T1574.001 | High | No rule for DLL sideloading/hijacking via suspicious DLL loads in unusual paths |
| 12 | Image File Execution Options | T1546.012 | High | No rule for IFEO debugger persistence |
| 13 | Boot/Logon Scripts | T1037.001 | High | No rule for UserInitMprLogon, Userinit registry key modifications |
| 14 | Startup Folder Persistence | T1547.001 | High | No rule for file drops into Startup folders |

#### Defense Evasion Gaps
| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 15 | Signed Binary Proxy Execution (LOLBins) | T1218 | **Critical** | No rules for certutil, mshta, regsvr32, rundll32, cmstp, msiexec abuse. Only PowerShell is covered |
| 16 | ETW Patching/Evasion | T1562.006 | **Critical** | No rule for EtwEventWrite patching to blind security tools |
| 17 | Masquerading | T1036.005 | High | No rule for executables mimicking legitimate process names from wrong directories |
| 18 | Timestomping | T1070.006 | High | No rule for modified file timestamps |
| 19 | Disable/Modify Sysmon | T1562.001 | High | No rule for Sysmon service/driver unload |

#### Lateral Movement Gaps
| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 20 | PsExec/SMB-based Execution | T1569.002 | **Critical** | No rule for PsExec-style service creation via SMB (Event ID 7045 with PSEXESVC) |
| 21 | Pass-the-Hash | T1550.002 | **Critical** | No rule for NTLM authentication from unusual sources (Event ID 4624 LogonType 9) |
| 22 | Pass-the-Ticket | T1550.003 | High | No rule for Kerberos ticket reuse from non-original endpoints |
| 23 | RDP Session Hijacking | T1563.002 | High | No rule for tscon.exe session hijacking |
| 24 | WMI Remote Execution | T1047 | High | No native rule for WMIC /node: remote command execution |
| 25 | Windows Admin Shares | T1021.002 | High | No rule for suspicious C$, ADMIN$ share access |

#### Execution Gaps
| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 26 | MSHTA Execution | T1218.005 | High | No rule for mshta.exe executing HTA files or inline scripts |
| 27 | BITSAdmin Abuse | T1197 | High | No rule for bitsadmin.exe downloading files or persistence |
| 28 | Certutil Download/Decode | T1140 | High | No rule for certutil.exe -urlcache or -decode |
| 29 | Windows Script Host | T1059.005 | High | No rule for suspicious wscript.exe/cscript.exe execution |

#### Discovery Gaps
| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 30 | Domain Trust Discovery | T1482 | High | No rule for nltest /domain_trusts, dsquery, or Get-ADTrust enumeration |

### macOS (24 rules)

**Current Coverage:** AppleScript abuse, application firewall, code signing, contacts/calendar access, developer tools abuse, directory service modifications, endpoint security bypass, FileVault tampering, full disk access, Gatekeeper bypass, Homebrew tampering, kernel extension loading, keychain access, launch agent/daemon persistence, location services, MDM profile removal, microphone/camera access, network extension abuse, notarization bypass, privacy preferences tampering, Safari extension abuse, screen recording, Spotlight manipulation, Time Machine destruction.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Dylib Hijacking | T1574.004 | **Critical** | No rule for dynamic library hijacking by placing malicious dylibs in search paths |
| 2 | Login Items Persistence | T1547.015 | High | No rule for Login Items modifications for startup persistence |
| 3 | Bash/Zsh Profile Modification | T1546.004 | High | No rule for .bash_profile, .bashrc, .zshrc modifications for persistence |
| 4 | Sudo Caching Abuse | T1548.003 | High | No rule for sudo timestamp modification or tty_tickets bypass |
| 5 | Keychain Dumping via security CLI | T1555.001 | High | No specific detection for `security dump-keychain` or `security find-generic-password -w` |
| 6 | SSH Key Theft | T1552.004 | High | No rule for access to ~/.ssh/id_rsa or other private key files |
| 7 | macOS Ransomware Indicators | T1486 | High | No rule for mass file encryption patterns |
| 8 | Osascript from Terminal | T1059.002 | High | No detection for osascript invoked from command line with encoded scripts |
| 9 | Cron Job Persistence | T1053.003 | High | No rule for macOS crontab modifications |
| 10 | Folder Action Scripts | T1546.002 | Medium | No rule for Folder Actions abuse |
| 11 | Hidden Users Creation | T1564.002 | Medium | No rule for creating hidden macOS user accounts |

### Linux (33 rules - Debian + RHEL)

**Current Coverage:** Init scripts, systemd services, user accounts, kernel modules, AppArmor, SELinux, firewall changes, log/time tampering, rootkits, kernel exploits, SUID/sudo, container escape, package/repository tampering, boot security, PAM config, SSH brute force, boot loader attacks, etc.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Crontab Persistence | T1053.003 | **Critical** | No rule for crontab -e, /etc/cron.d/ additions, or at job creation |
| 2 | SSH Authorized Keys Modification | T1098.004 | **Critical** | No rule for modifications to ~/.ssh/authorized_keys for persistent backdoor access |
| 3 | Reverse Shell Detection | T1059.004 | **Critical** | No rule for bash -i >& /dev/tcp, python -c 'import socket', nc -e, socat reverse shells |
| 4 | SUID/SGID Binary Creation | T1548.001 | **Critical** | No rule for chmod +s on executables |
| 5 | LD_PRELOAD Hijacking | T1574.006 | **Critical** | No rule for LD_PRELOAD or /etc/ld.so.preload modifications |
| 6 | Shell RC File Modifications | T1546.004 | **Critical** | No rule for .bashrc, .bash_profile, /etc/profile.d/ modifications |
| 7 | /etc/passwd and /etc/shadow Access | T1003.008 | **Critical** | No rule for unauthorized reading of /etc/shadow or modification of /etc/passwd |
| 8 | eBPF Rootkit Detection | T1014 | **Critical** | No rule for malicious eBPF programs being loaded |
| 9 | Container Escape Techniques (specific) | T1611 | **Critical** | No specific rules for Docker socket access, nsenter escapes, cgroups release_agent exploit |
| 10 | Suspicious Binary in /tmp or /dev/shm | T1059.004 | High | No rule for executables in /tmp, /dev/shm, /var/tmp |
| 11 | SSH Tunneling/Port Forwarding | T1572 | High | No rule for ssh -L, ssh -R, ssh -D port forwarding |
| 12 | Kernel Exploitation Indicators | T1068 | High | No detection for dirty pipe, dirty cow, overlayfs exploit artifacts |
| 13 | Process Masquerading | T1036.004 | High | No rule for processes renaming via /proc/self/comm |
| 14 | Systemd Timer Persistence | T1053.003 | High | No rule for systemd timer units as alternative to cron |
| 15 | Auditd/Syslog Disabling | T1562.001 | High | No rule for systemctl stop auditd or service rsyslog stop |
| 16 | Rootkit Files/Directories | T1564.001 | High | No general rule for common rootkit indicators |

### Antivirus (82 rules)

#### Bitdefender GZ (18 rules)
| # | Missing Detection | Priority | Description |
|---|-------------------|----------|-------------|
| 1 | Policy Override Detection | High | No rule for detecting when AV policies are weakened by administrators |
| 2 | Lateral Movement via AV Console | High | No rule for AV management console used to push malicious policies |
| 3 | Scan Failure Patterns | Medium | No rule for repeated scan failures indicating malware interfering with AV |

#### ESET ESMC (14 rules)
| # | Missing Detection | Priority | Description |
|---|-------------------|----------|-------------|
| 1 | ESET Agent Tampering | **Critical** | No rule for ESET agent being disabled, uninstalled, or tampered with |
| 2 | Remote Admin Console Abuse | High | No rule for suspicious ERA/ESMC console activity (mass policy changes) |
| 3 | Quarantine Failures | High | No rule for repeated quarantine failures indicating AV evasion |

#### Kaspersky (17 rules)
| # | Missing Detection | Priority | Description |
|---|-------------------|----------|-------------|
| 1 | Kaspersky Agent Tampering | High | No rule for agent being disabled or tampered with |
| 2 | Ransomware-specific Behavior | High | No dedicated ransomware detection rule |
| 3 | Rootkit Detection | High | No rootkit detection rule |

#### SentinelOne (17 rules)
| # | Missing Detection | Priority | Description |
|---|-------------------|----------|-------------|
| 1 | Exclusion/Allowlist Abuse | High | No rule for suspicious allowlist additions |
| 2 | Policy Downgrade Detection | High | No rule for policy weakened from "Protect" to "Detect" mode |

### HIDS/OSSEC (15 rules)

**Current Coverage:** Active response, agent disconnection, enrollment anomalies, API auth failures, command monitoring, custom rules, Docker security, FIM, incident response, network anomalies, process monitoring, registry monitoring, rootcheck, system integrity, threat intel.

| # | Missing Detection | Priority | Description |
|---|-------------------|----------|-------------|
| 1 | OSSEC Rule Tampering | **Critical** | No rule for modifications to OSSEC rules themselves (ossec.conf, local_rules.xml) |
| 2 | Mass Agent Disconnection | **Critical** | No correlation for mass simultaneous disconnections indicating infrastructure attack |
| 3 | OSSEC Log Evasion | High | No rule for event flooding to bury malicious events |
| 4 | Syscheck Exclusion Abuse | High | No rule for modifications to syscheck exclusion lists |
| 5 | Agent Key Compromise | High | No rule for stolen or replayed agent authentication keys |

### Endpoint Security: MITRE Coverage Assessment

| Tactic | Windows | macOS | Linux | Rating |
|--------|---------|-------|-------|--------|
| Initial Access | Low | Low | Low | WEAK |
| Execution | Medium | Low | Low | MODERATE |
| **Persistence** | **Very Low** | **Medium** | **Very Low** | **CRITICAL GAP** |
| Privilege Escalation | High | Medium | Low | MODERATE |
| Defense Evasion | Medium | High | Medium | MODERATE |
| Credential Access | High | Low | Very Low | NEEDS WORK |
| Discovery | Low | None | None | WEAK |
| Lateral Movement | Medium | None | None | NEEDS WORK |
| Collection | Low | Medium | None | WEAK |
| **Exfiltration** | **None** | **None** | **None** | **CRITICAL GAP** |
| **Command and Control** | **Low** | **None** | **None** | **CRITICAL GAP** |
| Impact | Medium | Low | None | MODERATE |

---

## 4. Application & Infrastructure

### Apache (21 rules)

**Current Coverage:** Directory traversal, request smuggling, response splitting, cache poisoning, CGI exploitation, WebDAV, brute force, authentication bypass, htaccess, log injection, DoS, modsecurity, rate limiting bypass, reverse proxy abuse, server status exposure, source code disclosure, backup file access, config file access, information disclosure, module loading, web server compromise indicators.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Server-Side Request Forgery (SSRF) | T1090 | **Critical** | No rule for requests to internal IPs/metadata endpoints (169.254.169.254) via request parameters |
| 2 | Server-Side Template Injection (SSTI) | T1059 | **Critical** | No rule for template injection payloads ({{7*7}}, ${7*7}) |
| 3 | Remote Code Execution via CVE | T1190 | **Critical** | No rule for Apache CVEs like CVE-2021-41773/42013 (path traversal RCE) |
| 4 | XML External Entity (XXE) | T1190 | High | No rule for XXE payloads in POST bodies |
| 5 | Web Shell Access Detection | T1505.003 | High | No rule for access to common web shell paths (cmd.php, c99.php, etc.) |
| 6 | API Abuse/Enumeration | T1087 | High | No rule for rapid sequential API endpoint requests with error responses |
| 7 | Slow HTTP DoS (Slowloris) | T1499.001 | Medium | No detection for slow HTTP attacks |

### Nginx (7 rules)

**Current Coverage:** Buffer overflow, cache poisoning, error log injection, FastCGI exploitation, location block bypasses, Lua script injection, SSL/TLS vulnerabilities.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Directory Traversal | T1083 | **Critical** | Apache has path traversal but Nginx does not |
| 2 | Request Smuggling | T1190 | **Critical** | Apache has request smuggling but Nginx does not |
| 3 | SSRF via Proxy Pass | T1090 | **Critical** | No rule for SSRF through nginx proxy_pass misconfiguration |
| 4 | Brute Force Detection | T1110 | High | No brute force rule for Nginx-served applications |
| 5 | Web Shell Access | T1505.003 | High | No detection for web shell access patterns |
| 6 | DoS Patterns | T1499 | High | No DoS detection |
| 7 | Authentication Bypass | T1212 | High | No rule for auth bypass attempts |
| 8 | Server-Side Template Injection | T1059 | High | No SSTI detection |

### IIS (14 rules)

**Current Coverage:** ASPX injection, config file access, double encoding, FTP attacks, handler mapping abuse, ISAPI filter exploits, request filtering evasion, short filename enumeration, TRACE method, unicode bypass, URL rewrite bypass, virtual directory traversal, web shell uploads, WebDAV vulnerabilities.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | SSRF Attacks | T1090 | **Critical** | No SSRF detection |
| 2 | Deserialization Attacks (ViewState) | T1190 | **Critical** | No rule for .NET ViewState tampering or ysoserial payloads |
| 3 | Exchange/OWA Attacks | T1190 | **Critical** | No rule for ProxyShell/ProxyLogon/ProxyNotShell patterns |
| 4 | SQL Injection via IIS | T1190 | High | No SQL injection detection in IIS access logs |
| 5 | Cross-Site Scripting (XSS) | T1189 | High | No XSS detection |
| 6 | Brute Force on IIS Authentication | T1110 | High | No brute force detection for Windows auth failures |

### MySQL (9 rules)

**Current Coverage:** SQL injection, privilege escalation, stored procedure abuse, UDF exploitation, trigger manipulation, plugin installation, file system access, information schema queries, user account manipulation.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Data Exfiltration via SELECT INTO OUTFILE | T1048 | **Critical** | No specific rule for high-volume data export via OUTFILE/DUMPFILE |
| 2 | Binary Log Tampering | T1070 | High | No rule for PURGE BINARY LOGS |
| 3 | Audit Log Disable | T1562.002 | High | No rule for UNINSTALL PLUGIN audit_log |
| 4 | Mass Data Deletion | T1485 | High | No rule for TRUNCATE TABLE or mass DELETE operations |
| 5 | Replication-Based Attacks | T1210 | High | No rule for rogue MySQL replication |

### PostgreSQL (4 rules)

**Current Coverage:** SQL injection, privilege escalation, Kerberos auth failures, LDAP auth bypass.

**NOTE: Only 4 rules -- severely undercovered for such a critical database.**

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | COPY TO/FROM PROGRAM RCE | T1005 | **Critical** | Allows arbitrary command execution via PostgreSQL |
| 2 | Extension-Based RCE | T1059 | **Critical** | CREATE EXTENSION for plpythonu/plperlu OS command execution |
| 3 | pg_read_file/pg_ls_dir Abuse | T1083 | High | Filesystem browsing via admin functions |
| 4 | Role/User Manipulation | T1098 | High | CREATE ROLE with SUPERUSER or GRANT ALL PRIVILEGES |
| 5 | Foreign Data Wrapper Abuse | T1048 | High | Data exfiltration to external servers |
| 6 | Stored Procedure Code Execution | T1059 | High | CREATE FUNCTION with plpythonu/plperlu/plsh |
| 7 | Data Exfiltration via dblink | T1048 | High | dblink_connect/dblink_send_query to external servers |

### MongoDB (9 rules)

**Current Coverage:** Authentication bypass, collection dropping, database enumeration, injection attacks, Kerberos auth failures, LDAP injection, role privilege escalation, SCRAM auth attacks, stored JavaScript execution.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Unauthorized Data Export | T1048 | **Critical** | No rule for mongoexport/mongodump from unauthorized sources |
| 2 | Replica Set Reconfiguration | T1210 | High | No rule for rs.reconfig() to inject rogue replica members |
| 3 | Audit Log Disable | T1562.002 | High | No rule for disabling MongoDB audit log |

### Redis (5 rules)

**Current Coverage:** Command injection, data exfiltration, Lua script injection, persistence mechanism abuse, unauthorized access patterns.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | SLAVEOF/REPLICAOF Rogue Replication | T1210 | **Critical** | Attacker-initiated replication to steal all data or load malicious modules |
| 2 | ACL Manipulation | T1098 | High | No rule for ACL SETUSER/DELUSER to create backdoor accounts |
| 3 | FLUSHALL/FLUSHDB Data Destruction | T1485 | High | No rule for mass data deletion commands |

### Kafka (2 rules) -- SEVERELY UNDERCOVERED

**Current Coverage:** Unauthorized topic access, schema registry tampering.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Consumer Group Hijacking | T1557 | **Critical** | No rule for unauthorized consumer group creation to intercept messages |
| 2 | Topic Deletion/Mass Purging | T1485 | **Critical** | No rule for topic deletion destroying message data |
| 3 | ACL Modification | T1098 | **Critical** | No rule for kafka-acls changes granting unauthorized access |
| 4 | Producer Spoofing/Message Injection | T1565.001 | High | No rule for unauthorized message production to sensitive topics |
| 5 | Connect Worker Exploitation | T1059 | High | No rule for malicious Kafka Connect connector deployment |
| 6 | Broker Configuration Changes | T1562 | High | No rule for dynamic broker config changes |
| 7 | Partition Reassignment | T1489 | High | No rule for partition reassignment causing data loss |

### NATS (2 rules) -- SEVERELY UNDERCOVERED

**Current Coverage:** JWT authentication failures, unauthorized subscription.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Subject Wildcard Abuse | T1040 | **Critical** | No rule for ">" wildcard subscriptions capturing all messages |
| 2 | JetStream Stream Deletion | T1485 | **Critical** | No rule for unauthorized stream/consumer deletion |
| 3 | Account Token Theft/Replay | T1528 | High | No rule for credential theft beyond JWT failures |
| 4 | Cluster Route Poisoning | T1557 | High | No rule for unauthorized route advertisements hijacking cluster traffic |

### Elasticsearch (6 rules)

**Current Coverage:** API key abuse, node compromise indicators, role mapping changes, security realm modifications, token service anomalies, unauthorized index access.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Snapshot Repository Exfiltration | T1048 | **Critical** | No rule for snapshot repos pointing to attacker-controlled S3/NFS |
| 2 | Painless Script RCE | T1059 | **Critical** | No rule for malicious Painless/Groovy script execution in queries |
| 3 | Index Deletion/Mass Deletion | T1485 | High | No rule for DELETE /{index} or _delete_by_query bulk operations |
| 4 | Reindex to External | T1048 | High | No rule for _reindex with remote dest pointing to external clusters |
| 5 | Cluster Settings Manipulation | T1562 | High | No rule for _cluster/settings disabling security features |

### Kibana (9 rules)

**Current Coverage:** API access violations, CSRF token bypasses, OIDC security, PKI auth failures, role mapping bypasses, SAML issues, session hijacking, unauthorized space access, XSS vulnerabilities.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Saved Object Import RCE | T1059 | **Critical** | No rule for malicious saved object imports with embedded scripts |
| 2 | Console/Dev Tools Abuse | T1059 | High | No rule for dangerous queries via Kibana Dev Tools |
| 3 | Report Generation for Exfiltration | T1048 | High | No rule for mass report/CSV export generation |

### Logstash (6 rules)

**Current Coverage:** Configuration tampering, file input path traversal, HTTP input vulnerabilities, JDBC input injection, monitoring API exposure, Ruby filter code execution.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Pipeline Injection via Central Management | T1565 | **Critical** | No rule for unauthorized pipeline deployment |
| 2 | Output Redirection | T1048 | **Critical** | No rule for output plugin modification to redirect data to attacker-controlled destinations |
| 3 | Malicious Plugin Installation | T1195.002 | High | No rule for logstash-plugin install of unofficial/malicious plugins |

### HAProxy (10 rules)

**Current Coverage:** ACL bypass, admin socket abuse, backend server manipulation, cache manipulation, configuration injection, HTTP request smuggling, rate limiting evasion, runtime API abuse, SSL/TLS vulnerabilities, stats page access.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | SSRF via Backend Selection | T1090 | High | No rule for SSRF via Host/X-Forwarded-Host manipulation |
| 2 | Connection Exhaustion DoS | T1499 | Medium | No rule for maxconn exhaustion attacks |
| 3 | Stick Table Poisoning | T1565 | Medium | No rule for stick table manipulation via runtime API |

### Traefik (7 rules)

**Current Coverage:** Access log injection, dynamic configuration attacks, metrics endpoint exposure, middleware bypass, rate limiter bypass, router rule injection, TLS configuration attacks.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Docker/Kubernetes Provider Exploitation | T1610 | **Critical** | No rule for exploiting Docker/K8s label-based service discovery to inject malicious routes |
| 2 | SSRF via Service Discovery | T1090 | High | No rule for SSRF through Consul/etcd/ZooKeeper manipulation |
| 3 | Dashboard API Exposure | T1082 | High | No rule for unauthorized Traefik dashboard/API access |

### VMware ESXi (21 rules)

**Current Coverage:** Hypervisor escape, VM escape, vCenter attacks, VMCI/VMHGFS exploitation, SSO bypasses, vSphere API abuse, PowerCLI execution, snapshot manipulation, template tampering, virtual hardware manipulation, distributed switch attacks, VLAN trunk attacks, VMotion security, VMkernel interface abuse, management network exposure, iSCSI/NFS storage attacks, OVF/OVA deployment risks, VMware Tools vulnerabilities, certificate validation, guest introspection bypass.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | ESXi Ransomware (Royal/BlackBasta/ESXiArgs) | T1486 | **Critical** | No rule for ESXi ransomware patterns (encrypting .vmdk/.vmx, vim-cmd mass VM shutdown) |
| 2 | SSH Shell Access to ESXi | T1021.004 | **Critical** | No rule for direct SSH login to ESXi host (should normally be disabled) |
| 3 | Unauthorized VIB Sideloading | T1195.002 | High | No dedicated rule for unsigned VIB installation (VIRTUALPITA technique) |
| 4 | ESXi Firewall Rule Modification | T1562.004 | High | No rule for esxcli network firewall ruleset modifications |
| 5 | VM Disk Theft via Datastore Browser | T1005 | High | No rule for .vmdk download via vSphere Web Client |
| 6 | ESXi Account Manipulation | T1098 | High | No rule for local ESXi user creation/modification |
| 7 | Syslog Forwarding Disruption | T1562.002 | Medium | No rule for disabling ESXi syslog forwarding |

### IBM AIX (18 rules)

**Current Coverage:** Core dump analysis, encrypted filesystem, RBAC, firewall policy, intrusion detection, IPSec, Kerberos auth, LDAP client, network security, privileged command execution, RBAC violations, security audit subsystem, security policy enforcement, system config changes, system integrity violations, trusted computing base, trusted execution violations, user administration.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Rootkit Detection | T1014 | **Critical** | No rule for AIX-specific rootkit indicators |
| 2 | Cron Job Persistence | T1053.003 | High | No rule for unauthorized crontab modifications |
| 3 | SSH Key Manipulation | T1098.004 | High | No rule for authorized_keys modifications |
| 4 | NIM (Network Installation Manager) Abuse | T1210 | High | No rule for NIM master exploitation |
| 5 | HMC Access | T1078 | High | No rule for unauthorized Hardware Management Console access |

### IBM AS/400 (6 rules) -- SIGNIFICANTLY UNDERCOVERED

**Current Coverage:** Object authority violations, password policy violations, security audit journal, special authority usage, system value changes, user profile modifications.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Program Adoption Authority Abuse | T1548 | **Critical** | No rule for *ALLOBJ program adopt authority for privilege escalation |
| 2 | Exit Point Tampering | T1562 | **Critical** | No rule for malicious exit point programs (ADDEXITPGM) |
| 3 | Audit Journal Disable | T1562.002 | **Critical** | No rule for CHGAUD to disable audit journaling |
| 4 | IFS Access | T1005 | High | No rule for unauthorized Integrated File System access |
| 5 | FTP/Remote Command Execution | T1021 | High | No rule for RCMD, RMTCMD, or DDM remote command execution |
| 6 | Job Queue Manipulation | T1053 | High | No rule for job submission with elevated authority |
| 7 | SQL via ODBC/JDBC | T1190 | High | No rule for SQL injection via ODBC/JDBC connections |
| 8 | Library List Manipulation | T1574 | Medium | No rule for CHGLIBL/ADDLIBLE (DLL hijacking equivalent on AS/400) |

### Auditd (18 rules)

**Current Coverage:** AppArmor violations, audit daemon failures, audit rule modifications, capability usage, configuration changes, executable monitoring, file access violations, group membership changes, kernel module loading, key-based correlation, log tampering, network connections, privilege escalation, process execution, resource limits, SELinux AVC denials, syscall filtering bypasses, system call anomalies, time changes, user authentication, watch rule violations.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Container Escape via Syscalls | T1611 | **Critical** | No rule for unshare CLONE_NEWUSER, nsenter, host filesystem mount from containers |
| 2 | eBPF-Based Attacks | T1014 | **Critical** | No rule for malicious eBPF program loading (bpf() syscall from non-standard processes) |
| 3 | ptrace-Based Process Injection | T1055.008 | High | No rule for PTRACE_ATTACH/PTRACE_POKETEXT |
| 4 | Memory-Only Malware | T1059 | High | No rule for memfd_create + fexecve fileless execution |
| 5 | Namespace Manipulation | T1611 | High | No rule for setns/unshare for namespace escape |

### System Linux Module (11 rules)

**Current Coverage:** AppArmor, firewall rules, hardware errors, kernel parameters, memory usage, network config, package installation, PAM config, process anomalies, SELinux, SSH brute force, sudo escalation, system boot, resource exhaustion, service failures, update failures, user account modifications.

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Reverse Shell Detection | T1059.004 | **Critical** | No rule for common reverse shell patterns |
| 2 | SSH Authorized Key Injection | T1098.004 | High | No rule for unauthorized authorized_keys modifications |
| 3 | Systemd Timer Persistence | T1053.006 | High | No rule for malicious systemd timer creation |
| 4 | LD_PRELOAD Hijacking | T1574.006 | High | No rule for /etc/ld.so.preload modifications |

### Osquery (6 rules)

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Agent Tampering/Uninstall | T1562.001 | **Critical** | No rule for osquery agent being stopped or tampered with |
| 2 | Fleet Server Compromise | T1199 | High | No rule for Fleet/Kolide server compromise indicators |

### Generic (9 rules)

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Log Source Health Monitoring | N/A | **Critical** | No rule for detecting when a log source stops sending data -- critical for detecting evasion |
| 2 | Timestamp Manipulation | T1070.006 | High | No rule for timestomping in log events |
| 3 | Cross-Source Lateral Movement Correlation | T1021 | High | No generic rule correlating auth events across sources |
| 4 | Log Volume Anomalies | T1562 | High | No rule for sudden log volume changes indicating compromise or evasion |

### JSON Input (10 rules)

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | NoSQL Injection via JSON | T1190 | **Critical** | No rule for MongoDB-style injection ($gt, $ne, $regex, $where operators) |
| 2 | GraphQL Abuse | T1190 | High | No rule for introspection queries, batching attacks, deeply nested queries |
| 3 | Mass Assignment | T1565 | High | No rule for unexpected fields like isAdmin in JSON body |
| 4 | JWT Algorithm Confusion | T1550 | High | No rule for alg:none or RS256->HS256 downgrade attacks |

### Syslog (23 rules)

| # | Missing Detection | MITRE ATT&CK | Priority | Description |
|---|-------------------|---------------|----------|-------------|
| 1 | Log Source Impersonation | T1036 | **Critical** | No rule for CEF messages with spoofed device vendor/product fields |
| 2 | Severity Manipulation | T1565 | High | No rule for systematic severity downgrading in forwarded messages |
| 3 | Syslog Relay Chain Manipulation | T1557 | High | No rule for unauthorized syslog relay additions |

---

## 5. Cross-Cutting Gaps

These gaps affect multiple technologies and should be addressed with cross-platform rules:

| # | Gap | Priority | Technologies Affected | Description |
|---|-----|----------|----------------------|-------------|
| 1 | **SSRF Detection** | **Critical** | Apache, Nginx, IIS, HAProxy, Traefik | Zero SSRF rules across any web server/proxy. Requests to 169.254.169.254, localhost, internal ranges are unmonitored |
| 2 | **Server-Side Template Injection** | **Critical** | Apache, Nginx, IIS | Zero SSTI detection across any web tier |
| 3 | **Log Source Health Monitoring** | **Critical** | All technologies | No rules to detect when any source stops sending events -- first thing attackers disable |
| 4 | **Firewall Policy Change Detection** | **Critical** | FortiGate, Palo Alto, pfSense, SonicWall, Sophos XG | Most network devices lack policy change detection |
| 5 | **Impossible Travel Detection** | **Critical** | AWS, Azure, O365, GitHub | No impossible travel detection for any cloud/SaaS platform |
| 6 | **Persistence Mechanisms** | **Critical** | Windows, macOS, Linux | Registry run keys, crontab, SSH authorized_keys, WMI subscriptions all undetected |
| 7 | **Supply Chain via Plugins** | High | Elasticsearch, Logstash, Kibana, Redis, Traefik, Osquery, PostgreSQL, MySQL | Most technologies lack malicious plugin/extension loading detection |
| 8 | **Data Exfiltration via Native Tools** | High | MySQL, PostgreSQL, MongoDB, Redis, Elasticsearch | Limited detection for data theft using native export/replication features |
| 9 | **Audit/Logging Disable** | High | MySQL, MongoDB, PostgreSQL, Elasticsearch, AIX, AS/400 | Most technologies lack detection for disabling their own audit logging |
| 10 | **XSS Detection** | Medium | Apache, Nginx, IIS | No cross-site scripting detection in web server access logs |
| 11 | **Browser Credential Theft** | High | Windows, macOS, Linux | No rule for access to browser credential stores on any OS |
| 12 | **Exfiltration (endpoint)** | **Critical** | Windows, macOS, Linux | Zero exfiltration detection rules on any endpoint OS |

---

## 6. Priority Implementation Roadmap

### Phase 1: Critical (Immediate -- addresses highest-risk gaps)

**Endpoint Persistence & Credential Theft (Windows)**
1. Registry Run Keys Persistence (T1547.001)
2. Scheduled Task Creation (T1053.005)
3. WMI Event Subscription Persistence (T1546.003)
4. Windows Service Creation Abuse (T1543.003)
5. Kerberoasting Detection (T1558.003)
6. Golden Ticket Detection (T1558.001)
7. LSASS Memory Protection Bypass (T1003.001)
8. Credential Dumping via Registry (T1003.004)
9. LOLBin Execution -- certutil, mshta, regsvr32, rundll32, bitsadmin (T1218)
10. PsExec/SMB-based Lateral Movement (T1569.002)
11. Pass-the-Hash Detection (T1550.002)
12. ETW Patching/Evasion (T1562.006)

**Linux Core Detections**
13. Crontab Persistence (T1053.003)
14. SSH Authorized Keys Modification (T1098.004)
15. Reverse Shell Detection (T1059.004)
16. SUID/SGID Binary Creation (T1548.001)
17. LD_PRELOAD Hijacking (T1574.006)
18. /etc/shadow Access (T1003.008)
19. eBPF Rootkit Detection (T1014)
20. Container Escape Techniques (T1611)

**Web Application (Cross-Server)**
21. SSRF Detection -- Apache, Nginx, IIS (T1090)
22. Server-Side Template Injection (T1059)
23. NoSQL Injection via JSON (T1190)

**Cloud Critical**
24. EC2/GCE Cryptomining Detection (T1496)
25. Impossible Travel -- AWS, Azure, O365 (T1078)
26. GCP Audit Log Disabling (T1562.008)
27. GCP Service Account Impersonation (T1550.001)
28. O365 Mailbox Delegation Abuse (T1098.002)

**Infrastructure**
29. Log Source Health Monitoring (Generic)
30. ESXi Ransomware Patterns (T1486)

**Network**
31. Firewall Policy Change Detection -- FortiGate, Palo Alto, pfSense, SonicWall, Sophos XG (T1562.004)
32. FortiGate SSL-VPN Brute Force (T1133)

### Phase 2: High (Near-term -- important gaps)

33. macOS Dylib Hijacking (T1574.004)
34. Shell RC File Modifications -- Linux/macOS (T1546.004)
35. MFA Fatigue / Push Spam -- O365 (T1621)
36. Azure AD App Registration Abuse (T1098.001)
37. Azure Golden SAML / Federation Abuse (T1606.002)
38. GCP Cloud Storage Exfiltration (T1530)
39. GitHub Self-Hosted Runner Compromise (T1195.002)
40. PostgreSQL COPY TO PROGRAM RCE (T1005)
41. Elasticsearch Snapshot Exfiltration & Painless Script RCE (T1048/T1059)
42. Redis SLAVEOF Rogue Replication (T1210)
43. Kafka ACL/Topic Deletion/Consumer Group Hijacking (T1098/T1485/T1557)
44. Nginx Directory Traversal & Request Smuggling (T1083/T1190)
45. IIS Deserialization Attacks & Exchange/OWA Attacks (T1190)
46. VMware SSH Access & VIB Sideloading (T1021.004/T1195.002)
47. MikroTik User Account Creation & SOCKS Proxy (T1136/T1090)
48. NetFlow SMB/RDP Lateral Movement (T1021)
49. Windows Admin Shares (T1021.002)
50. Browser Credential Theft -- all OSes (T1555.003)

### Phase 3: Medium (Scheduled -- important but lower urgency)

51-80. Remaining medium-priority gaps including: Azure PIM abuse, GCP BigQuery exfiltration, O365 Power BI export, IBM AS/400 exit point tampering, Auditd ptrace injection, NATS wildcard abuse, Logstash output redirection, Traefik Docker provider exploitation, HAProxy SSRF, Syslog source impersonation, macOS Login Items persistence, Windows Discovery techniques, etc.

---

## Appendix: Gap Counts by Technology

| Technology | Current Rules | Identified Gaps | Critical | High | Medium |
|-----------|:------------:|:---------------:|:--------:|:----:|:------:|
| Windows | 28 | 30 | 12 | 15 | 3 |
| Linux (Debian+RHEL) | 33 | 16 | 9 | 7 | 0 |
| macOS | 24 | 11 | 1 | 8 | 2 |
| AWS | 24 | 12 | 2 | 5 | 5 |
| Azure | 18 | 11 | 2 | 5 | 4 |
| GCP | 8 | 12 | 3 | 4 | 5 |
| Office 365 | 27 | 10 | 2 | 4 | 4 |
| GitHub | 20 | 8 | 1 | 4 | 3 |
| Cisco (all) | 41 | 26 | 2 | 13 | 11 |
| Fortinet (all) | 24 | 14 | 3 | 5 | 6 |
| Palo Alto | 8 | 7 | 1 | 4 | 2 |
| MikroTik | 17 | 6 | 2 | 4 | 0 |
| pfSense | 8 | 4 | 1 | 2 | 1 |
| SonicWall | 19 | 4 | 1 | 3 | 0 |
| Sophos | 24 | 4 | 1 | 3 | 0 |
| NetFlow | 14 | 7 | 1 | 4 | 2 |
| NIDS/Suricata | 17 | 6 | 0 | 4 | 2 |
| Apache | 21 | 7 | 3 | 3 | 1 |
| Nginx | 7 | 8 | 3 | 5 | 0 |
| IIS | 14 | 6 | 3 | 3 | 0 |
| MySQL | 9 | 5 | 1 | 4 | 0 |
| PostgreSQL | 4 | 7 | 2 | 5 | 0 |
| MongoDB | 9 | 3 | 1 | 2 | 0 |
| Redis | 5 | 3 | 1 | 2 | 0 |
| Kafka | 2 | 7 | 3 | 4 | 0 |
| NATS | 2 | 4 | 2 | 2 | 0 |
| Elasticsearch | 6 | 5 | 2 | 3 | 0 |
| Kibana | 9 | 3 | 1 | 2 | 0 |
| Logstash | 6 | 3 | 2 | 1 | 0 |
| HAProxy | 10 | 3 | 0 | 1 | 2 |
| Traefik | 7 | 3 | 1 | 2 | 0 |
| VMware ESXi | 21 | 7 | 2 | 4 | 1 |
| IBM AIX | 18 | 5 | 1 | 4 | 0 |
| IBM AS/400 | 6 | 8 | 3 | 4 | 1 |
| Auditd | 18 | 5 | 2 | 3 | 0 |
| System Linux | 11 | 4 | 1 | 3 | 0 |
| Osquery | 6 | 2 | 1 | 1 | 0 |
| Antivirus (all) | 82 | 15 | 1 | 10 | 4 |
| HIDS/OSSEC | 15 | 5 | 2 | 3 | 0 |
| Generic | 9 | 4 | 1 | 3 | 0 |
| JSON Input | 10 | 4 | 1 | 3 | 0 |
| Syslog | 23 | 3 | 1 | 2 | 0 |
