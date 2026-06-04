INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1136, 'ESET Repeated Quarantine Failures', 2, 3, 2, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects repeated quarantine failures in ESET, which may indicate malware actively resisting quarantine through file locks, permission manipulation, or rapid re-creation of malicious files.

Next Steps:
1. Identify the specific file or threat that cannot be quarantined
2. Check the file permissions and processes locking the file
3. Attempt manual quarantine or deletion in safe mode
4. Review the malware\'s persistence mechanisms
5. Consider isolating the endpoint for manual remediation
6. Run a boot-time scan if available
', '["https://help.eset.com/ees/8/en-US/","https://attack.mitre.org/techniques/T1562/001/"]', e'(contains("log.message", "quarantine") &&
 (contains("log.message", "failed") || contains("log.message", "error") ||
  contains("log.message", "unable") || contains("log.message", "denied"))) ||
(contains("log.message", "clean") && contains("log.message", "failed") &&
 contains("log.message", "threat"))
', '2026-03-02 16:02:21.835806', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-esmc-eset-*","with":[{"field":"log.headHostname","operator":"filter_term","value":"{{.log.headHostname}}"}],"or":null,"within":"now-1h","count":5}]', '["lastEvent.log.headHostname","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1137, 'ESET ERA/ESMC Console Suspicious Activity', 3, 3, 2, 'Lateral Movement', 'T1072 - Software Deployment Tools', e'Detects suspicious activity on the ESET ERA/ESMC management console including unauthorized policy changes, mass task deployments, or admin account modifications that could indicate console compromise.

Next Steps:
1. Verify the admin account performing console operations
2. Review recent policy changes and task deployments
3. Check admin login history for unauthorized access
4. Verify the content of any pushed policies or tasks
5. Suspend suspicious admin accounts
6. Audit endpoints affected by recent console changes
', '["https://help.eset.com/esmc_admin/70/en-US/","https://attack.mitre.org/techniques/T1072/"]', e'(contains("log.message", "policy") &&
 (contains("log.message", "modified") || contains("log.message", "assigned") ||
  contains("log.message", "created"))) ||
(contains("log.message", "client task") &&
 (contains("log.message", "executed") || contains("log.message", "deployed"))) ||
(contains("log.message", "administrator") &&
 ((contains("log.message", "created") || contains("log.message", "modified")) ||
  (contains("log.message", "login") && contains("log.message", "failed"))))
', '2026-03-02 16:02:23.011021', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-esmc-eset-*","with":[{"field":"log.headHostname","operator":"filter_term","value":"{{.log.headHostname}}"}],"or":null,"within":"now-30m","count":10}]', '["lastEvent.log.headHostname","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1138, 'Advanced Heuristic Detection Triggers', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1055 - Process Injection', e'Detects when ESET\'s advanced heuristic engine identifies suspicious behavior patterns that may indicate novel malware or zero-day threats. These detections use DNA signatures and behavioral analysis.

