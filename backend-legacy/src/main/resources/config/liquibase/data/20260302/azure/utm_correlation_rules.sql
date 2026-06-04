INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1259, 'Azure PIM Role Activation Anomaly', 3, 3, 1, 'Privilege Escalation', 'T1078 - Valid Accounts', e'Detects unusual Privileged Identity Management (PIM) role activation patterns including activation of high-privilege roles such as Global Administrator or Privileged Role Administrator. Repeated or unusual PIM activations may indicate an attacker leveraging compromised credentials to escalate privileges.

Next Steps:
1. Verify the user activating the PIM role has legitimate business justification
2. Review the specific role being activated and its scope
3. Check the activation justification message provided by the user
4. Review the activation duration and whether it exceeds normal patterns
5. Check for unusual source IP or device during the activation
6. If unauthorized, immediately deactivate the role and disable the user account
7. Review PIM audit logs for other suspicious activations by the same user
8. Implement PIM access reviews and require approval for critical roles
', '["https://learn.microsoft.com/en-us/azure/active-directory/privileged-identity-management/pim-configure","https://attack.mitre.org/techniques/T1078/"]', e'(contains("log.operationName", "Add member to role completed (PIM activation)") ||
 contains("log.operationName", "Add eligible member to role in PIM completed") ||
 contains("log.operationName", "Activate PIM role")) &&
equals("log.categoryValue", "Administrative")
', '2026-03-02 22:35:56.083827', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.user","operator":"filter_term","value":"{{.origin.user}}"},{"field":"log.categoryValue","operator":"filter_term","value":"Administrative"}],"or":null,"within":"now-4h","count":3}]', '["lastEvent.log.operationName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1260, 'Azure Managed Identity Token Abuse', 3, 3, 1, 'Credential Access', 'T1078.004 - Valid Accounts: Cloud Accounts', e'Detects suspicious token acquisition from Azure Instance Metadata Service (IMDS) by managed identities. Attackers who compromise an Azure VM can abuse managed identities to obtain access tokens for Azure resources without credentials, enabling lateral movement across the cloud environment.

Next Steps:
1. Identify the Azure resource (VM, App Service, Function) where the token was acquired
2. Review the target resource being accessed with the managed identity token
3. Check if the managed identity\'s permissions follow least privilege principles
4. Investigate the process or application that requested the token
5. Review Azure Activity logs for actions performed using the managed identity
6. If unauthorized, restrict the managed identity\'s role assignments immediately
7. Investigate the source VM for signs of compromise
8. Implement Conditional Access policies for workload identities
', '["https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview","https://attack.mitre.org/techniques/T1078/004/"]', e'contains("log.operationName", "Microsoft.ManagedIdentity") &&
equals("log.categoryValue", "Administrative") &&
(contains("log.properties.message", "token") ||
 contains("log.operationName", "tokens"))
', '2026-03-02 22:35:57.698411', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-1h","count":5}]', '["lastEvent.log.operationName","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1261, 'Azure Key Vault Excessive Access Detected', 3, 2, 1, 'Collection', 'T1530 - Data from Cloud Storage Object', e'Detects unusual spikes in Azure Key Vault access patterns. Monitors for multiple secret retrieval operations from the same source, which could indicate credential harvesting or data exfiltration attempts.

Next Steps:
1. Investigate the source IP address and verify if it\'s a legitimate system or user
2. Review the specific secrets/keys being accessed and their criticality
3. Check for any recent changes to Key Vault access policies
4. Correlate with user authentication logs to identify the account responsible
5. Verify if the access pattern aligns with normal business operations
6. Consider implementing additional access controls or monitoring if suspicious activity is confirmed
', '["https://learn.microsoft.com/en-us/azure/key-vault/general/logging","https://attack.mitre.org/techniques/T1530/"]', e'equals("log.category", "AuditEvent") &&
oneOf("log.operationName", ["SecretGet", "SecretList", "KeyGet"])
', '2026-03-02 22:35:59.272256', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.category","operator":"filter_term","value":"AuditEvent"}],"or":null,"within":"now-10m","count":20}]', '["lastEvent.log.resourceId","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1262, 'Azure AD Resource Owner Password Credentials Flow Detected', 2, 2, 1, 'Credential Access', 'T1078 - Valid Accounts', e'Detects use of the Resource Owner Password Credentials (ROPC) OAuth flow in Azure AD. ROPC sends plain-text credentials directly to the token endpoint, bypassing MFA and conditional access. It is commonly abused by attackers for credential stuffing and automated account compromise.

Next Steps:
1. Identify the application using ROPC flow and verify its legitimacy
2. Check if the application has a legitimate need for ROPC (legacy/headless apps)
3. Review the source IPs making ROPC requests for suspicious patterns
4. Check for high volumes of failed ROPC requests (credential stuffing)
5. Migrate the application to a modern auth flow (authorization code, device code)
6. If unauthorized, block the application and reset affected user passwords
', '["https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth-ropc","https://attack.mitre.org/techniques/T1078/"]', e'contains("log.properties", "urn:ietf:params:oauth:grant-type:password") ||
(contains("log.operationName", "Sign-in") && contains("log.properties", "ropc"))
', '2026-03-02 22:36:00.844215', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-1h","count":5}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1263, 'Azure AD LAPS Password Recovery', 3, 2, 1, 'Credential Access', 'T1003 - OS Credential Dumping', e'Detects Local Administrator Password Solution (LAPS) password recovery from Entra ID. While LAPS recovery is a legitimate admin operation, excessive or unauthorized recovery attempts indicate credential dumping for lateral movement.

Next Steps:
1. Verify the user recovering the LAPS password has legitimate need
2. Check the target device and whether the user is responsible for it
3. Review the frequency of LAPS password recoveries by this user
4. Correlate with subsequent RDP or SMB connections to the target device
5. If unauthorized, rotate the LAPS password and investigate the user\'s activities
6. Review RBAC for LAPS password read permissions
', '["https://learn.microsoft.com/en-us/entra/identity/devices/howto-manage-local-admin-passwords","https://attack.mitre.org/techniques/T1003/"]', e'contains("log.operationName", "Recover device local administrator password") ||
(contains("log.operationName", "Read device local administrator password") && exists("log.properties"))
', '2026-03-02 22:36:02.459874', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-1h","count":3}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1264, 'Azure Kubernetes Secret Write or Delete', 3, 3, 2, 'Credential Access', 'T1552.007 - Unsecured Credentials: Container API', e'Detects write or delete operations on Kubernetes Secrets in Azure Kubernetes Service. Secrets contain sensitive data like service account tokens, TLS certificates, and database credentials. Unauthorized access indicates potential credential theft or data tampering.

