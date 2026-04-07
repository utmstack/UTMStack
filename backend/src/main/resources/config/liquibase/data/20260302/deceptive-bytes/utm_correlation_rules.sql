INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1120, 'Zero-Day Behavior Patterns Detection', 3, 3, 3, 'Defense Evasion', 'T1211 - Exploitation for Defense Evasion', e'Identifies potential zero-day exploits and unknown malware through abnormal behavior patterns, deception interactions, and anomaly detection in endpoint activities.

Next Steps:
1. Immediately isolate the affected system from the network to prevent lateral movement
2. Capture memory dumps and process information for forensic analysis
3. Check for similar behavioral anomalies on other endpoints in the same network segment
4. Review the exploit technique and process chain to understand the attack vector
5. Submit samples to threat intelligence platforms for analysis
6. Update security controls based on the identified exploit patterns
7. Document all IOCs (file hashes, network connections, process behaviors) for threat hunting
', '["https://attack.mitre.org/techniques/T1211/","https://attack.mitre.org/techniques/T1055/","https://attack.mitre.org/techniques/T1620/"]', e'oneOf("log.eventType", ["unknown_threat", "behavioral_anomaly", "zero_day_suspect"]) &&
equals("log.threatSignature", "unknown") &&
equals("log.deceptionEnvironment", true) &&
(
  (greaterOrEqual("log.memoryAnomalyScore", 90)) ||
  (greaterOrEqual("log.processChainAnomalyScore", 85)) ||
  (greaterOrEqual("log.networkBehaviorScore", 88)) ||
  (greaterOrEqual("log.fileSystemAnomalyScore", 92))
) &&
equals("log.knownMalwareFamily", "") &&
exists("log.exploitTechnique")
', '2026-03-02 15:48:34.556929', true, true, 'origin', null, '[{"indexPattern":"v11-log-deceptive-bytes-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.processName","operator":"filter_term","value":"{{.log.processName}}"}],"or":null,"within":"now-30m","count":2}]', '["lastEvent.log.exploitTechnique","lastEvent.log.processHash","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1121, 'Ransomware Behavior Detected in Deception Environment', 3, 3, 3, 'Impact', 'T1486 - Data Encrypted for Impact', e'Detects ransomware-like behavior patterns when attackers interact with deceptive files, including rapid file enumeration, encryption attempts, and ransom note creation in the Deceptive Bytes deception environment.

Next Steps:
1. Immediately isolate the affected system from the network to prevent ransomware spread
2. Check the process name and path identified in the alert for known ransomware indicators
3. Review file system activity from the same process for encryption patterns
4. Check for shadow copy deletion attempts (vssadmin, wmic shadowcopy delete)
5. Look for network connections to potential C2 servers from the identified process
6. Preserve forensic evidence and memory dumps if possible
7. Verify if this is a deception environment interaction or production system compromise
8. Check for lateral movement attempts from the source IP address
9. Review backup integrity and availability before any restoration attempts
', '["https://attack.mitre.org/techniques/T1486/","https://attack.mitre.org/techniques/T1490/","https://deceptivebytes.com/solution/"]', e'equals("log.event_type", "ransomware_behavior") &&
oneOf("log.behavior_pattern", ["mass_encryption", "file_enumeration", "ransom_note_drop"])
', '2026-03-02 15:48:35.988006', true, true, 'origin', null, '[{"indexPattern":"v11-log-deceptive-bytes-*","with":[{"field":"log.process","operator":"filter_term","value":"{{.log.process}}"},{"field":"log.source_ip","operator":"filter_term","value":"{{.log.source_ip}}"}],"or":null,"within":"now-15m","count":10}]', '["lastEvent.log.hostname","lastEvent.log.process"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1122, 'Honey Table Query Detection', 3, 2, 1, 'Collection', 'T1005 - Data from Local System', e'Detects when an attacker queries honey tables or decoy database objects deployed by Deceptive Bytes. This indicates potential data exfiltration attempts or database reconnaissance. Honey tables are deliberately placed decoy data designed to attract and identify unauthorized access attempts.

Next Steps:
1. Identify the source IP and determine if it\'s an internal or external address
2. Check if the source IP has accessed other decoy resources or legitimate database tables
3. Review the specific honey table(s) that were queried to understand attacker interest
4. Correlate with authentication logs to identify the user account used
5. Check for any data exfiltration patterns following the honey table access
6. Isolate the compromised system or account if malicious activity is confirmed
7. Review database access logs for unauthorized queries to legitimate tables
8. Consider blocking the source IP if it\'s external and confirmed malicious
9. Document the incident and update security monitoring rules if needed
', '["https://attack.mitre.org/techniques/T1005/","https://deceptivebytes.com/solution/"]', 'equals("log.eventType", "decoy_access") && equals("log.resourceType", "database_table") && equals("log.action", "query") && exists("origin.ip")', '2026-03-02 15:48:37.349533', true, true, 'origin', null, '[{"indexPattern":"v11-log-deceptive-bytes-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.eventType","operator":"filter_term","value":"decoy_access"}],"or":null,"within":"now-1h","count":5}]', '["lastEvent.log.tableName","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1123, 'Decoy Share Access Monitoring', 3, 2, 1, 'Discovery', 'T1135 - Network Share Discovery', e'Detects when an attacker attempts to access decoy network shares set up by Deceptive Bytes. This indicates potential lateral movement or reconnaissance activity within the network. Any interaction with decoy shares is a high-confidence indicator of malicious activity since legitimate users should never access these resources.

Next Steps:
- Immediately investigate the source IP and verify if it belongs to an authorized user or system
- Check for other suspicious activities from the same source IP in the last 24-48 hours
- Review authentication logs to identify any compromised credentials associated with this IP
- Look for lateral movement attempts or privilege escalation from the same source
- Consider isolating the source system if it shows signs of compromise
- Document all accessed decoy resources for threat intelligence purposes
- Update security controls to block or monitor the attacker\'s techniques
', '["https://attack.mitre.org/techniques/T1135/","https://deceptivebytes.com/solution/"]', 'equals("log.eventType", "decoy_access") && equals("log.resourceType", "network_share") && exists("origin.ip")', '2026-03-02 15:48:38.654988', true, true, 'origin', '["adversary.ip","lastEvent.log.resourceType"]', '[{"indexPattern":"v11-log-deceptive-bytes-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.eventType","operator":"filter_term","value":"decoy_access"},{"field":"log.resourceType","operator":"filter_term","value":"network_share"}],"or":null,"within":"now-30m","count":3}]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1124, 'Deception Token Access Patterns', 3, 3, 1, 'Defense Evasion, Persistence, Privilege Escalation, Initial Access', 'T1078 - Valid Accounts: Credential Access', e'Detects when deception tokens or honeytokens are accessed, indicating potential unauthorized activity or insider threat. Multiple token accesses from the same source within a short timeframe suggest systematic reconnaissance or data harvesting attempts. Honeytokens are fake credentials or access tokens planted as traps to detect unauthorized access.

