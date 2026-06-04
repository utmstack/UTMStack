INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1168, 'Multiple Failed VPN Authentication Attempts', 3, 2, 1, 'Credential Access', 'T1110 - Brute Force', e'Detects multiple failed VPN authentication attempts from the same source IP address, which could indicate a brute force attack or password guessing attempt against VPN credentials.

Next Steps:
- Review the source IP address and check if it is known or authorized
- Check for successful authentication attempts from the same IP after failed attempts
- Verify if the targeted user accounts exist and are active
- Consider temporarily blocking the source IP if attack continues
- Review VPN access logs for any unusual patterns or other indicators
- Contact the user if the IP is associated with a legitimate user to verify activity
', '["https://attack.mitre.org/techniques/T1110/","https://www.cisco.com/c/en/us/td/docs/security/asa/syslog/b_syslog/syslogs1.html"]', e'(equals("log.messageId", "113015") ||
 equals("log.messageId", "113021") ||
 equals("log.messageId", "109034") ||
 equals("log.messageId", "611102")) &&
exists("origin.ip") &&
(regexMatch("log.reason", "(?i)(invalid|failed|rejected|authentication)") ||
 regexMatch("log.message", "(?i)(authentication.*failed|invalid.*password)"))
', '2026-03-02 16:23:36.293561', true, true, 'origin', null, '[{"indexPattern":"v11-log-firewall-cisco-asa-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.messageId","operator":"filter_term","value":"113015"}],"or":[{"indexPattern":"v11-log-firewall-cisco-asa-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.messageId","operator":"filter_term","value":"113021"}],"or":null,"within":"now-15m","count":10},{"indexPattern":"v11-log-firewall-cisco-asa-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.messageId","operator":"filter_term","value":"109034"}],"or":null,"within":"now-15m","count":10},{"indexPattern":"v11-log-firewall-cisco-asa-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.messageId","operator":"filter_term","value":"611102"}],"or":null,"within":"now-15m","count":10}],"within":"now-15m","count":10}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1169, 'IPS Signature Match - Malicious Pattern Detected', 3, 3, 2, 'Initial Access', 'T1190 - Exploit Public-Facing Application', e'Detects when ASA IPS features identify malicious patterns in network traffic. Message ID 108003 indicates ESMTP/SMTP connections terminated due to malicious patterns. Also monitors for general IPS/IDS signature matches and threat intelligence hits.

Next Steps:
1. Review the specific IPS signature that was triggered and understand its severity
2. Investigate the source IP address for reputation and previous malicious activity
3. Check if the target system shows any signs of compromise
4. Review firewall logs for any successful connections from the same source
5. Consider blocking the source IP if multiple signatures are triggered
6. Verify that IPS signatures are up-to-date
7. Document the incident and any actions taken
', '["https://www.cisco.com/c/en/us/td/docs/security/asa/syslog/b_syslog.html","https://attack.mitre.org/techniques/T1190/"]', e'equals("log.messageId", "108003")
|| (contains("log.message", "malicious pattern") && contains("log.message", ["detected", "terminated", "blocked"]))
|| (contains("log.message", "IPS") && contains("log.message", "signature") && contains("log.message", ["matched", "triggered", "detected"]))
|| oneOf("log.action", ["ips_alert", "ids_alert", "threat_detected"])
', '2026-03-02 16:23:37.737324', true, true, 'origin', null, '[{"indexPattern":"v11-log-firewall-cisco-asa-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-15m","count":3}]', '["adversary.ip","target.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1170, 'Botnet Command and Control Traffic Detected', 3, 2, 1, 'Command and Control', 'T1071 - Application Layer Protocol', e'Detects botnet command and control traffic identified by Cisco ASA\'s dynamic filter/botnet database. Message IDs 338001-338002 indicate blacklisted traffic from/to malicious addresses. This could indicate compromised hosts communicating with known botnet infrastructure.

Next Steps:
1. Immediately isolate the affected host(s) to prevent further communication with C2 infrastructure
2. Review the source IP address (origin.ip) to identify the compromised internal host
3. Check the destination IP/domain against threat intelligence sources to confirm malicious nature
4. Examine other logs from the affected host for signs of initial compromise or lateral movement
5. Run full antivirus/anti-malware scans on the affected system
6. Review DNS logs for additional suspicious queries from the same host
7. Check for any data exfiltration attempts or unusual outbound traffic patterns
8. Consider reimaging the affected system if compromise is confirmed
', '["https://www.cisco.com/c/en/us/td/docs/security/asa/special/botnet/asa-botnet.pdf","https://attack.mitre.org/techniques/T1071/"]', e'oneOf("log.messageId", ["338001", "338002"])
|| regexMatch("log.message", "botnet.*(detected|blocked|dropped)")
|| contains("log.message", "dynamic filter blacklisted")
|| contains("log.message", "malicious address")
', '2026-03-02 16:23:38.875569', true, true, 'origin', null, '[]', '["adversary.ip","target.ip"]');