Next Steps:
1. Identify the user or service account accessing the secrets
2. Review which secrets were accessed, modified, or deleted
3. Check if the operation was part of a legitimate deployment workflow
4. Audit the RBAC permissions of the identity performing the action
5. If unauthorized, rotate all affected secrets immediately
6. Review pod specifications for secrets mounted as volumes or environment variables
', '["https://kubernetes.io/docs/concepts/configuration/secret/","https://attack.mitre.org/techniques/T1552/007/"]', e'contains("log.operationName", "MICROSOFT.CONTAINERSERVICE") &&
contains("log.properties", "secrets") &&
(contains("log.properties", "create") || contains("log.properties", "update") || contains("log.properties", "delete") || contains("log.properties", "patch"))
', '2026-03-02 22:36:04.036942', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-30m","count":5}]', '["lastEvent.log.resourceId","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1265, 'Azure AD Bulk Privileged Role Assignment Changes', 3, 3, 2, 'Privilege Escalation', 'T1098 - Account Manipulation', e'Detects mass privileged role assignment changes in Azure AD. Multiple role assignments in a short time window indicate an attacker rapidly escalating privileges across multiple accounts for persistence and lateral movement.

Next Steps:
1. Review all role assignments made in the burst
2. Identify the admin account making the changes
3. Check if these changes were part of an approved onboarding or migration
4. Review the specific roles assigned (Global Admin, Exchange Admin, etc.)
5. If unauthorized, revert all role assignments and investigate the admin account
6. Enable Azure PIM for just-in-time role activation
', '["https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-resource-roles-assign-roles","https://attack.mitre.org/techniques/T1098/"]', e'contains("log.operationName", "Add member to role") ||
contains("log.operationName", "Add eligible member to role")
', '2026-03-02 22:36:05.649800', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-30m","count":10}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1266, 'Azure AD Password Spray Attack Detection', 3, 2, 1, 'Credential Access', 'T1110 - Brute Force', e'Detects password spray attacks against Azure AD by correlating failed sign-in attempts across multiple usernames from the same source IP within a short time window. Password spraying tries common passwords against many accounts to avoid account lockout thresholds.

Next Steps:
1. Identify the source IP and check threat intelligence feeds for known malicious sources
2. Review the list of targeted user accounts for patterns (executives, admins, service accounts)
3. Check if any of the targeted accounts subsequently had successful logins
4. Verify that account lockout policies are properly configured
5. Block the source IP at the network level if confirmed malicious
6. Enable Azure AD Smart Lockout for brute force protection
7. Implement Conditional Access policies requiring MFA
8. Review password policies and enforce complexity requirements
', '["https://learn.microsoft.com/en-us/azure/active-directory/identity-protection/concept-identity-protection-risks","https://attack.mitre.org/techniques/T1110/"]', e'contains("log.operationName", "Sign-in activity") &&
(equals("log.properties.status.errorCode", "50126") ||
 equals("log.properties.status.errorCode", "50053") ||
 equals("log.properties.status.errorCode", "50057")) &&
exists("origin.ip")
', '2026-03-02 22:36:07.222474', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.properties.status.errorCode","operator":"filter_match","value":"5005"}],"or":null,"within":"now-15m","count":15}]', '["lastEvent.log.operationName","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1267, 'Azure AD App Registration with High-Privilege API Permissions', 3, 3, 1, 'Persistence', 'T1098.001 - Account Manipulation: Additional Cloud Credentials', e'Detects creation of new Azure AD application registrations which may be used to establish persistence with high-privilege API permissions. Attackers create app registrations with permissions like Mail.ReadWrite, Directory.ReadWrite.All, or RoleManagement.ReadWrite.Directory to maintain access.

Next Steps:
1. Review the application registration and its requested API permissions
2. Verify the creator has authorization to register applications
3. Check if admin consent was granted for the application\'s permissions
4. Review the application\'s redirect URIs for suspicious external domains
5. Examine the application\'s credential types (secrets, certificates)
6. If unauthorized, delete the application registration and revoke any granted consents
7. Implement app registration policies to restrict who can create applications
8. Enable admin consent workflow for application permission requests
', '["https://learn.microsoft.com/en-us/azure/active-directory/develop/app-objects-and-service-principals","https://attack.mitre.org/techniques/T1098/001/"]', e'(contains("log.operationName", "Add application") ||
 contains("log.operationName", "Add service principal") ||
 contains("log.operationName", "Consent to application")) &&
equals("log.categoryValue", "Administrative")
', '2026-03-02 22:36:08.795179', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.user","operator":"filter_term","value":"{{.origin.user}}"},{"field":"log.categoryValue","operator":"filter_term","value":"Administrative"}],"or":null,"within":"now-1h","count":3}]', '["lastEvent.log.operationName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1268, 'Application Gateway WAF Security Alerts', 3, 3, 2, 'Initial Access', 'T1190 - Exploit Public-Facing Application', e'Detects Web Application Firewall alerts from Azure Application Gateway indicating potential web attacks or malicious activity. This rule triggers when WAF blocks or detects suspicious requests that match security rules.

**Next Steps:**
1. Review the specific WAF rule ID and message details to understand the attack type
2. Analyze the source IP address for reputation and geographic location
3. Examine the request URL, headers, and payload for attack indicators
4. Check for additional requests from the same source IP within the time window
5. Verify if this is a legitimate application behavior or actual attack attempt
6. Consider implementing additional WAF rules or IP blocking if confirmed malicious
7. Review application logs for any successful bypass attempts
', '["https://learn.microsoft.com/en-us/azure/web-application-firewall/ag/web-application-firewall-logs","https://attack.mitre.org/techniques/T1190/"]', e'(equals("log.operationName", "ApplicationGatewayFirewallLog") || equals("log.type", "ApplicationGatewayFirewallLog")) &&
equals("log.action", "Blocked") &&
exists("log.ruleId")
', '2026-03-02 22:36:10.423092', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-10m","count":5}]', '["lastEvent.log.ruleId","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1269, 'Azure AKS Container Security Threat Detection', 3, 3, 2, 'Execution', 'T1610 - Deploy Container', e'Detects suspicious container operations in Azure Kubernetes Service (AKS) including privileged pod creation, container exec commands, and potential container escape attempts. These activities may indicate an attacker attempting to deploy malicious workloads or escape container isolation.

Next Steps:
1. Review the Kubernetes audit logs for the specific pod or container operation
2. Check if the container image is from an approved registry
3. Verify the service account and RBAC permissions used for the operation
4. Examine pod security context for privileged flags, host network, or host PID access
5. Review the container command for suspicious payloads or reverse shells
6. If unauthorized, delete the pod and investigate the cluster for further compromise
7. Implement Azure Policy for AKS to enforce pod security standards
8. Enable Microsoft Defender for Containers for runtime protection
', '["https://learn.microsoft.com/en-us/azure/defender-for-cloud/defender-for-containers-introduction","https://attack.mitre.org/techniques/T1610/"]', e'(contains("log.operationName", "Microsoft.ContainerService") ||
 contains("log.operationName", "MICROSOFT.KUBERNETES")) &&