Next Steps:
- Review the affected hostname and user context to understand the scope
- Check the process name (if available) that triggered the detection
- Verify if the action taken (cleaned/deleted/quarantined) was successful
- Look for related alerts from the same host within the past 24 hours
- If multiple hosts show similar detections, investigate potential lateral movement
- Consider isolating the affected system if threat persists
- Review ESET console link (if available) for detailed threat information
- Check file hash against threat intelligence databases if available
- Capture and analyze the malicious file sample if quarantined
- Review system logs for any unusual activities before and after detection
- Update ESET signatures and run a full system scan
', '["https://help.eset.com/eea/8/en-US/idh_config_threat_sense.html","https://attack.mitre.org/techniques/T1055/"]', e'oneOf("log.msgType", ["EnterpriseInspectorAlert_Event", "threat_event", "FirewallAggregatedAlert_Event"]) &&
contains("log.jsonMessage", ["heuristic", "NewHeur", "suspicious behavior"]) &&
contains("log.jsonMessage", ["cleaned", "deleted", "quarantined", "blocked"])
', '2026-03-02 16:02:24.326768', true, true, 'origin', null, '[{"indexPattern":"v11-log-antivirus-esmc-eset-*","with":[{"field":"log.headHostname","operator":"filter_term","value":"{{.log.headHostname}}"}],"or":null,"within":"now-30m","count":3}]', '["lastEvent.log.headHostname","lastEvent.log.msgType"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1139, 'Suspicious Process Behavior Detection', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1055 - Process Injection', e'Detects suspicious process behaviors including injection attempts, privilege escalation, and abnormal process creation patterns identified by ESET\'s behavioral monitoring. This alert indicates potential malware activity or exploitation attempts on the affected system.

Next Steps:
1. Immediately review the alert details to identify:
   - Affected hostname (check log.headHostname)
   - Specific threat or behavior detected (check log.jsonMessage)
   - Process name and path if available
   - Time of detection (check log.deviceTime)
2. Investigate the process that triggered the alert:
   - Verify if it\'s a legitimate application or unknown/suspicious
   - Check process creation chain and parent-child relationships
   - Review file hash against threat intelligence sources
3. Check for related security events:
   - Look for other ESET alerts from the same host
   - Search for network connections from the suspicious process
   - Review authentication events around the same timeframe
4. Containment actions if malicious:
   - Isolate the affected host from the network
   - Kill the suspicious process if still running
   - Preserve forensic evidence (memory dump, logs)
5. Remediation steps:
   - Run full antivirus scan on the affected system
   - Check for persistence mechanisms (registry, scheduled tasks)
   - Update ESET signatures and perform system hardening
6. Prevention measures:
   - Review and update application control policies
   - Ensure ESET real-time protection is enabled
   - Consider implementing application whitelisting
', '["https://help.eset.com/ees/12/en-US/idh_dialog_epfw_ids_alert.html","https://attack.mitre.org/techniques/T1055/"]', e'oneOf("log.msgType", ["EnterpriseInspectorAlert_Event", "HIPS_Event"]) &&
exists("log.jsonMessage") &&
contains("log.jsonMessage", ["Process injection", "Suspicious behavior",
  "Anomalous process", "blocked", "terminated", "prevented"])
', '2026-03-02 16:02:25.460232', true, true, 'origin', null, '[]', '["lastEvent.log.headHostname","lastEvent.log.msgType"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1140, 'ESET Blocked Suspicious PowerShell Activity', 3, 3, 1, 'Execution', 'T1059.001 - Command and Scripting Interpreter: PowerShell', e'Detects when ESET blocks suspicious PowerShell commands or scripts that exhibit malicious behavior patterns, including obfuscated scripts, encoded commands, or attempts to bypass execution policies. This is a high-priority security event that indicates potential malicious activity was prevented.

Next Steps:
1. Review the blocked PowerShell command details in the log message
2. Identify the user account and process that attempted to execute PowerShell
3. Check if this is part of legitimate administrative activity or scripting
4. Investigate the source of the PowerShell execution (parent process, script location)
5. Look for other suspicious activities from the same host or user
6. Consider isolating the affected system if malicious intent is confirmed
7. Review and update PowerShell execution policies if needed
', '["https://help.eset.com/ees/8/en-US/idh_hips_main.html","https://attack.mitre.org/techniques/T1059/001/"]', 'regexMatch("log.message", "(?i)(powershell|pwsh)") && equals("log.action", "blocked") && exists("log.headHostname")', '2026-03-02 16:02:26.813800', true, true, 'origin', null, '[]', '["lastEvent.log.headHostname","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1141, 'Suspicious Encrypted File Activity', 3, 3, 2, 'Impact', 'T1486 - Data Encrypted for Impact', e'Detects suspicious encrypted file activities that may indicate ransomware encryption attempts or unauthorized file encryption operations. This rule triggers when ESET detects ransomware-related threats or file encryption activities.

Next Steps:
1. Immediately isolate the affected system to prevent spread
2. Check if backup systems are accessible and uncompromised
3. Review the threat details in log.jsonMessage for specific ransomware variant
4. Look for other systems showing similar encryption patterns
5. Preserve forensic evidence before remediation
6. Consider engaging incident response team for ransomware cases
7. Do not power off the system if encryption is in progress
', '["https://attack.mitre.org/techniques/T1486/","https://help.eset.com/protect_admin/10.1/en-US/events-exported-to-json-format.html"]', e'equals("log.msgType", "Threat_Event") &&
contains("log.jsonMessage", ["ransomware", "filecoder", "encrypted", ".encrypted"])
', '2026-03-02 16:02:27.993312', true, true, 'origin', null, '[]', '["lastEvent.log.msgType","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1142, 'Registry Modification Attempts Blocked', 2, 3, 2, 'Defense Evasion, Persistence', 'T1112 - Modify Registry', e'Identifies attempts to modify critical Windows registry keys that were blocked by ESET, indicating potential persistence or system tampering attempts. Registry modifications are a common technique used by malware to establish persistence, disable security features, or alter system behavior.

Next Steps:
1. Review the blocked action details to understand what registry key was targeted
2. Investigate the source process and user account involved in the attempt
3. Check for other security events from the same host around the same time
4. Verify if this is legitimate administrative activity or potential malicious behavior
5. If suspicious, isolate the affected system and perform a full malware scan
6. Review system logs for any successful registry modifications before the block occurred
', '["https://help.eset.com/esmc_admin/70/en-US/events-exported-to-json-format.html","https://attack.mitre.org/techniques/T1112/"]', e'exists("log.jsonMessage") &&
contains("log.jsonMessage", "registry") &&
oneOf("log.action", ["blocked", "denied", "prevented"]) &&
oneOf("log.severity", ["high", "medium"])
', '2026-03-02 16:02:29.178944', true, true, 'origin', null, '[]', '["lastEvent.log.action","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1143, 'ESET Network Attack Detection', 3, 2, 1, 'Initial Access', 'T1190 - Exploit Public-Facing Application', e'Detects network-based attacks and exploits blocked by ESET\'s Network Attack Protection (IDS). This includes attempts to exploit known vulnerabilities in network services and protocols.

Next Steps:
1. Review the attack details in ESET console to identify the specific vulnerability or attack pattern
2. Check if the source IP is known malicious using threat intelligence sources
3. Verify if other systems received similar attacks from the same source
4. Review firewall logs for additional suspicious activity from the source IP
5. Consider blocking the source IP at the perimeter firewall if attacks persist
6. Update network security policies and ensure all systems are patched
', '["https://help.eset.com/ees/7/en-US/idh_config_epfw_network_attack_protection.html","https://attack.mitre.org/techniques/T1190/"]', e'equals("log.event_type", "NetworkProtection_Event") &&
equals("log.action", "blocked") &&
exists("origin.ip")
', '2026-03-02 16:02:30.262729', true, true, 'origin', null, '[]', '["adversary.ip","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1144, 'Machine Learning Detection Anomalies', 3, 3, 2, 'Execution', 'T1204.002 - User Execution: Malicious File', e'Identifies threats detected by ESET\'s machine learning engine that analyzes file behavior patterns and characteristics to identify previously unknown malware variants. Machine learning detection indicates advanced malware that may evade signature-based detection methods.

Next Steps:
- Immediately investigate the affected host for signs of compromise
- Review the threat details in the log message to understand the malware type and behavior
- Check if the malware was successfully blocked or quarantined
- Look for similar detections across other hosts in your environment
- Consider isolating the affected system if the threat was not successfully contained
- Review process activity around the time of detection for suspicious behavior
- Collect and analyze the malware sample if available for threat intelligence
- Update security policies to prevent similar threats
- Check for any data exfiltration or lateral movement attempts from the affected host
', '["https://help.eset.com/protect_admin/11.0/en-US/events-exported-to-json-format.html","https://attack.mitre.org/techniques/T1204/002/"]', e'contains("log.message", "machine learning") &&
contains("log.message", ["threat", "detected", "found"]) &&
exists("log.msgType") &&
exists("log.headHostname")
', '2026-03-02 16:02:31.619452', true, true, 'origin', null, '[]', '["lastEvent.log.headHostname","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1145, 'ESET Host Intrusion Prevention System Triggered', 3, 3, 2, 'Defense Evasion, Privilege Escalation', 'T1055 - Process Injection', e'Detects when ESET\'s Host-based Intrusion Prevention System (HIPS) blocks suspicious behavior, including process manipulation, registry modifications, and file system changes that indicate potential malware activity. HIPS events indicate active attempts to compromise system integrity through various attack techniques.

Next Steps:
1. Review the blocked process or action details in the ESET console
2. Identify the source application attempting the blocked behavior
3. Check if the blocked action is from legitimate software (false positive)
4. If malicious, isolate the affected system and perform full malware scan
5. Review system logs for any successful compromise attempts before HIPS activation
6. Update HIPS rules if necessary to prevent similar attacks
7. Check for persistence mechanisms on the affected host
8. Review network connections from the suspicious process if applicable
', '["https://help.eset.com/ees/8/en-US/idh_hips_main.html","https://attack.mitre.org/techniques/T1055/"]', 'equals("log.actionResult", "HIPS_Event") && equals("log.action", "blocked") && oneOf("log.severity", ["medium", "high"])', '2026-03-02 16:02:32.971290', true, true, 'origin', null, '[]', '["lastEvent.log.objectname","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1146, 'ESET Exploit Detection Alert', 3, 3, 2, 'Privilege Escalation', 'T1068 - Exploitation for Privilege Escalation', e'Detects when ESET\'s Exploit Blocker identifies and blocks exploitation attempts targeting vulnerabilities in commonly exploited applications such as browsers, document readers, email clients, Flash, and Java.

Next Steps:
1. Identify the affected host and the specific exploit attempt details
2. Check for any successful exploitation attempts on the same host
3. Review process execution logs for suspicious activity following the exploit attempt
4. Verify that the exploit was successfully blocked and no compromise occurred
5. Update the vulnerable application if a patch is available
6. Consider isolating the host if exploitation may have succeeded
', '["https://www.eset.com/us/about/technology/","https://attack.mitre.org/techniques/T1068/"]', e'(contains("log.jsonMessage", "exploit") ||
 oneOf("log.msgType", ["Exploit_Blocked", "Exploit"])) &&
equals("actionResult", "blocked") &&
oneOf("log.severity", ["medium", "high"])
', '2026-03-02 16:02:34.327549', true, true, 'origin', null, '[]', '["lastEvent.log.jsonMessage","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1147, 'ESET Agent Disabled or Tampered', 3, 3, 3, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects when the ESET security agent is disabled, uninstalled, or tampered with. This is a critical defense evasion indicator as attackers commonly disable endpoint protection before executing their payload.

Next Steps:
1. Immediately investigate the affected endpoint
2. Determine who or what process disabled the agent
3. Check for concurrent malicious activity on the endpoint
4. Reinstall and re-enable the ESET agent
5. Review the endpoint for malware or unauthorized software
6. Check if similar tampering occurred on other endpoints
', '["https://help.eset.com/ees/8/en-US/idh_config_era_agent.html","https://attack.mitre.org/techniques/T1562/001/"]', e'(regexMatch("log.message", "(?i)(eset|ekrn|egui|agent)") &&
 regexMatch("log.message", "(?i)(disabled|stopped|uninstalled|removed|tampered|terminated)")) ||
(contains("log.message", "protection status") && contains("log.message", "disabled")) ||
(contains("log.message", "agent") && contains("log.message", "not responding")) ||
(equals("log.eventType", "AGENT_EVENT") && contains("log.message", "removed"))
', '2026-03-02 16:02:35.582991', true, true, 'origin', null, '[]', '["lastEvent.log.eventType","target.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1148, 'ESET Botnet Communication Detection', 3, 3, 2, 'Command and Control', 'T1071 - Application Layer Protocol', e'Detects attempts to communicate with known botnet command and control servers. ESET identifies typical communication patterns when a computer is infected and a bot is attempting to communicate with malicious C2 infrastructure.

Next Steps:
1. Immediately isolate the affected system from the network to prevent further C2 communication
2. Check the hostname (log.headHostname) to identify the affected system
3. Review the full log message content (log.jsonMessage) for additional threat details including target IPs and processes
4. Review process activity on the affected host to identify the malicious process
5. Scan the system with ESET for complete malware removal
6. Check other systems in the network for similar C2 communication attempts
7. Update firewall rules to block any identified C2 server IPs found in the logs
8. Consider reimaging the system if the infection persists
9. Review ESET logs for the time period around this detection to identify related malicious activity
', '["https://www.eset.com/us/botnet/","https://support.eset.com/en/kb7487-resolve-the-incomingattackgeneric-or-botnetcncgeneric-network-protection-alert","https://attack.mitre.org/techniques/T1071/"]', e'contains("log.jsonMessage", ["Botnet", "CnC.Generic", "botnet", "C&C", "command and control"]) &&
exists("log.headHostname")
', '2026-03-02 16:02:36.913305', true, true, 'origin', null, '[]', '["lastEvent.log.headHostname","lastEvent.log.jsonMessage"]');
