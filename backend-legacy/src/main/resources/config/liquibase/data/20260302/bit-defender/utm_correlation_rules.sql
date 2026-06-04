INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1100, 'USB-Based Malware Propagation', 3, 3, 2, 'Lateral Movement, Initial Access', 'T1091 - Replication Through Removable Media', e'Detects USB-based malware propagation attempts including autorun infections, removable media threats, and device control violations. This rule monitors for device control events and removable media access patterns that may indicate malware attempting to spread via USB devices.

Next Steps:
1. Isolate the affected endpoint immediately to prevent further spread
2. Check device control logs for unauthorized USB device connections
3. Scan all removable media that were connected to the affected system
4. Review file creation/modification events on removable drives (especially autorun.inf)
5. Verify if similar events occurred on other endpoints in the network
6. Update device control policies to restrict USB usage if necessary
7. Consider implementing USB device whitelisting for critical systems
', '["https://attack.mitre.org/techniques/T1091/","https://www.bitdefender.com/business/support/en/77209-135324-event-types.html"]', e'(oneOf("log.eventType", ["device-control", "dp"]) &&
 (contains("log.restData", ["malware", "threat", "infection", "autorun", "suspicious"]) ||
  oneOf("log.severity", ["high", "critical", "4", "5"]))) ||
(contains("log.requested", ["usb", "removable", "autorun"]) &&
 contains("log.restData", ["malware", "threat", "infection"]))
', '2026-03-02 15:08:08.501917', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-bitdefender-gz-*","with":[{"field":"log.hostId","operator":"filter_term","value":"{{.log.hostId}}"}],"or":null,"within":"now-30m","count":5}]', '["lastEvent.log.eventType","lastEvent.log.hostId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1101, 'Ransomware Behavior Detection', 3, 3, 3, 'Impact', 'T1486 - Data Encrypted for Impact', e'Detects ransomware behavior patterns including file encryption attempts, mass file modifications, and ransomware-specific malware types detected by Bitdefender GravityZone.

Next Steps:
1. Immediately isolate the affected system from the network to prevent spread
2. Check for recent backup availability and integrity
3. Review process execution history on the affected host
4. Look for suspicious file modifications or mass encryption activities
5. Check for ransomware notes or changed file extensions
6. Investigate the source of infection (email attachments, downloads, RDP compromise)
7. Scan other systems for similar indicators
8. Consider engaging incident response team for containment and recovery
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1486/"]', e'(contains("log.message", ["ransomware", "ransom", "locky", "cerber",
  "wannacry", "petya", "ryuk", "sodinokibi", "maze"]) ||
 contains("log.signatureID", "ransomware") ||
 (equals("log.severity", "10") && contains("log.eventType", "malware"))) &&
exists("log.severity")
', '2026-03-02 15:08:10.423627', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-bitdefender-gz-*","with":[{"field":"log.hostId","operator":"filter_term","value":"{{.log.hostId}}"}],"or":null,"within":"now-10m","count":5}]', '["lastEvent.log.hostId","lastEvent.log.signatureID"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1102, 'Network-Based Threat Detection', 3, 2, 3, 'Command and Control', 'T1071 - Application Layer Protocol: Command and Control', e'Detects network-based threats including C2 communications, malicious network activity, and suspicious network connections identified by Bitdefender GravityZone.

Next Steps:
1. Identify the affected host using log.hostId and check for other security events from this system
2. Review origin.ip to determine if it\'s a known malicious IP or C2 server
3. Check firewall logs for any blocked or allowed connections to/from the suspicious IP
4. Investigate running processes on the affected host for signs of malware
5. Review network traffic patterns for data exfiltration attempts
6. If ransomware is detected, immediately isolate the affected system
7. Collect network packet captures if available for deeper analysis
8. Check if other hosts have communicated with the same external IP address
9. Submit suspicious IPs to threat intelligence platforms for reputation checking
10. Document findings and update firewall rules to block confirmed malicious IPs
', '["https://attack.mitre.org/techniques/T1071/","https://www.bitdefender.com/business/support/en/77209-135324-event-types.html"]', e'(oneOf("log.eventType", ["network-sandboxing", "fw"]) &&
 oneOf("log.severity", ["high", "critical", "4", "5"])) ||
