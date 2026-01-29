insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1593, 'API Management Security Events', 3, 3, 2, 'Defense Evasion, Persistence, Privilege Escalation, Initial Access', 'T1078 - Valid Accounts', 'Detects suspicious API Management activities including authentication failures, unauthorized access attempts, or API policy violations in Azure API Management services.

Next Steps:
1. Review the specific API endpoint and operation that triggered the alert
2. Investigate the source IP address and user identity associated with the request
3. Check for patterns of failed authentication attempts from the same source
4. Verify if the API access request aligns with legitimate business needs
5. Review API Management policies and access controls
6. Check for any recent changes to API permissions or policies
7. Consider implementing additional rate limiting or IP restrictions if needed
', '["https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-azure-monitor","https://attack.mitre.org/techniques/T1078/"]', '(contains("log.operationName", "Microsoft.ApiManagement") || equals("log.category", "GatewayLogs")) && (oneOf("statusCode", [401, 403]) || equals("actionResult", "denied"))', '2026-01-29 16:18:52.034045', true, true, 'origin', '["origin.ip"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1594, 'Azure Event Hub Authorization Rule Created or Updated', 2, 3, 2, 'Cloud Storage Object', 'Collection', 'Identifies when an Event Hub Authorization Rule is created or updated in Azure.  An authorization rule is associated with specific rights (Listen, Send, Manage), and carries a pair of cryptographic keys.  When you create an Event Hubs namespace, a policy rule named RootManageSharedAccessKey is created for the namespace.  This has manage permissions for the entire namespace and it''s recommended that you treat this rule like an  administrative root account and don''t use it in your application.  Adversaries may create or modify authorization rules to establish persistence, exfiltrate data, or maintain access to Event Hub streams.', '["https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1537/","https://attack.mitre.org/tactics/TA0010/","https://learn.microsoft.com/en-us/azure/event-hubs/authorize-access-shared-access-signature"]', '(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) && (equalsIgnoreCase("log.operationName", "MICROSOFT.EVENTHUB/NAMESPACES/AUTHORIZATIONRULES/WRITE") || contains("log.operationName", "Microsoft.EventHub/namespaces/authorizationRules/write")) && (equals("log.resultType", "0"))', '2026-01-29 16:18:53.218299', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1595, 'Azure AD Conditional Access Policy Bypass Attempt', 3, 3, 1, 'Defense Evasion, Persistence, Privilege Escalation, Initial Access', 'T1078 - Valid Accounts', 'Detects potential attempts to bypass Azure AD Conditional Access policies through policy tampering or unauthorized modifications. Monitors for policy updates and deletions that could weaken security controls such as MFA requirements, device compliance checks, or location-based restrictions. Adversaries may modify or delete conditional access policies to facilitate unauthorized access, bypass security controls, or establish persistence.

Next Steps:
1. Immediately review the conditional access policy changes made and document all modifications
2. Verify the identity and authorization of the user who made the changes, including verifying they have legitimate administrative access
3. Check if the policy modifications align with approved change management processes and security approval workflows
4. Review Azure AD sign-in logs for any unusual authentication patterns or successful logins following the policy change
5. Assess if the modified policy creates security gaps, weakens access controls, or allows unauthorized access paths
6. Cross-reference the timing of policy changes with any recent security incidents or suspicious activities
7. Consider immediately reverting unauthorized changes and implementing stronger approval workflows for future policy modifications
8. Audit all other conditional access policies for similar unauthorized modifications
', '["https://danielchronlund.com/2022/01/07/the-attackers-guide-to-azure-ad-conditional-access/","https://learn.microsoft.com/en-us/entra/identity/conditional-access/overview","https://attack.mitre.org/techniques/T1078/"]', '(equalsIgnoreCase("log.category", "AuditLogs") || contains("log.category", "Audit")) && (oneOf("log.operationName", ["Update policy", "Delete policy", "Delete conditional access policy", "Update conditional access policy"]) || contains("log.operationName", "conditionalAccessPolicies")) && equals("log.resultType", "0")', '2026-01-29 16:18:54.353112', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1596, 'Azure Container Registry Critical Vulnerability Detected', 3, 3, 2, 'Initial Access', 'T1190 - Exploit Public-Facing Application', 'Detects critical or high severity vulnerabilities in container images within Azure Container Registry, including newly pushed images or recently scanned images with security issues.

Next Steps:
1. Review the vulnerability details and affected container image
2. Assess the impact and exploitability of the vulnerability
3. Update or patch the container image to address the vulnerability
4. Implement security scanning in CI/CD pipeline to prevent future issues
5. Consider quarantining affected images until patched
6. Monitor for any exploitation attempts against vulnerable containers
', '["https://learn.microsoft.com/en-us/azure/defender-for-cloud/defender-for-container-registries-introduction","https://attack.mitre.org/techniques/T1190/"]', 'contains("log.OperationName", "Microsoft.ContainerRegistry") && (equalsIgnoreCase("log.ResultType", "VulnerabilityFound") || equalsIgnoreCase("log.Category", "SecurityAssessment")) && (oneOf("severity", ["critical", "high"]) || greaterOrEqual("statusCode", "400"))', '2026-01-29 16:18:55.466506', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1597, 'Azure Key Vault Modified', 3, 3, 2, 'Unsecured Credentials', 'Credential Access', 'Identifies modifications to a Key Vault in Azure. The Key Vault is a service that safeguards encryption keys and secrets like certificates,  connection strings, and passwords. Because this data is sensitive and business critical, access to key vaults should be secured to allow  only authorized applications and users. Adversaries may modify Key Vault configurations to weaken security controls, add unauthorized access policies,  or change network rules to facilitate credential theft and unauthorized access to sensitive secrets.', '["https://attack.mitre.org/techniques/T1552/","https://attack.mitre.org/tactics/TA0006/","https://learn.microsoft.com/en-us/azure/key-vault/general/security-features"]', '(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) && (equalsIgnoreCase("log.operationName", "MICROSOFT.KEYVAULT/VAULTS/WRITE") || contains("log.operationName", "Microsoft.KeyVault/vaults/write")) && (equals("log.resultType", "0"))', '2026-01-29 16:18:56.863184', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1598, 'Azure Storage Account Key Regenerated', 3, 2, 2, 'Application Access Token', 'Credential Access', 'Identifies a rotation to storage account access keys in Azure. Regenerating access keys can affect any applications or Azure services that are dependent on the storage account key. Adversaries may regenerate a key as a means of acquiring credentials to access systems and resources, potentially locking out legitimate users while maintaining their own access. This technique can be used to establish persistence, disrupt operations, or facilitate data exfiltration from Azure Storage.', '["https://attack.mitre.org/techniques/T1528/","https://attack.mitre.org/tactics/TA0006/","https://learn.microsoft.com/en-us/azure/storage/common/storage-account-keys-manage"]', '(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) && (equalsIgnoreCase("log.operationName", "MICROSOFT.STORAGE/STORAGEACCOUNTS/REGENERATEKEY/ACTION") || contains("log.operationName", "Microsoft.Storage/storageAccounts/regeneratekey/action")) && (equals("log.resultType", "0"))', '2026-01-29 16:18:57.886582', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1599, 'Azure Defender for Cloud Critical Security Alert', 3, 3, 2, 'Intrusion Detection', 'T1001 - Initial Access', 'Detects critical and high severity alerts from Microsoft Defender for Cloud (formerly Azure Security Center) indicating potential active threats, malware infections, successful breach attempts, or suspicious activities that require immediate response. These alerts leverage advanced threat detection, behavioral analytics, and machine learning to identify security incidents across Azure resources.

Next Steps:
1. Review the full alert details in Microsoft Defender for Cloud portal
2. Verify the affected resource and assess the scope of potential compromise
3. Check for related suspicious activities on the affected resource and correlated events
4. Implement immediate containment measures if threat is confirmed
5. Review security policies and configurations for the affected resource
6. Document the incident and update security procedures as needed
7. Investigate the CompromisedEntity and Entities fields for IOCs
', '["https://learn.microsoft.com/en-us/azure/defender-for-cloud/alerts-overview","https://learn.microsoft.com/en-us/azure/defender-for-cloud/alerts-schemas","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/securityalert","https://attack.mitre.org/tactics/TA0001/"]', 'equalsIgnoreCase("log.AlertSeverity", "High") || equalsIgnoreCase("severity", "High") && (equalsIgnoreCase("log.ProductName", "Azure Security Center") || equalsIgnoreCase("log.ProductName", "Microsoft Defender for Cloud"))', '2026-01-29 16:18:59.231442', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1600, 'Azure Diagnostic Settings Deletion', 1, 3, 3, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', 'Detects the deletion of diagnostic settings in Azure, which are critical for sending platform logs, metrics, and activity data to destinations like Log Analytics workspaces, Event Hubs, or storage accounts. Adversaries delete diagnostic settings to evade detection by disabling security monitoring and audit logging capabilities.

This technique is commonly observed when attackers:
- Attempt to hide malicious activities from security teams
- Disable logging before executing destructive operations
- Remove evidence trails of their presence in the environment
- Prevent detection of lateral movement or data exfiltration

Legitimate deletions are rare and typically occur only during:
- Infrastructure decommissioning or major reconfigurations
- Cost optimization initiatives (but should be heavily scrutinized)
- Migration to new monitoring solutions

Next Steps:
1. Immediately verify if the deletion was authorized and documented
2. Identify who performed the operation and from which IP address
3. Check if diagnostic settings were immediately recreated (potential test)
4. Review recent activities on the affected resource for suspicious behavior
5. Verify if other resources had their diagnostic settings deleted
6. Restore diagnostic settings immediately to resume monitoring
7. Investigate the caller''s account for potential compromise
8. Check for other defense evasion techniques in the timeline
', '["https://attack.mitre.org/techniques/T1562/008/","https://attack.mitre.org/tactics/TA0005/","https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/diagnostic-settings","https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log"]', '(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) && (equalsIgnoreCase("log.operationName", "MICROSOFT.INSIGHTS/DIAGNOSTICSETTINGS/DELETE") || contains("log.operationName", "Delete diagnostic setting")) && (equalsIgnoreCase("log.resultType", "0"))', '2026-01-29 16:19:00.321909', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1601, 'Application Gateway WAF Security Alerts', 3, 3, 2, 'Initial Access', 'T1190 - Exploit Public-Facing Application', 'Detects Web Application Firewall alerts from Azure Application Gateway indicating potential web attacks or malicious activity. This rule triggers when WAF blocks or detects suspicious requests that match OWASP security rules, including SQL injection, cross-site scripting (XSS), command injection, and other common web exploits.

**Next Steps:**
1. Review the specific WAF rule ID and message details to understand the attack type
2. Analyze the source IP address for reputation and geographic location
3. Examine the request URL, headers, and payload for attack indicators
4. Check for additional requests from the same source IP within the time window
5. Verify if this is a legitimate application behavior or actual attack attempt
6. Consider implementing additional WAF rules or IP blocking if confirmed malicious
7. Review application logs for any successful bypass attempts
8. Investigate ruleId and ruleGroup to understand the specific OWASP rule triggered
', '["https://learn.microsoft.com/en-us/azure/web-application-firewall/ag/web-application-firewall-logs","https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-diagnostics","https://attack.mitre.org/techniques/T1190/"]', 'equalsIgnoreCase("log.category", "ApplicationGatewayFirewallLog") &&
(equalsIgnoreCase("log.properties.action", "Blocked") ||
equalsIgnoreCase("log.properties.action", "Matched") ||
!equals("log.properties.ruleId", ""))
', '2026-01-29 16:19:01.377231', true, true, 'origin', '["origin.ip","log.properties.ruleId"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1602, 'Azure Event Hub Deletion', 1, 3, 3, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', 'Detects the deletion of an Azure Event Hub, which is a critical event processing service that ingests and processes large volumes of events, logs, and telemetry data. Event Hubs are commonly used for security monitoring, log aggregation, and SIEM integration. Adversaries may delete Event Hubs to evade detection by disrupting log collection pipelines and preventing security events from reaching monitoring systems.

Threat Context:
- Event Hubs are often used to stream logs to SIEM solutions
- Deletion interrupts security monitoring and incident detection capabilities
- Can be part of anti-forensics activities to cover tracks
- May indicate an attempt to blind security operations before further attacks

Legitimate Use Cases:
- Decommissioning unused Event Hubs during cost optimization
- Infrastructure cleanup during application retirement
- Migration to new Event Hub namespaces or different logging solutions
- Testing and development environment cleanup

Suspicious Indicators:
- Event Hub actively receiving logs suddenly deleted
- Deletion performed by non-administrative accounts
- Multiple Event Hubs deleted in quick succession
- Deletion outside change management windows
- Deletion from unusual locations or IP addresses
- Event Hub connected to production SIEM or security monitoring

Next Steps:
1. Verify if the deletion was authorized via change management process
2. Identify who performed the deletion (caller) and their role
3. Check if the Event Hub was actively receiving security logs
4. Determine the impact on security monitoring and log collection
5. Review recent authentication activity for the caller account
6. Check for other suspicious activities in the timeline (diagnostic settings changes, etc.)
7. Verify if backups of the Event Hub configuration exist
8. If unauthorized, restore the Event Hub and investigate for account compromise
9. Review authorization rules and access policies for remaining Event Hubs
', '["https://attack.mitre.org/techniques/T1562/008/","https://attack.mitre.org/tactics/TA0005/","https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs","https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log"]', '(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) && (equalsIgnoreCase("log.operationName", "MICROSOFT.EVENTHUB/NAMESPACES/EVENTHUBS/DELETE") || contains("log.operationName", "Delete EventHub")) && (equalsIgnoreCase("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))', '2026-01-29 16:19:02.478107', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1603, 'Azure Firewall Policy Deletion', 1, 3, 3, 'Defense Evasion', 'T1562.004 - Impair Defenses: Disable or Modify System Firewall', 'Detects the deletion of an Azure Firewall Policy, which defines network and application rules, threat intelligence settings, and security configurations for Azure Firewall. Adversaries delete firewall policies to disable network security controls, eliminate barriers to lateral movement, enable unrestricted outbound communication, or facilitate data exfiltration without detection.

Threat Context:
- Firewall policies control network traffic filtering and security rules
- Deletion removes critical network security controls and visibility
- Often precedes lateral movement, data exfiltration, or command-and-control establishment
- Can be part of ransomware attacks to disable network segmentation
- May indicate preparation for destructive attacks by removing protective barriers

Azure Firewall Policies Control:
- Network filtering rules (allow/deny traffic between networks)
- Application rules (L7 filtering based on FQDNs, URLs, HTTP/HTTPS)
- NAT rules (destination NAT for inbound connections)
- Threat intelligence-based filtering
- IDPS (Intrusion Detection and Prevention) settings
- TLS inspection configurations

Legitimate Use Cases:
- Migration to new firewall policies or consolidated policy structures
- Decommissioning of test/development environments
- Replacement during security architecture redesign
- Cleanup of unused or deprecated policies

Suspicious Indicators:
- Active production firewall policy deleted
- Deletion by non-network/security administrators
- Multiple firewall policies deleted in sequence
- Deletion during off-hours or outside change windows
- Policy protecting critical workloads suddenly removed
- Deletion followed by suspicious network activity

Next Steps:
1. Immediately verify if the deletion was authorized via change management
2. Identify who performed the deletion and verify their administrative role
3. Determine which Azure Firewalls were using the deleted policy
4. Check if affected firewalls now have no policy (completely unprotected)
5. Review the deleted policy''s rules to understand security impact
6. Verify if a replacement policy was immediately applied
7. Check for suspicious network traffic patterns after deletion
8. Look for other security control modifications in the timeline
9. If unauthorized, restore the policy from backup immediately
10. Investigate the caller''s account for potential compromise
', '["https://attack.mitre.org/techniques/T1562/004/","https://attack.mitre.org/tactics/TA0005/","https://learn.microsoft.com/en-us/azure/firewall/policy-rule-sets","https://learn.microsoft.com/en-us/azure/firewall-manager/policy-overview"]', '(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) && (equalsIgnoreCase("log.operationName", "MICROSOFT.NETWORK/FIREWALLPOLICIES/DELETE") || contains("log.operationName", "Delete Firewall Policy")) && (equalsIgnoreCase("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))', '2026-01-29 16:19:03.534595', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1604, 'Azure Network Watcher Deletion', 1, 3, 3, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', 'Detects the deletion of an Azure Network Watcher instance, which provides critical network monitoring, diagnostic, and visibility tools for Azure virtual networks. Network Watcher enables flow logs, packet capture, connection monitoring, network topology visualization, and NSG diagnostics. Adversaries delete Network Watcher instances to blind network monitoring capabilities, hide lateral movement, evade detection of data exfiltration, and eliminate network forensic capabilities.

Threat Context:
- Network Watcher provides visibility into network traffic patterns and connections
- Deletion disables NSG flow logs, stopping network traffic logging
- Removes packet capture capabilities needed for incident investigation
- Eliminates connection monitoring and network topology visibility
- Often precedes lateral movement or data exfiltration activities
- Can be part of anti-forensics to eliminate evidence of network activity

Network Watcher Capabilities Lost:
- NSG Flow Logs: Track all traffic flowing through network security groups
- Packet Capture: Capture network packets for forensic analysis
- Connection Monitor: Monitor connectivity and latency between resources
- IP Flow Verify: Test NSG rules and diagnose connectivity issues
- Next Hop: Determine routing paths and identify routing problems
- VPN Diagnostics: Troubleshoot VPN gateway and connection issues
- Network Topology: Visualize network architecture and dependencies
- Traffic Analytics: Analyze flow log data for security insights

Legitimate Use Cases:
- Region consolidation or migration to different monitoring solutions
- Cost optimization by disabling in unused regions
- Decommissioning of development/test environments
- Replacement during infrastructure redesign

Suspicious Indicators:
- Network Watcher deleted in production regions
- Deletion by non-network/security administrators
- Multiple Network Watchers deleted across regions
- Deletion during off-hours or outside maintenance windows
- Deletion preceded or followed by suspicious network activity
- NSG flow logs were actively collecting security data
- No replacement monitoring solution configured

Next Steps:
1. Immediately verify if deletion was authorized via change management
2. Identify who performed the deletion and verify their role
3. Determine which regions lost Network Watcher coverage
4. Check if NSG flow logs were active and where they were stored
5. Verify if any security incidents occurred around deletion time
6. Review recent network activity for suspicious patterns
7. Check for other security control modifications in timeline
8. Examine the caller''s account for potential compromise
9. If unauthorized, immediately recreate Network Watcher instances
10. Restore NSG flow logging and verify log retention
11. Review stored flow logs for evidence of malicious activity before deletion
', '["https://attack.mitre.org/techniques/T1562/008/","https://attack.mitre.org/tactics/TA0005/","https://learn.microsoft.com/en-us/azure/network-watcher/network-watcher-overview","https://learn.microsoft.com/en-us/azure/network-watcher/network-watcher-monitoring-overview"]', '(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) && (equalsIgnoreCase("log.operationName", "MICROSOFT.NETWORK/NETWORKWATCHERS/DELETE") || contains("log.operationName", "Delete Network Watcher")) && (equalsIgnoreCase("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))', '2026-01-29 16:19:04.640528', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1605, 'Azure Blob Container Access Level Modification', 3, 2, 1, 'Exfiltration', 'T1537 - Transfer Data to Cloud Account', 'Detects modifications to Azure Blob Storage container access levels, particularly changes that enable anonymous public read access. While anonymous public access is a legitimate feature for sharing data broadly (e.g., CDN content, public downloads), it presents a critical security risk when applied to containers with sensitive data. Adversaries may modify container access levels to exfiltrate data, establish command-and-control infrastructure, or expose confidential information without authentication.

Threat Context:
- Container access level changes can expose sensitive data publicly
- Anonymous access allows data access without authentication
- Commonly used for data exfiltration in breached environments
- Can be leveraged to stage malware or C2 infrastructure
- May indicate unauthorized data disclosure or insider threats

Azure Blob Container Access Levels:
- Private (No public access): Default, requires authentication
- Blob (Anonymous read for blobs): Individual blobs accessible publicly
- Container (Anonymous read for container and blobs): Full container browsing enabled

Legitimate Use Cases:
- Publishing static website content or assets
- Sharing large datasets for public consumption
- CDN origin configuration for public content delivery
- Open-source project artifact hosting
- Public documentation or media distribution

Suspicious Indicators:
- Production storage accounts suddenly made public
- Containers with "confidential", "private", "backup", "logs" in names made public
- Access level changes during off-hours
- Changes by non-storage/DevOps administrators
- Multiple containers made public in succession
- Recently created containers immediately made public
- Containers in finance, HR, or sensitive workload resource groups

Next Steps:
1. Immediately verify if the access change was authorized
2. Identify which container was modified and review its contents
3. Determine if the container contains sensitive or confidential data
4. Check who performed the modification (caller) and their role
5. Review recent access logs for anonymous access attempts
6. If unauthorized and sensitive, immediately revert to private access
7. Audit other containers in the same storage account for similar changes
8. Check for data download activity from public IPs after the change
9. Review the caller''s account for potential compromise
10. Verify if allowBlobPublicAccess is disabled at storage account level
11. Consider implementing Azure Policy to prevent public access
12. Document incident if sensitive data was exposed
', '["https://attack.mitre.org/techniques/T1537/","https://attack.mitre.org/tactics/TA0010/","https://learn.microsoft.com/en-us/azure/storage/blobs/anonymous-read-access-configure","https://learn.microsoft.com/en-us/azure/storage/blobs/anonymous-read-access-prevent"]', '(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) && (equalsIgnoreCase("log.operationName", "MICROSOFT.STORAGE/STORAGEACCOUNTS/BLOBSERVICES/CONTAINERS/WRITE") || contains("log.operationName", "Update container")) && (equalsIgnoreCase("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))', '2026-01-29 16:19:05.781178', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1606, 'ExpressRoute Configuration Changes', 2, 3, 3, 'Discovery', 'T1046 - Network Service Discovery', 'Detects configuration changes to Azure ExpressRoute circuits which could indicate unauthorized network modifications or attempts to bypass security controls. ExpressRoute circuits provide private connectivity between Azure and on-premises infrastructure, making unauthorized changes particularly concerning for network security.

Next Steps:
1. Verify the change was authorized and performed by legitimate administrators
2. Review the specific configuration changes made to the ExpressRoute circuit
3. Check if the change aligns with documented change management procedures
4. Investigate the source IP and user account that performed the change
5. Validate that no unauthorized access to critical network segments occurred
6. Review related network logs for any suspicious activity following the change
', '["https://learn.microsoft.com/en-us/azure/expressroute/monitor-expressroute","https://attack.mitre.org/techniques/T1046/"]', 'oneOf("log.operationName", ["Microsoft.Network/expressRouteCircuits/write", "Microsoft.Network/expressRouteCircuits/delete"]) || (contains("log.resourceId", "/expressRouteCircuits/") && contains("log.operationName", "write"))', '2026-01-29 16:19:07.019042', true, true, 'origin', '["log.resourceId"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1607, 'Azure Service Principal Credentials Added', 3, 3, 2, 'Persistence', 'Account Manipulation: Additional Cloud Credentials', 'Detects when new credentials (certificates or secrets) are added to Azure service principals through Azure AD/Entra ID Audit Logs.

**Security Context:**
Adversaries may add credentials to service principals to maintain persistent access to victim Azure accounts. By hijacking an application with granted permissions through adding rogue secrets or certificates, attackers can access protected data and bypass MFA requirements. This technique is commonly used after initial compromise to establish long-term persistence.

**Detection Logic:**
This rule monitors AuditLogs for successful "Add service principal" operations, which indicate new credentials being added to service principals. The operation captures both certificate and secret additions.

**Investigation Steps:**
1. Identify the actor who added the credentials: Check log.propertiesInitiatedBy for the user or service principal
2. Review the target service principal: Examine log.propertiesTargetResources for the affected service principal name and ID
3. Verify if the action was authorized: Correlate with change management tickets
4. Check service principal permissions: Review what resources this service principal can access
5. Examine recent sign-in activity: Look for unusual authentication patterns using the service principal
6. Review credential type: Determine if a certificate or secret was added via log.propertiesModifiedProperties

**Recommended Actions:**
- If unauthorized, immediately revoke the newly added credentials
- Review and rotate all credentials for the affected service principal
- Audit all resources accessible by the service principal for signs of compromise
- Enable alerts for future credential additions to critical service principals
- Implement conditional access policies and privileged identity management

**MITRE ATT&CK Reference:** T1098.001 - Account Manipulation: Additional Cloud Credentials

**Azure Documentation:**
- AuditLogs table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs
- Service Principal credentials: https://learn.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal
', '["https://attack.mitre.org/techniques/T1098/001/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal"]', 'equalsIgnoreCase("log.category", "AuditLogs") &&
contains("log.operationName", "Add service principal") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-01-29 16:19:08.188259', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1608, 'Azure Function App Security Alert', 3, 3, 2, 'Credential Access', 'T1078 - Valid Accounts', 'Detects security-related errors and exceptions in Azure Function Apps from the FunctionAppLogs table, including authentication failures, authorization denials, execution exceptions, and suspicious patterns. This rule identifies potential security incidents such as credential access attempts, unauthorized function invocations, code injection attempts, or misconfigured security settings.

Threat Context:
- Function Apps often have access to sensitive data and backend systems
- Authentication/authorization errors may indicate credential stuffing or brute force
- Exception details can reveal code injection attempts or exploitation
- Failed executions may indicate tampering with function code or configurations
- Function Apps can be abused for lateral movement or data exfiltration

Azure Functions Log Categories (host.json):
- Host.Results: Function execution results and performance metrics
- Host.Aggregator: Aggregated performance and invocation metrics
- Function: Individual function execution logs and custom logging
- Host.Executor: Function host execution details

What This Rule Detects:
- Exceptions with populated ExceptionDetails field
- Warning level logs (Level >= 3) in Function and Host.Results categories
- Error level logs (Level >= 4) in Function and Host.Results categories
- Authentication and authorization related errors
- Function execution failures that may indicate security issues

Legitimate Scenarios (Reduce False Positives):
- Transient connectivity issues to dependencies
- Expected validation errors from user input
- Planned maintenance or deployment errors
- Development/testing activities in non-production environments

Suspicious Indicators:
- Repeated authentication failures from same source
- Authorization errors accessing sensitive functions
- SQL injection or command injection patterns in exceptions
- Unusual error rates or patterns during off-hours
- Errors from unexpected geographic locations
- Function invocations with suspicious payloads

Investigation and Response Steps:
1. Review the ExceptionDetails and Message fields for error context
2. Identify the specific FunctionName experiencing errors
3. Check the source IP address and correlate with threat intelligence
4. Query for error frequency and patterns:
   FunctionAppLogs | where FunctionName == "<name>" | summarize count() by bin(TimeGenerated, 5m)
5. Review recent code deployments or configuration changes
6. Check Function App authentication settings (Easy Auth, API keys, managed identities)
7. Verify if function keys or connection strings were recently exposed
8. Examine the HostInstanceId to identify affected instances
9. Review Application Insights for correlated traces and dependencies
10. If credential abuse suspected, rotate function keys and connection strings
11. Check for unauthorized code deployments via ARM templates or DevOps pipelines
12. Review network security group rules and private endpoint configurations
13. Enable detailed logging temporarily for deeper investigation if needed
', '["https://learn.microsoft.com/en-us/azure/azure-functions/monitor-functions","https://learn.microsoft.com/en-us/azure/azure-functions/configure-monitoring","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/functionapplogs","https://attack.mitre.org/techniques/T1078/"]', '(contains("log.Category", "Host.Results") || contains("log.Category", "Function")) && (greaterOrEqual("log.Level", 3) || exists("log.ExceptionDetails"))', '2026-01-29 16:19:09.247701', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1609, 'Azure Resource Group Deletion', 2, 3, 3, 'Data Destruction', 'Data Destruction', 'Detects the successful deletion of Azure resource groups through Azure Activity Logs.

**Security Context:**
Resource group deletion is a high-impact administrative action that permanently removes all contained resources. Adversaries may delete resource groups to destroy evidence, disrupt operations, or cause financial impact. This is an irreversible action that can result in significant data loss and service disruption.

**Detection Logic:**
This rule monitors Activity Logs for successful resource group deletion operations. The operation name "MICROSOFT.RESOURCES/SUBSCRIPTIONS/RESOURCEGROUPS/DELETE" specifically identifies when an entire resource group is removed from a subscription.

**Investigation Steps:**
1. Identify the actor: Check log.propertiesCaller for the user or service principal who performed the deletion
2. Review the deleted resource group: Examine log.resourceId and log.resourceGroupName for the affected resources
3. Check authorization: Verify if the deletion was authorized through change management procedures
4. Assess impact: Determine what resources were contained in the deleted resource group
5. Review recent activity: Look for suspicious authentication or privilege escalation events before the deletion
6. Check for bulk operations: Identify if multiple resource groups were deleted in a short timeframe
7. Examine timing: Verify if the deletion occurred during normal business hours

**Recommended Actions:**
- If unauthorized, immediately investigate the actor''s account for compromise
- Review Azure Resource Graph changes history to identify all deleted resources
- Check if resource group locks were bypassed or removed before deletion
- Verify backup availability and initiate recovery procedures if needed
- Enable resource locks on critical resource groups to prevent accidental or malicious deletion
- Implement approval workflows for resource group deletion operations
- Enable Azure Policy to enforce retention policies

**Note:** Azure Resource Groups use soft-delete for certain resource types. Recovery may be possible within the retention period depending on the resources contained.

**MITRE ATT&CK Reference:** T1485 - Data Destruction

**Azure Documentation:**
- AzureActivity table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/azureactivity
- Resource Group deletion: https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/delete-resource-group
', '["https://attack.mitre.org/techniques/T1485/","https://attack.mitre.org/tactics/TA0040/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/azureactivity","https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/delete-resource-group"]', 'equalsIgnoreCase("log.category", "Administrative") &&
equalsIgnoreCase("log.operationName", "MICROSOFT.RESOURCES/SUBSCRIPTIONS/RESOURCEGROUPS/DELETE") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-01-29 16:19:10.352945', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1610, 'Azure Active Directory High Risk Sign-in', 3, 3, 2, 'Valid Accounts', 'Initial Access', 'Identifies high risk Azure Active Directory (AD) sign-ins by leveraging Microsoft''s Identity Protection machine learning and heuristics. Identity Protection categorizes risk into three tiers: low, medium, and high. While Microsoft does not provide specific details about how risk is calculated, each level brings higher confidence that the user or sign-in is compromised. This rule triggers on ''high'' risk level sign-ins, which indicate strong indicators of compromise such as impossible travel, anonymous IP usage, or leaked credentials.', '["https://attack.mitre.org/techniques/T1078/","https://learn.microsoft.com/en-us/entra/id-protection/concept-identity-protection-risks","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/signinlogs"]', 'equalsIgnoreCase("log.category", "SignInLogs") && equalsIgnoreCase("log.properties.RiskLevelDuringSignIn", "high") && equalsIgnoreCase("log.propertiesTokenIssuerType", "AzureAD") && equals("log.resultType", "0")', '2026-01-29 16:19:11.521069', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1611, 'Azure Active Directory PowerShell Sign-in', 1, 2, 3, 'Valid Accounts', 'Initial Access', 'Identifies a sign-in using the Azure Active Directory PowerShell module. PowerShell for Azure Active Directory allows for managing settings from the command line, which is intended for users who are members of an admin role. This activity could indicate legitimate administrative access or potential unauthorized access if the account has been compromised.', '["https://attack.mitre.org/techniques/T1078/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/signinlogs","https://learn.microsoft.com/en-us/entra/identity/monitoring-health/reference-azure-monitor-sign-ins-log-schema"]', 'equalsIgnoreCase("log.category", "SignInLogs") && equalsIgnoreCase("log.propertiesAppDisplayName", "Azure Active Directory PowerShell") && equalsIgnoreCase("log.propertiesTokenIssuerType", "AzureAD") && equals("log.resultType", "0") && equalsIgnoreCase("log.resultSignature", "SUCCESS")', '2026-01-29 16:19:12.599594', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1612, 'Azure Application Credential Modification', 3, 3, 2, 'Defense Evasion', 'T1098.001 - Account Manipulation: Additional Cloud Credentials', 'Detects when a new credential (certificate or secret) is added to an Azure AD application. Applications can use certificates or secret strings to authenticate when requesting tokens. Adversaries may add additional authentication credentials to existing applications to establish persistence, evade defenses, or enable privilege escalation by impersonating legitimate applications.

This technique is commonly used in post-compromise scenarios where attackers:
- Add secrets to high-privilege applications to maintain access
- Create backdoor authentication methods to evade MFA requirements
- Establish persistence mechanisms that survive password resets
- Enable token-based authentication for automated attacks

Next Steps:
1. Verify if the credential modification was authorized and expected
2. Identify who performed the operation (check InitiatedBy field)
3. Review the affected application''s permissions and access scope
4. Check for subsequent suspicious sign-in activity using the application
5. Audit other applications for similar unauthorized modifications
6. If unauthorized, immediately remove the suspicious credentials
7. Review application usage logs for potential abuse
8. Investigate the source IP address and user agent of the modification
', '["https://attack.mitre.org/techniques/T1098/001/","https://attack.mitre.org/tactics/TA0005/","https://learn.microsoft.com/en-us/azure/active-directory/reports-monitoring/concept-audit-logs","https://learn.microsoft.com/en-us/entra/identity/monitoring-health/reference-audit-activities"]', '(equalsIgnoreCase("log.category", "AuditLogs") || contains("log.category", "Audit")) && (contains("log.operationName", "Certificates and secrets management") || equalsIgnoreCase("log.operationName", "Add service principal credentials") || equalsIgnoreCase("log.operationName", "Update application") || equalsIgnoreCase("log.operationName", "Update application - Certificates and secrets management")) && (equalsIgnoreCase("log.resultType", "0"))', '2026-01-29 16:19:13.655604', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1613, 'Possible Consent Grant Attack via Azure-Registered Application', 3, 3, 2, 'Phishing', 'Initial Access', 'Detects when a user grants permissions to an Azure-registered application or when an administrator grants tenant-wide permissions to an application. An adversary may create an Azure-registered application that requests access to data such as contact information, email, or documents. Consent grant attacks are commonly used in phishing campaigns where malicious OAuth applications trick users into granting excessive permissions, enabling data exfiltration or unauthorized access to organizational resources.', '["https://attack.mitre.org/techniques/T1566/","https://learn.microsoft.com/en-us/entra/identity/enterprise-apps/manage-consent-requests","https://learn.microsoft.com/en-us/defender-cloud-apps/investigate-risky-oauth"]', '(equalsIgnoreCase("log.category", "AuditLogs") || contains("log.category", "Audit")) && equalsIgnoreCase("log.operationName", "Consent to application") && equals("log.resultType", "0")', '2026-01-29 16:19:14.772015', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1614, 'Azure Key Vault Excessive Access Detected', 3, 2, 1, 'Collection', 'T1530 - Data from Cloud Storage Object', 'Detects unusual spikes in Azure Key Vault access patterns. Monitors for multiple secret retrieval operations from the same source, which could indicate credential harvesting or data exfiltration attempts.

Next Steps:
1. Investigate the source IP address and verify if it''s a legitimate system or user
2. Review the specific secrets/keys being accessed and their criticality
3. Check for any recent changes to Key Vault access policies
4. Correlate with user authentication logs to identify the account responsible
5. Verify if the access pattern aligns with normal business operations
6. Consider implementing additional access controls or monitoring if suspicious activity is confirmed
', '["https://learn.microsoft.com/en-us/azure/key-vault/general/logging","https://attack.mitre.org/techniques/T1530/"]', 'equalsIgnoreCase("log.category", "AuditEvent") && oneOf("log.operationName", ["SecretGet", "SecretList", "KeyGet"]) && exists("origin.ip")', '2026-01-29 16:19:15.960290', true, true, 'origin', '["origin.ip","log.resourceId"]', '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.category.keyword","operator":"filter_term","value":"AuditEvent"}],"or":null,"within":"now-10m","count":20}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1615, 'Azure External Guest User Invitation', 2, 2, 1, 'Valid Accounts', 'Initial Access', 'Identifies an invitation to an external user in Azure Active Directory (Azure AD / Microsoft Entra ID). Azure AD B2B collaboration allows you to invite people from outside your organization to be guest users in your cloud account and grant them access to resources. Unless there is a business need to provision guest access, it is best practice to avoid creating guest users. Guest users could potentially be overlooked indefinitely, leading to a potential security vulnerability. Adversaries may leverage guest accounts to establish initial access, maintain persistence, or move laterally within the organization.', '["https://attack.mitre.org/techniques/T1078/","https://learn.microsoft.com/en-us/entra/external-id/what-is-b2b","https://learn.microsoft.com/en-us/entra/identity/users/users-restrict-guest-permissions"]', '(equalsIgnoreCase("log.category", "AuditLogs") || contains("log.category", "Audit")) && oneOf("log.operationName", ["Invite external user", "Invite user"]) && equals("log.resultType", "0")', '2026-01-29 16:19:17.137532', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1616, 'Azure Automation Account Created', 2, 3, 2, 'Persistence', 'Valid Accounts: Cloud Accounts', 'Detects the creation of Azure Automation accounts through Azure Activity Logs.

**Security Context:**
Azure Automation accounts provide a platform to automate management tasks and orchestrate actions across Azure and hybrid environments. Adversaries may create Automation accounts to establish persistence by deploying malicious runbooks, webhooks, or scheduled tasks that execute with privileged credentials. This allows them to maintain long-term access and execute commands without direct interactive login.

**Detection Logic:**
This rule monitors Activity Logs for successful creation or update operations on Automation accounts. The operation "MICROSOFT.AUTOMATION/AUTOMATIONACCOUNTS/WRITE" captures both new account creation and modifications to existing accounts.

**Investigation Steps:**
1. Identify the creator: Check log.propertiesCaller for the user or service principal who created the account
2. Review account configuration: Examine the Automation account name and resource group (log.resourceId)
3. Verify authorization: Confirm if the creation was part of legitimate infrastructure deployment
4. Inspect runbooks: Check if any runbooks have been created or imported into the new account
5. Review webhooks: Identify any webhooks configured that could be triggered externally
6. Check schedules: Look for scheduled tasks that may execute malicious code
7. Examine credentials: Verify if credentials or certificates have been added to the account
8. Review RBAC assignments: Check what permissions were granted to the Automation account''s managed identity
9. Correlate with other events: Look for suspicious authentication or privilege escalation before creation

**Recommended Actions:**
- If unauthorized, immediately disable the Automation account
- Review and audit all runbooks, webhooks, and schedules within the account
- Check Run As accounts and credential assets for suspicious additions
- Examine execution history for any jobs that have already run
- Revoke any managed identity or Run As account permissions
- Enable diagnostic logging for the Automation account
- Implement approval workflows for Automation account creation
- Use Azure Policy to restrict Automation account creation to authorized users

**Note:** The WRITE operation captures both creation and updates. To distinguish new accounts, correlate with Resource Graph changes or check for absence of previous activity.

**MITRE ATT&CK Reference:** T1078.004 - Valid Accounts: Cloud Accounts

**Azure Documentation:**
- AzureActivity table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/azureactivity
- Azure Automation security: https://learn.microsoft.com/en-us/azure/automation/automation-security-overview
', '["https://attack.mitre.org/techniques/T1078/004/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/azureactivity","https://learn.microsoft.com/en-us/azure/automation/automation-security-overview"]', 'equalsIgnoreCase("log.category", "Administrative") &&
equalsIgnoreCase("log.operationName", "MICROSOFT.AUTOMATION/AUTOMATIONACCOUNTS/WRITE") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-01-29 16:19:18.236122', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1617, 'Azure Automation Runbook Created or Modified', 2, 3, 2, 'Persistence', 'Command and Scripting Interpreter', 'Detects creation, modification, or publishing of Azure Automation runbooks through Azure Activity Logs.

**Security Context:**
Azure Automation runbooks are scripts (PowerShell, Python, etc.) that execute automated tasks with assigned credentials and permissions. Adversaries may create malicious runbooks or modify existing ones to execute arbitrary code, establish persistence, perform lateral movement, or exfiltrate data. Since runbooks can run on schedules or be triggered via webhooks, they provide a powerful mechanism for maintaining long-term access without interactive login.

**Detection Logic:**
This rule monitors Activity Logs for three critical runbook operations:
- DRAFT/WRITE: Creating or updating a draft runbook
- WRITE: Creating or updating a published runbook
- PUBLISH/ACTION: Publishing a draft runbook to make it executable

All three operations indicate potential malicious activity when performed by unauthorized actors.

**Investigation Steps:**
1. Identify the actor: Check log.propertiesCaller for who created/modified the runbook
2. Review runbook details: Examine log.resourceId for runbook name and Automation account
3. Verify authorization: Confirm if the action was part of legitimate automation deployment
4. Inspect runbook content: Review the actual script code for malicious commands
5. Check runbook type: Identify if it''s PowerShell, Python, or other scripting language
6. Review execution history: Look for any jobs that have already executed this runbook
7. Examine triggers: Check for webhooks or schedules that will execute the runbook
8. Analyze credentials used: Verify what Run As accounts or credential assets the runbook accesses
9. Check network activity: Look for suspicious connections if the runbook has executed
10. Correlate timing: Check if creation/modification follows suspicious authentication events

**Recommended Actions:**
- If unauthorized, immediately unpublish or delete the malicious runbook
- Review runbook code for indicators of compromise (C2 connections, data exfiltration, etc.)
- Disable any associated webhooks or schedules
- Revoke credentials that the runbook may have accessed
- Enable version control for runbooks to track all changes
- Implement code review requirements for runbook modifications
- Use Azure Policy to enforce runbook creation restrictions
- Enable diagnostic logging for all Automation accounts
- Monitor runbook job execution logs for suspicious activity

**Common Malicious Patterns:**
- Runbooks that create new users or modify permissions
- Scripts that exfiltrate data to external storage
- Code that establishes reverse shells or C2 connections
- Runbooks that disable security controls or delete logs

**MITRE ATT&CK Reference:** T1059 - Command and Scripting Interpreter

**Azure Documentation:**
- AzureActivity table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/azureactivity
- Runbook management: https://learn.microsoft.com/en-us/azure/automation/manage-runbooks
', '["https://attack.mitre.org/techniques/T1059/","https://azure.microsoft.com/en-in/blog/azure-automation-runbook-management/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/azureactivity","https://learn.microsoft.com/en-us/azure/automation/manage-runbooks"]', 'equalsIgnoreCase("log.category", "Administrative") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS")) &&
oneOf("log.operationName", [
  "MICROSOFT.AUTOMATION/AUTOMATIONACCOUNTS/RUNBOOKS/DRAFT/WRITE",
  "MICROSOFT.AUTOMATION/AUTOMATIONACCOUNTS/RUNBOOKS/WRITE",
  "MICROSOFT.AUTOMATION/AUTOMATIONACCOUNTS/RUNBOOKS/PUBLISH/ACTION"
])
', '2026-01-29 16:19:19.325349', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1618, 'Azure Automation Webhook Created', 2, 3, 2, 'Persistence', 'Command and Scripting Interpreter', 'Detects the creation of Azure Automation webhooks through Azure Activity Logs.

**Security Context:**
Azure Automation webhooks provide an HTTP endpoint that enables external systems to trigger runbook execution. Each webhook has a unique URL that can execute runbooks with passed parameters, making it a powerful automation mechanism. Adversaries can abuse webhooks to establish persistence by creating backdoor triggers that execute malicious runbooks from external locations, bypassing interactive authentication and logging requirements. This technique is well-documented in offensive Azure toolkits like PowerZure.

**Detection Logic:**
This rule monitors Activity Logs for webhook creation and update operations:
- WEBHOOKS/ACTION: Generate or regenerate webhook URL
- WEBHOOKS/WRITE: Create or update webhook configuration

Both operations indicate potential malicious activity when creating backdoor access to runbook execution.

**Investigation Steps:**
1. Identify the creator: Check log.propertiesCaller for who created the webhook
2. Review webhook details: Examine log.resourceId for webhook name and associated Automation account
3. Verify authorization: Confirm if webhook creation was part of legitimate automation workflow
4. Identify target runbook: Determine which runbook the webhook is configured to trigger
5. Review runbook content: Inspect the linked runbook code for malicious commands
6. Check webhook expiry: Verify the expiration date (long-lived webhooks are suspicious)
7. Examine webhook URL: Determine if the URL has been accessed or shared externally
8. Analyze timing: Check if creation follows suspicious authentication or privilege escalation
9. Review network logs: Look for external HTTP POST requests to the webhook URL
10. Correlate with runbook jobs: Check execution history for suspicious job runs

**Recommended Actions:**
- If unauthorized, immediately disable or delete the webhook
- Rotate the webhook URL if compromise is suspected
- Review and audit the associated runbook for malicious code
- Check webhook execution logs for any triggered jobs
- Examine firewall logs for external connections attempting to trigger the webhook
- Implement IP restrictions on webhook access when possible
- Enable diagnostic logging for Automation accounts
- Use Azure Policy to restrict webhook creation to authorized personnel
- Implement webhook URL lifecycle management with short expiration periods
- Monitor for webhook regeneration attempts

**Common Attack Patterns:**
- Creating webhooks linked to malicious runbooks for command execution
- Using webhooks as covert C2 channels for remote access
- Establishing persistence through externally-triggered automation
- Bypassing MFA by triggering privileged runbooks via webhooks
- Creating long-lived (multi-year) webhook URLs for sustained access

**Known Threat Tools:**
- PowerZure: Create-Backdoor function specifically targets webhook creation
- SpecterOps documented webhook abuse for Azure persistence

**MITRE ATT&CK Reference:** T1059 - Command and Scripting Interpreter

**Azure Documentation:**
- AzureActivity table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/azureactivity
- Webhook documentation: https://learn.microsoft.com/en-us/azure/automation/automation-webhooks
', '["https://attack.mitre.org/techniques/T1059/","https://powerzure.readthedocs.io/en/latest/Functions/operational.html#create-backdoor","https://github.com/hausec/PowerZure","https://posts.specterops.io/attacking-azure-azure-ad-and-introducing-powerzure-ca70b330511a","https://www.ciraltos.com/webhooks-and-azure-automation-runbooks/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/azureactivity","https://learn.microsoft.com/en-us/azure/automation/automation-webhooks"]', 'equalsIgnoreCase("log.category", "Administrative") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS")) &&
oneOf("log.operationName", [
  "MICROSOFT.AUTOMATION/AUTOMATIONACCOUNTS/WEBHOOKS/ACTION",
  "MICROSOFT.AUTOMATION/AUTOMATIONACCOUNTS/WEBHOOKS/WRITE"
])
', '2026-01-29 16:19:20.472132', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1619, 'Network Security Group Modifications', 3, 3, 2, 'Defense Evasion', 'T1562.007 - Impair Defenses: Disable or Modify Cloud Firewall', 'Detects modifications to Azure Network Security Groups which could indicate attempts to bypass network security controls or create backdoor access. Network Security Groups control traffic flow to Azure resources and unauthorized changes could expose critical infrastructure.

Next Steps:
1. Verify the legitimacy of the NSG modification with the responsible administrator
2. Review the specific security rules that were added, modified, or deleted
3. Check if the modification aligns with approved change management processes
4. Investigate the source IP and user account that performed the change
5. Review other recent Azure activity from the same user or IP address
6. Validate that the NSG changes don''t expose critical resources to unauthorized access
7. Check if the changes affect production workloads or sensitive environments
', '["https://learn.microsoft.com/en-us/azure/virtual-network/virtual-network-nsg-manage-log","https://attack.mitre.org/techniques/T1562/007/"]', 'contains("log.operationName", "Microsoft.Network/networkSecurityGroups") && contains("log.operationName", ["/write", "/delete", "/securityRules/write"]) && equalsIgnoreCase("log.category", "Administrative") && equalsIgnoreCase("log.resultType", "Success")', '2026-01-29 16:19:21.527856', true, true, 'origin', '["origin.ip","log.resourceId"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1620, 'Azure Conditional Access Policy Modified', 3, 3, 2, 'Persistence', 'Account Manipulation: Additional Cloud Credentials', 'Detects modifications to Azure Conditional Access policies through Azure AD/Entra ID Audit Logs.

**Security Context:**
Azure Conditional Access policies are critical security controls that enforce access requirements such as multi-factor authentication (MFA), device compliance, location restrictions, and application-specific rules. Adversaries who gain sufficient privileges may modify these policies to weaken security controls, create exceptions for their compromised accounts, exclude malicious users from MFA requirements, or establish persistent access by bypassing security mechanisms.

**Detection Logic:**
This rule monitors AuditLogs for successful "Update policy" operations in Azure AD/Entra ID. These operations indicate changes to existing Conditional Access policy configurations, including modifications to:
- User and group inclusions/exclusions
- Application scope changes
- Location-based access rules
- Grant controls (MFA, device compliance, etc.)
- Session controls
- Policy state (enabled/disabled/report-only)

**Investigation Steps:**
1. Identify the modifier: Check log.propertiesInitiatedBy for who modified the policy
2. Review policy details: Examine log.propertiesTargetResources for the affected policy name and ID
3. Verify authorization: Confirm if the modification was part of approved security changes
4. Compare policy versions: Review log.propertiesModifiedProperties to identify specific changes made
5. Check policy before/after states: Look for weakening of security controls (MFA removed, users excluded, etc.)
6. Analyze timing: Determine if modification follows suspicious authentication or privilege escalation
7. Review affected users/apps: Identify which users, groups, or applications are impacted by the change
8. Check for exclusions: Look for specific users or groups being excluded from security requirements
9. Examine policy state: Verify if policy was disabled or moved to report-only mode
10. Correlate with sign-ins: Check for unusual sign-in activity after policy modification

**Recommended Actions:**
- If unauthorized, immediately revert the policy to its previous secure state
- Review all Conditional Access policies for unauthorized modifications
- Enable change tracking for Conditional Access policies
- Implement privileged access management for policy modification rights
- Use PIM (Privileged Identity Management) with approval for Conditional Access Administrator role
- Enable alerts for all Conditional Access policy changes
- Maintain documented baseline configurations for all policies
- Implement policy-as-code for version control and change management
- Review and audit accounts with Conditional Access Administrator role
- Consider emergency access accounts and their exclusions

**Common Malicious Modifications:**
- Excluding attacker-controlled accounts from MFA requirements
- Disabling location-based restrictions
- Removing device compliance requirements
- Adding exceptions for legacy authentication protocols
- Changing policy state from "enabled" to "report-only" or "disabled"
- Expanding policy exclusions to include compromised accounts

**MITRE ATT&CK Reference:** T1098.001 - Account Manipulation: Additional Cloud Credentials

**Azure Documentation:**
- AuditLogs table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs
- Conditional Access: https://learn.microsoft.com/en-us/entra/identity/conditional-access/overview
', '["https://attack.mitre.org/techniques/T1098/001/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/entra/identity/conditional-access/overview","https://learn.microsoft.com/en-us/entra/identity/conditional-access/howto-conditional-access-policy-all-users-mfa"]', 'equalsIgnoreCase("log.category", "AuditLogs") &&
equalsIgnoreCase("log.operationName", "Update policy") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-01-29 16:19:22.706259', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1621, 'Azure Global Administrator Role Addition to PIM User', 3, 3, 3, 'Persistence', 'Account Manipulation: Additional Cloud Credentials', 'Detects when users are granted Global Administrator (Company Administrator) role assignments through Azure AD/Entra ID Privileged Identity Management (PIM).

**Security Context:**
The Global Administrator role is the most powerful administrative role in Azure AD/Entra ID, granting complete control over all aspects of the directory and services that use Azure AD identities. PIM enables just-in-time privileged access through eligible (requires activation) or time-bound assignments. Adversaries who gain sufficient privileges may add themselves or other compromised accounts to this role to establish persistence and maintain full administrative control over the tenant.

**Detection Logic:**
This rule monitors AuditLogs for successful PIM role assignments specifically for the Global Administrator role. It detects both:
- **Eligible assignments (permanent)**: User can activate the role when needed
- **Active assignments (time-bound)**: Role is directly active for a specified duration

The rule identifies these assignments through the operation names and filters for the Global Administrator role specifically.

**Investigation Steps:**
1. Identify the assignor: Check log.propertiesInitiatedBy for who made the role assignment
2. Identify the assignee: Examine log.propertiesTargetResources for the user receiving the role
3. Verify authorization: Confirm if this assignment was part of approved privileged access request
4. Check assignment type: Determine if it''s eligible (requires activation) or time-bound (direct)
5. Review duration: For time-bound assignments, check the duration of the assignment
6. Analyze timing: Determine if assignment follows suspicious authentication or compromise indicators
7. Review justification: Check if a business justification was provided in log.propertiesAdditionalDetails
8. Check user history: Review the assignee''s account for recent suspicious activity
9. Examine recent actions: Look for privileged operations performed immediately after assignment
10. Correlate with sign-ins: Check for unusual authentication patterns before/after assignment

**Recommended Actions:**
- If unauthorized, immediately revoke the Global Administrator role assignment
- Review all recent PIM role assignments for anomalies
- Enable PIM approval workflows for Global Administrator role assignments
- Implement maximum assignment duration limits for time-bound assignments
- Require MFA and justification for all Global Administrator activations
- Enable PIM alerts for high-privilege role assignments
- Audit accounts with Privileged Role Administrator permissions
- Review and limit the number of permanent Global Administrator assignments
- Enable Azure AD Identity Protection to detect compromised credentials
- Implement break-glass emergency access accounts following best practices

**PIM Assignment Types:**
- **Eligible (permanent)**: User must activate the role when needed, typically with MFA and justification
- **Active (time-bound)**: Role is directly assigned for a limited duration without activation required
- Both types should be monitored as adversaries may use either for persistence

**Common Attack Patterns:**
- Compromised Privileged Role Administrator adding backdoor accounts
- Insider threat establishing persistent administrative access
- Privilege escalation from lower-privilege administrative roles
- Adding service principals or managed identities to Global Administrator role
- Creating long-duration time-bound assignments for sustained access

**MITRE ATT&CK Reference:** T1098.001 - Account Manipulation: Additional Cloud Credentials

**Azure Documentation:**
- AuditLogs table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs
- PIM for Azure AD roles: https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-configure
', '["https://attack.mitre.org/techniques/T1098/001/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-configure","https://learn.microsoft.com/en-us/entra/identity/role-based-access-control/permissions-reference#global-administrator"]', 'equalsIgnoreCase("log.category", "AuditLogs") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS")) &&
(contains("log.operationName", "Add eligible member to role") || contains("log.operationName", "Add member to role")) &&
(contains("log.properties.targetResources.displayName", "Global Administrator") || contains("log.properties.targetResources.displayName", "Company Administrator"))
', '2026-01-29 16:19:23.832434', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1622, 'Azure Privileged Identity Management Role Settings Modified', 3, 3, 2, 'Persistence', 'Account Manipulation: Additional Cloud Credentials', 'Detects modifications to Azure AD/Entra ID Privileged Identity Management (PIM) role settings through Audit Logs.

**Security Context:**
PIM role settings define critical security controls for privileged role assignments, including:
- Approval requirements for role activation
- Multi-factor authentication (MFA) enforcement
- Maximum activation duration
- Justification requirements
- Notification settings
- Eligibility and assignment duration limits

Adversaries with sufficient privileges may modify these settings to weaken security controls, remove approval requirements, extend activation durations, or disable MFA requirements, making it easier to abuse privileged roles for persistent access.

**Detection Logic:**
This rule monitors AuditLogs for successful "Update role setting in PIM" operations, which capture modifications to role configuration policies that govern how privileged roles are activated and managed.

**Investigation Steps:**
1. Identify the modifier: Check log.propertiesInitiatedBy for who changed the role settings
2. Identify affected role: Examine log.propertiesTargetResources for which role''s settings were modified
3. Review specific changes: Analyze log.propertiesModifiedProperties to identify what settings changed (before/after values)
4. Verify authorization: Confirm if the modification was part of approved policy changes
5. Check for security weakening: Look for:
   - Removal or reduction of approval requirements
   - Disabling MFA for activation
   - Extending maximum activation durations
   - Removing justification requirements
   - Disabling notifications to administrators
6. Analyze timing: Determine if modification follows suspicious authentication or privilege escalation
7. Review role sensitivity: Assess the criticality of the affected role (Global Admin, Privileged Role Admin, etc.)
8. Check for pattern: Look for multiple role setting modifications in short timeframe
9. Examine subsequent activations: Monitor for role activations after settings were weakened
10. Correlate with user behavior: Check if modifier has history of legitimate administrative actions

**Recommended Actions:**
- If unauthorized, immediately revert role settings to secure baseline
- Review all PIM role settings for unauthorized modifications
- Enable change notifications for PIM role setting updates
- Implement approval workflows for modifying PIM role settings
- Use PIM for Privileged Role Administrator role itself
- Maintain documented baseline configurations for all PIM role settings
- Enable alerts for PIM configuration changes
- Audit accounts with permissions to modify PIM settings
- Implement policy-as-code for PIM role configurations
- Review and document approved security baselines for each privileged role

**Common Malicious Modifications:**
- Removing MFA requirement for role activation
- Disabling approval workflows for high-privilege roles
- Extending maximum activation duration from hours to days
- Removing justification requirements
- Disabling email notifications to security teams
- Extending maximum eligible assignment duration
- Removing requirement for assignment end date
- Allowing permanent assignments without expiration

**PIM Role Settings Categories:**
- **Activation**: MFA, approval, justification, duration
- **Assignment**: Maximum duration, expiration requirements, permanent assignments
- **Notification**: Alerts to admins, assignees, and approvers

**MITRE ATT&CK Reference:** T1098.001 - Account Manipulation: Additional Cloud Credentials

**Azure Documentation:**
- AuditLogs table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs
- PIM role settings: https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-how-to-change-default-settings
', '["https://attack.mitre.org/techniques/T1098/001/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-how-to-change-default-settings","https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-configure"]', 'equalsIgnoreCase("log.category", "AuditLogs") &&
contains("log.operationName", "Update role setting") &&
(equals("log.resultType", "0") || equalsIgnoreCase("lactionResult", "SUCCESS"))
', '2026-01-29 16:19:24.831514', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1623, 'Azure Service Principal Addition', 2, 3, 2, 'Persistence', 'T1136.003 - Create Account: Cloud Account', 'Detects when a new service principal is created in Azure Active Directory (Entra ID). A service principal is an identity created for use with applications, hosted services, and automated tools to access Azure resources. While service principals are legitimate and necessary for automation, adversaries may create rogue service principals to establish persistent access, escalate privileges, or move laterally within an Azure environment.

Threat Context:
- Service principals can be granted powerful permissions across Azure subscriptions
- Unlike user accounts, service principals often lack MFA protection
- Credentials (secrets/certificates) can persist for years without rotation
- Service principals can be used for automated attacks without triggering user behavior analytics

Legitimate Use Cases:
- DevOps pipelines and CI/CD automation
- Application authentication and service-to-service communication
- Terraform/Bicep/ARM template deployments
- Monitoring and management tools

Suspicious Indicators:
- Creation by non-administrative users
- Creation outside business hours
- Service principal granted high privileges immediately after creation
- Multiple service principals created in quick succession
- Creation from unusual IP addresses or locations

Next Steps:
1. Verify if the service principal creation was authorized and documented
2. Identify who created it (check InitiatedBy field) and verify their role
3. Review the permissions/roles assigned to the new service principal
4. Check if credentials (secrets/certificates) were added immediately after
5. Examine the source IP address and location of the creation event
6. Verify if the service principal has been used for authentication
7. Cross-reference with change management tickets or DevOps records
8. If unauthorized, immediately disable the service principal and rotate credentials
', '["https://attack.mitre.org/techniques/T1136/003/","https://attack.mitre.org/tactics/TA0003/","https://learn.microsoft.com/en-us/entra/identity/monitoring-health/reference-audit-activities","https://learn.microsoft.com/en-us/entra/identity-platform/app-objects-and-service-principals"]', '(equalsIgnoreCase("log.category", "AuditLogs") || contains("log.category", "Audit")) && (equalsIgnoreCase("log.operationName", "Add service principal") || contains("log.operationName", "Create service principal")) && (equalsIgnoreCase("log.resultType", "0") || equalsIgnoreCase("actionResult", "success"))', '2026-01-29 16:19:25.821239', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1624, 'User Added as Owner for Azure Application', 3, 3, 2, 'Persistence', 'Account Manipulation: Additional Cloud Credentials', 'Detects when a user or service principal is added as an owner for an Azure AD/Entra ID application registration through Audit Logs.

**Security Context:**
Azure AD application registrations represent identities used for authentication to Azure and other Microsoft services. Application owners have full administrative control over the application, including the ability to:
- Add or rotate credentials (certificates and secrets)
- Modify API permissions
- Add additional owners
- Configure redirect URIs and authentication settings
- Delete the application

Adversaries may add themselves or compromised accounts as application owners to establish persistence, as this grants them the ability to generate new credentials for the application''s service principal, enabling long-term access even if the original compromise vector is remediated.

**Detection Logic:**
This rule monitors AuditLogs for successful "Add owner to application" operations, which capture when ownership permissions are granted on Azure AD application registrations.

**Investigation Steps:**
1. Identify the adder: Check log.propertiesInitiatedBy for who added the owner
2. Identify new owner: Examine log.propertiesTargetResources for the user/principal being added as owner
3. Identify application: Review log.propertiesTargetResources for the affected application details
4. Verify authorization: Confirm if the owner addition was part of legitimate administrative action
5. Review application sensitivity: Check what API permissions and resources the application has access to
6. Check application credentials: Look for credential additions after owner was added
7. Analyze timing: Determine if owner addition follows suspicious authentication or privilege escalation
8. Review new owner privileges: Assess if the new owner already has elevated permissions
9. Check for pattern: Look for multiple ownership additions across different applications
10. Examine subsequent actions: Monitor for credential generation, permission changes, or authentication using the application

**Recommended Actions:**
- If unauthorized, immediately remove the malicious owner from the application
- Review and rotate all credentials (secrets and certificates) for the affected application
- Audit all API permissions granted to the application
- Review authentication activity using the application''s service principal
- Check for any configuration changes made after the owner was added
- Enable application owner change alerts for critical applications
- Implement approval workflows for adding owners to sensitive applications
- Audit accounts with permissions to modify application ownership
- Review and document expected owners for all applications
- Consider using managed identities instead of application registrations where possible

**Application Owner Capabilities:**
Application owners can perform all configuration actions on the application, including:
- **Credential Management**: Add/remove certificates and client secrets
- **Permission Management**: Request and consent to API permissions
- **Authentication Configuration**: Modify redirect URIs, token settings
- **Ownership Management**: Add/remove other owners
- **Application Deletion**: Remove the application entirely

**Common Attack Patterns:**
- Adding backdoor account as owner after initial compromise
- Privilege escalation by compromising application with broad permissions
- Establishing persistence through credential rotation capabilities
- Insider threats adding personal accounts as owners
- Service principal abuse by adding owners to high-privilege applications
- Lateral movement by gaining control of applications with cross-tenant access

**Related Detections:**
- Service principal credential additions
- Application permission changes
- Service principal authentication anomalies
- Cross-tenant application access

**MITRE ATT&CK Reference:** T1098.001 - Account Manipulation: Additional Cloud Credentials

**Azure Documentation:**
- AuditLogs table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs
- Application owners: https://learn.microsoft.com/en-us/entra/identity/role-based-access-control/permissions-reference#application-administrator
', '["https://attack.mitre.org/techniques/T1098/001/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/entra/identity/role-based-access-control/permissions-reference#application-administrator","https://learn.microsoft.com/en-us/entra/identity-platform/howto-create-service-principal-portal"]', 'equalsIgnoreCase("log.category", "AuditLogs") &&
equalsIgnoreCase("log.operationName", "Add owner to application") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-01-29 16:19:26.937445', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1625, 'MFA Disabled for Privileged Azure AD User', 3, 3, 1, 'Credential Access, Defense Evasion, Persistence', 'T1556 - Modify Authentication Process', 'Detects when Multi-Factor Authentication (MFA) is disabled for privileged users in Azure AD. This could indicate an attempt to weaken security controls for unauthorized access.

Next Steps:
1. Verify if the MFA disable action was authorized and legitimate
2. Check who initiated the change and from which IP address
3. Review the user''s recent login activity and permissions
4. Ensure the user account has not been compromised
5. Re-enable MFA if the change was unauthorized
6. Consider implementing conditional access policies to prevent unauthorized MFA changes
', '["https://learn.microsoft.com/en-us/entra/identity/authentication/howto-mfa-reporting","https://attack.mitre.org/techniques/T1556/"]', 'oneOf("log.operationName", ["Disable Strong Authentication", "Update user"]) && equalsIgnoreCase("log.service", "Authentication Methods") && contains("target.user", ["admin", "globaladmin"])', '2026-01-29 16:19:27.953421', true, true, 'origin', '["target.user"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1626, 'User Added as Owner for Azure Service Principal', 3, 3, 2, 'Persistence', 'Account Manipulation: Additional Cloud Credentials', 'Detects when a user or service principal is added as an owner for an Azure AD/Entra ID service principal (Enterprise Application) through Audit Logs.

**Security Context:**
Service principals are the local representation of application objects in a specific Azure AD tenant, also known as Enterprise Applications. They define what an application can actually do in that specific tenant, including:
- API permissions and consented scopes
- Role assignments to Azure resources
- Authentication and access policies
- Conditional Access policy applicability

Service principal owners have significant control over the identity, including the ability to:
- Manage credentials for the service principal
- Modify application assignments and permissions
- Configure authentication settings
- Add or remove additional owners
- Delete the service principal

Adversaries may add themselves or compromised accounts as service principal owners to establish persistence, as this grants the ability to authenticate as the service principal and access all resources it has permissions to.

**Key Difference: Application vs Service Principal:**
- **Application Registration**: The global definition of the app across all tenants
- **Service Principal (Enterprise App)**: The local instance/identity in a specific tenant
- Application owners control the app definition; Service Principal owners control the tenant-specific instance

**Detection Logic:**
This rule monitors AuditLogs for successful "Add owner to service principal" operations, which capture when ownership permissions are granted on service principal objects.

**Investigation Steps:**
1. Identify the adder: Check log.propertiesInitiatedBy for who added the owner
2. Identify new owner: Examine log.propertiesTargetResources for the user/principal being added as owner
3. Identify service principal: Review log.propertiesTargetResources for the affected service principal details
4. Verify authorization: Confirm if the owner addition was part of legitimate administrative action
5. Review service principal permissions: Check what API permissions and Azure role assignments the SP has
6. Check for credential additions: Look for certificate or secret additions after owner was added
7. Analyze timing: Determine if owner addition follows suspicious authentication or privilege escalation
8. Review new owner privileges: Assess if the new owner already has elevated permissions elsewhere
9. Check authentication history: Monitor for authentication attempts using the service principal
10. Examine subsequent actions: Look for permission changes, role assignments, or configuration modifications

**Recommended Actions:**
- If unauthorized, immediately remove the malicious owner from the service principal
- Review and rotate all credentials associated with the service principal
- Audit all API permissions and Azure role assignments for the service principal
- Review authentication activity using the service principal''s credentials
- Check for any configuration changes made after the owner was added
- Enable service principal owner change alerts for critical applications
- Implement approval workflows for adding owners to sensitive service principals
- Audit accounts with permissions to modify service principal ownership
- Review and document expected owners for all service principals
- Consider implementing Managed Identities where service principals are currently used

**Service Principal Owner Capabilities:**
Service principal owners can perform critical configuration actions:
- **Credential Management**: Add/remove certificates and client secrets
- **Permission Management**: View and manage API permissions (consent may require admin)
- **Authentication Configuration**: Modify authentication policies and settings
- **Ownership Management**: Add/remove other owners
- **Assignment Management**: Manage user/group assignments to the application
- **Service Principal Deletion**: Remove the service principal from the tenant

**Common Attack Patterns:**
- Adding backdoor account as owner after compromising admin credentials
- Privilege escalation by controlling high-permission service principals
- Establishing persistence through credential generation capabilities
- Lateral movement by gaining control of service principals with cross-resource access
- Insider threats adding personal accounts as owners
- Compromising service principals with privileged Azure RBAC roles
- Abusing service principals with delegated API permissions

**Related Detections:**
- Service principal credential additions (certificates/secrets)
- Application owner additions (companion detection)
- Service principal permission changes
- Service principal authentication anomalies
- Azure RBAC role assignments to service principals

**MITRE ATT&CK Reference:** T1098.001 - Account Manipulation: Additional Cloud Credentials

**Azure Documentation:**
- AuditLogs table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs
- Service principals: https://learn.microsoft.com/en-us/entra/identity-platform/app-objects-and-service-principals
', '["https://attack.mitre.org/techniques/T1098/001/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/entra/identity-platform/app-objects-and-service-principals","https://learn.microsoft.com/en-us/entra/identity/enterprise-apps/overview-assign-app-owners"]', 'equalsIgnoreCase("log.category", "AuditLogs") &&
equalsIgnoreCase("log.operationName", "Add owner to service principal") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-01-29 16:19:29.489211', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1627, 'Multi-Factor Authentication Disabled for an Azure User', 3, 3, 2, 'Persistence', 'Modify Authentication Process', 'Detects when multi-factor authentication (MFA) is disabled for an Azure AD/Entra ID user account through Audit Logs.

**Security Context:**
Multi-factor authentication is a critical security control that requires users to provide additional verification beyond just a password. Disabling MFA for user accounts significantly weakens authentication security and is a common technique used by adversaries to maintain persistent access. Once MFA is disabled, attackers can authenticate using only compromised credentials without triggering additional verification steps, making detection more difficult.

**Detection Logic:**
This rule monitors AuditLogs for successful "Disable Strong Authentication" operations, which represent the per-user MFA setting being turned off in Azure AD/Entra ID. This operation is distinct from Conditional Access MFA policies and represents the legacy per-user MFA enforcement method.

**Investigation Steps:**
1. Identify the disabler: Check log.propertiesInitiatedBy for who disabled MFA
2. Identify affected user: Examine log.propertiesTargetResources for the user whose MFA was disabled
3. Verify authorization: Confirm if the MFA disabling was part of legitimate administrative action
4. Review user privilege: Determine if the affected user has elevated permissions (admins, privileged roles)
5. Check timing: Analyze if MFA was disabled after suspicious authentication events
6. Review authentication history: Look for failed authentication attempts before MFA disabling
7. Check for compromise indicators: Search for unusual sign-in patterns, impossible travel, or risky sign-ins
8. Examine subsequent logins: Monitor for authentication activity immediately after MFA disabling
9. Review MFA methods: Check what MFA methods the user had registered before disabling
10. Correlate with other events: Look for privilege escalation or data access after MFA disabling

**Recommended Actions:**
- If unauthorized, immediately re-enable MFA for the affected user
- Force password reset for the affected account
- Review all authentication activity for the affected user
- Check for compromised credentials using Azure AD Identity Protection
- Revoke all active sessions for the affected user
- Enable Conditional Access policies instead of per-user MFA for better control
- Implement PIM approval workflows for modifying MFA settings
- Enable alerts for MFA changes on privileged accounts
- Audit accounts with permissions to modify user authentication settings
- Review and restrict who can disable MFA (typically requires User Administrator or higher)

**Modern MFA Management:**
- **Per-user MFA (legacy)**: This detection targets the legacy per-user MFA setting
- **Conditional Access**: Modern approach using policies instead of per-user settings
- **Authentication Methods Policy**: Newer method for managing FIDO2, passwordless, etc.

Organizations should migrate from per-user MFA to Conditional Access policies for more granular control.

**Common Attack Patterns:**
- Disabling MFA after compromising an administrator account
- Removing MFA from privileged accounts for easier persistent access
- Disabling MFA before credential harvesting or lateral movement
- Insider threats removing MFA from their own accounts
- Disabling MFA on service accounts to enable automated authentication attacks

**Related Detections:**
- MFA method removal/changes
- Conditional Access policy modifications
- Authentication methods policy changes
- Privileged role assignments without MFA

**MITRE ATT&CK Reference:** T1556 - Modify Authentication Process

**Azure Documentation:**
- AuditLogs table: https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs
- Per-user MFA: https://learn.microsoft.com/en-us/entra/identity/authentication/howto-mfa-userstates
', '["https://attack.mitre.org/techniques/T1556/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/entra/identity/authentication/howto-mfa-userstates","https://learn.microsoft.com/en-us/entra/identity/authentication/concept-mfa-licensing"]', 'equalsIgnoreCase("log.category", "AuditLogs") &&
equalsIgnoreCase("log.operationName", "Disable Strong Authentication") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-01-29 16:19:30.685602', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1628, 'Azure AD Privilege Escalation Attempt Detected', 3, 3, 1, 'Defense Evasion, Persistence, Privilege Escalation, Initial Access', 'T1078 - Valid Accounts', 'Detects attempts to escalate privileges in Azure AD through role assignments. Monitors for the Microsoft.Authorization/roleAssignments/write operation which indicates a user or service principal is being granted additional permissions.

Next Steps:
1. Verify the legitimacy of the role assignment by checking with the requesting user or administrator
2. Review the specific role being assigned and ensure it follows the principle of least privilege
3. Check if this is part of a scheduled maintenance or approved change request
4. Investigate the source IP address and user context for any suspicious patterns
5. Review Azure AD audit logs for any other suspicious activities from the same user or IP
6. If unauthorized, immediately revoke the role assignment and investigate potential account compromise
', '["https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-alert","https://attack.mitre.org/techniques/T1078/"]', 'equalsIgnoreCase("log.operationName", "Microsoft.Authorization/roleAssignments/write") && equalsIgnoreCase("log.category", "Administrative")', '2026-01-29 16:19:32.384127', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1629, 'Resource Group Mass Modifications', 2, 3, 3, 'Impact', 'T1496 - Resource Hijacking', 'Detects mass modifications to Azure resource groups which could indicate unauthorized infrastructure changes or resource hijacking attempts. This rule triggers when multiple resource group write or delete operations are performed by the same user from the same IP address within a 15-minute window.

Next Steps:
1. Investigate the user account (log.aadObjectId) performing the modifications
2. Review the specific resource groups being modified
3. Check if the operations align with scheduled maintenance or legitimate business activities
4. Verify the source IP address and geolocation for suspicious activity
5. Review Azure Activity Logs for the full scope of changes made
6. Check for any privilege escalation or unauthorized access to the account
7. If malicious, immediately revoke access and assess impact on affected resources
', '["https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log-schema","https://attack.mitre.org/techniques/T1496/"]', 'contains("log.operationName", "Microsoft.Resources/subscriptions/resourceGroups") && contains("log.operationName", ["/write", "/delete"]) && equalsIgnoreCase("log.category", "Administrative") && equalsIgnoreCase("log.resultSignature", "Succeeded")', '2026-01-29 16:19:34.088773', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1630, 'Azure Service Principal Multiple Failed Authentications', 3, 3, 2, 'Initial Access', 'T1078.004 - Valid Accounts: Cloud Accounts', 'Detects potential brute force or credential guessing attacks against Azure service principals by identifying multiple failed authentication attempts from the same source IP within a 2-hour window.

**Security Context:**
Service principals are non-interactive identities used by applications and services to authenticate to Azure AD. Multiple failed authentication attempts may indicate:
- Brute force attacks trying to guess client secrets
- Replay attacks with expired or invalid credentials
- Misconfigured applications repeatedly failing to authenticate
- Compromised credentials being tested

**Detection Logic:**
- **Trigger**: A failed service principal authentication (resultType != "0")
- **Correlation**: Looks back 2 hours for 5+ failed authentications from the same service principal and source IP

This correlation approach reduces false positives by only alerting when there''s a pattern of multiple failures, not just a single authentication error.

**Investigation Steps:**
1. Identify the service principal: Check log.propertiesServicePrincipalId and log.propertiesAppDisplayName
2. Review failure reasons: Examine log.propertiesResultDescription for specific error codes
3. Analyze source location: Check origin.ip and log.propertiesLocation for geographic anomalies
4. Review failure pattern: Check timestamps and frequency of authentication attempts
5. Verify credential type: Determine if certificate or secret authentication was attempted
6. Check service principal ownership: Review who owns the application/service principal
7. Examine recent credential changes: Look for recent secret/certificate additions in AuditLogs
8. Review service principal permissions: Check API permissions and Azure RBAC roles
9. Look for successful authentications: Check if any attempts succeeded after the failures
10. Correlate with other events: Search for related suspicious activities (permission changes, resource access)

**Recommended Actions:**
- If legitimate: Review application configuration and fix authentication issues
- If suspicious: Immediately rotate all credentials (secrets and certificates) for the service principal
- Review and audit all API permissions and Azure RBAC role assignments
- Check for unauthorized credential additions in AuditLogs
- Implement IP restrictions or Conditional Access for service principals where possible
- Enable Azure AD Identity Protection for service principal risk detection
- Consider implementing Managed Identities instead of service principals where possible
- Review authentication logs for successful breaches after failed attempts

**Common Failure Result Types:**
- **50126**: Invalid credentials (wrong password/secret)
- **50053**: Account locked due to too many sign-in attempts
- **50057**: Account disabled
- **700016**: Application not found in directory
- **7000215**: Invalid client secret provided

**MITRE ATT&CK Reference:** T1078.004 - Valid Accounts: Cloud Accounts
', '["https://attack.mitre.org/techniques/T1078/004/","https://www.cloud-architekt.net/auditing-of-msi-and-service-principals/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/signinlogs"]', 'equalsIgnoreCase("log.category", "SignInLogs") &&
exists("log.propertiesServicePrincipalId") &&
!equals("log.resultType", "0")
', '2026-01-29 16:19:35.088874', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"log.propertiesServicePrincipalId.keyword","operator":"filter_term","value":"{{.log.propertiesServicePrincipalId}}"},{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.category.keyword","operator":"filter_term","value":"SignInLogs"}],"or":null,"within":"now-2h","count":5}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1631, 'SQL Database Firewall Rule Modifications', 3, 2, 1, 'Lateral Movement', 'Remote Services', 'Detects modifications to Azure SQL Database firewall rules which could allow unauthorized access to sensitive data. This includes both creation and deletion of firewall rules that control network access to SQL databases.

Next Steps:
1. Verify the legitimacy of the firewall rule modification
2. Check if the change was authorized and documented
3. Review the source IP and user making the modification
4. Assess if the new firewall rule creates security risks
5. Monitor for subsequent database access attempts from newly allowed IPs
6. Review Azure Activity Logs for related database activities
', '["https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log-schema","https://attack.mitre.org/techniques/T1021/"]', 'contains("log.operationName", "Microsoft.Sql/servers") && (contains("log.operationName", "/firewallRules/write") || contains("log.operationName", "/firewallRules/delete")) && equalsIgnoreCase("log.category", "Administrative") && equalsIgnoreCase("actionResult", "accepted")', '2026-01-29 16:19:36.077882', true, true, 'origin', '["origin.ip","log.resourceId"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1632, 'Azure Storage Account Public Access Enabled', 3, 2, 1, 'Collection', 'T1530 - Data from Cloud Storage Object', 'Detects when public (anonymous) access is enabled on Azure Storage Accounts or Blob containers, creating a critical security risk by allowing unauthenticated access to potentially sensitive data.

**Security Context:**
Azure Storage Accounts can be configured to allow public access at two levels:
1. **Account Level**: `allowBlobPublicAccess` property enables/disables public access for the entire storage account
2. **Container Level**: Individual blob containers can be set to "Blob" or "Container" public access levels

Public access levels:
- **None (Private)**: No anonymous access - requires authentication
- **Blob**: Anonymous read access for blobs only
- **Container**: Anonymous read access for blobs and container metadata

**Risk Scenarios:**
- Data exfiltration without authentication
- Exposure of sensitive files, databases, backups, or credentials
- Compliance violations (GDPR, HIPAA, PCI-DSS)
- Ransomware actors scanning for exposed storage accounts
- Unauthorized data modification if write permissions misconfigured

**Detection Logic:**
Monitors Activity Logs for successful WRITE operations on:
- Storage account properties (allowBlobPublicAccess setting)
- Blob service configurations (publicAccess on containers)

**Investigation Steps:**
1. Identify the storage account: Check log.resourceId for the full resource path
2. Review who made the change: Check log.identityClaimUid (user/service principal)
3. Verify the specific change: Check log.propertiesAllowBlobPublicAccess or log.propertiesPublicAccess values
4. Determine change justification: Review log.callerIpAddress and check if change request was approved
5. Audit current configuration: Use Azure CLI/Portal to check current publicAccess settings
6. Scan for exposed data: Review containers and blobs for sensitive information
7. Check access logs: Look for anonymous requests in storage analytics logs
8. Review network rules: Verify if firewall/VNET rules provide additional protection
9. Verify encryption: Ensure encryption at rest and in transit is enabled
10. Check for data exfiltration: Review Storage Analytics logs for unusual download patterns

**Recommended Actions:**
- **Immediate**: If unauthorized, disable public access immediately via Azure Portal or CLI
- **CLI Command**: `az storage account update --name <account> --allow-blob-public-access false`
- Review all containers for public access: `az storage container list --account-name <account>`
- Enable Azure Defender for Storage for threat detection
- Implement Azure Private Link to restrict access to private networks
- Configure network rules (firewall/VNET) to limit access
- Enable storage account access logs and monitoring
- Implement Shared Access Signatures (SAS) with expiration for temporary access
- Use Azure RBAC instead of public access for authorized users
- Enable soft delete and versioning for blob protection
- Set up alerts for anonymous access attempts in storage analytics

**Azure Security Best Practices:**
- Disable public access at the account level by default
- Use Azure Private Endpoints for private connectivity
- Require secure transfer (HTTPS) for all operations
- Enable Azure AD authentication instead of shared keys
- Implement least privilege access with Azure RBAC
- Enable Azure Policy to prevent public access across subscriptions

**Common Legitimate Use Cases:**
- Hosting static website content (images, CSS, JS)
- Distributing public software/packages
- Sharing public datasets or documentation

Even for legitimate cases, consider alternatives like Azure CDN with authentication or Azure Static Web Apps.

**MITRE ATT&CK Reference:** T1530 - Data from Cloud Storage Object
', '["https://attack.mitre.org/techniques/T1530/","https://learn.microsoft.com/en-us/azure/storage/blobs/anonymous-read-access-configure","https://learn.microsoft.com/en-us/azure/storage/common/storage-network-security","https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log-schema"]', 'equalsIgnoreCase("log.category", "Administrative") &&
contains("log.operationName", "Microsoft.Storage/storageAccounts") &&
(contains("log.operationName", "write") || contains("log.operationName", "blobServices")) &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-01-29 16:19:37.423512', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1633, 'Virtual Machine Suspicious Activities', 2, 2, 3, 'Defense Evasion', 'T1578 - Modify Cloud Compute Infrastructure', 'Detects suspicious activities on Azure Virtual Machines including rapid creation, deletion, or configuration changes that could indicate compromise or abuse. This rule triggers when multiple VM operations are performed from the same IP address within a short timeframe.

Next Steps:
1. Review the specific VM operations performed and verify if they are legitimate business activities
2. Check the user account and IP address associated with the activities for any signs of compromise
3. Examine the timing and frequency of operations to determine if they follow normal usage patterns
4. Verify if the operations were performed during expected business hours
5. Check for any associated alerts or anomalies in authentication logs
6. Review VM configurations and access logs for any unauthorized changes
7. Contact the resource owner to confirm if the activities were authorized
', '["https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log-schema","https://attack.mitre.org/techniques/T1578/"]', 'contains("log.operationName", "Microsoft.Compute/virtualMachines") && (contains("log.operationName", "/write") || contains("log.operationName", "/delete") || contains("log.operationName", "/restart/action") || contains("log.operationName", "/powerOff/action")) && equalsIgnoreCase("log.category", "Administrative") && equalsIgnoreCase("log.resultSignature", "Succeeded") && exists("origin.ip")', '2026-01-29 16:19:38.433747', true, true, 'origin', '["origin.ip"]', '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-10m","count":8}]', null);

insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1593, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1594, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1595, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1596, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1597, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1598, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1599, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1600, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1601, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1602, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1603, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1604, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1605, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1606, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1607, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1608, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1609, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1610, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1611, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1612, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1613, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1614, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1615, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1616, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1617, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1618, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1619, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1620, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1621, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1622, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1623, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1624, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1625, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1626, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1627, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1628, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1629, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1630, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1631, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1632, 3, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1633, 3, null);