(contains("log.operationName", "write") ||
 contains("log.operationName", "create") ||
 contains("log.operationName", "exec")) &&
equals("log.resultType", "Success")
', '2026-03-02 22:36:12.096678', true, true, 'origin', null, '[{"indexPattern":"v11-log-azure-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.operationName","operator":"filter_match","value":"Container"}],"or":null,"within":"now-30m","count":10}]', '["lastEvent.log.operationName","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1270, 'Azure Subscription Ownership Transfer Detected', 3, 3, 2, 'Identity and Access Management', 'T1078 - Valid Accounts', e'Detects when ownership of an Azure subscription is transferred by monitoring role assignment changes at the subscription level. This could indicate unauthorized access or insider threat activity.

Next Steps:
1. Verify the legitimacy of the ownership transfer with the subscription administrator
2. Check if the user performing the transfer is authorized for this action
3. Review the timing and context of the transfer (business hours, planned change)
4. Examine other recent activities by the same user or from the same source IP
5. Validate that proper change management procedures were followed
6. Check for any unusual activity following the ownership transfer
7. If unauthorized, immediately revoke the new owner\'s access and escalate to security team
', '["https://learn.microsoft.com/en-us/azure/role-based-access-control/change-history-report","https://attack.mitre.org/techniques/T1078/"]', e'equals("log.operationName", "Microsoft.Authorization/roleAssignments/write") &&
contains("log.properties", "Owner") &&
equals("log.category", "Administrative") &&
contains("log.resourceId", "/subscriptions/") &&
!contains("log.resourceId", "/resourceGroups/")
', '2026-03-02 22:36:13.559108', true, true, 'origin', null, '[]', '["lastEvent.log.correlationId","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1271, 'Storage Account Public Access Enabled', 3, 2, 1, 'Collection', 'T1530 - Data from Cloud Storage Object', e'Detects when public access is enabled on Azure Storage Accounts which could lead to unauthorized data exposure.
This configuration change creates a significant security risk as it allows anonymous access to stored data.

Next Steps:
1. Immediately review the affected storage account configuration
2. Verify if public access was intentionally enabled and properly authorized
3. Check if any sensitive data is stored in the account
4. Review access logs for any unauthorized access attempts
5. Consider disabling public access if not required for business operations
6. Implement network restrictions and access policies if public access is necessary
7. Monitor for any data exfiltration activities
', '["https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log-schema","https://attack.mitre.org/techniques/T1530/"]', e'contains("log.operationName", "Microsoft.Storage/storageAccounts") &&
(contains("log.operationName", "/write") || contains("log.operationName", "/blobServices/write")) &&
equals("log.category", "Administrative") &&
equals("log.actionResult", "accepted") &&
(contains("log.properties", "allowBlobPublicAccess") || contains("log.properties", "publicAccess"))
', '2026-03-02 22:36:15.084397', true, true, 'origin', null, '[]', '["lastEvent.log.aadObjectId","lastEvent.log.resourceId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1272, 'Multi-Factor Authentication Disabled for an Azure User', 3, 3, 2, 'Persistence', 'T1556 - Modify Authentication Process', e'Detects when multi-factor authentication (MFA) is disabled for an Azure AD/Entra ID user account through Audit Logs.

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
', '["https://attack.mitre.org/techniques/T1556/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/entra/identity/authentication/howto-mfa-userstates","https://learn.microsoft.com/en-us/entra/identity/authentication/concept-mfa-licensing"]', e'equalsIgnoreCase("log.category", "AuditLogs") &&
equalsIgnoreCase("log.operationName", "Disable Strong Authentication") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-03-02 22:36:16.382203', true, true, 'target', null, '[]', '["target.ip","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1273, 'Azure Global Administrator Role Addition to PIM User', 3, 3, 3, 'Persistence', 'T1098.001 - Account Manipulation: Additional Cloud Credentials', e'Detects when users are granted Global Administrator (Company Administrator) role assignments through Azure AD/Entra ID Privileged Identity Management (PIM).

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
4. Check assignment type: Determine if it\'s eligible (requires activation) or time-bound (direct)
5. Review duration: For time-bound assignments, check the duration of the assignment
6. Analyze timing: Determine if assignment follows suspicious authentication or compromise indicators
7. Review justification: Check if a business justification was provided in log.propertiesAdditionalDetails
8. Check user history: Review the assignee\'s account for recent suspicious activity
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
', '["https://attack.mitre.org/techniques/T1098/001/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-configure","https://learn.microsoft.com/en-us/entra/identity/role-based-access-control/permissions-reference#global-administrator"]', e'equalsIgnoreCase("log.category", "AuditLogs") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS")) &&
(contains("log.operationName", "Add eligible member to role") || contains("log.operationName", "Add member to role")) &&
(contains("log.properties.targetResources.displayName", "Global Administrator") || contains("log.properties.targetResources.displayName", "Company Administrator"))
', '2026-03-02 22:36:17.750986', true, true, 'target', null, '[]', '["target.ip","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1274, 'MFA Disabled for Privileged Azure AD User', 3, 3, 1, 'Defense Evasion', 'T1556 - Modify Authentication Process', e'Detects when Multi-Factor Authentication (MFA) is disabled for privileged users in Azure AD. This could indicate an attempt to weaken security controls for unauthorized access.

Next Steps:
1. Verify if the MFA disable action was authorized and legitimate
2. Check who initiated the change and from which IP address
3. Review the user\'s recent login activity and permissions
4. Ensure the user account has not been compromised
5. Re-enable MFA if the change was unauthorized
6. Consider implementing conditional access policies to prevent unauthorized MFA changes
', '["https://learn.microsoft.com/en-us/entra/identity/authentication/howto-mfa-reporting","https://attack.mitre.org/techniques/T1556/"]', e'(equals("log.operationName", "Disable Strong Authentication") ||
 (equals("log.operationName", "Update user") && contains("log.properties", "StrongAuthenticationMethod"))) &&
equals("log.categoryValue", "Administrative")
', '2026-03-02 22:36:19.154309', true, true, 'origin', null, '[]', '["lastEvent.log.correlationId","lastEvent.log.targetUserPrincipalName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1275, 'Possible Consent Grant Attack via Azure-Registered Application', 3, 3, 2, 'Initial Access', 'T1078 - Valid Accounts', 'Detects when a user grants permissions to an Azure-registered application or when an administrator grants tenant-wide permissions to an application. An adversary may create an Azure-registered application that requests access to data such as contact information, email, or documents. Consent grant attacks are commonly used in phishing campaigns where malicious OAuth applications trick users into granting excessive permissions, enabling data exfiltration or unauthorized access to organizational resources.', '["https://attack.mitre.org/techniques/T1566/","https://learn.microsoft.com/en-us/entra/identity/enterprise-apps/manage-consent-requests","https://learn.microsoft.com/en-us/defender-cloud-apps/investigate-risky-oauth"]', e'(equalsIgnoreCase("log.category", "AuditLogs") || contains("log.category", "Audit")) &&
equalsIgnoreCase("log.operationName", "Consent to application") &&
equals("log.resultType", "0")
', '2026-03-02 22:36:20.530093', true, true, 'target', null, '[]', '["target.ip","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1276, 'Azure Active Directory High Risk Sign-in', 3, 3, 2, 'Initial Access', 'T1078 - Valid Accounts', 'Identifies high risk Azure Active Directory (AD) sign-ins by leveraging Microsoft''s Identity Protection machine learning and heuristics. Identity Protection categorizes risk into three tiers: low, medium, and high. While Microsoft does not provide specific details about how risk is calculated, each level brings higher confidence that the user or sign-in is compromised. This rule triggers on ''high'' risk level sign-ins, which indicate strong indicators of compromise such as impossible travel, anonymous IP usage, or leaked credentials.', '["https://attack.mitre.org/techniques/T1078/","https://learn.microsoft.com/en-us/entra/id-protection/concept-identity-protection-risks","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/signinlogs"]', e'equalsIgnoreCase("log.category", "SignInLogs") &&
equalsIgnoreCase("log.properties.RiskLevelDuringSignIn", "high") &&
equalsIgnoreCase("log.propertiesTokenIssuerType", "AzureAD") &&
equals("log.resultType", "0")
', '2026-03-02 22:36:21.903345', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1277, 'Azure Service Principal Credentials Added', 3, 3, 2, 'Persistence', 'T1098.001 - Account Manipulation: Additional Cloud Credentials', e'Detects when new credentials (certificates or secrets) are added to Azure service principals through Azure AD/Entra ID Audit Logs.

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
', '["https://attack.mitre.org/techniques/T1098/001/","https://learn.microsoft.com/en-us/azure/azure-monitor/reference/tables/auditlogs","https://learn.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal"]', e'equalsIgnoreCase("log.category", "AuditLogs") &&
contains("log.operationName", "Add service principal") &&
(equals("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-03-02 22:36:23.393358', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1278, 'Azure AD Golden SAML and Federation Domain Abuse', 3, 3, 2, 'Credential Access', 'T1606.002 - Forge Web Credentials: SAML Tokens', e'Detects additions or modifications of federated domains in Azure AD which could indicate Golden SAML attacks. Attackers who compromise AD FS signing certificates or add rogue federation domains can forge SAML tokens to impersonate any user in the organization.

Next Steps:
1. Immediately verify if the federation domain change was authorized
2. Review the domain being added and its federation metadata endpoint
3. Check the AD FS signing certificate for unauthorized modifications
4. Verify the identity of the administrator making the change
5. Review Azure AD audit logs for other suspicious tenant-level changes
6. If unauthorized, immediately remove the federated domain and revoke all active sessions
7. Rotate the AD FS token signing certificate
8. Enable Certificate Authority revocation checking for federation certificates
', '["https://learn.microsoft.com/en-us/azure/active-directory/hybrid/whatis-fed","https://attack.mitre.org/techniques/T1606/002/"]', e'(contains("log.operationName", "Set federation settings on domain") ||
 contains("log.operationName", "Set domain authentication") ||
 contains("log.operationName", "Add unverified domain") ||
 contains("log.operationName", "Add verified domain") ||
 contains("log.operationName", "Set DomainFederationSettings")) &&
equals("log.categoryValue", "Administrative")
', '2026-03-02 22:36:24.966927', true, true, 'origin', null, '[]', '["lastEvent.log.operationName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1279, 'Azure Diagnostic Settings Tampering', 2, 3, 2, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', e'Detects deletion or modification of Azure diagnostic settings which are used to route platform logs and metrics to monitoring destinations. Attackers may disable diagnostic settings to prevent their activities from being logged and detected.

Next Steps:
1. Verify if the diagnostic settings change was authorized through change management
2. Identify which resources lost their diagnostic logging
3. Review the identity performing the change and confirm authorization
4. Check if any suspicious activities occurred after logging was disabled
5. Restore diagnostic settings for affected resources immediately
6. Implement Azure Policy to enforce diagnostic settings on all resources
7. Set up alerts for diagnostic settings modifications
8. Review Azure Activity Log for other defense evasion activities by the same identity
', '["https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/diagnostic-settings","https://attack.mitre.org/techniques/T1562/008/"]', e'contains("log.operationName", "Microsoft.Insights/diagnosticSettings") &&
(contains("log.operationName", "delete") ||
 contains("log.operationName", "Delete")) &&
equals("log.resultType", "Success")
', '2026-03-02 22:36:26.578967', true, true, 'origin', null, '[]', '["lastEvent.log.resourceId","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1280, 'Azure Event Hub Deletion', 1, 3, 3, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', e'Detects the deletion of an Azure Event Hub, which is a critical event processing service that ingests and processes large volumes of events, logs, and telemetry data. Event Hubs are commonly used for security monitoring, log aggregation, and SIEM integration. Adversaries may delete Event Hubs to evade detection by disrupting log collection pipelines and preventing security events from reaching monitoring systems.

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
', '["https://attack.mitre.org/techniques/T1562/008/","https://attack.mitre.org/tactics/TA0005/","https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs","https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log"]', e'(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) &&
(equalsIgnoreCase("log.operationName", "MICROSOFT.EVENTHUB/NAMESPACES/EVENTHUBS/DELETE") ||
contains("log.operationName", "Delete EventHub")) &&
(equalsIgnoreCase("log.resultType", "0") || equalsIgnoreCase("actionResult", "SUCCESS"))
', '2026-03-02 22:36:28.109574', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1281, 'Azure Diagnostic Settings Deletion', 1, 3, 3, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', e'Detects the deletion of diagnostic settings in Azure, which are critical for sending platform logs, metrics, and activity data to destinations like Log Analytics workspaces, Event Hubs, or storage accounts. Adversaries delete diagnostic settings to evade detection by disabling security monitoring and audit logging capabilities.

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
7. Investigate the caller\'s account for potential compromise
8. Check for other defense evasion techniques in the timeline
', '["https://attack.mitre.org/techniques/T1562/008/","https://attack.mitre.org/tactics/TA0005/","https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/diagnostic-settings","https://learn.microsoft.com/en-us/azure/azure-monitor/essentials/activity-log"]', e'(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) &&
(equalsIgnoreCase("log.operationName", "MICROSOFT.INSIGHTS/DIAGNOSTICSETTINGS/DELETE") ||
contains("log.operationName", "Delete diagnostic setting")) &&
equalsIgnoreCase("log.resultType", "0")
', '2026-03-02 22:36:29.724542', true, true, 'target', null, '[]', '["target.ip","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1282, 'Azure Defender for Cloud Critical Security Alert', 3, 3, 2, 'Intrusion Detection', 'TA0001 - Initial Access', e'Detects critical severity alerts from Azure Defender for Cloud indicating potential active threats, malware infections, or successful breach attempts that require immediate response.

Next Steps:
1. Review the full alert details in Azure Defender for Cloud portal
2. Verify the affected resource and assess the scope of potential compromise
3. Check for related suspicious activities on the affected resource
4. Implement immediate containment measures if threat is confirmed
5. Review security policies and configurations for the affected resource
6. Document the incident and update security procedures as needed
', '["https://learn.microsoft.com/en-us/azure/defender-for-cloud/alerts-overview","https://learn.microsoft.com/en-us/azure/defender-for-cloud/alerts-schemas","https://attack.mitre.org/tactics/TA0001/"]', e'(equals("log.eventName", "Microsoft.Security/locations/alerts/Activate/action") || contains("log.operationName", "Microsoft.Security")) &&
equals("log.category", "Security") &&
oneOf("log.level", ["Critical", "High", "Error"]) &&
(equals("log.properties.severity", "High") || equals("log.properties.alertSeverity", "High"))
', '2026-03-02 22:36:31.298397', true, true, 'origin', null, '[]', '["lastEvent.log.correlationId","lastEvent.log.eventDataId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1283, 'Azure Key Vault Modified', 3, 3, 2, 'Credential Access', 'T1552 - Unsecured Credentials', 'Identifies modifications to a Key Vault in Azure. The Key Vault is a service that safeguards encryption keys and secrets like certificates,  connection strings, and passwords. Because this data is sensitive and business critical, access to key vaults should be secured to allow  only authorized applications and users. Adversaries may modify Key Vault configurations to weaken security controls, add unauthorized access policies,  or change network rules to facilitate credential theft and unauthorized access to sensitive secrets.', '["https://attack.mitre.org/techniques/T1552/","https://attack.mitre.org/tactics/TA0006/","https://learn.microsoft.com/en-us/azure/key-vault/general/security-features"]', e'(equalsIgnoreCase("log.category", "Administrative") || contains("log.category", "Activity")) &&
(equalsIgnoreCase("log.operationName", "MICROSOFT.KEYVAULT/VAULTS/WRITE") ||
contains("log.operationName", "Microsoft.KeyVault/vaults/write")) &&
equals("log.resultType", "0")
', '2026-03-02 22:36:32.910598', true, true, 'target', null, '[]', '["target.ip","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1284, 'Azure Subscription Permission Elevation via ElevateAccess', 3, 3, 3, 'Privilege Escalation', 'T1078.004 - Valid Accounts: Cloud Accounts', e'Detects the MICROSOFT.AUTHORIZATION/ELEVATEACCESS/ACTION operation which grants a Global Administrator access to ALL Azure subscriptions in the tenant. This is an extremely high-impact action that should be very rare and carefully monitored.

Next Steps:
1. Immediately verify this action was authorized by a known Global Administrator
2. Check if a change request or emergency procedure exists for this action
3. Review what subscription-level changes were made after the elevation
4. Check for new role assignments at the management group or subscription level
5. If unauthorized, remove the User Access Administrator role and audit all changes
6. Enable Azure PIM (Privileged Identity Management) if not already in use
', '["https://learn.microsoft.com/en-us/azure/role-based-access-control/elevate-access-global-admin","https://attack.mitre.org/techniques/T1078/004/"]', e'regexMatch("log.operationName", "(?i)MICROSOFT\\\\.AUTHORIZATION/ELEVATEACCESS/ACTION")
', '2026-03-02 22:36:34.355261', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1285, 'Azure AD Temporary Access Pass Registration', 3, 2, 1, 'Credential Access', 'T1078.004 - Valid Accounts: Cloud Accounts', e'Detects registration of Temporary Access Pass (TAP) in Azure AD. TAPs can be used to bypass MFA requirements and are a growing attack vector for initial access and MFA circumvention.

Next Steps:
1. Verify the TAP was requested through legitimate channels (IT helpdesk)
2. Check the admin user who created the TAP for legitimacy
3. Review the target user and reason for TAP issuance
4. Check for sign-ins using the TAP, especially from unusual locations
5. Verify MFA registration events following the TAP usage
6. If unauthorized, revoke the TAP immediately and investigate
7. Review TAP policy settings for appropriate lifetime and usage limits
', '["https://learn.microsoft.com/en-us/entra/identity/authentication/howto-authentication-temporary-access-pass","https://attack.mitre.org/techniques/T1078/004/"]', e'(contains("log.operationName", "Admin registered security info") && contains("log.properties", "Temporary Access Pass")) ||
(contains("log.operationName", "Update user") && contains("log.properties", "TemporaryAccessPass"))
', '2026-03-02 22:36:35.767739', true, true, 'origin', null, '[]', '["lastEvent.log.properties.targetResources","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1286, 'Azure Sentinel High/Critical Alert Pattern Detection', 3, 3, 2, 'Threat Detection', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects high-severity or critical alerts from Azure Sentinel that may indicate coordinated attack activity or serious security incidents requiring immediate investigation. This rule identifies new alerts with High or Critical severity levels from Microsoft Sentinel that could represent active threats.

Next Steps:
1. Review the alert details and affected resources immediately
2. Correlate with other security events in the environment
3. Check for signs of lateral movement or privilege escalation
4. Verify if the alert represents a true positive through manual investigation
5. Implement containment measures if attack activity is confirmed
6. Document findings and update incident response procedures
', '["https://learn.microsoft.com/en-us/azure/sentinel/security-alert-schema","https://attack.mitre.org/techniques/T1562/"]', e'oneOf("log.AlertSeverity", ["High", "Critical"]) &&
equals("log.VendorName", "Microsoft Sentinel") &&
equals("log.Status", "New")
', '2026-03-02 22:36:37.309930', true, true, 'origin', null, '[]', '["lastEvent.log.AlertType","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1287, 'Azure AD Password Spray Attack Detected', 3, 2, 1, 'Credential Access', 'T1110.003 - Brute Force: Password Spraying', e'Detects Azure Identity Protection password spray attack signals. Microsoft\'s ML-based detection identifies distributed password spray attempts across multiple accounts using common passwords.

Next Steps:
1. Identify all affected user accounts in the password spray
2. Check if any accounts were successfully compromised
3. Force password resets for all targeted accounts
4. Review source IPs for known attack infrastructure
5. Check for successful sign-ins from the same source IPs
6. Enable smart lockout policies if not already configured
7. Review MFA enforcement across all targeted accounts
', '["https://learn.microsoft.com/en-us/entra/id-protection/concept-identity-protection-risks","https://attack.mitre.org/techniques/T1110/003/"]', e'contains("log.operationName", "Password Spray") ||
(contains("log.properties", "riskEventType") && contains("log.properties", "passwordSpray"))
', '2026-03-02 22:36:38.834034', true, true, 'origin', null, '[]', '["lastEvent.log.operationName","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1288, 'Azure Primary Refresh Token Access Attempt', 3, 3, 1, 'Credential Access', 'T1528 - Steal Application Access Token', e'Detects attempts to access the Primary Refresh Token (PRT) in Azure AD. PRT theft is a high-confidence compromise indicator as PRTs provide SSO access across all Azure AD-integrated applications and can be used to bypass conditional access policies.

Next Steps:
1. Immediately investigate the user account associated with this alert
2. Check the device from which the PRT access was attempted
3. Review sign-in logs for the affected user for anomalous patterns
4. Check for token replay attacks or sessions from unexpected locations
5. If compromise is confirmed, revoke all refresh tokens for the user
6. Re-register the device and force re-authentication
7. Review conditional access policies for PRT-based bypass vulnerabilities
', '["https://learn.microsoft.com/en-us/entra/identity/devices/concept-primary-refresh-token","https://attack.mitre.org/techniques/T1528/"]', e'contains("log.operationName", "Primary Refresh Token") ||
(contains("log.properties", "PRT") && contains("log.properties", "access"))
', '2026-03-02 22:36:40.294393', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1289, 'Azure AD New Root Certificate Authority Added', 3, 3, 2, 'Persistence', 'T1556 - Modify Authentication Process', e'Detects when a new root certificate authority is added to the TrustedCAsForPasswordlessAuth configuration in Azure AD. Adding a rogue root CA enables persistent passwordless authentication backdoor access.

Next Steps:
1. Immediately verify the root CA addition was authorized
2. Review the certificate details and issuing authority
3. Check the user identity performing the change
4. Validate the CA against your organization\'s known PKI infrastructure
5. If unauthorized, remove the root CA immediately
6. Audit all certificate-based authentications since the CA was added
7. Review Azure AD authentication methods policies
', '["https://learn.microsoft.com/en-us/entra/identity/authentication/concept-certificate-based-authentication","https://attack.mitre.org/techniques/T1556/"]', e'contains("log.operationName", "TrustedCAsForPasswordlessAuth") ||
(contains("log.operationName", "Update organization settings") && contains("log.properties", "certificateAuthorities"))
', '2026-03-02 22:36:41.980438', true, true, 'origin', null, '[]', '["lastEvent.log.operationName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1290, 'Azure Application Credential Modification', 3, 3, 2, 'Defense Evasion', 'T1098.001 - Account Manipulation: Additional Cloud Credentials', e'Detects when a new credential (certificate or secret) is added to an Azure AD application. Applications can use certificates or secret strings to authenticate when requesting tokens. Adversaries may add additional authentication credentials to existing applications to establish persistence, evade defenses, or enable privilege escalation by impersonating legitimate applications.

This technique is commonly used in post-compromise scenarios where attackers:
- Add secrets to high-privilege applications to maintain access
- Create backdoor authentication methods to evade MFA requirements
- Establish persistence mechanisms that survive password resets
- Enable token-based authentication for automated attacks

Next Steps:
1. Verify if the credential modification was authorized and expected
2. Identify who performed the operation (check InitiatedBy field)
3. Review the affected application\'s permissions and access scope
4. Check for subsequent suspicious sign-in activity using the application
5. Audit other applications for similar unauthorized modifications
6. If unauthorized, immediately remove the suspicious credentials
7. Review application usage logs for potential abuse
8. Investigate the source IP address and user agent of the modification
', '["https://attack.mitre.org/techniques/T1098/001/","https://attack.mitre.org/tactics/TA0005/","https://learn.microsoft.com/en-us/azure/active-directory/reports-monitoring/concept-audit-logs","https://learn.microsoft.com/en-us/entra/identity/monitoring-health/reference-audit-activities"]', e'(equalsIgnoreCase("log.category", "AuditLogs") || contains("log.category", "Audit")) &&
(contains("log.operationName", "Certificates and secrets management") ||
equalsIgnoreCase("log.operationName", "Add service principal credentials") ||
equalsIgnoreCase("log.operationName", "Update application") ||
equalsIgnoreCase("log.operationName", "Update application - Certificates and secrets management")) &&
equalsIgnoreCase("log.resultType", "0")
', '2026-03-02 22:36:43.553631', true, true, 'target', null, '[]', '["target.ip","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1291, 'Azure AD Leaked Credentials Detection', 3, 3, 2, 'Credential Access', 'T1078 - Valid Accounts', e'Detects Azure Identity Protection alerts for leaked credentials found on dark web, paste sites, or other sources. This indicates user credentials have been exposed and may be used for unauthorized access.

Next Steps:
1. Immediately force a password reset for the affected user
2. Revoke all active sessions and refresh tokens
3. Review recent sign-in activity for unauthorized access
4. Check for any data access or configuration changes after the leak
5. Enable MFA if not already required for the user
6. Investigate how the credentials were leaked (phishing, malware, reuse)
7. Check if the same password was used across other services
', '["https://learn.microsoft.com/en-us/entra/id-protection/concept-identity-protection-risks","https://attack.mitre.org/techniques/T1078/"]', e'contains("log.operationName", "Leaked Credentials") ||
(contains("log.properties", "riskEventType") && contains("log.properties", "leakedCredentials"))
', '2026-03-02 22:36:45.168922', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1292, 'Azure Kubernetes Events Deleted', 1, 3, 2, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects deletion of Kubernetes events in Azure Kubernetes Service (AKS). Attackers delete events to cover traces of their activities within the cluster.

Next Steps:
1. Identify the user or service account that deleted the events
2. Check for other suspicious Kubernetes operations from the same identity
3. Review AKS audit logs for activities that occurred before the event deletion
4. Verify if this was part of a legitimate cluster maintenance operation
5. If unauthorized, investigate the cluster for signs of compromise
6. Review RBAC policies to restrict event deletion permissions
', '["https://learn.microsoft.com/en-us/azure/aks/monitor-aks","https://attack.mitre.org/techniques/T1562/001/"]', e'contains("log.operationName", "MICROSOFT.CONTAINERSERVICE") &&
(contains("log.properties", "events") && contains("log.properties", "delete"))
', '2026-03-02 22:36:46.747455', true, true, 'origin', null, '[]', '["lastEvent.log.resourceId","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1293, 'Azure Kubernetes Admission Webhook Modified', 3, 3, 2, 'Persistence', 'T1078.004 - Valid Accounts: Cloud Accounts', e'Detects creation or modification of MutatingAdmissionWebhook or ValidatingAdmissionWebhook configurations in Azure Kubernetes Service. Attackers use admission controllers to inject malicious containers or modify workload specifications.

Next Steps:
1. Review the webhook configuration and its target service
2. Verify the webhook was created as part of a legitimate deployment
3. Check the webhook\'s namespace selector and object selector
4. Examine what resources the webhook intercepts (pods, deployments, etc.)
5. If unauthorized, delete the webhook and audit all recent pod deployments
6. Review cluster RBAC for excessive admission controller permissions
', '["https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/","https://attack.mitre.org/techniques/T1078/004/"]', e'contains("log.operationName", "MICROSOFT.CONTAINERSERVICE") &&
(contains("log.properties", "MutatingWebhookConfiguration") || contains("log.properties", "ValidatingWebhookConfiguration")) &&
(contains("log.properties", "create") || contains("log.properties", "update") || contains("log.properties", "patch"))
', '2026-03-02 22:36:48.115412', true, true, 'origin', null, '[]', '["lastEvent.log.resourceId","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1294, 'Azure AD Impossible Travel Sign-In', 3, 2, 1, 'Initial Access', 'T1078 - Valid Accounts', e'Detects Azure Identity Protection impossible travel alerts where a user signs in from geographically distant locations in a timeframe that makes physical travel impossible. This strongly indicates credential theft or session hijacking.

Next Steps:
1. Contact the user to verify both sign-in locations
2. Check if a VPN or proxy could explain the geolocation discrepancy
3. Review the sign-in details (device, browser, app) for both locations
4. If unauthorized, force password reset and revoke all sessions
5. Review data access and actions from the suspicious location
6. Enable location-based conditional access policies
7. Check for other users with similar patterns from the same locations
', '["https://learn.microsoft.com/en-us/entra/id-protection/concept-identity-protection-risks","https://attack.mitre.org/techniques/T1078/"]', e'contains("log.operationName", "Impossible Travel") ||
(contains("log.properties", "riskEventType") && contains("log.properties", "impossibleTravel"))
', '2026-03-02 22:36:49.648902', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1295, 'Azure AD Federation Settings Modified', 3, 3, 3, 'Credential Access', 'T1556 - Modify Authentication Process', e'Detects modifications to Azure AD domain federation settings. Changing federation configuration is a critical attack technique that enables Golden SAML attacks and domain takeover, allowing attackers to forge authentication tokens for any user.

Next Steps:
1. Immediately verify the federation modification was authorized
2. Check the user identity and source IP performing the change
3. Review the new federation settings for suspicious IdP configurations
4. Validate the signing certificate in the federation configuration
5. Check for subsequent sign-ins using federated authentication
6. If unauthorized, revert the federation changes and investigate all federated sessions
7. Review all privileged role assignments that occurred after the change
', '["https://learn.microsoft.com/en-us/entra/identity/hybrid/connect/whatis-fed","https://attack.mitre.org/techniques/T1556/"]', e'contains("log.operationName", "Set federation settings on domain") ||
(contains("log.operationName", "Set domain authentication") && contains("log.properties", "Federated"))
', '2026-03-02 22:36:50.988512', true, true, 'origin', null, '[]', '["lastEvent.log.operationName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1296, 'Azure Disk Snapshot Exfiltration', 3, 2, 1, 'Data Exfiltration', 'T1537 - Transfer Data to Cloud Account', e'Detects Azure disk snapshot operations that could be used for data exfiltration, including sharing snapshots across subscriptions, generating SAS URIs for download, or copying snapshots to external storage accounts.

Next Steps:
1. Identify the disk snapshot and the virtual machine it was taken from
2. Review the target location or account where the snapshot is being shared
3. Verify the operator has authorization for cross-subscription snapshot operations
4. Check if a SAS URI was generated that could allow external download
5. Review the data sensitivity of the affected virtual machine\'s disk
6. If unauthorized, revoke any generated SAS tokens and delete shared snapshots
7. Implement Azure Policy to restrict snapshot sharing across subscriptions
8. Enable diagnostic logging for disk operations
', '["https://learn.microsoft.com/en-us/azure/virtual-machines/disks-incremental-snapshots","https://attack.mitre.org/techniques/T1537/"]', e'(contains("log.operationName", "Microsoft.Compute/snapshots") ||
 contains("log.operationName", "Microsoft.Compute/disks")) &&
(contains("log.operationName", "beginGetAccess") ||
 contains("log.operationName", "export")) &&
equals("log.resultType", "Success")
', '2026-03-02 22:36:52.358497', true, true, 'origin', null, '[]', '["lastEvent.log.operationName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1297, 'Azure AD Device Code Authentication Flow Detected', 3, 2, 1, 'Initial Access', 'T1078 - Valid Accounts', e'Detects OAuth device code flow authentication in Azure AD. Device code phishing is a growing attack vector where attackers trick users into authenticating on a device the attacker controls, granting the attacker access tokens.

Next Steps:
1. Verify the device code authentication was initiated by the user on a legitimate device
2. Check the application requesting the device code for legitimacy
3. Review the source IP where the token was redeemed
4. Check for subsequent suspicious activities using the obtained token
5. If unauthorized, revoke the session and all refresh tokens
6. Consider blocking device code flow via conditional access policies
7. Educate users about device code phishing attacks
', '["https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-device-code","https://attack.mitre.org/techniques/T1078/"]', e'(contains("log.properties", "deviceCode") && contains("log.operationName", "Sign-in")) ||
contains("log.properties", "urn:ietf:params:oauth:grant-type:device_code")
', '2026-03-02 22:36:53.925728', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1298, 'AzureHound Reconnaissance Tool Detected', 2, 1, 0, 'Discovery', 'T1087.004 - Account Discovery: Cloud Account', e'Detects AzureHound user agent in Azure AD sign-in logs. AzureHound is the Azure AD data collector for BloodHound, used to enumerate all users, groups, roles, apps, and relationships in the tenant for attack path analysis.

Next Steps:
1. Identify the user account running AzureHound
2. Determine if this is an authorized security assessment
3. Review the scope of data collected (users, groups, roles, apps)
4. Check for lateral movement or privilege escalation following the enumeration
5. If unauthorized, revoke the user\'s tokens and investigate
6. Review API permissions that allowed the enumeration
7. Consider implementing Graph API rate limiting or monitoring
', '["https://bloodhound.readthedocs.io/en/latest/data-collection/azurehound.html","https://attack.mitre.org/techniques/T1087/004/"]', e'contains("log.properties", "azurehound") ||
contains("log.properties", "AzureHound")
', '2026-03-02 22:36:55.414706', true, true, 'origin', '["adversary.user","adversary.ip"]', '[]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1299, 'Azure AD Privileged App Role Assignment', 3, 3, 2, 'Privilege Escalation', 'T1098.003 - Account Manipulation: Additional Cloud Roles', e'Detects privileged app role assignments to service principals in Azure AD, which is the mechanism used in illicit consent grant attacks. Attackers create or modify applications with high-privilege API permissions to access organizational data.

Next Steps:
1. Review the application and the specific API permissions granted
2. Verify the consent was authorized by a legitimate administrator
3. Check if the application is known and trusted
4. Review the application publisher and redirect URIs
5. Check for data access using the application\'s permissions
6. If unauthorized, remove the role assignment and revoke application consent
7. Review and restrict user consent settings in Azure AD
', '["https://learn.microsoft.com/en-us/entra/identity-platform/app-objects-and-service-principals","https://attack.mitre.org/techniques/T1098/003/"]', e'contains("log.operationName", "Add app role assignment to service principal") ||
(contains("log.operationName", "Consent to application") && contains("log.properties", "AppRoleAssignment"))
', '2026-03-02 22:36:56.907396', true, true, 'origin', null, '[]', '["lastEvent.log.properties.targetResources","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1300, 'Azure AD Application Credential Added', 3, 3, 2, 'Persistence', 'T1098.001 - Account Manipulation: Additional Cloud Credentials', e'Detects when new certificates or client secrets are added to Azure AD application registrations. This is the primary Azure AD persistence technique - attackers add credentials to existing apps to maintain access even after password resets.

Next Steps:
1. Verify the credential addition was authorized by the application owner
2. Identify the application and its permissions (especially Graph API permissions)
3. Check the user identity adding the credential for legitimacy
4. Review the credential type (certificate vs secret) and expiration
5. Check for subsequent sign-ins using the new application credential
6. If unauthorized, remove the credential and rotate all app secrets
7. Review the application\'s API permissions for excessive access
', '["https://learn.microsoft.com/en-us/entra/identity-platform/howto-create-service-principal-portal","https://attack.mitre.org/techniques/T1098/001/"]', e'oneOf("log.operationName", ["Add service principal credentials", "Update application - Certificates and secrets management"]) ||
(contains("log.operationName", "application") && contains("log.properties", "KeyCredentials"))
', '2026-03-02 22:36:58.474218', true, true, 'origin', null, '[]', '["lastEvent.log.properties.targetResources","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1301, 'Azure AD Anomalous Token Detection', 3, 2, 1, 'Credential Access', 'T1528 - Steal Application Access Token', e'Detects Azure Identity Protection alerts for anomalous tokens with unusual lifetime, unfamiliar locations, or other suspicious properties. These indicate potential token theft or manipulation.

Next Steps:
1. Review the token properties that triggered the anomaly detection
2. Check the user\'s recent sign-in activity for suspicious patterns
3. Verify the source IP and device used for the authentication
4. Check for impossible travel or unfamiliar location patterns
5. If compromise is suspected, revoke all refresh tokens for the user
6. Force MFA re-registration if MFA token was compromised
7. Review conditional access policies for token protection gaps
', '["https://learn.microsoft.com/en-us/entra/id-protection/concept-identity-protection-risks","https://attack.mitre.org/techniques/T1528/"]', e'contains("log.operationName", "Anomalous Token") ||
(contains("log.properties", "riskEventType") && contains("log.properties", "anomalousToken"))
', '2026-03-02 22:36:59.892966', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1302, 'Azure Security Alert Suppression Rule Created', 2, 3, 2, 'Defense Evasion', 'T1562 - Impair Defenses', e'Detects creation of alert suppression rules in Azure Defender / Microsoft Defender for Cloud. Attackers create suppression rules to hide security alerts generated by their activities.

Next Steps:
1. Review the suppression rule and what alert types it suppresses
2. Verify the rule creation was part of an authorized security operations workflow
3. Check the user identity for legitimate security team membership
4. Review recent security alerts that may have been suppressed
5. If unauthorized, delete the suppression rule and review suppressed alerts
6. Check for other defense evasion activities from the same user
', '["https://learn.microsoft.com/en-us/azure/defender-for-cloud/alerts-suppression-rules","https://attack.mitre.org/techniques/T1562/"]', e'regexMatch("log.operationName", "(?i)MICROSOFT\\\\.SECURITY/ALERTSSUPPRESSIONRULES/WRITE")
', '2026-03-02 22:37:01.427523', true, true, 'origin', null, '[]', '["lastEvent.log.operationName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1303, 'Azure AD Impossible Travel Sign-In Detection', 3, 2, 1, 'Credential Access', 'T1078 - Valid Accounts', e'Detects Azure AD sign-ins flagged as risky due to impossible travel, anonymous IP usage, or unfamiliar locations. These risk detections indicate potential credential compromise when a user authenticates from geographically impossible locations or through anonymizing services.

Next Steps:
1. Review the sign-in details including IP addresses and geographic locations
2. Check if the user employs VPN services that could explain different locations
3. Verify with the user whether the sign-in attempts are legitimate
4. Review the risk level and risk detail provided by Azure AD Identity Protection
5. Check for MFA challenges and their outcomes during the sign-in
6. If compromised, immediately reset user credentials and revoke active sessions
7. Enable Conditional Access policies requiring MFA for risky sign-ins
8. Review Azure AD sign-in logs for other accounts from the same suspicious IPs
', '["https://learn.microsoft.com/en-us/azure/active-directory/identity-protection/concept-identity-protection-risks","https://attack.mitre.org/techniques/T1078/"]', e'contains("log.operationName", "Sign-in activity") &&
(equals("log.properties.riskLevelDuringSignIn", "high") ||
 equals("log.properties.riskState", "atRisk") ||
 contains("log.properties.riskEventTypes", "impossibleTravel") ||
 contains("log.properties.riskEventTypes", "anonymizedIPAddress"))
', '2026-03-02 22:37:02.995700', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1304, 'Azure Automation Runbook Abuse', 3, 3, 2, 'Execution', 'T1059 - Command and Scripting Interpreter', e'Detects creation or modification of Azure Automation runbooks which can be abused for code execution with managed identity privileges. Attackers may create runbooks to execute arbitrary code, establish persistence, or perform lateral movement using the automation account\'s managed identity.

Next Steps:
1. Review the runbook content for malicious scripts or commands
2. Verify the automation account\'s managed identity permissions
3. Check the user creating or modifying the runbook has authorization
4. Review the runbook schedule for unauthorized execution times
5. Examine the runbook\'s Run As account credentials
6. If unauthorized, disable the runbook and revoke the automation account\'s permissions
7. Review execution history for already-executed malicious runbooks
8. Implement RBAC to restrict automation account management
', '["https://learn.microsoft.com/en-us/azure/automation/automation-runbook-types","https://attack.mitre.org/techniques/T1059/"]', e'contains("log.operationName", "Microsoft.Automation") &&
(contains("log.operationName", "runbooks/write") ||
 contains("log.operationName", "runbooks/publish") ||
 contains("log.operationName", "jobs/write") ||
 contains("log.operationName", "schedules/write")) &&
equals("log.resultType", "Success")
', '2026-03-02 22:37:04.524865', true, true, 'origin', null, '[]', '["lastEvent.log.operationName","adversary.user"]');