(exists("origin.ip") && contains("log.eventType", "network") &&
 contains("log.restData", ["malware", "threat", "blocked", "c2", "botnet"])) ||
(equals("log.severity", "critical") && contains("log.product", "network"))
', '2026-03-02 15:08:11.725118', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-bitdefender-gz-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":[{"indexPattern":"v11-log-antivirus-bitdefender-gz-*","with":[{"field":"log.hostId","operator":"filter_term","value":"{{.log.hostId}}"},{"field":"log.eventType","operator":"filter_term","value":"network-sandboxing"}],"or":null,"within":"now-4h","count":3}],"within":"now-2h","count":5}]', '["lastEvent.log.hostId","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1103, 'Multiple Malware Detections from Single Source', 3, 3, 2, 'Command and Control', 'T1105 - Ingress Tool Transfer', e'Detects when multiple malware threats are detected on a single host within a short time period. This could indicate a compromised system actively spreading malware or an attacker launching multiple malware variants.

Next Steps:
1. Investigate the affected host:
   - Identify the system using the hostId field
   - Check if it\'s a critical system or server
   - Review recent user activity on the host
2. Analyze the detected malware:
   - Review the malware types and names detected (signatureID field)
   - Check file paths and processes involved
   - Determine if malware was successfully quarantined
3. Check for lateral movement:
   - Look for connections from the affected host to other internal systems
   - Review authentication logs for suspicious activity
   - Check for file share access patterns
4. Remediation actions:
   - Isolate the affected system if confirmed compromised
   - Run full system scans on potentially affected systems
   - Update antivirus signatures and definitions
   - Consider reimaging if system is severely compromised
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1105/"]', e'equals("log.eventType", "AntiMalware") &&
oneOf("log.severity", ["4", "5"]) &&
exists("log.hostId")
', '2026-03-02 15:08:12.946636', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-bitdefender-gz-*","with":[{"field":"log.hostId","operator":"filter_term","value":"{{.log.hostId}}"},{"field":"log.eventType","operator":"filter_term","value":"AntiMalware"}],"or":null,"within":"now-1h","count":5}]', '["lastEvent.log.hostId","lastEvent.log.signatureID"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1104, 'Malware Outbreak Detection - Multiple Hosts Infected', 3, 3, 3, 'Command and Control', 'T1105 - Ingress Tool Transfer', e'Detects when the same malware signature or threat is detected on multiple endpoints within a short time window. This pattern indicates a potential malware outbreak spreading across the network environment.

Next Steps:
1. Immediately isolate affected endpoints to prevent further spread
2. Identify the malware signature ID and research its capabilities and impact
3. Check network logs for lateral movement patterns between infected hosts
4. Review the initial infection vector - check email logs, web proxy logs, and USB device usage
5. Verify antivirus definitions are up-to-date on all endpoints
6. Conduct memory and disk forensics on patient zero if identifiable
7. Check for persistence mechanisms on infected systems
8. Review domain controller and authentication logs for credential compromise
9. Document all affected systems and timeline for incident response
10. Consider engaging incident response team if outbreak involves critical systems
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1105/"]', e'equals("log.eventType", "AntiMalware") &&
oneOf("log.severity", ["4", "5"]) &&
exists("log.signatureID") &&
exists("log.syslogHostIP")
', '2026-03-02 15:08:14.258459', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-bitdefender-gz-*","with":[{"field":"log.signatureID","operator":"filter_term","value":"{{.log.signatureID}}"},{"field":"log.eventType","operator":"filter_term","value":"AntiMalware"}],"or":null,"within":"now-2h","count":10}]', '["lastEvent.log.signatureID","lastEvent.log.syslogHostIP"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1105, 'Bitdefender AV Policy Weakened', 3, 3, 2, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects when Bitdefender GravityZone antivirus policies are weakened by administrators, such as disabling real-time protection, reducing scan aggressiveness, or adding broad exclusions. This could indicate a compromised admin account or insider threat.

