# Sigma Rules Delta Report - Missing High-Value Detections

**Generated:** 2026-02-05
**Comparison:** UTMStack Correlation Rules (692 rules) vs [SigmaHQ/sigma](https://github.com/SigmaHQ/sigma) repository
**Focus:** Real threat detections only - no informational/operational/high-FP rules

---

## Executive Summary

Analysis identified **~154 high-value detection gaps** across 4 technology areas where Sigma has rules we lack. The largest gaps are:

1. **Windows** (25 gaps): Zero LOLBIN detection, zero Kerberos attack detection, zero persistence mechanism detection (scheduled tasks, services, registry, WMI, COM), no Impacket/framework-specific lateral movement
2. **Linux/macOS** (50 gaps): Zero reverse shell detection, no webshell activity detection, no Linux hack tool detection, limited credential access and persistence coverage
3. **Cloud/SaaS** (47 gaps): No Golden SAML detection, no Azure Identity Protection integration, missing defense evasion rules (alert suppression, finding evasion), missing Azure AD-specific attacks
4. **Network/Web** (32 gaps): Zero proxy-level C2/malware user agent detection, no IOC-based DNS rules, missing SSTI/JNDI exploitation signatures, no rclone/exfiltration tool detection

---

## Table of Contents

- [1. Windows Detection Gaps](#1-windows-detection-gaps)
  - [1.1 Critical Priority](#11-critical-priority)
  - [1.2 High Priority](#12-high-priority)
  - [1.3 Medium Priority](#13-medium-priority)
- [2. Linux/macOS Detection Gaps](#2-linuxmacos-detection-gaps)
  - [2.1 Critical Priority](#21-critical-priority)
  - [2.2 High Priority](#22-high-priority)
  - [2.3 Medium Priority](#23-medium-priority)
- [3. Cloud/SaaS Detection Gaps](#3-cloudsaas-detection-gaps)
  - [3.1 Critical Priority](#31-critical-priority)
  - [3.2 High Priority](#32-high-priority)
  - [3.3 Medium Priority](#33-medium-priority)
- [4. Network/Web/Proxy Detection Gaps](#4-networkwebproxy-detection-gaps)
  - [4.1 Critical Priority](#41-critical-priority)
  - [4.2 High Priority](#42-high-priority)
  - [4.3 Medium Priority](#43-medium-priority)

---

## 1. Windows Detection Gaps

**Current coverage:** 28 rules | **Gaps identified:** 25

Our Windows rules have strong AD attack coverage (DCShadow, DCSync, Certificate Services, BloodHound) and good PowerShell/process injection detection. However, we have **zero coverage** for LOLBINs, Kerberos attacks, persistence mechanisms, and specific lateral movement frameworks.

### 1.1 Critical Priority

These are the most exploited attack techniques in real-world Windows compromises.

| # | Gap | Sigma Rule | MITRE | Why Missing This Matters |
|---|-----|-----------|-------|--------------------------|
| W1 | **Kerberoasting** | [win_security_kerberoasting_activity.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/builtin/security/win_security_kerberoasting_activity.yml) | T1558.003 | #1 AD credential theft technique. Detects TGS requests with RC4 encryption targeting service accounts. Used in nearly every AD compromise. |
| W2 | **AS-REP Roasting** | [win_security_kerberos_asrep_roasting.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/builtin/security/win_security_kerberos_asrep_roasting.yml) | T1558.004 | Companion to Kerberoasting. Targets accounts without pre-authentication. |
| W3 | **Scheduled Task Persistence** | [win_security_susp_scheduled_task_creation.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/builtin/security/win_security_susp_scheduled_task_creation.yml) | T1053.005 | One of the most common persistence mechanisms. Detects tasks created from suspicious paths (Temp, Downloads, AppData) running PowerShell/certutil/rundll32. |
| W4 | **Suspicious Service Installation** | [win_security_cobaltstrike_service_installs.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/builtin/security) + metasploit/meterpreter variants | T1543.003 | Malicious service patterns from Cobalt Strike, Metasploit/Impacket PSExec, Meterpreter getsystem (Event 4697/7045). |
| W5 | **Impacket Lateral Movement** | [proc_creation_win_hktl_impacket_lateral_movement.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/process_creation/proc_creation_win_hktl_impacket_lateral_movement.yml) | T1047, T1021.003 | THE most-used lateral movement framework in real attacks. Detects wmiexec/dcomexec/atexec/smbexec patterns (cmd.exe /Q /c with 127.0.0.1 pipe redirection). |
| W6 | **Certutil LOLBIN Abuse** | [proc_creation_win_certutil_download.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/process_creation) + 9 more certutil rules | T1027, T1105 | Top LOLBIN for file download and payload staging. Used for downloading files from URLs, encoding/decoding Base64 payloads, NTLM coercion. |
| W7 | **Registry Run Key Persistence** | [registry_set_susp_reg_persist_explorer_run.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/registry/registry_set) + powershell variant | T1547.001 | THE most fundamental Windows persistence mechanism. Detects suspicious entries in HKCU/HKLM Run/RunOnce keys. |
| W8 | **LSASS Dumping (Multiple Techniques)** | [9+ rules in process_access/](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/process_access) | T1003.001 | Our mimikatz rule covers ONE tool. Need detection for comsvcs.dll, procdump, nanodump, pypykatz, WinRM, Windows Error Reporting, direct NT API calls. |

### 1.2 High Priority

| # | Gap | Sigma Rule | MITRE | Why Missing This Matters |
|---|-----|-----------|-------|--------------------------|
| W9 | **MSHTA Execution** | [proc_creation_win_mshta_http.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/process_creation) + 6 more | T1218.005 | Top initial access and defense evasion vector. Detects remote HTA execution, inline VBScript/JavaScript. |
| W10 | **BITSAdmin Abuse** | [proc_creation_win_bitsadmin_download.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/process_creation) + 5 more | T1197, T1105 | BITS jobs commonly abused for both download and persistence. |
| W11 | **MSIExec Abuse** | [proc_creation_win_msiexec_web_install.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/process_creation) + 6 more | T1218.007 | Remote payload delivery and DLL execution via MSI packages from URLs. |
| W12 | **Cobalt Strike Process Patterns** | [proc_creation_win_hktl_cobaltstrike_process_patterns.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/process_creation) + pipe rules | T1059 | Distinct CS beacon process creation patterns and named pipe patterns beyond our existing generic rule. |
| W13 | **WMI Event Subscription Persistence** | [sysmon_wmi_event_subscription.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/wmi_event/sysmon_wmi_event_subscription.yml) | T1546.003 | Stealthy and common persistence mechanism via WMI event subscriptions. |
| W14 | **BYOVD (Vulnerable Driver Loading)** | [driver_load_win_vuln_drivers.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/driver_load) | T1543.003, T1068 | Loading known vulnerable drivers to disable security tools. Increasingly used by ransomware groups and APTs. 1000+ driver hashes. |
| W15 | **DPAPI Domain Key Extraction** | [win_security_dpapi_domain_backupkey_extraction.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/builtin/security/win_security_dpapi_domain_backupkey_extraction.yml) | T1003.004 | Domain DPAPI backup key extraction from DCs enables decryption of all domain credentials. |
| W16 | **PetitPotam NTLM Relay** | [win_security_petitpotam_network_share.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/builtin/security) | T1187 | Critical NTLM relay attack via EFS RPC abuse, frequently used against AD CS. |

### 1.3 Medium Priority

| # | Gap | Sigma Rule | MITRE | Description |
|---|-----|-----------|-------|-------------|
| W17 | **Rundll32 Abuse** | [proc_creation_win_hktl_cobaltstrike_load_by_rundll32.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/process_creation) | T1218.011 | Rundll32 loading malicious DLLs, comsvcs.dll MiniDump for LSASS. |
| W18 | **CMSTP UAC/AppLocker Bypass** | [proc_creation_win_cmstp_execution_by_creation.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/process_creation/proc_creation_win_cmstp_execution_by_creation.yml) | T1218.003 | CMSTP.exe executing arbitrary commands via .inf files, bypassing AppLocker and UAC. |
| W19 | **Sliver C2 Framework** | [proc_creation_win_hktl_sliver_c2_execution_pattern.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/process_creation/proc_creation_win_hktl_sliver_c2_execution_pattern.yml) | T1059 | Increasingly popular open-source C2 replacing Cobalt Strike. Characteristic PowerShell pattern. |
| W20 | **COM Hijacking Persistence** | [registry_set_persistence_com_hijacking_builtin.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/registry/registry_set) + 4 more | T1546.015 | COM object registration modifications for stealthy persistence. |
| W21 | **DLL Sideloading/Hijacking** | [image_load_side_load_from_non_system_location.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/image_load) | T1574.001/002 | Known DLL sideloading patterns. Sigma has 99 rules covering specific vulnerable apps. |
| W22 | **Shadow Credentials Attack** | [win_security_susp_possible_shadow_credentials_added.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/builtin/security/win_security_susp_possible_shadow_credentials_added.yml) | T1556 | Addition of Key Credentials to AD objects (msDS-KeyCredentialLink) for passwordless auth. |
| W23 | **LaZagne Credential Harvesting** | [proc_creation_win_hktl_lazagne.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/process_creation/proc_creation_win_hktl_lazagne.yml) | T1003, T1555 | Multi-credential harvesting tool targeting 25+ credential stores. |
| W24 | **Tunnel Service C2** | [dns_query_win_domain_ngrok.yml](https://github.com/SigmaHQ/sigma/tree/master/rules/windows/dns_query) + cloudflared, devtunnels, localtonet | T1572, T1090 | DNS queries to tunneling services (ngrok, cloudflared, devtunnels) from non-browser processes. |
| W25 | **NTLM Downgrade Attack** | [win_security_net_ntlm_downgrade.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/windows/builtin/security/win_security_net_ntlm_downgrade.yml) | T1562.001 | NTLM auth downgrade via LMCompatibilityLevel registry changes. |

---

## 2. Linux/macOS Detection Gaps

**Current coverage:** ~85 rules | **Gaps identified:** ~50

Our Linux rules cover auditd events, systemd, package management, and kernel security well. However, we have **zero reverse shell detection**, no webshell activity detection, no hack tool signatures, and limited credential access/persistence coverage.

### 2.1 Critical Priority

| # | Gap | Sigma Rule | MITRE | Why Missing This Matters |
|---|-----|-----------|-------|--------------------------|
| L1 | **Comprehensive Reverse Shell Detection** | [lnx_shell_susp_rev_shells.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/builtin/lnx_shell_susp_rev_shells.yml) | T1059.004 | Covers 20+ reverse shell patterns: awk, bash, perl, python, ruby, socat, nc, telnet, xterm variants. **CRITICAL gap - we have zero reverse shell detection.** |
| L2 | **Netcat Reverse Shell** | [proc_creation_lnx_netcat_reverse_shell.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_netcat_reverse_shell.yml) | T1059 | nc/ncat with -c or -e flags spawning shells. Most common reverse shell technique. |
| L3 | **Python Reverse Shell** | [proc_creation_lnx_python_reverse_shell.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_python_reverse_shell.yml) | T1059 | Python with socket+pty import pattern. Very common post-webshell. |
| L4 | **Bash /dev/tcp Reverse Shell** | [lnx_susp_dev_tcp.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/builtin/lnx_susp_dev_tcp.yml) | T1059.004 | Classic `bash -i >& /dev/tcp/` pattern. |
| L5 | **PHP Reverse Shell** | [proc_creation_lnx_php_reverse_shell.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_php_reverse_shell.yml) | T1059 | PHP -r with fsockopen. Common on compromised web servers. |
| L6 | **Webshell Detection (Web Server Child Processes)** | [proc_creation_lnx_webshell_detection.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_webshell_detection.yml) | T1505.003 | Web servers (httpd, nginx, apache2, tomcat) spawning recon commands (whoami, ifconfig, uname, cat, crontab). Active exploitation indicator. |
| L7 | **Linux HackTool Execution** | [proc_creation_lnx_susp_hktl_execution.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_susp_hktl_execution.yml) | T1587 | 50+ known tools: crackmapexec, sliver, havoc, merlin, msfconsole, msfvenom, pspy, linpeas, hydra, john, hashcat, sqlmap, ncrack, bloodhound-python, evil-winrm. |
| L8 | **Java Child Processes (Log4Shell/Confluence)** | [proc_creation_lnx_susp_java_children.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_susp_java_children.yml) | T1059 | Java spawning shells, curl, wget, python. Covers Log4Shell, Confluence exploits, Java deserialization attacks. |
| L9 | **ld.so.preload Injection** | [lnx_ldso_preload_injection.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/builtin/lnx_ldso_preload_injection.yml) | T1574.006 | Modification of /etc/ld.so.preload for shared library injection. Major persistence/privesc vector. |
| L10 | **Base64 Decode to Shell Execution** | [proc_creation_lnx_base64_execution.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_base64_execution.yml) | T1140 | base64 -d piped to bash/sh. Extremely common in malware stagers and initial access. |

### 2.2 High Priority

| # | Gap | Sigma Rule | MITRE | Description |
|---|-----|-----------|-------|-------------|
| L11 | **Clear Linux Logs** | [proc_creation_lnx_clear_logs.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_clear_logs.yml) | T1070.002 | rm/shred/unlink targeting /var/log. Direct anti-forensics. |
| L12 | **History File Deletion** | [proc_creation_lnx_susp_history_delete.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_susp_history_delete.yml) | T1565.001 | rm/unlink/shred targeting .bash_history, .zsh_history. Attacker covering tracks. |
| L13 | **Credential File Access (passwd/shadow)** | [proc_creation_lnx_susp_sensitive_file_access.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_susp_sensitive_file_access.yml) | T1565.001 | Editing/modifying /etc/passwd, /etc/sudoers, /etc/crontab, /boot/, /bin/, /sbin/. |
| L14 | **Copy passwd/shadow to /tmp** | [proc_creation_lnx_cp_passwd_or_shadow_tmp.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_cp_passwd_or_shadow_tmp.yml) | T1552.001 | cp command staging credential files. Direct credential theft indicator. |
| L15 | **Shell Profile Persistence** | [lnx_auditd_unix_shell_configuration_modification.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/auditd/path/lnx_auditd_unix_shell_configuration_modification.yml) | T1546.004 | Modification of .bashrc, .bash_profile, .zshrc, /etc/profile. |
| L16 | **Cron File Persistence** | [file_event_lnx_persistence_cron_files.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/file_event/file_event_lnx_persistence_cron_files.yml) | T1053.003 | File creation in /etc/cron.d/, cron.daily, /var/spool/cron/crontabs. |
| L17 | **Sudoers Persistence** | [file_event_lnx_persistence_sudoers_files.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/file_event/file_event_lnx_persistence_sudoers_files.yml) | T1053.003 | File creation in /etc/sudoers.d/. Direct privilege persistence. |
| L18 | **SUID/SGID Bit Setting** | [proc_creation_lnx_setgid_setuid.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_setgid_setuid.yml) | T1548.001 | chown root + chmod u+s/g+s creating SUID/SGID backdoors. |
| L19 | **Execution from /tmp** | [proc_creation_lnx_susp_execution_tmp_folder.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_susp_execution_tmp_folder.yml) | T1036 | Any process with Image path starting with /tmp/. Common malware staging location. |
| L20 | **Download and Execute from /tmp** | [proc_creation_lnx_curl_wget_exec_tmp.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_curl_wget_exec_tmp.yml) | T1059.004, T1203 | curl/wget downloading to /tmp or /dev/shm then executing. Multi-stage stager pattern. |
| L21 | **Security Tools Disabling** | [proc_creation_lnx_security_tools_disabling.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_security_tools_disabling.yml) | T1562.004 | Stopping iptables, firewalld, cbdaemon, falcon-sensor, SELinux via service/systemctl. |
| L22 | **Crypto Mining Indicators** | [proc_creation_lnx_crypto_mining.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_crypto_mining.yml) | T1496 | Mining arguments (--cpu-priority, --donate-level, stratum+tcp://, --algo=rx/0). |
| L23 | **Crypto Mining Pool Connections** | [net_connection_lnx_crypto_mining_indicators.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/network_connection/net_connection_lnx_crypto_mining_indicators.yml) | T1496 | Connections to 22+ known Monero mining pool hostnames. |
| L24 | **ESXi VM Kill (Ransomware)** | [proc_creation_lnx_esxcli_vm_kill.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_esxcli_vm_kill.yml) | T1529 | esxcli vm process kill. Pre-encryption ransomware step (ESXiArgs). |
| L25 | **Python PTY Spawn (Shell Upgrade)** | [proc_creation_lnx_python_pty_spawn.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_python_pty_spawn.yml) | T1059 | Python import pty; pty.spawn. Used to upgrade shell to full interactive TTY. |
| L26 | **BPFDoor Backdoor** | [lnx_auditd_bpfdoor_file_accessed.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/auditd/path/lnx_auditd_bpfdoor_file_accessed.yml) | T1106 | Access to BPFDoor-specific files. Known APT backdoor. |
| L27 | **AWS SSM Agent Hijacking** | [proc_creation_lnx_ssm_agent_abuse.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_ssm_agent_abuse.yml) | T1219.002 | Unauthorized re-registration of amazon-ssm-agent. Real AWS attack technique. |
| L28 | **Setuid Capability via Setcap** | [proc_creation_lnx_cap_setuid.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_cap_setuid.yml) | T1548 | setcap cap_setuid on binaries. Capability-based privilege escalation. |
| L29 | **Ngrok Tunnel Communication** | [net_connection_lnx_ngrok_tunnel.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/network_connection/net_connection_lnx_ngrok_tunnel.yml) | T1572, T1090 | Connections to tunnel.*.ngrok.com. Popular tunneling tool abused by attackers. |
| L30 | **Malware Callback Port Communication** | [net_connection_lnx_susp_malware_callback_port.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/network_connection/net_connection_lnx_susp_malware_callback_port.yml) | T1571 | Outbound to known C2 ports (4444, 6789, 8531, etc.) to non-private IPs. |

### 2.3 macOS-Specific Gaps

| # | Gap | Sigma Rule | MITRE | Description |
|---|-----|-----------|-------|-------------|
| M1 | **GUI Input Capture (Fake Dialogs)** | [proc_creation_macos_gui_input_capture.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_gui_input_capture.yml) | T1056.002 | osascript displaying fake password dialogs for credential theft. |
| M2 | **Hidden User Account Creation** | [proc_creation_macos_create_hidden_account.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_create_hidden_account.yml) | T1564.002 | dscl create with UniqueID < 500 or IsHidden=true. |
| M3 | **Admin Group Addition via dscl** | [proc_creation_macos_dscl_add_user_to_admin_group.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_dscl_add_user_to_admin_group.yml) | T1078.003 | dscl -append /Groups/admin GroupMembership. |
| M4 | **Security Tools Disabling (macOS)** | [proc_creation_macos_disable_security_tools.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_disable_security_tools.yml) | T1562.001 | launchctl unload of LuLu, BlockBlock, Santa, CarbonBlack, CrowdStrike, etc. |
| M5 | **System Log Clearing** | [proc_creation_macos_clear_system_logs.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_clear_system_logs.yml) | T1070.002 | rm/unlink/shred targeting /var/log or ~/Library/Logs/. |
| M6 | **JXA In-Memory Execution** | [proc_creation_macos_jxa_in_memory_execution.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_jxa_in_memory_execution.yml) | T1059.002 | osascript -e with eval + NSData.dataWithContentsOfURL. Fileless attack. |
| M7 | **Suspicious Browser Child Process** | [proc_creation_macos_susp_browser_child_process.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_susp_browser_child_process.yml) | T1189, T1203 | Safari/Chrome/Firefox spawning bash, curl, wget, python. Browser exploitation. |
| M8 | **Suspicious Office Child Process** | [proc_creation_macos_office_susp_child_processes.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_office_susp_child_processes.yml) | T1204.002 | Word/Excel spawning bash, curl, wget, python. Macro exploitation on macOS. |
| M9 | **PlistBuddy Persistence** | [proc_creation_macos_persistence_via_plistbuddy.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_persistence_via_plistbuddy.yml) | T1543.001 | PlistBuddy setting RunAtLoad=true in LaunchAgents/LaunchDaemons. |
| M10 | **Keychain Credential Dumping** | [proc_creation_macos_creds_from_keychain.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_creds_from_keychain.yml) | T1555.001 | security find-certificate, export, dump-keychain commands. |

### 2.4 Medium Priority (Linux/macOS)

| # | Gap | Sigma Rule | MITRE | Description |
|---|-----|-----------|-------|-------------|
| L31 | Perl reverse shell | [proc_creation_lnx_perl_reverse_shell.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_perl_reverse_shell.yml) | T1059 | Perl fdopen+Socket reverse shell pattern |
| L32 | Ruby reverse shell | [proc_creation_lnx_ruby_reverse_shell.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_ruby_reverse_shell.yml) | T1059 | Ruby TCPSocket reverse shell |
| L33 | Immutable flag removal (chattr -i) | [proc_creation_lnx_chattr_immutable_removal.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_chattr_immutable_removal.yml) | T1222.002 | Used by ransomware/rootkits |
| L34 | Timestomping service files | [proc_creation_lnx_touch_susp.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_touch_susp.yml) | T1070.006 | touch -t on .service files |
| L35 | DD process injection | [proc_creation_lnx_dd_process_injection.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_dd_process_injection.yml) | T1055.009 | dd writing to /proc/PID/mem |
| L36 | Named pipe in /tmp (mkfifo) | [proc_creation_lnx_mkfifo_named_pipe_creation_susp_location.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_mkfifo_named_pipe_creation_susp_location.yml) | T1059 | Associated with Barracuda ESG exploitation |
| L37 | Container/Docker discovery | [proc_creation_lnx_susp_container_residence_discovery.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_susp_container_residence_discovery.yml) | T1082 | Reading /proc/*/cgroup for container detection |
| L38 | Shellshock expression | [lnx_shellshock.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/builtin/lnx_shellshock.yml) | T1505.003 | () { :; }; pattern in logs |
| L39 | OMIGOD SCX vulnerability | [proc_creation_lnx_omigod_scx_runasprovider_executeshellcommand.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_omigod_scx_runasprovider_executeshellcommand.yml) | T1068, T1190 | CVE-2021-38647 Azure OMI exploitation |
| L40 | Triple Cross eBPF rootkit | [proc_creation_lnx_triple_cross_rootkit_install.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_triple_cross_rootkit_install.yml) | T1014 | sudo tc + qdisc/filter pattern |
| L41 | Root certificate installation | [proc_creation_lnx_install_root_certificate.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_install_root_certificate.yml) | T1553.004 | update-ca-certificates execution |
| L42 | Suspicious package installation | [proc_creation_lnx_install_suspicious_packages.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/process_creation/proc_creation_lnx_install_suspicious_packages.yml) | T1553.004 | Installing nmap, netcat, wireshark, proxychains, socat |
| L43 | Credential search in files | [lnx_auditd_find_cred_in_files.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/auditd/execve/lnx_auditd_find_cred_in_files.yml) | T1552.001 | grep/find searching for passwords/tokens |
| L44 | Profile.d script persistence | [file_event_lnx_susp_shell_script_under_profile_directory.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/linux/file_event/file_event_lnx_susp_shell_script_under_profile_directory.yml) | T1546.004 | Shell scripts in /etc/profile.d/ |
| M11 | macOS Gatekeeper bypass via xattr | [proc_creation_macos_xattr_gatekeeper_bypass.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_xattr_gatekeeper_bypass.yml) | T1553.001 | xattr -d com.apple.quarantine |
| M12 | macOS root account enable | [proc_creation_macos_dsenableroot_enable_root_account.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_dsenableroot_enable_root_account.yml) | T1078 | dsenableroot execution |
| M13 | macOS Time Machine backup deletion | [proc_creation_macos_tmutil_delete_backup.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_tmutil_delete_backup.yml) | T1490 | tmutil delete - ransomware preparation |
| M14 | macOS WizardUpdate malware | [proc_creation_macos_wizardupdate_malware_infection.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_wizardupdate_malware_infection.yml) | C2 | Known macOS trojan patterns |
| M15 | macOS XCSSET malware | [proc_creation_macos_xcsset_malware_infection.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/macos/process_creation/proc_creation_macos_xcsset_malware_infection.yml) | C2 | Xcode-targeting malware |

---

## 3. Cloud/SaaS Detection Gaps

**Current coverage:** ~97 rules | **Gaps identified:** 47

Good coverage for basic AWS/Azure/GCP operations. Main gaps: Golden SAML attacks, Azure Identity Protection signals, defense evasion (alert/finding suppression), Azure AD-specific persistence, container/Kubernetes attacks.

### 3.1 Critical Priority

| # | Gap | Sigma Rule | MITRE | Why Missing This Matters |
|---|-----|-----------|-------|--------------------------|
| C1 | **AWS Golden SAML Attack** | [aws_susp_saml_activity.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_susp_saml_activity.yml) | T1078.004, T1556 | AssumeRoleWithSAML and UpdateSAMLProvider. This is the technique used in the SolarWinds attack. |
| C2 | **Azure Federation Modified** | [azure_federation_modified.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/audit_logs/azure_federation_modified.yml) | T1556 | Domain federation modification. Critical Golden SAML / domain takeover technique. |
| C3 | **Azure Subscription Permission Elevation** | [azure_subscription_permissions_elevation_via_activitylogs.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/activity_logs/azure_subscription_permissions_elevation_via_activitylogs.yml) | T1078.004 | ELEVATEACCESS action granting access to ALL Azure subscriptions. Extremely high impact. |
| C4 | **AWS KMS Key Material Import** | [aws_kms_import_key_material.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_kms_import_key_material.yml) | T1486 | ImportKeyMaterial/DeleteImportedKeyMaterial. Very rare operation, high ransomware correlation. |
| C5 | **Azure PRT Access Attempt** | [azure_identity_protection_prt_access.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/identity_protection/azure_identity_protection_prt_access.yml) | T1528 | Primary Refresh Token access. Rare, high-confidence compromise indicator. |
| C6 | **M365 Federated Domain Added** | [microsoft365_new_federated_domain_added_audit.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/m365/audit/microsoft365_new_federated_domain_added_audit.yml) | T1556 | Adding federated domains enables Golden SAML and persistent backdoor access. |
| C7 | **AWS S3 Versioning Disabled** | [aws_disable_bucket_versioning.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_disable_bucket_versioning.yml) | T1490 | PutBucketVersioning with "Suspended". Cloud ransomware preparation. |
| C8 | **Azure App Credential Added** | [azure_app_credential_added.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/audit_logs/azure_app_credential_added.yml) | T1098.001 | New certs/secrets added to Azure AD apps. Primary Azure persistence technique. |
| C9 | **AWS SecurityHub Evasion** | [aws_securityhub_finding_evasion.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_securityhub_finding_evasion.yml) | T1562 | BatchUpdateFindings, DeleteInsight. Attackers suppressing security findings. |
| C10 | **Azure Alert Suppression** | [azure_suppression_rule_created.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/activity_logs/azure_suppression_rule_created.yml) | T1562 | Creating rules to suppress security alerts. Critical defense evasion. |

### 3.2 High Priority

| # | Gap | Sigma Rule | MITRE | Description |
|---|-----|-----------|-------|-------------|
| C11 | **Azure New Root CA Added** | [azure_ad_new_root_ca_added.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/audit_logs/azure_ad_new_root_ca_added.yml) | T1556 | Root CA for passwordless auth persistence. |
| C12 | **Azure Anomalous Token** | [azure_identity_protection_anomalous_token.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/identity_protection/azure_identity_protection_anomalous_token.yml) | T1528 | Tokens with unusual lifetime or unfamiliar locations. |
| C13 | **Azure Leaked Credentials** | [azure_identity_protection_leaked_credentials.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/identity_protection/azure_identity_protection_leaked_credentials.yml) | T1078 | Credentials found on dark web / paste sites. |
| C14 | **Azure Device Code Auth Abuse** | [azure_app_device_code_authentication.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/signin_logs/azure_app_device_code_authentication.yml) | T1078 | Device code phishing. Growing attack vector. |
| C15 | **Azure App Privileged Permissions** | [azure_app_privileged_permissions.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/audit_logs/azure_app_privileged_permissions.yml) | T1098.003 | Illicit consent grant attacks. |
| C16 | **AWS Config Disabled** | [aws_config_disable_recording.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_config_disable_recording.yml) | T1562.008 | DeleteDeliveryChannel, StopConfigurationRecorder. |
| C17 | **AWS SSO IdP Change** | [aws_sso_idp_change.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_sso_idp_change.yml) | T1556 | IdP changes enable user impersonation. |
| C18 | **GCP Domain API Access Granted** | [gcp_gworkspace_granted_domain_api_access.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/gcp/gworkspace/gcp_gworkspace_granted_domain_api_access.yml) | T1098 | Domain-wide delegation to service accounts. |
| C19 | **AWS IAM Login Profile Modification** | [aws_update_login_profile.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_update_login_profile.yml) | T1098 | Password change for another account (modifier != target). |
| C20 | **AWS EC2 Startup Script Modification** | [aws_ec2_startup_script_change.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_ec2_startup_script_change.yml) | T1059 | ModifyInstanceAttribute targeting userData. Persistent root/SYSTEM execution. |
| C21 | **Azure Password Spray** | [azure_identity_protection_password_spray.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/identity_protection/azure_identity_protection_password_spray.yml) | T1110 | Microsoft ML-based password spray detection. |
| C22 | **Azure Impossible Travel** | [azure_identity_protection_impossible_travel.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/identity_protection/azure_identity_protection_impossible_travel.yml) | T1078 | Geographically impossible sign-in locations. |
| C23 | **Azure Cert-Based Auth Enabled** | [azure_ad_certificate_based_authencation_enabled.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/audit_logs/azure_ad_certificate_based_authencation_enabled.yml) | T1556 | Passwordless persistence in Azure AD. |
| C24 | **Azure TAP Added** | [azure_tap_added.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/audit_logs/azure_tap_added.yml) | T1078.004 | Temporary Access Pass bypass for MFA circumvention. |
| C25 | **AzureHound Discovery** | [azure_ad_azurehound_discovery.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/signin_logs/azure_ad_azurehound_discovery.yml) | T1087.004, T1526 | BloodHound for Azure AD reconnaissance. |
| C26 | **Azure ROPC Authentication** | [azure_app_ropc_authentication.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/signin_logs/azure_app_ropc_authentication.yml) | T1078 | Resource Owner Password Credentials flow. Credentials exposed to apps. |

### 3.3 Medium Priority

| # | Gap | Sigma Rule | MITRE | Description |
|---|-----|-----------|-------|-------------|
| C27 | AWS GetSigninToken abuse | [aws_console_getsignintoken.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_console_getsignintoken.yml) | T1550.001 | Federated credential pivot, MFA bypass |
| C28 | AWS ECS credential theft | [aws_ecs_task_definition_cred_endpoint_query.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_ecs_task_definition_cred_endpoint_query.yml) | T1525 | Container credential endpoint abuse |
| C29 | AWS Glue privesc | [aws_passed_role_to_glue_development_endpoint.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_passed_role_to_glue_development_endpoint.yml) | T1078.004 | Glue dev endpoint privilege escalation |
| C30 | AWS Snapshot exfiltration | [aws_snapshot_backup_exfiltration.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_snapshot_backup_exfiltration.yml) | T1537 | EC2 snapshot sharing with external accounts |
| C31 | AWS SSM SendCommand | [aws_cloudtrail_ssm_malicious_usage.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_cloudtrail_ssm_malicious_usage.yml) | T1566 | Remote command execution on instances |
| C32 | AWS Lambda URL creation | [aws_lambda_function_url.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_lambda_function_url.yml) | T1078.004 | Internet-accessible Lambda with IAM role |
| C33 | AWS RDS public restore | [aws_rds_public_db_restore.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_rds_public_db_restore.yml) | T1020 | DB snapshot restored as publicly accessible |
| C34 | AWS TruffleHog scanning | [aws_cloudtrail_pua_trufflehog.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/aws/cloudtrail/aws_cloudtrail_pua_trufflehog.yml) | T1555 | Automated credential scanning tool |
| C35 | Azure K8s events deleted | [azure_kubernetes_events_deleted.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/activity_logs/azure_kubernetes_events_deleted.yml) | T1562.001 | K8s event deletion for defense evasion |
| C36 | Azure K8s admission controller | [azure_kubernetes_admission_controller.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/activity_logs/azure_kubernetes_admission_controller.yml) | T1078.004 | Malicious container injection |
| C37 | Azure K8s secret access | [azure_kubernetes_secret_or_config_object_access.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/activity_logs/azure_kubernetes_secret_or_config_object_access.yml) | T1485 | K8s Secrets write/delete |
| C38 | Azure LAPS credential dump | [azure_auditlogs_laps_credential_dumping.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/audit_logs/azure_auditlogs_laps_credential_dumping.yml) | T1098.005 | Local admin password recovery from Entra ID |
| C39 | Azure bulk role changes | [azure_priviledged_role_assignment_bulk_change.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/azure/audit_logs/azure_priviledged_role_assignment_bulk_change.yml) | T1098 | Mass privilege assignment changes |
| C40 | GCP break-glass container | [gcp_breakglass_container_workload_deployed.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/gcp/audit/gcp_breakglass_container_workload_deployed.yml) | T1548 | Binary Authorization bypass |
| C41 | GCP DLP re-identification | [gcp_dlp_re_identifies_sensitive_information.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/gcp/audit/gcp_dlp_re_identifies_sensitive_information.yml) | T1565 | Reversing data de-identification |
| C42 | GCP packet capture | [gcp_full_network_traffic_packet_capture.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/gcp/audit/gcp_full_network_traffic_packet_capture.yml) | T1074 | PacketMirrorings for credential capture |
| C43 | GCP K8s admission controller | [gcp_kubernetes_admission_controller.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/gcp/audit/gcp_kubernetes_admission_controller.yml) | T1078.004 | GKE admission webhook modification |
| C44 | GCP Workspace MFA disabled | [gcp_gworkspace_mfa_disabled.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/gcp/gworkspace/gcp_gworkspace_mfa_disabled.yml) | T1556 | MFA enforcement disabled |
| C45 | M365 data exfiltration to unsanctioned apps | [microsoft365_data_exfiltration_to_unsanctioned_app.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/m365/threat_management/microsoft365_data_exfiltration_to_unsanctioned_app.yml) | T1537 | CASB alert for exfiltration |
| C46 | M365 PST export | [microsoft365_pst_export_alert.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/m365/threat_management/microsoft365_pst_export_alert.yml) | T1114 | Bulk email exfiltration |
| C47 | M365 OAuth app file downloads | [microsoft365_susp_oauth_app_file_download_activities.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/cloud/m365/threat_management/microsoft365_susp_oauth_app_file_download_activities.yml) | T1530 | Malicious OAuth app mass downloads |

---

## 4. Network/Web/Proxy Detection Gaps

**Current coverage:** ~200+ rules | **Gaps identified:** 32

Our network coverage is extensive for firewall events, IDS/IPS, web attacks (SQLi, XSS, CSRF), and DNS anomalies. Main gaps: **proxy-level C2/malware user agent detection** (biggest gap), IOC-based DNS rules, web exploitation signatures (SSTI, JNDI), and exfiltration tool detection.

### 4.1 Critical Priority

| # | Gap | Sigma Rule | MITRE | Why Missing This Matters |
|---|-----|-----------|-------|--------------------------|
| N1 | **Cobalt Strike DNS Beacon** | [net_dns_mal_cobaltstrike.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/network/dns/net_dns_mal_cobaltstrike.yml) | T1071.004 | Specific CS DNS beacon patterns (`aaa.stage.*`, `post.1*`). Catches C2 before encrypted payload. |
| N2 | **DNS OOB Interaction Domains** | [net_dns_external_service_interaction_domains.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/network/dns/net_dns_external_service_interaction_domains.yml) | T1190, T1595.002 | Queries to Burp Collaborator, interactsh, dnslog.cn, canarytokens.com, ~25 more. Almost exclusively active exploitation callbacks. |
| N3 | **CS Malleable C2 Profiles (Proxy)** | [proxy_hktl_cobalt_strike_malleable_c2_requests.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_hktl_cobalt_strike_malleable_c2_requests.yml) | T1071.001 | Known CS malleable C2 profile URI/UA/Host combos (Amazon, OCSP, OneDrive profiles). Very low FP. |
| N4 | **Exploit Framework User Agents** | [proxy_ua_frameworks.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_frameworks.yml) | T1071.001 | UAs from CS, Metasploit, Empire, Havoc, 25+ C2 frameworks. |
| N5 | **Malware User Agent Strings** | [proxy_ua_malware.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_malware.yml) | T1071.001 | UAs from 50+ malware families: PlugX, njRAT, Emotet, Lockbit, Raccoon, Quasar, AntSword, SparkRAT, Latrodectus. |
| N6 | **APT User Agent Strings** | [proxy_ua_apt.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_apt.yml) | T1071.001 | UAs from 30+ APT groups: APT28, APT17, Winnti, Mustang Panda, OceanLotus, RedCurl. Nation-state signatures. |
| N7 | **Rclone Exfiltration** | [proxy_ua_rclone.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_rclone.yml) | T1567.002 | Rclone/v user agent. Most commonly used exfil tool in ransomware operations. |

### 4.2 High Priority

| # | Gap | Sigma Rule | MITRE | Description |
|---|-----|-----------|-------|-------------|
| N8 | **DNS TXT Execution Strings** | [net_dns_susp_txt_exec_strings.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/network/dns/net_dns_susp_txt_exec_strings.yml) | T1071.004 | TXT records containing IEX, Invoke-Expression, cmd.exe. DNS-based C2 payload delivery. |
| N9 | **Base64 DNS Queries** | [net_dns_susp_b64_queries.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/network/dns/net_dns_susp_b64_queries.yml) | T1048.003 | DNS queries with `==.` indicating base64-encoded exfiltration. |
| N10 | **Hack Tool User Agents** | [proxy_ua_hacktool.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_hacktool.yml) | T1190, T1110 | UAs from Hydra, dirbuster, sqlmap, masscan, Nikto. 40+ scanner signatures. |
| N11 | **Raw Paste Service Access** | [proxy_raw_paste_service_access.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_raw_paste_service_access.yml) | T1102.001 | Raw content from pastebin, paste.ee, hastebin. Second-stage payload download. |
| N12 | **SSTI Detection** | [web_ssti_in_access_logs.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/webserver_generic/web_ssti_in_access_logs.yml) | T1221 | Template injection payloads (`={{`, `=${`, freemarker.template.utility.Execute). |
| N13 | **JNDI Exploit Kit (Log4Shell)** | [web_jndi_exploit.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/webserver_generic/web_jndi_exploit.yml) | T1190 | JNDI-Exploit-Kit patterns (/Basic/Command/Base64/, /Deserialization/*). |
| N14 | **Java Payload Strings** | [web_java_payload_in_access_logs.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/webserver_generic/web_java_payload_in_access_logs.yml) | T1190 | OGNL injection, getRuntime().exec(), Confluence CVEs. |
| N15 | **ReGeorg Webshell** | [web_webshell_regeorg.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/webserver_generic/web_webshell_regeorg.yml) | T1505.003 | ReGeorg tunneling webshell commands (cmd=read, cmd=connect). |
| N16 | **Webshell Command Strings** | [web_win_webshells_in_access_logs.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/webserver_generic/web_win_webshells_in_access_logs.yml) | T1505.003 | Webshell commands in GET requests (=whoami, =net%20user, =cmd%20/c). |
| N17 | **BITSAdmin to Uncommon TLD** | [proxy_ua_bitsadmin_susp_tld.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_bitsadmin_susp_tld.yml) | T1197 | Microsoft BITS/ user agent to non-standard TLDs. |
| N18 | **Download from DynDNS** | [proxy_download_susp_dyndns.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_download_susp_dyndns.yml) | T1105, T1568 | Executable downloads from 70+ DynDNS providers. |
| N19 | **Tor Proxy DNS Lookups** | [zeek_dns_torproxy.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/network/zeek/zeek_dns_torproxy.yml) | T1048 | 45+ Tor proxy/gateway domains. |
| N20 | **DNS Mining Pool Lookups** | [zeek_dns_mining_pools.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/network/zeek/zeek_dns_mining_pools.yml) | T1496 | 60+ mining pool domains (moneropool, xmrpool, minergate, nanopool). |
| N21 | **WebDAV External Execution** | [proxy_webdav_external_execution.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_webdav_external_execution.yml) | T1566 | Executable download via external WebDAV (CVE-2024-21412 technique). |

### 4.3 Medium Priority

| # | Gap | Sigma Rule | MITRE | Description |
|---|-----|-----------|-------|-------------|
| N22 | Download from suspicious TLDs | [proxy_download_susp_tlds_blacklist.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_download_susp_tlds_blacklist.yml) | T1566 | Executables from .xyz, .top, .tk, .zip TLDs |
| N23 | NKN blockchain C2 | [zeek_dns_nkn.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/network/zeek/zeek_dns_nkn.yml) | T1071 | Blockchain-based C2 via NKN seed nodes |
| N24 | WannaCry killswitch | [net_dns_wannacry_killswitch_domain.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/network/dns/net_dns_wannacry_killswitch_domain.yml) | T1071.001 | Active WannaCry infection indicator |
| N25 | PwnDrop file hosting | [proxy_pwndrop.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_pwndrop.yml) | T1102 | Red team/criminal file hosting tool |
| N26 | Baby Shark C2 | [proxy_hktl_baby_shark_default_agent_url.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_hktl_baby_shark_default_agent_url.yml) | T1071.001 | Specific C2 framework detection |
| N27 | Suspicious/malformed user agents | [proxy_ua_susp.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_susp.yml) | T1071.001 | Typos ("Mozila"), CertUtil, truncated UAs |
| N28 | PowerShell user agent | [proxy_ua_powershell.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_powershell.yml) | T1071.001 | WindowsPowerShell/ in proxy logs |
| N29 | Crypto miner user agent | [proxy_ua_cryptominer.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_cryptominer.yml) | T1071.001 | XMRig/CCMiner UAs |
| N30 | Base64 encoded user agent | [proxy_ua_base64_encoded.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_ua_base64_encoded.yml) | T1071.001 | YamaBot-style base64 UA evasion |
| N31 | Git repository exposure | [web_source_code_enumeration.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/webserver_generic/web_source_code_enumeration.yml) | T1083 | .git/ directory access on web servers |
| N32 | IPFS credential harvesting | [proxy_susp_ipfs_cred_harvest.yml](https://github.com/SigmaHQ/sigma/blob/master/rules/web/proxy_generic/proxy_susp_ipfs_cred_harvest.yml) | T1056 | Phishing via IPFS URLs with email parameters |

---

## Summary Statistics

| Category | Existing Rules | Gaps Found | Critical | High | Medium |
|----------|---------------|------------|----------|------|--------|
| Windows | 28 | 25 | 8 | 8 | 9 |
| Linux/macOS | ~85 | ~50 | 10 | 20 | 20 |
| Cloud/SaaS | ~97 | 47 | 10 | 16 | 21 |
| Network/Web | ~200+ | 32 | 7 | 14 | 11 |
| **Total** | **~410** | **~154** | **35** | **58** | **61** |

## Top 20 Rules to Implement First

Based on real-world attack frequency and detection confidence:

1. **W1** - Kerberoasting (every AD compromise)
2. **W3** - Scheduled Task Persistence (every attacker uses this)
3. **L1** - Comprehensive Reverse Shell Detection (Linux/macOS fundamental gap)
4. **N5** - Malware User Agent Strings (50+ malware families, high confidence)
5. **N6** - APT User Agent Strings (nation-state IOCs)
6. **W5** - Impacket Lateral Movement (most-used framework)
7. **C1** - AWS Golden SAML (SolarWinds-class attack)
8. **C3** - Azure Subscription Permission Elevation (full tenant access)
9. **L6** - Webshell Detection (active exploitation)
10. **N1** - Cobalt Strike DNS Beacon (specific C2 signature)
11. **W4** - Suspicious Service Installation (persistence/privesc)
12. **W6** - Certutil LOLBIN Abuse (top download tool)
13. **N2** - DNS OOB Interaction Domains (RCE indicator)
14. **L7** - Linux HackTool Execution (50+ tools)
15. **W7** - Registry Run Key Persistence (fundamental persistence)
16. **C7** - AWS S3 Versioning Disabled (ransomware prep)
17. **N7** - Rclone Exfiltration (ransomware data theft)
18. **W8** - LSASS Dumping variants (beyond Mimikatz)
19. **L8** - Java Child Processes (Log4Shell/Confluence)
20. **C2** - Azure Federation Modified (domain takeover)
