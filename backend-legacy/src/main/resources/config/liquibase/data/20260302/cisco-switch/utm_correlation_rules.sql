INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1171, 'MAC Address Spoofing Detection', 2, 3, 1, 'Initial Access', 'MAC Spoofing', e'Detects potential MAC address spoofing attempts by monitoring for MAC address flapping between ports, duplicate MAC addresses, or MAC addresses appearing on unexpected ports. This could indicate an attacker attempting to impersonate legitimate devices.

Next Steps:
1. Identify the affected MAC address and ports involved in the flapping
2. Check if the MAC address belongs to a legitimate device that may be moving between ports
3. Review switch logs for any unauthorized configuration changes
4. Verify if port security or dynamic ARP inspection is properly configured
5. Investigate the source device and check for signs of ARP spoofing tools
6. Consider implementing port security to limit MAC addresses per port
7. Enable DHCP snooping and dynamic ARP inspection if not already configured
', '["https://www.cisco.com/c/en/us/support/docs/switches/catalyst-3750-series-switches/72846-layer2-secftrs-catl3fixed.html","https://attack.mitre.org/techniques/T1200/"]', e'(equals("log.facility", "SW_MATM") && equals("log.facilityMnemonic", "MACFLAP_NOTIF"))
|| (equals("log.facility", "SW_DAI") && oneOf("log.facilityMnemonic", ["INVALID_ARP", "DHCP_SNOOPING_DENY"]))
|| regexMatch("log.message", "(?i)(mac.*flap|duplicate.*mac|mac.*move.*between.*port)")
|| regexMatch("log.message", "(?i)(Host [0-9a-fA-F:.]+.*is flapping between port)")
|| (lessOrEqual("log.severity", 4) && regexMatch("log.message", "(?i)(mac.*address.*conflict|duplicate.*address.*detected)"))
', '2026-03-02 16:31:52.224411', true, true, 'origin', null, '[{"indexPattern":"v11-log-cisco-switch-*","with":[{"field":"origin.mac","operator":"filter_term","value":"{{.origin.mac}}"}],"or":null,"within":"now-10m","count":3}]', '["adversary.mac"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1172, 'ARP Poisoning Attack Detection', 3, 3, 2, 'Credential Access, Collection', 'T1557.002 - Adversary-in-the-Middle: ARP Cache Poisoning', e'Detects potential ARP poisoning attacks by monitoring for invalid ARP packets, DHCP snooping violations, and gratuitous ARP abuse. These attacks can enable man-in-the-middle attacks by corrupting the ARP cache of network devices and redirecting network traffic through an attacker-controlled system.

Next Steps:
1. Identify the source MAC and IP addresses involved in the suspicious ARP activity
2. Check if the source device is authorized to be on the network segment
3. Review DHCP snooping and dynamic ARP inspection logs for additional violations
4. Verify if legitimate network changes (new devices, IP changes) may have triggered the alert
5. If confirmed malicious, immediately isolate the affected switch port and investigate the compromised device
6. Review network traffic for signs of data interception, credential harvesting, or traffic redirection
7. Update switch security configurations (enable port security, DHCP snooping, DAI if not already enabled)
8. Consider implementing additional network segmentation to limit attack impact
', '["https://www.cisco.com/c/en/us/td/docs/switches/lan/catalyst4500/12-2/25ew/configuration/guide/conf/dynarp.html","https://attack.mitre.org/techniques/T1557/002/"]', e'(equals("log.facility", "SW_DAI") && oneOf("log.facilityMnemonic", ["INVALID_ARP", "DHCP_SNOOPING_DENY", "ACL_DENY"]))
|| (equals("log.facility", "IP") && oneOf("log.facilityMnemonic", ["DUPADDR", "SOURCEGUARD"]))
|| contains("log.message", ["invalid arp", "arp inspection drop", "dhcp snooping deny", "gratuitous arp", "arp reply not request", "duplicate ip address", "IP source guard deny", "arp packet validation failed"])
|| (lessOrEqual("log.severity", 3) && contains("log.message", ["arp spoofing", "arp poison", "man in the middle"]))
', '2026-03-02 16:31:53.525673', true, true, 'origin', null, '[{"indexPattern":"v11-log-cisco-switch-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-10m","count":5}]', '["adversary.ip","adversary.mac"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1173, 'VLAN Hopping Attack Detection', 3, 3, 2, 'Defense Evasion', 'T1599 - Network Boundary Bridging', e'Detects potential VLAN hopping attacks through switch spoofing or double tagging. Monitors for DTP negotiation attempts, trunk port changes, or multiple VLAN tags that could indicate an attacker trying to gain unauthorized access to other VLANs.

Next Steps:
1. Immediately identify the affected switch port and connected device
2. Review switch configuration for DTP enabled ports and disable where not needed
3. Check trunk port configurations and ensure proper native VLAN settings
4. Verify VLAN access lists and ensure proper segmentation
5. Investigate the source MAC address for any previous suspicious activity
6. Review network topology to assess potential lateral movement paths
7. Consider implementing VLAN ACLs or private VLANs for additional protection
8. Document the incident and update switch hardening procedures
', '["https://www.cisco.com/c/en/us/support/docs/switches/catalyst-3750-series-switches/72846-layer2-secftrs-catl3fixed.html","https://attack.mitre.org/techniques/T1599/"]', e'(equals("log.facility", "SW_VLAN") && oneOf("log.facilityMnemonic", ["VLAN_INCONSISTENCY", "MACFLAP_NOTIF", "TRUNK_MODE_CHANGE"]))
|| (equals("log.facility", "DTP") && oneOf("log.facilityMnemonic", ["NONTRUNKPORTON", "DOMAINMISMATCH", "TRUNKPORTON"]))
|| regexMatch("log.message", "(?i)(received 802.1Q BPDU on non trunk|native vlan mismatch|inconsistent vlan|double tag)")
|| (lessOrEqual("log.severity", 4) && regexMatch("log.message", "(?i)(vlan.*tag.*tag|switch.*spoofing|dtp.*negotiation)"))
', '2026-03-02 16:31:54.914197', true, true, 'origin', null, '[]', '["adversary.ip","adversary.mac"]');
