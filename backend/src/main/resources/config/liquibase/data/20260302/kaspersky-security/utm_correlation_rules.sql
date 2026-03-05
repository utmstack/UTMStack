INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1160, 'Lateral Movement Indicators Detection', 3, 3, 2, 'Lateral Movement', 'T1021 - Remote Services', e'Detects indicators of lateral movement attempts within the network through Kaspersky antivirus logs.
Attackers use various techniques including PSExec, WMI, RDP, SMB shares, and exploitation tools to move
from one compromised system to others, expanding their access and control across the network. This rule
identifies blocked or detected activities that may indicate lateral movement attempts.

Next Steps:
1. Investigate the source IP and hostname for signs of compromise
2. Review authentication logs for the same time period to identify potential credential theft
3. Check if the detected tools (PSExec, WMI, RDP) are authorized for use in your environment
4. Examine network traffic between the source and destination systems
5. Look for other suspicious activities from the same source host
6. Consider isolating affected systems if lateral movement is confirmed
7. Review similar patterns from the same source within the detection window
', '["https://attack.mitre.org/tactics/TA0008/","https://support.kaspersky.com/KESWin/11/en-us/151065.htm"]', e'(equals("log.cn1", "3") || equals("log.cs1", "DETECT") || equals("log.act", "blocked")) &&
(contains("log.msg", ["psexec", "wmi", "rdp", "smb", "admin$", "ipc$", "c$",
  "remote", "lateral", "pivot"]) ||
 contains("log.cs4", ["exploit", "mimikatz", "bloodhound", "sharphound", "propagat"])) &&
exists("log.dst") &&
exists("log.src") &&
safe(log.src, "") != safe(log.dst, "")
', '2026-03-02 16:10:26.625080', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-kaspersky-*","with":[{"field":"log.src","operator":"filter_term","value":"{{.log.src}}"}],"or":null,"within":"now-2h","count":3}]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1161, 'Kaspersky Rootkit Detection', 3, 3, 3, 'Defense Evasion', 'T1014 - Rootkit', e'Detects rootkit activity identified by Kaspersky security, including hidden processes, kernel-level modifications, and rootkit-specific malware classifications that indicate a deeply compromised system.

Next Steps:
1. Immediately isolate the affected system
2. Do not trust any output from the compromised system
3. Perform offline forensic analysis
4. Plan for full system reimaging
5. Check for lateral movement from the compromised host
6. Determine the initial infection vector
', '["https://support.kaspersky.com/","https://attack.mitre.org/techniques/T1014/"]', e'exists("log.signatureID") &&
(regexMatch("log.msg", "(?i)(rootkit|bootkit|Rootkit|hidden.*process|hidden.*module)") ||
 contains("log.cs2", "Rootkit") || contains("log.cs4", "Rootkit") ||
 contains("log.cs2", "Bootkit") || contains("log.cs4", "Bootkit") ||
 (contains("log.msg", "System Analysis") && contains("log.msg", "hidden")))
', '2026-03-02 16:10:27.971187', true, true, 'origin', null, '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1162, 'Kaspersky Ransomware Behavior Detection', 3, 3, 3, 'Impact', 'T1486 - Data Encrypted for Impact', e'Detects ransomware behavior patterns identified by Kaspersky including mass file encryption, ransom note creation, and ransomware-specific malware classifications.

Next Steps:
1. Immediately isolate the affected system from the network
2. Identify the ransomware variant from Kaspersky\'s classification
3. Check backup availability and integrity
4. Do not pay the ransom
5. Engage incident response team
6. Scan other systems for the same indicators
7. Determine the initial infection vector
', '["https://support.kaspersky.com/","https://attack.mitre.org/techniques/T1486/"]', e'exists("log.signatureID") &&
(regexMatch("log.msg", "(?i)(ransomware|ransom|trojan-ransom|cryptolocker|locky|cerber|wannacry|ryuk|conti|lockbit|blackcat)") ||
 contains("log.cs2", "Trojan-Ransom") || contains("log.cs4", "Trojan-Ransom") ||
 (contains("log.msg", "encrypt") && contains("log.msg", "mass")) ||
 (contains("log.msg", "System Watcher") && contains("log.msg", "rollback")))