Next Steps:
1. Identify the administrator who modified the policy
2. Verify the policy change was authorized through change management
3. Review the specific settings that were weakened
4. Check for concurrent suspicious activity on managed endpoints
5. Restore the previous policy configuration if unauthorized
6. Review admin account access logs for compromise indicators
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1562/001/"]', e'(contains("log.message", ["policy", "configuration", "setting"]) &&
 (contains("log.message", ["disabled", "weakened", "reduced", "lowered", "excluded"]) ||
  (contains("log.message", "real-time") && contains("log.message", "off")) ||
  (contains("log.message", "exclusion") && contains("log.message", "added")) ||
  (contains("log.message", "protection") && contains("log.message", "disabled")))) &&
exists("log.severity")
', '2026-03-02 15:08:15.561087', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-bitdefender-gz-*","with":[{"field":"log.hostId","operator":"filter_term","value":"{{.log.hostId}}"}],"or":null,"within":"now-1h","count":3}]', '["lastEvent.log.eventType","lastEvent.log.hostId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1106, 'Bitdefender Console Used for Lateral Movement', 3, 3, 3, 'Lateral Movement', 'T1072 - Software Deployment Tools', e'Detects when the Bitdefender GravityZone management console is potentially being used to push malicious policies, scripts, or tasks to managed endpoints, indicating a compromised admin account being leveraged for lateral movement.

Next Steps:
1. Review all recent task and policy deployments from the console
2. Identify the admin account used and verify its legitimacy
3. Check for unusual login patterns to the GravityZone console
4. Review the content of pushed policies for malicious configurations
5. Suspend the admin account if compromise is suspected
6. Audit all managed endpoints for signs of compromise
', '["https://attack.mitre.org/techniques/T1072/","https://www.bitdefender.com/business/support/en/77212-237089-event-types.html"]', e'(contains("log.message", ["remote task", "deploy", "push policy", "execute script"]) ||
 (contains("log.message", "task") && contains("log.message", "created") &&
  (contains("log.message", "scan") || contains("log.message", "install") ||
   contains("log.message", "uninstall") || contains("log.message", "execute")))) &&
exists("log.severity")
', '2026-03-02 15:08:16.926735', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-bitdefender-gz-*","with":[{"field":"log.hostId","operator":"filter_term","value":"{{.log.hostId}}"}],"or":null,"within":"now-30m","count":10}]', '["lastEvent.log.eventType","lastEvent.log.hostId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1107, 'Bitdefender GravityZone Zero-Day Malware Detection', 3, 3, 2, 'Execution', 'T1203 - Exploitation for Client Execution', e'Detects potential zero-day malware identified by Bitdefender\'s advanced threat detection capabilities including HyperDetect and Sandbox Analyzer. These detection methods use behavioral analysis and machine learning to identify previously unknown threats.

