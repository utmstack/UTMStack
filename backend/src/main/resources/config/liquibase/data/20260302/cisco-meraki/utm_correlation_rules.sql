INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1179, 'Meraki Client VPN Brute Force Attempts', 3, 2, 1, 'Credential Access', 'T1110 - Brute Force', e'Detects multiple failed client VPN authentication attempts from the same source IP on Meraki MX appliances, indicating potential brute force attacks against VPN credentials.

Next Steps:
1. Review the source IP address and check geographic location
2. Verify if the targeted user account exists and is active
3. Check for any successful VPN connections from the same IP
4. Consider blocking the source IP at the MX appliance
5. Review VPN authentication settings and ensure MFA is enabled
6. Notify the targeted user if the account is legitimate
', '["https://documentation.meraki.com/MX/Client_VPN/Client_VPN_Overview","https://attack.mitre.org/techniques/T1110/"]', e'(oneOf("log.eventType", ["vpn_auth_failure", "client_vpn_auth_failure"]) ||
 (contains("log.message", "VPN") && contains("log.message", ["auth fail", "authentication failed", "invalid credentials"]))) &&
exists("origin.ip")
', '2026-03-02 16:46:57.549445', true, true, 'origin', null, '[{"indexPattern":"v11-log-firewall-meraki-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-15m","count":10}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1180, 'Wireless Intrusion Attempts', 3, 3, 2, 'Reconnaissance', 'T1595.002 - Active Scanning: Vulnerability Scanning', e'Detects wireless intrusion attempts including deauthentication attacks, association floods, and other wireless-specific attack patterns that could compromise the wireless network integrity.

Next Steps:
1. Review the wireless access point logs for the affected device
2. Identify the source MAC address and physical location if possible
3. Check for unauthorized devices or rogue access points in the vicinity
4. Verify wireless security configurations and update if necessary
5. Consider implementing additional wireless monitoring and detection capabilities
6. Document the incident and update security policies if needed
', '["https://documentation.meraki.com/General_Administration/Monitoring_and_Reporting/Syslog_Event_Types_and_Log_Samples","https://attack.mitre.org/techniques/T1595/002/"]', e'equals("log.eventType", "wids_alerted") ||
(equals("log.type", "airmarshal_events") &&
 (contains("log.subtype", "attack") ||
  contains("log.subtype", "flood") ||
  contains("log.subtype", "deauth"))) ||
(contains("log.message", "deauthentication attack") ||
 contains("log.message", "association flood") ||
 contains("log.message", "wireless intrusion"))
', '2026-03-02 16:46:59.169784', true, true, 'origin', '["adversary.ip","target.mac"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1181, 'Rogue SSID Detection', 3, 3, 2, 'Initial Access', 'T1200 - Hardware Additions', e'Detects when a rogue SSID is identified in the wireless environment. This could indicate an evil twin attack or unauthorized access point deployment attempting to intercept wireless traffic or credentials.

Next Steps:
1. Immediately investigate the rogue access point\'s physical location using the MAC address
2. Check if the rogue SSID name matches legitimate corporate SSIDs (potential evil twin attack)
3. Verify if the rogue AP is broadcasting from an unauthorized location
4. Review wireless client connection logs for any devices that may have connected to the rogue SSID
5. Consider performing a physical sweep of the area to locate and remove the unauthorized device
6. Update wireless intrusion detection policies if needed
7. Notify security team and facilities management for potential physical security breach
', '["https://documentation.meraki.com/General_Administration/Monitoring_and_Reporting/Syslog_Event_Types_and_Log_Samples","https://attack.mitre.org/techniques/T1200/"]', e'equals("log.eventType", "rogue_ssid_detected") ||
(equals("log.type", "airmarshal_events") &&
 equals("log.subtype", "rogue_ssid_detected")) ||
(contains("log.message", "rogue") &&
 contains("log.message", "SSID"))
', '2026-03-02 16:47:00.736696', true, true, 'origin', null, '[]', '["adversary.mac"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1182, 'Meraki IDS High Priority Intrusion Alert', 3, 3, 2, 'Initial Access', 'T1190 - Exploit Public-Facing Application', e'Detects high and medium priority intrusion detection alerts from Meraki IDS/IPS system. These alerts indicate potential exploitation attempts, malicious traffic patterns, or known attack signatures detected by the Snort engine.

Next Steps:
1. Review the specific signature that triggered the alert and assess its severity
2. Investigate the source IP for additional malicious activity or reputation
3. Check if the destination system shows signs of compromise
4. Verify if this is part of a larger attack campaign by correlating with other security events
5. Consider blocking the source IP if confirmed malicious
6. Review firewall rules and IPS signatures for potential tuning
7. Document the incident and update threat intelligence feeds if applicable
', '["https://documentation.meraki.com/General_Administration/Monitoring_and_Reporting/Syslog_Event_Types_and_Log_Samples","https://attack.mitre.org/techniques/T1190/"]', e'equals("log.eventType", "security_event") &&
equals("log.alertType", "ids_alerted") &&
lessOrEqual("log.priority", 2) &&
exists("origin.ip") &&
exists("target.ip")
', '2026-03-02 16:47:02.274134', true, true, 'origin', null, '[]', '["lastEvent.log.signature","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1183, 'Evil Twin Access Point Detection', 3, 3, 1, 'Wireless Security', 'T1557 - Adversary-in-the-Middle', e'Detects evil twin attacks where a rogue access point mimics a legitimate corporate SSID to intercept wireless traffic. Meraki Air Marshal identifies spoofed SSIDs that match corporate network names but originate from unauthorized hardware.

Next Steps:
1. Verify the detected SSID against authorized access point inventory
2. Check the BSSID (MAC address) against known Meraki access points
3. Use Air Marshal containment features to prevent client connections
4. Physically locate the rogue AP using signal strength triangulation
5. Check if any clients have connected to the rogue AP
6. Review network traffic from affected clients for signs of credential theft
', '["https://documentation.meraki.com/MR/Monitoring_and_Reporting/Air_Marshal","https://attack.mitre.org/techniques/T1557/"]', e'equals("log.eventType", "airmarshal_events") &&
(equals("log.type", "ssid_spoofing") ||
 equals("log.type", "rogue_ssid_detected") ||
 (contains("log.message", "SSID Spoofing") || contains("log.message", "Evil Twin"))) &&
exists("log.bssid")
', '2026-03-02 16:47:03.654329', true, true, 'origin', null, '[]', '["lastEvent.log.bssid","adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1184, 'Air Marshal Rogue Access Point Detection', 3, 3, 2, 'Initial Access', 'T1200 - Hardware Additions', e'Detects when Meraki Air Marshal identifies rogue access points or unauthorized SSIDs in the wireless environment. This could indicate malicious wireless infrastructure attempting to intercept traffic or perform man-in-the-middle attacks.

Next Steps:
1. Verify if the detected BSSID and SSID are known legitimate access points that may not be properly registered
2. Check the RSSI value to determine proximity - higher values indicate the rogue AP is closer to your infrastructure
3. Use wireless scanning tools to physically locate the rogue access point using the BSSID
4. Review network traffic logs for any suspicious connections to unknown wireless networks
5. Check if any sensitive data might have been exposed through connections to the rogue AP
6. Consider implementing MAC address filtering or 802.1X authentication to prevent unauthorized connections
7. Document the incident and update the wireless security policy if needed
', '["https://documentation.meraki.com/MR/Monitoring_and_Reporting/Air_Marshal","https://attack.mitre.org/techniques/T1200/"]', e'equals("log.eventType", "airmarshal_events") &&
equals("log.type", "rogue_ssid_detected") &&
exists("log.bssid") &&
greaterOrEqual("log.rssi", -50)
', '2026-03-02 16:47:05.132607', true, true, 'origin', null, '[]', '["adversary.host"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1185, 'Meraki Advanced Malware Protection Alert', 3, 3, 2, 'Execution', 'T1204.002 - User Execution: Malicious File', e'Detects when Meraki Advanced Malware Protection (AMP) identifies malicious files being downloaded or executed on the network. This includes retrospective alerts where files previously considered safe are later identified as malicious.

Next Steps:
1. Immediately isolate the affected system(s) from the network to prevent lateral movement
2. Review the malware details including file hash, name, and threat severity in the Meraki dashboard
3. Check if the malicious file was executed or only downloaded
4. Scan other systems for the same file hash to identify additional infections
5. Review network traffic logs from the affected IP for suspicious communications
6. If file was executed, perform full system scan and consider reimaging the affected device
7. Update endpoint protection signatures and ensure all systems are patched
8. Document the incident and update security policies if needed
', '["https://documentation.meraki.com/MX/Content_Filtering_and_Threat_Protection/Advanced_Malware_Protection_(AMP)","https://attack.mitre.org/techniques/T1204/002/"]', e'equals("log.eventType", "security_event") &&
(contains("log.message", "malware") ||
 contains("log.message", "AMP") ||
 contains("log.message", "malicious") ||
 equals("log.action", "malware_blocked") ||
 contains("log.eventName", "Advanced Malware Protection")) &&
exists("origin.ip")
', '2026-03-02 16:47:06.658050', true, true, 'origin', null, '[]', '["adversary.hostname","adversary.ip"]');