', '2026-03-02 16:10:29.339649', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-kaspersky-*","with":[{"field":"log.src","operator":"filter_term","value":"{{.log.src}}"}],"or":null,"within":"now-10m","count":3}]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1163, 'Kaspersky Agent Disabled or Tampered', 3, 3, 3, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects when the Kaspersky security agent is disabled, stopped, or tampered with. This is a critical indicator of defense evasion as attackers disable endpoint protection to execute malware undetected.

Next Steps:
1. Immediately investigate the affected endpoint
2. Identify the user or process that disabled the agent
3. Check for concurrent malicious activity
4. Re-enable the Kaspersky agent
5. Perform a full system scan
6. Check for similar events on other endpoints
', '["https://support.kaspersky.com/","https://attack.mitre.org/techniques/T1562/001/"]', e'exists("log.signatureID") &&
(regexMatch("log.msg", "(?i)(kaspersky|klnagent|kavfs|kesl).*( disabled| stopped| removed| tampered| uninstalled)") ||
 regexMatch("log.msg", "(?i)(protection|self-defense).*(disabled|off|stopped)") ||
 (contains("log.cs1", "PROTECTION") && contains("log.message", "disabled")) ||
 (contains("log.msg", "agent") && contains("log.message", "not running")))
', '2026-03-02 16:10:30.733359', true, true, 'origin', null, '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1164, 'Kaspersky Data Exfiltration Attempts Detection', 3, 2, 1, 'Exfiltration', 'T1048 - Exfiltration Over Alternative Protocol', e'Detects potential data exfiltration attempts identified by Kaspersky through suspicious network traffic patterns, large data transfers, or connections to suspicious external destinations. This rule monitors for network threats, trojan/backdoor detections, and suspicious data transfer patterns that may indicate data exfiltration.

Next Steps:
1. Immediately identify the source host (origin.ip) and any associated user accounts on the affected system
2. Check if the destination IP (target.ip) is known malicious using threat intelligence sources
3. Review the volume and frequency of data transfers to this destination in the last 24-48 hours
4. Search for any other malware detections (especially Trojans/Backdoors) on the same host
5. Analyze network traffic logs for unusual patterns or protocols from the source IP
6. Check if other hosts in your network have connected to the same destination
7. If confirmed malicious:
   - Block the destination IP at firewall/proxy level
   - Isolate the affected system from network
   - Initiate full incident response procedures
   - Preserve evidence for forensic analysis
8. Document all findings and actions taken for compliance and future reference
', '["https://attack.mitre.org/techniques/T1048/","https://support.kaspersky.com/KLMS/8.2/en-US/151684.htm"]', e'(equals("log.cat", "NetworkThreat") ||
 regexMatch("log.cs2", "(?i).*(trojan|backdoor).*") ||
 regexMatch("log.msg", "(?i).*(data.*transfer|exfiltrat|upload.*suspicious|unauthorized.*transfer).*") ||
 regexMatch("log.msg", "(?i).*(data.*exfiltration|suspicious.*upload|unauthorized.*transfer).*")) &&
exists("target.ip") &&
greaterOrEqual("log.cefDeviceSeverity", "3")
', '2026-03-02 16:10:32.000991', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-kaspersky-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.cat","operator":"filter_term","value":"NetworkThreat"}],"or":null,"within":"now-30m","count":5}]', '["origin.ip","target.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1165, 'Kaspersky Command and Control Communication Detection', 3, 3, 2, 'Command and Control', 'T1071 - Application Layer Protocol', e'Detects potential command and control (C2) communication attempts identified by Kaspersky, including suspicious outbound connections, malware callbacks, and botnet communication patterns. This rule triggers when Kaspersky identifies network threats related to botnet activity, C2 communications, or malware beaconing that was not successfully blocked.

Next Steps:
1. Immediately isolate the affected system from the network to prevent further C2 communication
2. Review the target IP address against threat intelligence feeds to confirm malicious activity
3. Check if other systems have communicated with the same C2 server
4. Analyze the process or malware that initiated the connection
5. Review Kaspersky logs for additional context about the threat
6. Perform a full system scan and forensic analysis on the affected machine
7. Update antivirus signatures and ensure real-time protection is enabled
8. Consider reimaging the system if compromise is confirmed
', '["https://attack.mitre.org/techniques/T1071/","https://support.kaspersky.com/KLMS/8.2/en-US/151504.htm"]', e'(contains("log.cs2", ["Bot", "bot", "C2", "Command", "command"]) ||
 contains("log.msg", ["callback", "beacon", "botnet", "command and control"]) ||
 equals("log.cat", "NetworkThreat")) &&