Next Steps:
1. Identify the source IP and user account associated with the token access
2. Review access logs to determine if this is legitimate testing or actual malicious activity
3. Check for lateral movement from the same source IP across the network
4. Investigate any data access or exfiltration attempts following the token access
5. Consider immediately blocking the source IP if confirmed malicious
6. Review and rotate any potentially compromised credentials in the environment
7. Alert security team immediately as honeytoken access is a high-confidence indicator of compromise
8. Document the incident and update detection rules based on observed attack patterns
9. Verify the integrity of the deception infrastructure to ensure it wasn\'t compromised
', '["https://attack.mitre.org/techniques/T1078/","https://deceptivebytes.com/"]', e'equals("log.eventType", "token_access") &&
equals("log.deceptionType", "honeytoken") &&
exists("origin.ip") &&
oneOf("log.severity", ["high", "critical"])
', '2026-03-02 15:48:40.056094', true, true, 'origin', null, '[{"indexPattern":"v11-log-deceptive-bytes-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.eventType","operator":"filter_term","value":"token_access"}],"or":null,"within":"now-1h","count":3}]', '["lastEvent.log.tokenId","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1125, 'Data Theft Attempt on Decoy Files', 3, 2, 1, 'Collection', 'T1005 - Data from Local System', e'Detects attempts to access, copy, or exfiltrate deceptive decoy files and honeypot data, indicating potential data theft activities by an attacker. This rule triggers when an attacker interacts with high-sensitivity decoy files planted by Deceptive Bytes.

Next Steps:
- Immediately isolate the affected endpoint to prevent lateral movement
- Review the source IP and user account for suspicious activity patterns
- Check for other decoy interactions from the same source in the past 24 hours
- Examine network traffic logs for potential data exfiltration attempts
- Verify if the user account has been compromised or if this is insider threat activity
- Consider resetting credentials for the affected user account
- Document all decoy files accessed for forensic analysis
', '["https://attack.mitre.org/techniques/T1005/","https://attack.mitre.org/techniques/T1567/","https://deceptivebytes.com/solution/"]', e'equals("log.event_type", "decoy_accessed") &&
oneOf("log.action", ["file_read", "file_copy", "file_download"]) &&
equals("log.decoy_sensitivity", "high")
', '2026-03-02 15:48:41.366195', true, true, 'origin', null, '[{"indexPattern":"v11-log-deceptive-bytes-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.event_type","operator":"filter_term","value":"decoy_accessed"}],"or":null,"within":"now-2h","count":3}]', '["lastEvent.log.decoy_file","adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1126, 'Advanced Threat Tactic Identification', 3, 3, 3, 'Advanced Persistent Threat', 'Multiple Tactics', e'Detects advanced threat tactics including initial access, execution, and persistence techniques by monitoring deception environment interactions and behavioral patterns. This rule triggers when deceptive assets are accessed with high behavior scores indicating sophisticated attack patterns.

Next Steps:
1. Immediately isolate the affected endpoint(s) associated with the source IP
2. Review the specific tactic name to understand the attack phase (initial access, execution, persistence, etc.)
3. Check all deception assets that were triggered to map the attacker\'s movement
4. Analyze the behavior score details to understand the sophistication level
5. Look for related alerts from the same source IP across different systems
6. Collect forensic data from the endpoint before any remediation
7. Review authentication logs for any credential abuse from this source
8. Check network logs for lateral movement attempts
9. Update security controls to block the identified tactics
10. Consider deploying additional deception assets in the path of the attacker
', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/tactics/TA0002/","https://attack.mitre.org/tactics/TA0003/"]', e'equals("log.eventType", "advanced_threat_detected") &&
equals("log.threatLevel", "critical") &&
(oneOf("log.tacticName", ["initial_access", "execution", "persistence", "privilege_escalation", "defense_evasion"])) &&
equals("log.deceptionTriggered", true) &&
greaterOrEqual("log.behaviorScore", 80)
', '2026-03-02 15:48:42.717733', true, true, 'origin', null, '[{"indexPattern":"v11-log-deceptive-bytes-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.tacticName","operator":"filter_term","value":"{{.log.tacticName}}"}],"or":null,"within":"now-15m","count":3}]', '["lastEvent.log.tacticName","lastEvent.log.threatLevel","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1127, 'Privilege Escalation Bait Accessed', 3, 3, 2, 'Privilege Escalation', 'T1068 - Exploitation for Privilege Escalation', e'Detects when an attacker accesses deceptive privileged account baits or attempts to escalate privileges using trapped credentials, indicating active exploitation attempts. This is a high-priority alert as it indicates an active attacker who has progressed beyond initial access and is attempting to gain elevated privileges.

Next Steps:
- Immediately isolate the affected system to prevent lateral movement
- Review authentication logs for the source IP and user account to identify scope of compromise
- Check for other deception bait interactions from the same source in the past 24 hours
- Investigate any legitimate user activity that may have been compromised
- Collect forensic data from the endpoint including running processes and network connections
- Review SIEM/EDR alerts for related suspicious activities from the same source
- Document the attacker\'s TTPs for threat intelligence sharing
- Consider resetting credentials for any accounts that may have been exposed
- Update firewall rules to block the attacker\'s source IP if confirmed malicious
', '["https://attack.mitre.org/techniques/T1068/","https://attack.mitre.org/techniques/T1078/","https://www.checkpoint.com/cyber-hub/cyber-security/what-is-deception-technology/"]', e'equals("log.event_type", "bait_accessed") &&
equals("log.bait_type", "privileged_account") &&
oneOf("log.target_privilege", ["admin", "system", "administrator"])
', '2026-03-02 15:48:43.983729', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1128, 'Threat Actor Attribution', 2, 2, 2, 'Threat Intelligence', 'T1583 - Acquire Infrastructure', e'Correlates observed attack patterns, tools, techniques, and infrastructure with known threat actor profiles to provide attribution intelligence and identify potential threat actors based on high-confidence indicators.

Next Steps:
1. Review the attributed threat actor profile and historical campaigns for context
2. Analyze the specific TTPs (Tactics, Techniques, and Procedures) that triggered the attribution
3. Check for related activity from the same actor across other systems or time periods
4. Correlate with threat intelligence feeds to validate attribution confidence
5. Document observed infrastructure and tooling for future threat hunting
6. Consider implementing specific detections for this actor\'s known techniques
7. Share attribution indicators with security teams for enhanced monitoring
8. Escalate to incident response team if high-profile threat actor is identified
', '["https://attack.mitre.org/groups/","https://malpedia.caad.fkie.fraunhofer.de/"]', e'equals("log.eventType", "threat_attribution") &&
greaterOrEqual("log.attributionConfidence", 70) &&
exists("log.actorProfile") &&
equals("log.deceptionTriggered", true) &&
(greaterOrEqual("log.ttpsMatched", 3) ||
 equals("log.infrastructureMatch", true) ||
 exists("log.toolingFingerprint")) &&
equals("log.historicalCampaignMatch", true)
', '2026-03-02 15:48:45.260156', true, true, 'origin', null, '[]', '["lastEvent.log.actorProfile","lastEvent.log.campaignId","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1129, 'Nation-State Tactic Detection', 3, 3, 3, 'Advanced Persistent Threat', 'T1595 - Active Scanning / Nation-State Attack Patterns', e'Detects sophisticated attack patterns and techniques commonly associated with nation-state actors including advanced persistence mechanisms, custom tooling, and strategic lateral movement.

Next Steps:
1) Immediately isolate affected systems and preserve forensic evidence
2) Review all decoy interactions and identify compromised credentials
3) Check for lateral movement attempts from the source IP across all systems
4) Analyze custom tools or malware samples if detected
5) Engage incident response team for potential APT activity
6) Review network traffic for command & control communications
7) Implement enhanced monitoring on high-value targets identified in the attack
', '["https://attack.mitre.org/groups/","https://www.cisa.gov/topics/cyber-threats-and-advisories/advanced-persistent-threats"]', e'oneOf("log.event_type", ["decoy_interaction", "honeypot_access", "deception_triggered"]) &&
equals("log.threat_level", "critical") &&
(equals("log.attack_sophistication", "advanced") || greaterOrEqual("log.threat_score", 85)) &&
(equals("log.apt_indicators", true) ||
 equals("log.custom_malware", true) ||
 equals("log.advanced_ttps", true) ||
 greaterThan("log.targeted_decoys", 1) ||
 equals("log.persistence_attempt", true))