Next Steps:
1. Immediately isolate the affected system from the network to prevent lateral movement
2. Review the detection details including file path, process information, and threat indicators
3. Check if similar detections occurred on other systems in the environment
4. Collect the suspicious file/process for further analysis in a sandbox environment
5. Review system logs for any suspicious activities before and after the detection
6. Update security policies to block similar threats across the organization
7. Consider submitting the sample to Bitdefender for further analysis
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1203/"]', e'oneOf("log.eventType", ["HyperDetect Activity", "Sandbox Analyzer Detection", "hyperdetect"]) ||
(equals("log.eventType", "avc") && equals("log.severity", "High"))
', '2026-03-02 15:08:18.064111', true, true, 'origin', null, '[]', '["lastEvent.log.hostId","lastEvent.log.syslogHostIP"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1108, 'Bitdefender GravityZone Suspicious Exclusion Added', 3, 3, 1, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects when exclusions are added to Bitdefender GravityZone that may allow malware to operate undetected. Attackers often add exclusions to antivirus software to prevent detection of their malicious tools and activities.

Next Steps:
1. Review the exclusion details to determine what files, folders, or processes were excluded
2. Verify if the exclusion was authorized by security team or IT administrators
3. Check if the excluded path contains any suspicious executables or scripts
4. Review recent activity from the user who added the exclusion
5. If unauthorized, immediately remove the exclusion and scan the excluded locations
6. Consider implementing a change control process for antivirus exclusions
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1562/001/"]', e'equals("log.eventType", "exclusion_added") ||
(oneOf("log.eventType", ["policy_change", "configuration_change"]) &&
 contains("log.requested", "exclusion"))
', '2026-03-02 15:08:19.240633', true, true, 'origin', null, '[]', '["lastEvent.log.eventType","lastEvent.log.hostId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1109, 'Rootkit Detection', 3, 3, 2, 'Defense Evasion', 'T1014 - Rootkit', e'Detects rootkit infections and kernel-level threats that attempt to hide malicious activity at the system level using Bitdefender GravityZone\'s advanced detection capabilities.

Next Steps:
1. Immediately isolate the affected system from the network to prevent lateral movement
2. Capture a memory dump and disk image for forensic analysis
3. Check for signs of privilege escalation or kernel-level modifications
4. Review system logs for any suspicious driver installations or kernel module loading
5. Scan other systems in the same network segment for similar infections
6. Consider rebuilding the system from a known clean state as rootkits can be difficult to fully remove
7. Review how the rootkit was initially delivered (email attachment, exploit kit, etc.)
8. Update all security software and operating system patches
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1014/"]', e'equals("log.eventType", "malware_detected") &&
oneOf("log.severity", ["high", "critical"]) &&
(
  contains("log.restData", ["rootkit", "kernel", "tdss", "zeroaccess",
    "necurs", "bootkit", "alureon", "rustock", "sinowal"]) ||
  contains("log.requested", "rootkit") ||
  equals("log.signatureID", "rootkit")
)
', '2026-03-02 15:08:20.430868', true, true, 'origin', null, '[]', '["lastEvent.log.hostId","lastEvent.log.signatureID"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1110, 'Real-time Protection Disabled', 3, 3, 2, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects when real-time protection features are disabled on an endpoint. This is a critical security event as it leaves the system vulnerable to malware infections and requires immediate investigation.

Next Steps:
1. Immediately investigate who disabled the real-time protection and why
2. Check if the action was authorized by IT security team
3. Review recent activity on the affected endpoint for signs of compromise
4. Re-enable real-time protection if the action was unauthorized
5. Check for any malware infections that may have occurred while protection was disabled
6. Review system logs for any suspicious activities during the protection downtime
7. Consider implementing additional controls to prevent unauthorized disabling of security tools
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1562/001/"]', e'exists("log.syslogHostIP") &&
(
  (equals("log.eventType", "modules") &&
   equals("log.product", "av") &&
   contains("log.restData", "real-time")) ||
  (equals("log.eventType", "Product ModulesStatus") &&
   oneOf("log.severity", ["4", "5"]) &&
   (contains("log.restData", "protection disabled") ||
    contains("log.restData", "real-time scanning disabled")))
)
', '2026-03-02 15:08:21.638489', true, true, 'origin', null, '[]', '["lastEvent.log.eventType","lastEvent.log.syslogHostIP"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1111, 'Bitdefender GravityZone Quarantine Failure Detection', 3, 3, 2, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects when Bitdefender GravityZone fails to quarantine detected malware. This could indicate that the malware is actively resisting remediation attempts or that there are permission issues preventing proper quarantine.

Next Steps:
- Immediately isolate the affected system from the network
- Check if the malware process is still running and attempt manual termination
- Verify antivirus permissions and ensure it has necessary privileges
- Review system logs for signs of privilege escalation or rootkit activity
- Consider reimaging the system if quarantine continues to fail
- Check for similar failures on other systems in the environment
- Investigate the specific malware detected and research its capabilities
- Review quarantine configuration and storage capacity
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1562/001/"]', e'oneOf("log.eventType", ["quarantine_failed", "quarantine_failure"]) ||
(equals("log.eventType", "AntiMalware") &&
  (containsAll("log.requestToParse", ["quarantine", "fail"]) ||
   contains("log.restData", ["quarantine failed", "unable to quarantine", "failed to quarantine"]) ||
   (equals("log.severity", "failure") && contains("log.requestToParse", "quarantine"))))
', '2026-03-02 15:08:22.904073', true, true, 'origin', null, '[]', '["lastEvent.log.hostId","lastEvent.log.signatureID"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1112, 'Memory-Based Threat Detection', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1055 - Process Injection', e'Detects memory-based threats including process injection, memory manipulation, and fileless malware executing in memory based on Bitdefender GravityZone event types.

Next Steps:
1. Identify the affected process and host using log.hostId and origin.path fields
2. Check if the process is legitimate or if it shows signs of compromise
3. Review the process tree to identify parent-child relationships
4. Look for other suspicious activities on the same host in the last hour
5. Collect memory dump if the process is still running
6. Analyze network connections from the affected process
7. Check for persistence mechanisms on the affected system
8. Isolate the host if active malicious behavior is confirmed
', '["https://attack.mitre.org/techniques/T1055/","https://www.bitdefender.com/business/support/en/77209-135324-event-types.html"]', e'exists("log.eventType") &&
(oneOf("log.eventType", ["aph", "antiexploit", "hd"]) ||
 (exists("origin.path") && contains("origin.path", "memory"))) &&
(oneOf("log.severity", ["critical", "high"]) ||
 exists("log.malwareName") ||
 exists("log.threatName"))
', '2026-03-02 15:08:24.053530', true, true, 'origin', null, '[]', '["lastEvent.log.hostId","adversary.path"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1113, 'Bitdefender GravityZone High Severity Threat Detection', 3, 3, 2, 'Execution', 'T1204.002 - User Execution: Malicious File', e'Detects high-severity malware threats identified by Bitdefender GravityZone that require immediate attention. This rule triggers on severity levels 8-10, which indicate critical threats such as trojans, ransomware, rootkits, or other advanced malware.

Next Steps:
1. Immediately isolate the affected endpoint(s) from the network to prevent lateral movement
2. Review the threat details in Bitdefender GravityZone console:
   - Check threat name and malware type from the event details
   - Verify the affected file path and process information
   - Review the action taken by Bitdefender (quarantine, delete, etc.)
3. Investigate the source of infection:
   - Check origin.ip and origin.path for the malware source
   - Review recent user activity and email attachments
   - Look for similar threats on other endpoints
4. Perform forensic analysis:
   - Collect memory dumps if rootkit or fileless malware is suspected
   - Check for persistence mechanisms (registry, scheduled tasks, services)
   - Review network connections from the affected host
5. Remediation actions:
   - Ensure Bitdefender has successfully cleaned/quarantined the threat
   - Run full system scan on affected and neighboring systems
   - Update antivirus signatures and security policies
   - Consider reimaging if system integrity is compromised
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1204/002/","https://attack.mitre.org/techniques/T1055/"]', 'oneOf("log.severity", ["8", "9", "10"]) && oneOf("log.eventType", ["avc", "malware_detected", "av"])', '2026-03-02 15:08:25.191189', true, true, 'origin', null, '[]', '["lastEvent.log.signatureID","lastEvent.log.syslogHostIP"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1114, 'Fileless Malware Detection', 2, 2, 2, 'Defense Evasion, Privilege Escalation', 'T1055 - Process Injection', e'Detects fileless malware attacks including PowerShell-based attacks, memory injection, and living-off-the-land techniques using Bitdefender GravityZone\'s HyperDetect and Command-Line Scanner modules. These attacks execute malicious code directly in memory without writing to disk, making them harder to detect with traditional antivirus.

Next Steps:
- Isolate the affected endpoint immediately to prevent lateral movement
- Review process tree to identify the parent process and initial attack vector
- Check for PowerShell command history and script blocks (Event ID 4104)
- Look for suspicious WMI activity or unusual process spawning patterns
- Examine network connections from the affected process
- Collect memory dump if the process is still running
- Review user activity to determine if account is compromised
- Apply security patches if exploitation of vulnerability is suspected
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1055/","https://www.bitdefender.com/en-us/business/gravityzone-platform/fileless-attack-defense"]', e'oneOf("log.eventType", ["fileless_attack", "hyperdetect_fileless", "command_line_scanner"]) ||
(
  equals("log.eventType", "malware_detected") &&
  contains("log.restData", ["fileless", "memory injection", "powershell",
    "wscript", "cscript", "mshta", "regsvr32", "rundll32"])
) ||
(
  oneOf("log.severity", ["HIGH", "CRITICAL"]) &&
  contains("log.restData", "code injection")
)
', '2026-03-02 15:08:26.361516', true, true, 'origin', null, '[]', '["lastEvent.log.eventType","lastEvent.log.hostId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1115, 'Email-Based Threat Spreading', 3, 3, 2, 'Initial Access', 'T1566 - Phishing', e'Detects email-based malware spreading including phishing attempts, malicious attachments, and email-borne threats through Bitdefender\'s Exchange protection. This rule triggers on Exchange-specific malware events and monitors for patterns of email-based threats.

Next Steps:
1. Investigate the affected email and sender:
   - Check the sender\'s email address and domain reputation
   - Review email headers for spoofing indicators
   - Analyze attachment hash values if present in log.restData
   - Check log.severity for threat level assessment
2. Review related events:
   - Look for similar events from the same sender or to other recipients
   - Check if the email was delivered or blocked
   - Verify if any users clicked links or opened attachments
   - Search for the same signatureID across other hosts
3. Remediation actions:
   - If delivered, recall the email from all recipients immediately
   - Reset credentials if phishing was successful
   - Block sender domain/IP at email gateway
   - Update email security policies if needed
   - Scan affected endpoints for malware if attachments were opened
   - Update Bitdefender Exchange protection rules
4. Investigation commands:
   - Check host status: Verify log.hostId endpoint protection status
   - Review product version: Ensure log.productVersion is up to date
   - Analyze event patterns: Look for unusual log.eventType combinations
', '["https://attack.mitre.org/techniques/T1566/","https://www.bitdefender.com/business/support/en/77209-135324-event-types.html"]', e'oneOf("log.eventType", ["exchange-malware", "exchange-user-credentials", "exchange-organization-info"]) ||
(contains("log.eventType", "exchange") && equals("log.severity", "High")) ||
(contains("log.product", "Exchange") && contains("log.eventType", ["malware", "phishing"]))
', '2026-03-02 15:08:27.623193', true, true, 'origin', null, '[]', '["lastEvent.log.hostId","lastEvent.log.signatureID"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1116, 'Crypto-Mining Detection', 2, 2, 3, 'Impact', 'T1496 - Resource Hijacking', e'Detects cryptocurrency mining activities including miners, coin miners, and cryptojacking attempts detected by Bitdefender GravityZone.

Next Steps:
- Review the affected endpoint details (hostname, IP) to identify the compromised system
- Check CPU and memory usage patterns on the affected system for unusual spikes
- Look for network connections to known mining pools or suspicious outbound traffic
- Search for related processes running with names like xmrig, minerd, cgminer, or bfgminer
- Review recent file downloads and installations on the affected system
- Check for persistence mechanisms (scheduled tasks, startup items, services)
- Isolate the affected system if active mining is confirmed
- Run a full system scan with updated definitions
- Consider reimaging the system if compromise is extensive', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1496/"]', e'equals("log.productVendor", "Bitdefender") &&
equals("log.product", "GravityZone") &&
(
  contains("log.eventType", ["miner", "coin", "crypto", "CoinMiner"]) ||
  contains("log.requested", ["miner", "coin", "xmr", "monero", "bitcoin",
    "ethereum", "xmrig", "minerd", "cgminer", "bfgminer", "coinhive"])
)
', '2026-03-02 15:08:28.945981', true, true, 'origin', null, '[]', '["adversary.host","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1117, 'Bootkit/UEFI Threat Detection', 3, 3, 3, 'Defense Evasion, Persistence', 'T1542.001 - Boot or Logon Autostart Execution: System Firmware', e'Detects bootkit or UEFI-level threats that attempt to persist at the firmware level and compromise the boot process. These threats can survive system reinstalls and bypass traditional security measures by infecting the system firmware.

Next Steps:
- Isolate the affected system immediately to prevent spread
- Review system boot logs and firmware settings for modifications
- Check for other malware detections on the same host in the past 24-48 hours
- Verify system integrity using offline scanning tools
- Consider reimaging the system and updating firmware/UEFI
- Enable Secure Boot if not already enabled
- Review user activity and recently installed software on the affected system
- Document the infection for incident response reporting
- Check if other systems with similar hardware/firmware versions are affected
', '["https://attack.mitre.org/techniques/T1542/001/","https://www.bitdefender.com/business/support/en/77209-135324-event-types.html"]', e'equals("log.eventType", "av") &&
greaterOrEqual("log.severity", 8) &&
(
  contains("log.requested", ["boot", "uefi", "rootkit", "firmware"]) ||
  contains("log.restData", ["boot", "uefi", "rootkit", "firmware",
    "\\\\EFI\\\\", "/EFI/", "\\\\boot\\\\", "/boot/"])
)
', '2026-03-02 15:08:30.211693', true, true, 'origin', null, '[]', '["lastEvent.log.hostId","lastEvent.log.severity"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1118, 'Advanced Persistent Threat (APT) Detection', 3, 3, 2, 'Command and Control', 'TA0011 - Application Layer Protocol', e'Detects indicators of Advanced Persistent Threats including targeted attacks, sophisticated malware, and persistent threats detected by Bitdefender GravityZone\'s HyperDetect module.

Next Steps:
- Investigate the affected endpoint to determine the scope of compromise
- Review process execution history and network connections from the affected system
- Check for lateral movement by examining authentication logs from the same source IP
- Isolate the affected system if active threat is confirmed
- Collect forensic artifacts including memory dumps and event logs
- Search for similar malware indicators across the environment
- Review user account activities for signs of credential compromise
- Contact security operations center if threat actors match known APT groups
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/tactics/TA0011/"]', e'equals("log.product", "Bitdefender GravityZone") &&
greaterOrEqual("log.severity", 8) &&
(
  contains("log.eventType", ["apt", "targeted", "advanced", "persistent", "hyperdetect"]) ||
  contains("log.restData", ["apt", "targeted attack", "advanced persistent",
    "lazarus", "equation", "sofacy", "cozy bear", "fancy bear",
    "panda", "kitten", "carbanak", "fin7", "fileless"]) ||
  equals("log.signatureID", "hyperdetect")
) &&
exists("log.hostId")
', '2026-03-02 15:08:31.607874', true, true, 'origin', null, '[]', '["lastEvent.log.eventType","lastEvent.log.hostId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1119, 'Antivirus Service Stopped or Disabled', 2, 3, 3, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects when the Bitdefender antivirus service or critical security modules are stopped, disabled, or experiencing failures. This is a critical security event that could indicate malicious tampering or system issues.

Next Steps:
- Verify if the service was intentionally stopped by authorized personnel
- Check system logs for any errors or crashes that may have caused the service to stop
- Look for signs of malware or unauthorized access attempts around the time of the event
- Review recent system changes or updates that might have affected the antivirus service
- If tampering is suspected, isolate the affected system and perform a forensic analysis
- Restart the Bitdefender service and ensure all modules are functioning properly
- Monitor for recurring issues that might indicate persistent threats
', '["https://www.bitdefender.com/business/support/en/77212-237089-event-types.html","https://attack.mitre.org/techniques/T1562/001/"]', e'(equals("log.eventType", "modules") ||
 equals("log.eventType", "Product ModulesStatus") ||
 equals("log.eventType", "registration")) &&
(oneOf("log.severity", ["high", "5"]) ||
 contains("log.product", "disabled") ||
 contains("log.product", "stopped") ||
 (contains("log.restData", "module") && contains("log.restData", "stopped")) ||
 (contains("log.restData", "module") && contains("log.restData", "disabled")) ||
 (contains("log.restData", "av") && contains("log.restData", "failure")))
', '2026-03-02 15:08:32.822661', true, true, 'origin', null, '[]', '["lastEvent.log.eventType","lastEvent.log.hostId"]');
