INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1174, 'Command and Control on Non-Standard Ports', 3, 2, 1, 'Command and Control', 'T1571 - Non-Standard Port', e'Detects connections on non-standard ports that may indicate command and control (C2) communication. Identifies HTTP traffic on non-HTTP ports, encrypted traffic on unexpected ports, and application protocol mismatches detected by Firepower\'s application identification engine.

Next Steps:
1. Investigate the internal host initiating the suspicious connection
2. Review the destination IP against threat intelligence feeds
3. Analyze the application identification results for protocol anomalies
4. Check if the destination port is commonly used for C2 frameworks
5. Examine the connection duration and data transfer patterns
6. Consider blocking the destination IP and scanning the internal host
', '["https://www.cisco.com/c/en/us/td/docs/security/secure-firewall/management-center/device-config/710/management-center-device-config-71/connection-log-fields.html","https://attack.mitre.org/techniques/T1571/"]', e'exists("log.appProto") &&
exists("origin.ip") &&
exists("target.ip") &&
((contains("log.appProto", "HTTP") && !oneOf("target.port", [80, 443, 8080, 8443, 8000, 8888])) ||
 (contains("log.appProto", "SSL") && !oneOf("target.port", [443, 8443, 993, 995, 465, 636])) ||
 equals("log.appProto", "unknown-tcp")) &&