exists("target.ip") &&
action != "blocked" && action != "Blocked"
', '2026-03-02 16:10:33.401026', true, true, 'origin', null, '[]', '["origin.host","target.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1166, 'Code Injection Attempt Detection', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1055 - Process Injection', e'Detects attempts to inject malicious code into legitimate processes. This technique is commonly used by malware to evade detection and gain elevated privileges by running within trusted processes.

Next Steps:
1. Immediately isolate the affected system to prevent lateral movement
2. Identify the source process that attempted the injection
3. Check if the malware was successfully quarantined or if manual removal is needed
4. Review system logs for any suspicious activities around the same timeframe
5. Scan the system with updated antivirus definitions
6. Check for persistence mechanisms (scheduled tasks, registry keys, services)
7. Review network connections from the affected host for C2 communications
8. Consider reimaging the system if critical processes were compromised
', '["https://attack.mitre.org/techniques/T1055/","https://support.kaspersky.com/KESWin/11/en-us/151065.htm"]', e'(regexMatch("log.msg", "(?i).*(inject|injection|CreateRemoteThread|SetWindowsHookEx|WriteProcessMemory).*") ||
 (contains("log.cs4", ["inject", "hooking", "trojan", "backdoor"]) &&
  contains("action", ["terminate", "delete", "quarantine"]))) &&
contains("log.msg", ["lsass", "csrss", "winlogon", "services", "svchost", "explorer"])
', '2026-03-02 16:10:34.586175', true, true, 'origin', '["origin.host","log.cs4"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1167, 'Kaspersky Critical Object Detection', 3, 3, 2, 'Execution', 'T1204 - User Execution: Malicious File', e'Detects when Kaspersky identifies critical threats including malware, trojans, or other dangerous objects that require immediate attention. High severity detections often indicate active threats.

Next Steps:
1. Immediately isolate the affected system from the network to prevent lateral movement
2. Identify the malware name/signature from log.cs1, log.cs2, or log.cs4 fields
3. Check if Kaspersky successfully quarantined or removed the threat
4. Scan other systems in the same network segment for similar infections
5. Review recent user activity and email attachments that could be the infection vector
6. Collect and preserve forensic artifacts if needed for incident response
7. Update antivirus signatures and run a full system scan
8. Consider reimaging the system if the infection is severe or persistent
', '["https://support.kaspersky.com/ScanEngine/1.0/en-US/186767.htm","https://attack.mitre.org/techniques/T1204/"]', e'(greaterOrEqual("log.cefDeviceSeverity", "7") ||
 equalsIgnoreCase("log.cefDeviceSeverity", "High") ||
 equalsIgnoreCase("log.cefDeviceSeverity", "Very-High")) &&
(contains("log.cs1", ["INFECTED", "MALWARE", "TROJAN"]) ||
 contains("log.cs2", ["Trojan", "HEUR:", "PDM:", "UDS:"]) ||
 contains("log.cs4", ["Trojan", "HEUR:", "PDM:", "UDS:"]) ||
 contains("log.msg", ["infected", "malicious", "dangerous"]))
', '2026-03-02 16:10:35.811944', true, true, 'origin', null, '[]', '["origin.host","log.signatureID"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1149, 'Kaspersky WMI Abuse Detection', 3, 3, 2, 'Execution', 'T1047 - Windows Management Instrumentation', e'Detects potential Windows Management Instrumentation (WMI) abuse identified by Kaspersky, including suspicious WMI queries, event subscriptions, or process creation via WMI. WMI is a legitimate Windows component often abused by attackers for lateral movement, persistence, and code execution.

Next Steps:
1. Identify the affected host and user account involved in the WMI activity
2. Review the specific WMI commands or queries that triggered the alert
3. Check for any unauthorized scheduled tasks or startup items created via WMI
4. Look for other indicators of compromise on the affected system
5. Verify if this is legitimate administrative activity or potential malicious behavior
6. If confirmed malicious, isolate the system and perform incident response procedures
', '["https://attack.mitre.org/techniques/T1047/","https://support.kaspersky.com/KLMS/8.2/en-US/151684.htm"]', e'(contains("log.msg", ["WMI", "wmi", "wmic", "winmgmt", "scrcons.exe"]) ||
 contains("log.msg", "WMI")) &&
(greaterOrEqual("log.cefDeviceSeverity", "3") || equals("log.cat", "blocked"))
', '2026-03-02 16:10:11.855340', true, true, 'origin', null, '[]', '["origin.host","origin.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1150, 'Kaspersky Trusted Application Compromise Detection', 3, 3, 2, 'Defense Evasion', 'T1218 - Signed Binary Proxy Execution', e'Identifies when legitimate or trusted applications exhibit malicious behavior, potentially indicating compromise or exploitation. This includes detecting when signed binaries are used for malicious purposes or when trusted processes perform suspicious activities. This is a critical security event that indicates an attacker may be using living-off-the-land techniques to evade detection.

Next Steps:
1. Immediately isolate the affected system to prevent lateral movement
2. Identify the compromised trusted application and its process chain
3. Review recent system changes and user activities on the affected host
4. Check for persistence mechanisms (scheduled tasks, services, registry keys)
5. Analyze network connections from the compromised application
6. Look for data exfiltration indicators from the affected system
7. Consider reimaging the system if compromise is confirmed
8. Update security policies to monitor the exploited application more closely
', '["https://www.kaspersky.com/enterprise-security/wiki-section/products/kaspersky-anti-targeted-attack-platform","https://attack.mitre.org/techniques/T1218/","https://attack.mitre.org/techniques/T1574/"]', e'exists("log.signatureID") &&
(contains("log.msg", ["trusted application", "signed binary", "legitimate program"]) ||
 contains("log.cs1", "TRUSTED_COMP") ||
 contains("log.cs4", "TrustedApp") ||
 contains("log.msg", "whitelisted") ||
 (equals("log.cat", "Exploit Prevention") && contains("log.msg", "exploit")) ||
 containsAll("log.msg", ["behavior", "trusted"])) &&
oneOf("log.cefDeviceSeverity", ["High", "Medium"])
', '2026-03-02 16:10:13.161179', true, true, 'origin', null, '[]', '["origin.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1151, 'Kaspersky System File Tampering Detection', 2, 3, 1, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects attempts to tamper with critical system files, Windows services, or protected system components. This includes unauthorized modifications to system binaries, service configurations, or attempts to manipulate security-critical files.

Next Steps:
1. Identify the affected system file or component from the log details
2. Check if the modification was authorized (planned maintenance, legitimate software update)
3. Review process information to identify the source of the tampering attempt
4. Look for additional indicators of compromise on the affected system
5. Investigate any parent processes or scripts that initiated the modification
6. Check for persistence mechanisms that may have been established
7. Consider isolating the system if unauthorized tampering is confirmed
', '["https://support.kaspersky.com/kwts/6.1/267200","https://attack.mitre.org/techniques/T1562/001/","https://attack.mitre.org/techniques/T1036/"]', e'exists("log.signatureID") &&
(contains("log.msg", ["system file", "critical file", "protected file", "service tamper"]) ||
 contains("log.cs1", "SYSTEM_MOD") ||
 contains("log.cs4", "SystemFile") ||
 contains("log.msg", ["system modification", "unauthorized change"]) ||
 (equals("log.cat", "Behavior Detection") && contains("log.msg", "modify")))
', '2026-03-02 16:10:14.568249', true, true, 'origin', null, '[]', '["origin.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1152, 'Kaspersky Suspicious Service Installation Detection', 2, 2, 2, 'Persistence, Privilege Escalation', 'T1543.003 - Create or Modify System Process: Windows Service', e'Detects suspicious Windows service installation or modification attempts identified by Kaspersky, which could indicate malware persistence mechanisms or privilege escalation attempts. Service manipulation is a common technique used by malware to maintain persistence on compromised systems.

Next Steps:
1. Identify the service name and executable path from the alert details
2. Verify if the service installation was authorized and legitimate
3. Check the digital signature and reputation of the service executable
4. Review parent process that initiated the service installation
5. Look for other suspicious activities on the affected host around the same time
6. If confirmed malicious, stop and remove the service, quarantine associated files
7. Perform full system scan and check for additional compromise indicators
', '["https://attack.mitre.org/techniques/T1543/003/","https://support.kaspersky.com/ScanEngine/2.1/en-US/186767.htm"]', e'(containsAll("log.msg", ["Service", "install"]) ||
 containsAll("log.msg", ["sc.exe", "create"]) ||
 containsAll("log.msg", ["New", "Service"]) ||
 contains("log.fname", "\\\\services.exe") ||
 contains("log.cs2", "Service")) &&
(oneOf("log.cs1", ["infected", "suspicious"]) ||
 greaterOrEqual("log.cefDeviceSeverity", "3"))
', '2026-03-02 16:10:15.873009', true, true, 'origin', null, '[]', '["origin.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1153, 'Kaspersky Suspicious Scheduled Tasks Detection', 3, 3, 2, 'Execution, Persistence, Privilege Escalation', 'T1053 - Scheduled Task/Job', e'Detects suspicious scheduled task creation or modification identified by Kaspersky, which could indicate persistence mechanisms used by malware or attackers. Scheduled tasks are commonly abused by attackers to maintain persistence, execute malicious code at specific times, or escalate privileges.

Next Steps:
1. Review the scheduled task details in log.msg, log.fname, and log.descMsg fields
2. Check the specific threat signature in log.signatureID to understand the detection
3. Examine log.cs1 and log.cs2 fields for additional threat context and classification
4. Verify if the task creation was part of legitimate administrative activity
5. Check the affected host (origin.host) for other persistence mechanisms:
   - Registry run keys
   - Startup folder items
   - Services
   - WMI event subscriptions
6. Review log.deviceTime for timeline analysis and correlate with other security events
7. If confirmed malicious:
   - Disable or remove the scheduled task immediately
   - Scan the system for additional malware components
   - Check if the malware has spread to other systems
   - Preserve evidence and initiate incident response procedures
', '["https://attack.mitre.org/techniques/T1053/","https://support.kaspersky.com/ScanEngine/1.0/en-US/186767.htm"]', e'(containsAll("log.msg", ["scheduled", "task"]) ||
 contains("log.msg", ["schtasks", "schedule"]) ||
 contains("log.msg", ["scheduled", "task"]) ||
 contains("log.cs2", "persist") ||
 contains("log.fname", "\\\\Tasks\\\\") ||
 contains("log.cat", "persistence")) &&
(exists("log.signatureID") ||
 oneOf("log.cs1", ["infected", "suspicious"]) ||
 exists("log.cefDeviceSeverity"))
', '2026-03-02 16:10:17.226855', true, true, 'origin', null, '[]', '["origin.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1154, 'Suspicious Packed Executable Detection', 3, 3, 2, 'Defense Evasion', 'T1027.002 - Obfuscated Files or Information: Software Packing', e'Detects when Kaspersky identifies suspicious packed executables, which are often used by malware to evade detection and analysis. Packed executables use compression or encryption to hide their true content and make reverse engineering more difficult.

Next Steps:
1. Identify the affected system from origin.hostname and origin.ip fields
2. Review the detected threat details from log.descMsg and log.msg fields
3. Check the action taken by the antivirus (blocked/detected) in the action field
4. Verify if the file is legitimate software that uses packing for protection
5. If malicious, isolate the affected system immediately
6. Perform a full system scan to identify additional threats
7. Review process execution logs for suspicious child processes spawned by packed executables
8. Check network connections initiated by the suspicious executable
9. Submit the sample to Kaspersky or third-party sandbox for detailed analysis
10. Update antivirus signatures and ensure real-time protection is enabled
', '["https://www.kaspersky.com/resource-center/threats/suspicious-packers","https://attack.mitre.org/techniques/T1027/002/"]', e'oneOf("action", ["blocked", "detected"]) &&
(contains("log.msg", ["Packed", "packer"]) ||
 contains("log.msg", ["packed", "Packed"]) ||
 contains("log.msg", ["NSAnti", "Themida", "VMProtect", "ASPack", "UPX",
   "PECompact", "Enigma", "Armadillo"]) ||
 contains("log.msg", ["NSAnti", "Themida", "VMProtect", "ASPack", "UPX",
   "PECompact", "Enigma", "Armadillo"]) ||
 contains("log.cat", ["Trojan.Packed", "Packed"]))
', '2026-03-02 16:10:18.535932', true, true, 'origin', null, '[]', '["origin.host","origin.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1155, 'Kaspersky Suspicious Network Activity Detection', 3, 2, 2, 'Command and Control', 'T1071 - Application Layer Protocol', e'Detects suspicious network activities including unusual connections, potential C2 communications, or network-based attacks identified by Kaspersky security monitoring. This rule triggers when Kaspersky blocks network connections that match suspicious patterns and multiple similar events occur from the same host.

Next Steps:
1. Review the source and destination IP addresses for known malicious indicators using threat intelligence feeds
2. Check if the blocked connection was attempting to reach known C2 servers or suspicious domains
3. Examine the process that initiated the network connection (check log.processName or log.filePath if available)
4. Review other security events from the same host within the last hour for additional IOCs
5. Verify if multiple hosts are exhibiting similar network behavior (potential lateral movement or outbreak)
6. Check firewall logs for any successful connections to the same destination that may have bypassed Kaspersky
7. Consider isolating the affected system if C2 communication is confirmed
8. Run a full system scan on the affected host and check for persistence mechanisms
9. Review network traffic logs for data exfiltration attempts to the same destination
10. Document the incident and update block lists with confirmed malicious IPs/domains
', '["https://support.kaspersky.com/kwts/6.1/267200","https://attack.mitre.org/techniques/T1071/","https://attack.mitre.org/techniques/T1043/"]', e'exists("log.signatureID") &&
(contains("log.msg", ["suspicious connection", "network attack", "port scan", "unusual traffic"]) ||
 contains("log.msg", "network") ||
 contains("log.cs1", "NETWORK") ||
 contains("log.cs2", "Net-Worm") ||
 contains("log.cs4", "Net-Worm") ||
 (exists("log.target.ip") && exists("log.dpt"))) &&
equals("action", "Blocked")
', '2026-03-02 16:10:19.847840', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-kaspersky-*","with":[{"field":"log.src","operator":"filter_term","value":"{{.log.src}}"},{"field":"log.dstIP","operator":"filter_term","value":"{{.log.dstIP}}"}],"or":null,"within":"now-30m","count":5}]', '["target.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1156, 'Kaspersky Sandbox Evasion Attempts Detection', 3, 3, 2, 'Defense Evasion, Discovery', 'T1497 - Virtualization/Sandbox Evasion', e'Identifies malware attempting to detect and evade sandbox environments. This includes time-based evasion, environment checks, anti-VM techniques, and other behaviors designed to avoid analysis in controlled environments.

Next Steps:
1. Immediately isolate the affected system to prevent potential malware spread
2. Review the process that triggered the sandbox evasion detection
3. Check for any suspicious parent processes or child processes
4. Collect memory dumps and samples for deeper analysis
5. Review recent file downloads and email attachments on the affected system
6. Check if similar detection occurred on other systems in the network
7. Consider submitting the sample to Kaspersky for further analysis
', '["https://www.kaspersky.com/enterprise-security/malware-sandbox","https://attack.mitre.org/techniques/T1497/","https://attack.mitre.org/techniques/T1497/001/"]', e'exists("log.signatureID") &&
(contains("log.msg", ["sandbox", "evasion", "anti-VM", "virtualization"]) ||
 contains("log.msg", ["sandbox", "evasion"]) ||
 contains("log.cs1", "SANDBOX_") ||
 contains("log.cs4", ["Evasion", "AntiVM", "environment check", "time delay", "VM detection"]) ||
 (equals("log.cat", "Behavior Detection") &&
  contains("log.msg", ["delay", "sleep"])))
', '2026-03-02 16:10:21.163229', true, true, 'origin', '["origin.host"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1157, 'Process Hollowing Detection', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1055.012 - Process Injection: Process Hollowing', e'Detects process hollowing attempts where malware creates a new process in suspended state, unmaps its memory, and replaces it with malicious code. This advanced technique is used to evade detection by hiding malicious code within legitimate processes.

Next Steps:
1. Immediately isolate the affected system to prevent lateral movement
2. Identify the parent process that initiated the hollowing attempt
3. Check for persistence mechanisms on the affected host
4. Review process creation events around the same timeframe
5. Collect memory dumps for forensic analysis
6. Search for similar patterns across other endpoints
7. Update endpoint protection policies to block the identified threat
', '["https://attack.mitre.org/techniques/T1055/012/","https://www.kaspersky.com/enterprise-security/wiki-section/products/behavior-based-protection"]', e'(equals("log.signatureID", "3") || equals("log.cs1", "DETECT")) &&
(contains("log.msg", ["hollow", "RunPE", "suspend", "NtUnmapViewOfSection", "ZwUnmapViewOfSection"]) ||
 contains("log.cs4", ["hollow", "RunPE", "replace"]) ||
 contains("log.msg", ["hollow", "suspended", "unmap"])) &&
greaterOrEqual("log.cefDeviceSeverity", "3")
', '2026-03-02 16:10:22.559264', true, true, 'origin', null, '[]', '["log.cs5","origin.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1158, 'Kaspersky Application Privilege Escalation Detection', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1055/T1134 - Process Injection and Access Token Manipulation', e'Detects attempts to escalate privileges through application manipulation, process injection, or unauthorized elevation of permissions monitored by Kaspersky endpoint protection. These techniques are commonly used by attackers to gain higher-level permissions on compromised systems.

Next Steps:
1. Immediately isolate the affected system to prevent lateral movement
2. Review the process that triggered the alert and its parent process chain
3. Check if the source process is legitimate or potentially malicious
4. Look for other suspicious activities on the same host in the last hour
5. Collect memory dumps if possible for forensic analysis
6. Review user account permissions and recent changes
7. Check for any unauthorized scheduled tasks or services
8. Update Kaspersky signatures and run a full system scan
', '["https://support.kaspersky.com/KLMS/8.2/en-US/151684.htm","https://attack.mitre.org/techniques/T1055/","https://attack.mitre.org/techniques/T1134/"]', e'exists("log.signatureID") &&
!equals("action", "Allowed") &&
(contains("log.msg", ["privilege", "elevation", "EXPLOIT", "Exploit",
   "process injection", "token manipulation"]) ||
 contains("log.cs1", "EXPLOIT") ||
 contains("log.cs2", "Exploit") ||
 contains("log.cs4", "Exploit") ||
 contains("log.msg", ["privilege", "elevation"]))
', '2026-03-02 16:10:23.961675', true, true, 'origin', null, '[]', '["log.signatureID","origin.host","origin.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1159, 'Living Off the Land Binaries (LOLBins) Abuse Detection', 3, 3, 2, 'Defense Evasion', 'T1218 - System Binary Proxy Execution', e'Detects the malicious use of legitimate Windows system binaries (LOLBins) to execute malicious code. Attackers abuse these trusted tools to bypass security controls and maintain persistence. LOLBins are particularly dangerous because they are signed Microsoft binaries that are trusted by most security products.

Next Steps:
1. Immediately isolate the affected system from the network to prevent lateral movement
2. Review the full context of the detection including command line parameters and parent processes
3. Check for any network connections or file downloads initiated by the LOLBin process
4. Look for persistence mechanisms (scheduled tasks, registry keys, services) created around the same time
5. Scan for additional indicators of compromise on the affected system
6. Review user account activity for signs of compromise or privilege escalation
7. Consider reimaging the system if fileless malware is confirmed
', '["https://attack.mitre.org/techniques/T1218/","https://lolbas-project.github.io/","https://www.kaspersky.com/enterprise-security/wiki-section/products/fileless-threats-protection"]', e'(equals("log.signatureID", "3") || equals("log.cs1", "DETECT")) &&
(regexMatch("log.msg", "(?i).*(rundll32|regsvr32|mshta|certutil|bitsadmin|powershell|wmic|cscript|wscript|msiexec|installutil|regasm|regsvcs).*") ||
 contains("log.cs4", ["fileless", "LOLBin", "LOLBas"])) &&
(contains("log.msg", ["download", "execute", "bypass", "encoded", "obfuscat", "hidden", "malicious"]) ||
 exists("log.actionResult"))
', '2026-03-02 16:10:25.187528', true, true, 'origin', null, '[]', '["log.cs4","origin.host"]');