', '2026-03-02 15:48:46.393702', true, true, 'origin', null, '[]', '["adversary.host","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1130, 'Living Off The Land Attack Using Deceptive Resources', 3, 3, 2, 'Defense Evasion', 'T1218 - Signed Binary Proxy Execution', e'Detects when attackers use legitimate system tools and binaries to interact with deceptive resources, indicating Living Off The Land (LOLBins) attack techniques. This is a high-confidence indicator of malicious activity as legitimate users should not be accessing deceptive resources with system binaries.

Next Steps:
1. Immediately isolate the affected system to prevent lateral movement
2. Review the process execution chain to identify the parent process and any child processes
3. Check if the user account is compromised by reviewing recent authentication logs
4. Examine command line arguments and scripts executed by the LOLBin
5. Search for other deceptive resource interactions from the same user or system
6. Collect memory dump if possible for forensic analysis
7. Review network connections made by the process for C2 communication
8. Check for persistence mechanisms (scheduled tasks, registry modifications, services)
', '["https://attack.mitre.org/techniques/T1218/","https://attack.mitre.org/techniques/T1053/","https://lolbas-project.github.io/","https://deceptivebytes.com/solution/"]', 'equals("log.event_type", "lolbin_trap") && oneOf("log.process_name", ["powershell.exe", "cmd.exe", "wmic.exe", "mshta.exe", "rundll32.exe", "regsvr32.exe", "certutil.exe", "bitsadmin.exe"]) && exists("log.deceptive_target")', '2026-03-02 15:48:47.671466', true, true, 'origin', null, '[]', '["lastEvent.log.deceptive_target","adversary.host","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1131, 'Lateral Movement Trap Triggered', 3, 3, 2, 'Lateral Movement', 'T1021 - Remote Services', e'Detects when an attacker triggers a deceptive trap while attempting lateral movement across the network. This indicates potential compromise and active threat movement within the environment.

Next Steps:
1. Immediately isolate the source IP address to prevent further lateral movement
2. Review all activities from the source IP in the last 24-48 hours
3. Check if the source system shows signs of compromise (unusual processes, new services, etc.)
4. Identify what credentials or methods were used in the lateral movement attempt
5. Review network logs for any successful connections from this source to other systems
6. Initiate incident response procedures for potential active threat
7. Consider deploying additional deception tokens around critical assets
', '["https://attack.mitre.org/techniques/T1021/","https://deceptivebytes.com/solution/"]', 'equals("log.event_type", "trap_triggered") && equals("log.trap_type", "lateral_movement") && exists("origin.ip")', '2026-03-02 15:48:48.784797', true, true, 'origin', null, '[]', '["lastEvent.log.trap_type","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1132, 'Fake User Authentication Attempts', 3, 3, 1, 'Credential Access', 'T1110 - Brute Force', e'Detects authentication attempts using decoy user accounts created by Deceptive Bytes. This indicates an attacker has obtained what they believe are valid credentials and is attempting to use them.

Next Steps:
- Immediately investigate the source IP address for other suspicious activities
- Check if the same IP has triggered other deception alerts or security events
- Review how the attacker obtained the decoy credentials (phishing, credential dumping, insider threat)
- Examine network logs for lateral movement attempts from this IP
- Consider blocking the source IP if confirmed malicious
- Check for any legitimate user accounts that may have been compromised
- Review authentication logs for attempts using real credentials from the same source
- Notify the security team for potential active breach investigation
', '["https://attack.mitre.org/techniques/T1110/","https://deceptivebytes.com/solution/"]', 'equals("log.eventType", "authentication") && equals("log.isDecoyUser", true) && exists("log.authResult") && exists("origin.ip")', '2026-03-02 15:48:50.062435', true, true, 'origin', null, '[]', '["lastEvent.log.username","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1133, 'Decoy System Enumeration', 3, 2, 1, 'Discovery', 'T1082 - System Information Discovery', e'Detects when an attacker performs system enumeration activities on decoy systems or services. This includes port scanning, service discovery, or system information gathering on deception assets.

Next Steps:
- Immediately investigate the source IP address for other suspicious activities
- Check if the source IP has attempted to access other decoy or real systems
- Review network logs for lateral movement attempts from this source
- Consider blocking the source IP if malicious intent is confirmed
- Document the attack pattern for threat intelligence sharing
- Verify if the attacker has discovered any real assets alongside decoys
', '["https://attack.mitre.org/techniques/T1082/","https://deceptivebytes.com/solution/"]', e'equals("log.eventType", "system_enumeration") &&
(equals("log.isDecoy", true) || equals("log.isDecoy", "true")) &&
oneOf("log.action", ["port_scan", "service_discovery", "system_info"]) &&
exists("origin.ip")
', '2026-03-02 15:48:51.284971', true, true, 'origin', '["adversary.ip","lastEvent.log.targetHost","lastEvent.log.decoyName"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1134, 'Deception API Call Tracking', 2, 2, 1, 'Execution', 'T1106 - Native API', e'Tracks suspicious API calls made to decoy services or endpoints. This behavior indicates an attacker is attempting to interact with what they believe are legitimate services but are actually deception assets.

Next Steps:
- Review the source IP address and check if it\'s from a known legitimate source
- Examine the API endpoint accessed and the HTTP method used
- Look for other activity from the same IP address across all log sources
- Check if the source IP has accessed multiple decoy endpoints (indicating reconnaissance)
- Investigate any authentication tokens or credentials used in the API calls
- Consider blocking the source IP if malicious intent is confirmed
- Document the attack pattern for threat intelligence sharing
', '["https://attack.mitre.org/techniques/T1106/","https://deceptivebytes.com/solution/"]', 'equals("log.eventType", "api_call") && equals("log.isDecoy", "true") && exists("log.httpMethod") && exists("origin.ip")', '2026-03-02 15:48:52.639498', true, true, 'origin', null, '[]', '["lastEvent.log.apiEndpoint","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1135, 'Criminal Group Signatures', 3, 3, 2, 'Organized Crime Activity', 'Criminal Group TTPs', e'Identifies attack signatures and behavioral patterns associated with known criminal groups including ransomware operators, financial crime syndicates, and organized cybercrime operations. Deception technology detects malicious activity by monitoring interactions with decoy assets that should never be accessed in legitimate workflows.

Next Steps:
1. IMMEDIATE: Isolate the affected endpoint to prevent lateral movement
2. Verify the criminal group signature or toolset identified in the alert details
3. Check if the source IP/domain appears in threat intelligence feeds or previous incidents
4. Review all activity from the affected endpoint in the last 24-48 hours
5. Search for indicators of lateral movement or data staging activities
6. Scan other endpoints for similar patterns or IoCs
7. If ransomware indicators are present, activate ransomware response playbook
8. Collect forensic evidence: process creation logs, network connections, file modifications
9. Check for data exfiltration attempts to external IPs or cloud services
10. Review all user account activity associated with the endpoint for signs of compromise
11. Document all findings and coordinate with incident response team
12. Consider threat hunting across the environment for related criminal group activities
', '["https://attack.mitre.org/groups/","https://www.ic3.gov/Media/PDF/AnnualReport/2023_IC3Report.pdf","https://www.acalvio.com/cyber-deception/the-role-of-deception-technology-in-the-endpoint-security-reference-architecture/","https://www.cisa.gov/news-events/cybersecurity-advisories/aa23-320a"]', e'oneOf("log.eventType", ["threat_detected", "deception_triggered", "malicious_activity"]) &&
oneOf("log.threatType", ["criminal_group", "ransomware", "organized_crime"]) &&
(exists("log.signature") || exists("log.toolset") || exists("log.groupName")) &&
(oneOf("log.action", ["blocked", "detected", "prevented"]) ||
 oneOf("log.severity", ["high", "critical"])) &&
(oneOf("log.indicatorType", ["ransomware", "financial_theft", "cryptomining", "data_exfiltration"]) ||
 greaterOrEqual("log.threatScore", 70))
', '2026-03-02 15:48:53.794345', true, true, 'origin', null, '[]', '["lastEvent.log.groupName","lastEvent.log.signature","adversary.host"]');