equals("log.initiatorPackets", true)
', '2026-03-02 16:38:33.554409', true, true, 'origin', null, '[{"indexPattern":"v11-log-firewall-cisco-firepower-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"target.ip","operator":"filter_term","value":"{{.target.ip}}"}],"or":null,"within":"now-1h","count":5}]', '["adversary.ip","target.ip","target.port"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1175, 'Firepower IOC (Indicator of Compromise) Detection', 3, 3, 2, 'Initial Access', 'T1566 - Phishing', e'Detects when Firepower identifies an Indicator of Compromise (IOC), indicating a host may be infected with malware or compromised. IOCs can include file hashes, malware signatures, or behavioral patterns that suggest malicious activity.

Next Steps:
1. Immediately isolate the affected host from the network to prevent lateral movement
2. Review the specific IOC details including threat name, SHA256 hash, and file path
3. Search for the same IOC across other endpoints in your environment
4. Check if the affected host has made any suspicious network connections recently
5. Collect memory dumps and disk images for forensic analysis if required
6. Review user activity logs to identify potential initial compromise vector
7. Update antivirus signatures and threat intelligence feeds with new IOC data
8. Perform deep scan of the affected system and related network segments
9. Consider reimaging the affected system after complete evidence collection
10. Update security controls to prevent similar future compromises
', '["https://www.cisco.com/c/en/us/td/docs/security/firepower/70/configuration/guide/fpmc-config-guide-v70/file_malware_events_and_network_file_trajectory.html","https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1566/"]', '(equals("log.eventType", "AMP_IOC") || equals("log.eventType", "IOC_DETECTED") || contains("log.message", "indication of compromise") || contains("log.message", "IOC")) && exists("origin.ip") && (exists("log.threatName") || exists("log.sha256") || exists("log.fileName"))', '2026-03-02 16:38:34.877284', true, true, 'origin', null, '[]', '["adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1176, 'Threat Intelligence Director (TID) Alert Detection', 3, 3, 2, 'Command and Control', 'T1071.001 - Application Layer Protocol: Web Protocols', e'Detects when Cisco Firepower Threat Intelligence Director identifies connections to known malicious indicators including IPs, domains, URLs, and SHA256 hashes from threat feeds. This rule triggers when TID blocks or would block connections based on threat intelligence matches with high confidence scores.

Next Steps:
- Immediately isolate the affected system if the connection was not blocked
- Review the specific threat indicator (IP/domain/URL/hash) that triggered the alert
- Check the threat category and score to understand the severity
- Investigate all recent network activity from the source IP address
- Search for similar indicators across other systems in the network
- Review endpoint logs for signs of malware or suspicious processes
- If a file hash triggered the alert, locate and quarantine the file
- Check if other systems have communicated with the same malicious indicator
- Update firewall rules to ensure the indicator is blocked network-wide
- Report the incident to the security team for further investigation
', '["https://www.cisco.com/c/en/us/td/docs/security/firepower/70/configuration/guide/fpmc-config-guide-v70/tid_overview.html","https://attack.mitre.org/techniques/T1071/"]', e'equals("log.eventType", "TID_EVENT") &&
(equals("log.action", "BLOCK") ||
 equals("log.action", "WOULD_BLOCK") ||
 exists("log.tidIndicatorType")) &&
(exists("log.tidCategory") ||
 greaterOrEqual("log.threatScore", 80))
', '2026-03-02 16:38:36.144371', true, true, 'origin', null, '[]', '["lastEvent.log.tidIndicator","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1177, 'Intrusion Prevention System High Priority Events', 3, 3, 3, 'Execution', 'T1203 - Exploitation for Client Execution', e'Detects high priority IPS events from Cisco Firepower indicating potential exploitation attempts, zero-day attacks, or critical vulnerabilities being targeted. These events represent immediate threats that require urgent investigation.

Next Steps:
1. Immediately isolate the affected system if the attack was successful
2. Review the specific signature ID and classification to understand the attack vector
3. Check if the target system shows signs of compromise (unusual processes, network connections, file modifications)
4. Analyze firewall logs to determine if the attack was blocked or if any malicious traffic passed through
5. Search for similar attempts from the same source IP across other systems
6. Update IPS signatures and ensure all systems are patched against the exploited vulnerability
7. Consider blocking the source IP if it shows persistent malicious behavior
8. Document the incident and update security controls based on findings
', '["https://www.cisco.com/c/en/us/td/docs/security/secure-firewall/management-center/device-config/710/management-center-device-config-71/intrusion-overview.html","https://attack.mitre.org/techniques/T1203/"]', e'equals("log.eventType", "IPS_EVENT") &&
(equals("log.priority", 1) ||
 lessOrEqual("log.severity", 2) ||
 equals("log.impact", "HIGH") ||
 contains("log.classification", "attempted-admin") ||
 contains("log.classification", "attempted-user") ||
 contains("log.classification", "web-application-attack") ||
 contains("log.classification", "exploit-kit"))
', '2026-03-02 16:38:37.496057', true, true, 'origin', null, '[]', '["adversary.ip","target.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1178, 'Advanced Malware Protection (AMP) Alert Detection', 3, 3, 2, 'Initial Access', 'T1566 - Phishing', e'Detects when Cisco Firepower Advanced Malware Protection (AMP) identifies malware or malicious files, including retrospective detections and high threat score files. This rule triggers on various malware dispositions including confirmed malware, custom detections, retrospective malware (files later identified as malicious), and files with high threat scores (>=70).

Next Steps:
1. Identify the affected host using the origin.ip and log.deviceName fields
2. Review the file hash (log.sha256) in threat intelligence databases
3. Check if the malware was successfully blocked or if remediation is needed
4. Look for lateral movement attempts from the affected host
5. Verify if other hosts accessed the same malicious file
6. Consider isolating the affected system if malware execution is confirmed
7. Review the file trajectory to understand the infection vector
8. Update endpoint protection rules to prevent similar infections
', '["https://www.cisco.com/c/en/us/td/docs/security/firepower/70/configuration/guide/fpmc-config-guide-v70/file_malware_events_and_network_file_trajectory.html","https://attack.mitre.org/techniques/T1566/"]', e'equals("log.eventType", "MALWARE_EVENT") &&
(equals("log.disposition", "MALWARE") ||
 equals("log.disposition", "CUSTOM_DETECTION") ||
 equals("log.disposition", "RETROSPECTIVE_MALWARE") ||
 greaterOrEqual("log.threatScore", 70))
', '2026-03-02 16:38:38.796270', true, true, 'origin', null, '[]', '["lastEvent.log.sha256","adversary.ip"]');
