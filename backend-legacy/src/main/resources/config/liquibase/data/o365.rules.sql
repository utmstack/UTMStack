insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1546, 'Office 365 Anti-Phishing Policy Bypass Detected', 3, 3, 1, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', 'Detects potential bypasses or modifications to anti-phishing policies including changes to safe sender lists, domain exclusions, or policy disabling that could allow phishing emails to reach users. This rule identifies successful administrative actions on anti-phishing configurations that could weaken email security defenses.

Next Steps:
1. Review the specific anti-phishing policy changes made by the user
2. Verify if the changes were authorized and documented in change management
3. Check if the user has legitimate administrative privileges for email security policies
4. Examine the timing and context of the changes (e.g., during business hours vs. off-hours)
5. Look for any subsequent phishing emails that may have bypassed detection
6. Consider rolling back unauthorized changes and implementing additional approval workflows
7. Monitor for any unusual email activity following the policy modifications
', '["https://learn.microsoft.com/en-us/defender-office-365/anti-phishing-policies-about","https://attack.mitre.org/techniques/T1562/001/"]', '(contains("action", "AntiPhish") || equals("action", "Set-AntiPhishPolicy") || equals("action", "Remove-AntiPhishPolicy") || equals("action", "New-AntiPhishPolicy") || equals("action", "Disable-AntiPhishRule") || contains("action","SafeSender") || contains("action","BypassedSender")) && equals("actionResult", "Success") && exists("origin.user")', '2026-01-28 22:54:21.386731', true, true, 'origin', null, '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"log.Operation.keyword","operator":"filter_match","value":"AntiPhish"}],"or":null,"within":"now-4h","count":2}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1547, 'Office 365 App Consent Grants Detected', 3, 3, 1, 'Persistence, Privilege Escalation', 'T1098.003 - Account Manipulation: Additional Cloud Credentials', 'Detects when OAuth consent is granted to an application in Office 365. Attackers may use malicious OAuth apps to gain persistent access to user data without requiring credentials. This technique allows attackers to maintain access even after password changes.

Next Steps:
1. Review the application that received consent and verify its legitimacy
2. Check the permissions granted to the application
3. Investigate the user who granted consent for suspicious activity
4. Review application audit logs for any unauthorized data access
5. If malicious, revoke the application consent and remove the app registration
6. Consider implementing application consent policies to prevent unauthorized app installations
', '["https://learn.microsoft.com/en-us/defender-office-365/detect-and-remediate-illicit-consent-grants","https://attack.mitre.org/techniques/T1098/003/"]', 'equals("action", "Consent to application") && equals("actionResult", "Success")', '2026-01-28 22:54:22.331652', true, true, 'origin', '["origin.user","log.appAccessContextClientAppId"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1548, 'Audit Log Tampering Detection', 3, 3, 2, 'Defense Evasion', 'T1070.001 - Indicator Removal: Clear Windows Event Logs', 'Detects attempts to tamper with audit logs by disabling auditing, modifying audit configurations, or clearing audit data. This could indicate an attempt to hide malicious activities.

Next Steps:
1. Investigate the user account that performed the audit configuration changes
2. Review recent activities by this user to identify potential malicious actions
3. Check if the audit configuration was restored after being disabled
4. Correlate with other security events around the same timeframe
5. Verify if this was an authorized administrative action
6. Review any recent privilege escalation or account compromise indicators
', '["https://learn.microsoft.com/en-us/purview/audit-log-enable-disable","https://attack.mitre.org/techniques/T1070/001/"]', '(oneOf("action", ["Set-AdminAuditLogConfig", "Remove-AdminAuditLogConfig", "Disable-OrganizationCustomization", "Set-OrganizationConfig"]) ||
(equals("log.Workload", "Exchange") && contains("log.ObjectId", "AdminAuditLog")) ||
(contains("log.Parameters", "UnifiedAuditLogIngestionEnabled") && contains("log.Parameters", "false"))) &&
exists("origin.user") &&
equals("actionResult", "Succeeded")
', '2026-01-28 22:54:23.212737', true, true, 'origin', null, '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"Set-AdminAuditLogConfig"}],"or":[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"Remove-AdminAuditLogConfig"}],"or":null,"within":"now-24h","count":1}],"within":"now-24h","count":1}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1549, 'Azure AD Integration Suspicious Activity', 3, 3, 2, 'Persistence, Privilege Escalation', 'T1098 - Account Manipulation', 'Detects suspicious Azure Active Directory integration events including multiple failed authentication attempts, unusual role assignments, or bulk user modifications that could indicate an attempted compromise of identity management systems. This rule identifies patterns of authentication failures, privilege escalations, and bulk account modifications that may suggest malicious activity targeting the organization''s identity infrastructure.

Next Steps:
1. Review the specific Azure AD activity details and affected user accounts
2. Analyze the frequency and timing of the detected events for patterns
3. Verify the legitimacy of any role assignments or user modifications
4. Check for concurrent suspicious activities from the same IP address or user
5. Review Azure AD sign-in logs for additional context around failed authentications
6. Validate whether the activities align with known business processes or authorized administrative tasks
7. Consider implementing additional monitoring for the affected accounts
8. If malicious activity is confirmed, immediately review and revoke any unauthorized permissions or access
', '["https://learn.microsoft.com/en-us/purview/audit-log-activities","https://attack.mitre.org/techniques/T1098/"]', 'equals("log.Workload", "AzureActiveDirectory") &&
(
  (equals("action", "UserLoginFailed") && equals("actionResult", "Failed")) ||
  (equals("action", "Add member to role") && equals("actionResult", "Success")) ||
  (equals("action", "Update user") && equals("actionResult", "Success")) ||
  (equals("action", "Delete user") && equals("actionResult", "Success")) ||
  (equals("action", "Add service principal") && equals("actionResult", "Success"))
) && exists("origin.user")
', '2026-01-28 22:54:24.351490', true, true, 'origin', '["origin.user","origin.ip"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"log.Workload.keyword","operator":"filter_term","value":"AzureActiveDirectory"}],"or":null,"within":"now-15m","count":10}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1550, 'Unauthorized Calendar Sharing Modification', 3, 2, 1, 'Collection', 'T1213 - Data from Information Repositories', 'Detects modifications to calendar sharing permissions that could expose sensitive scheduling information to unauthorized users. This rule identifies when users modify calendar delegation or folder permissions on calendar folders, which could indicate unauthorized access attempts or data exposure risks.

Next Steps:
1. Verify if the calendar sharing modification was authorized by the calendar owner
2. Review the specific permissions granted and recipient of the sharing permissions
3. Check if the user performing the action has legitimate business need for calendar access
4. Investigate any unusual patterns of calendar sharing modifications by the same user
5. Review related authentication logs for the user account
6. Consider implementing additional approval workflows for calendar sharing modifications
', '["https://docs.microsoft.com/en-us/microsoft-365/compliance/audit-log-activities","https://attack.mitre.org/techniques/T1213/"]', 'oneOf("action", ["UpdateCalendarDelegation", "AddFolderPermissions", "ModifyFolderPermissions", "RemoveFolderPermissions", "Set-MailboxFolderPermission", "Add-MailboxFolderPermission"]) &&
equals("actionResult", "Success") &&
contains("log.folderPath", "Calendar")
', '2026-01-28 22:54:25.212572', true, true, 'origin', '["lastEvent.origin.user","log.Item_FolderPath"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1551, 'Communication Compliance Alert', 2, 2, 1, 'Discovery', 'Account Discovery', 'Detects communication compliance policy violations including potentially threatening, harassing, or discriminatory language in messages, sensitive information sharing, or regulatory compliance violations in communications.

Next Steps:
1. Review the specific compliance violation details and message content
2. Investigate the user''s recent communication patterns
3. Check if this is part of a pattern of policy violations
4. Escalate to HR or compliance team if inappropriate content is confirmed
5. Consider additional training or disciplinary action based on severity
6. Review and update communication policies if needed
', '["https://learn.microsoft.com/en-us/purview/communication-compliance","https://attack.mitre.org/techniques/T1087/"]', 'equals("action", "CommunicationComplianceAlert") || (equals("log.ComplianceType", "CommunicationCompliance") && equals("actionResult", "PolicyMatch")) || (contains("log.PolicyType", "Communication") && oneOf("log.Severity", ["Medium", "High", "Critical"]))', '2026-01-28 22:54:26.109483', true, true, 'origin', '["origin.user"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1552, 'Suspicious Compliance Alert Activity', 3, 3, 2, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', 'Detects suspicious patterns in compliance-related activities including alert suppression, policy modifications, or audit log tampering that could indicate attempts to evade security monitoring or hide malicious activities in Office 365 Security & Compliance Center.

Next Steps:
1. Review the specific compliance action performed and verify if it was authorized
2. Check if the user has legitimate administrative privileges for compliance operations
3. Analyze the timing and frequency of compliance changes - multiple rapid changes may indicate malicious activity
4. Examine what compliance policies, rules, or alerts were modified and assess the security impact
5. Review audit logs for any related suspicious activities before and after this event
6. Verify if the changes align with documented change management processes
7. Check for any correlation with other security alerts or unusual user behavior
8. If unauthorized, immediately review and restore appropriate compliance settings
9. Consider implementing additional monitoring on compliance configuration changes
10. Escalate to security team if evidence suggests malicious intent or privilege abuse
', '["https://learn.microsoft.com/en-us/purview/audit-log-activities","https://attack.mitre.org/techniques/T1562/001/"]', 'equals("log.Workload", "SecurityComplianceCenter") &&
(
  equals("action", "AlertTriggered") ||
  equals("action", "AlertEntityGenerated") ||
  equals("action", "AlertUpdated") ||
  equals("action", "ComplianceSettingChanged") ||
  equals("action", "Set-ComplianceSecurityFilter") ||
  equals("action", "New-ComplianceSecurityFilter") ||
  equals("action", "Remove-ComplianceSecurityFilter") ||
  equals("action", "Set-AdminAuditLogConfig") ||
  equals("action", "Set-OrganizationConfig") ||
  contains("action", "CompliancePolicy") ||
  contains("action", "ComplianceRule") ||
  contains("action", "ComplianceTag")
) &&
equals("actionResult", "Success")
', '2026-01-28 22:54:26.973857', true, true, 'origin', '["origin.user","action","log.ObjectId"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1553, 'Conditional Access Bypass Detection', 3, 3, 2, 'Defense Evasion, Lateral Movement', 'T1550 - Use Alternate Authentication Material', 'Detects attempts to bypass conditional access policies through various methods including legacy authentication, trusted location manipulation, or session token abuse. This rule identifies when users successfully authenticate despite conditional access policy failures or use legacy authentication methods that may bypass modern security controls.

Next Steps:
1. Verify the legitimacy of the authentication attempt with the user
2. Review conditional access policy configuration and compliance
3. Check if the authentication method used is approved for the user''s role
4. Investigate the source IP address and device details
5. Review recent policy changes that might have caused legitimate bypass
6. Consider implementing stricter conditional access policies for high-risk users
7. Monitor for additional suspicious activities from the same user or IP
', '["https://learn.microsoft.com/en-us/azure/active-directory/conditional-access/overview","https://attack.mitre.org/techniques/T1550/"]', '(oneOf("action", ["UserLoggedIn", "UserLoginFailed"]) &&
 ((oneOf("log.propertiesConditionalAccessStatus", ["failure", "notApplied"]) && equals("actionResult", "Success")) ||
  (oneOf("log.AuthenticationMethod", ["Legacy Authentication", "Basic Authentication"])) ||
  (oneOf("log.ClientAppUsed", ["Exchange ActiveSync", "IMAP4", "POP3", "SMTP Auth"])) ||
  (equals("log.DeviceDetail", "{}") && contains("log.Location", "trusted")) ||
  (equals("log.IsInteractive", "false") && equals("log.appAccessContextClientAppId", "")))) &&
!equals("origin.user", "") &&
!equals("origin.ip", "")
', '2026-01-28 22:54:27.768609', true, true, 'origin', '["origin.user","origin.ip"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1554, 'Data Loss Prevention Policy Violation', 3, 2, 1, 'Collection', 'T1213 - Data from Information Repositories', 'Detects violations of Data Loss Prevention (DLP) policies including attempts to share, access, or exfiltrate sensitive information such as credit card numbers, social security numbers, or confidential business data.

This rule triggers when Office 365 DLP policies detect unauthorized handling of sensitive data across Exchange, SharePoint, OneDrive, Teams, or Security Compliance Center workloads.

**Next Steps:**
1. **Immediate Response:** Review the specific DLP policy violation details and sensitive data types involved
2. **User Investigation:** Verify if the user action was intentional and authorized, check user''s role and data access permissions
3. **Data Assessment:** Determine what sensitive information was involved and potential exposure scope
4. **Policy Review:** Evaluate if DLP policy settings are appropriate or need adjustment
5. **Incident Documentation:** Record the violation details, investigation findings, and remediation actions taken
6. **User Training:** If unintentional, provide additional data handling training to the user
7. **System Monitoring:** Monitor for additional violations from the same user or similar patterns
', '["https://learn.microsoft.com/en-us/purview/dlp-learn-about-dlp","https://attack.mitre.org/techniques/T1213/"]', '(
  equals("action", "DLPRuleMatch") ||
  equals("action", "DlpPolicyMatch") ||
  equals("action", "DLPRuleUndo") ||
  contains("log.PolicyDetails", "DLP") ||
  contains("log.ExceptionInfo", "DLP")
) &&
(
  equals("log.Workload", "Exchange") ||
  equals("log.Workload", "SharePoint") ||
  equals("log.Workload", "OneDrive") ||
  equals("log.Workload", "Teams") ||
  equals("log.Workload", "SecurityComplianceCenter")
) &&
!equals("actionResult", "Failed")
', '2026-01-28 22:54:28.714965', true, true, 'origin', '["lastEvent.origin.user","log.PolicyId","log.SensitiveInfoTypeData"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1555, 'eDiscovery Abuse Detection', 3, 2, 1, 'Collection', 'T1074 - Data Staged', 'Detects potential abuse of eDiscovery features including excessive searches, exports, or unauthorized access to sensitive data through eDiscovery operations. This rule identifies users performing multiple eDiscovery actions within a short timeframe, which could indicate malicious data staging or exfiltration attempts.

Next Steps:
1. Review the user''s authorization and role assignments for eDiscovery operations
2. Examine the scope and content of the searches performed
3. Verify if the searches align with legitimate business requirements or ongoing legal matters
4. Check for unusual patterns in search queries or export activities
5. Investigate if sensitive data was accessed or exported without proper justification
6. Review data classification and sensitivity of accessed content
7. Correlate with other suspicious activities from the same user account
8. Consider implementing additional monitoring for the user if activity appears unauthorized
', '["https://learn.microsoft.com/en-us/purview/ediscovery-search-for-activities-in-the-audit-log","https://attack.mitre.org/techniques/T1074/"]', 'oneOf("action", ["SearchStarted", "SearchExported", "SearchCreated", "CaseAdded", "HoldCreated", "SearchExportDownloaded", "SearchPreviewed", "SearchResultsPurged", "RemoveSearchResultsSentToZoom", "RemoveSearchExported", "RemoveSearchPreviewed", "RemoveSearchResultsPurged", "SearchResultsSentToZoom", "ViewedSearchExported", "ViewedSearchPreviewed"]) &&
!equals("origin.user", "") &&
equals("actionResult", "Succeeded") &&
exists("log.action")
', '2026-01-28 22:54:29.551792', true, true, 'origin', '["origin.user"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"log.action.keyword","operator":"filter_term","value":"{{.log.action}}"}],"or":null,"within":"now-1h","count":10}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1556, 'Exchange Admin Configuration Changes', 3, 3, 2, 'Persistence, Privilege Escalation', 'T1098 - Account Manipulation', 'Detects changes to Exchange administrative configuration that could impact security settings, user permissions, or mail flow policies. This rule monitors for successful execution of critical Exchange administrative cmdlets that can affect organizational security posture.

Next Steps:
1. Review the specific administrative action performed and verify it was authorized
2. Check if the user performing the action has appropriate privileges
3. Examine the timing and frequency of administrative changes
4. Validate configuration changes against change management policies
5. Review any related audit logs for the time period around the change
6. Confirm the source IP and location of the administrative session
7. If unauthorized, immediately review affected configurations and revert if necessary
', '["https://docs.microsoft.com/en-us/exchange/security-and-compliance/exchange-auditing-reports/view-administrator-audit-log","https://attack.mitre.org/techniques/T1098/"]', 'oneOf("action", ["Set-AdminAuditLogConfig", "Set-TransportRule", "Set-MalwareFilterPolicy", "Set-HostedContentFilterPolicy", "Set-DkimSigningConfig", "Set-OrganizationConfig", "Set-RoleGroup", "Add-RoleGroupMember", "Remove-RoleGroupMember", "New-ManagementRoleAssignment", "Remove-ManagementRoleAssignment"]) &&
equals("actionResult", "Succeeded")
', '2026-01-28 22:54:30.428878', true, true, 'origin', '["origin.user","action"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1557, 'Suspicious External Sharing Activity', 3, 2, 1, 'Exfiltration', 'T1567 - Exfiltration Over Web Service', 'Detects unauthorized or suspicious external sharing activities in SharePoint and OneDrive that could indicate data exfiltration attempts or policy violations when sharing sensitive content with external parties.

Next Steps:
1. Review the shared content and assess its sensitivity level
2. Verify if the external sharing was authorized by data owners
3. Check if the recipient has legitimate business need for access
4. Validate against organizational data sharing policies
5. Examine user''s recent activity for other suspicious sharing patterns
6. Consider revoking access if sharing was unauthorized
7. Update DLP policies if needed to prevent future violations
8. Document findings and coordinate with data protection team
', '["https://learn.microsoft.com/en-us/purview/audit-log-activities","https://attack.mitre.org/techniques/T1567/"]', '(equals("log.Workload", "SharePoint") || equals("log.Workload", "OneDrive")) &&
(
  equals("action", "SharingInvitationCreated") ||
  equals("action", "AnonymousLinkCreated") ||
  equals("action", "AnonymousLinkUsed") ||
  equals("action", "SecureLinkCreated") ||
  equals("action", "SharingSet") ||
  equals("action", "CompanyLinkCreated") ||
  equals("action", "AddedToSecureLink")
) &&
equals("actionResult", "Success") &&
(
  equals("log.TargetUserOrGroupType", "Guest") ||
  contains("log.SiteUrl", "external") ||
  contains("log.EventData", "AllowExternalSharing")
)
', '2026-01-28 22:54:31.270262', true, true, 'origin', '["origin.user","log.ObjectId"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1558, 'Abnormal Guest User Invitation Activity', 3, 2, 1, 'Persistence', 'Account Creation', 'Detects unusual spikes in guest user invitations which could indicate an attempt to establish persistence through external accounts or potential data exfiltration preparation by adding unauthorized external collaborators.

Next Steps:
1. Review the user account initiating the invitations for any signs of compromise
2. Verify the legitimacy of the invited external users and their business justification
3. Check if the inviting user has appropriate permissions for guest invitations
4. Examine the invited users'' email domains for suspicious or unexpected organizations
5. Review any subsequent activity by the invited guest users
6. Validate that the invitation frequency aligns with normal business processes
7. Consider implementing additional approval workflows for guest user invitations
', '["https://learn.microsoft.com/en-us/purview/audit-log-activities","https://attack.mitre.org/techniques/T1136/"]', 'equals("log.Workload", "AzureActiveDirectory") &&
(
  equals("action", "Invite external user") ||
  equals("action", "InviteGuest") ||
  equals("action", "Add guest to group") ||
  equals("action", "Guest user invite redeemed")
) &&
equals("actionResult", "Success") && exists("origin.user")
', '2026-01-28 22:54:32.172440', true, true, 'origin', '["origin.user"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"Invite external user"}],"or":null,"within":"now-1h","count":5}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1559, 'Office 365 Mail Flow Rule Modified', 3, 3, 2, 'Defense Evasion', 'T1564.008 - Hide Artifacts: Email Hiding Rules', 'Detects modifications to mail flow rules (transport rules) in Office 365. Attackers may create or modify mail flow rules to redirect, delete, or hide emails, bypassing security controls or exfiltrating data. Mail flow rules can be used to automatically delete emails containsing specific keywords, forward sensitive emails to external addresses, or modify email content.

Next Steps:
1. Review the specific mail flow rule that was modified, including its conditions and actions
2. Verify if the change was authorized and performed by a legitimate administrator
3. Check if the rule involves email forwarding to external domains or deletion of emails
4. Examine recent email traffic to identify any emails that may have been affected by the rule
5. Review mailbox audit logs for the affected user accounts
6. Check for other administrative changes made by the same user around the same time
7. If unauthorized, disable the malicious rule immediately and investigate the compromised account
', '["https://admindroid.com/how-to-audit-transport-rule-changes-report-in-microsoft-365","https://attack.mitre.org/techniques/T1564/008/"]', '(contains("action", "TransportRule") || equals("action", "New-TransportRule") || equals("action", "Set-TransportRule") || equals("action", "Remove-TransportRule") || equals("action", "Enable-TransportRule") || equals("action", "Disable-TransportRule")) && equals("actionResult", "Succeeded")', '2026-01-28 22:54:33.181131', true, true, 'origin', '["origin.user","log.action"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1560, 'Mass Email Deletion Detected', 2, 3, 2, 'Collection', 'T1114 - Email Collection', 'Detects when a user performs mass deletion of emails which could indicate data destruction, covering tracks, or malicious insider activity. Monitors for multiple HardDelete or SoftDelete operations within a short time window.

Next Steps:
1. Verify the legitimacy of the user performing the deletions
2. Check if this aligns with any scheduled maintenance or cleanup activities
3. Investigate the content and importance of deleted emails if possible
4. Review user''s recent access patterns and behavior
5. Check for any concurrent suspicious activities from the same user
6. Validate if the user has appropriate permissions for bulk email operations
7. Consider temporarily restricting the user''s access pending investigation
', '["https://learn.microsoft.com/en-us/purview/audit-mailboxes","https://attack.mitre.org/techniques/T1114/"]', 'oneOf("action", ["HardDelete", "SoftDelete"]) && !equals("origin.user", "") && !equals("origin.ip", "")', '2026-01-28 22:54:34.228904', true, true, 'origin', '["origin.user"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"HardDelete"}],"or":[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"SoftDelete"}],"or":null,"within":"now-15m","count":25}],"within":"now-15m","count":25}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1561, 'Suspicious Mail Forwarding Rule Creation', 3, 2, 1, 'Collection', 'T1114.001 - Email Collection: Local Email Collection', 'Detects creation or modification of inbox rules that forward emails to external recipients, which could indicate data exfiltration attempts. This rule monitors for the creation of new inbox rules or modifications to existing rules that contains forwarding parameters.

Next Steps:
1. Review the specific forwarding parameters in the log.Parameters field to identify the destination email address
2. Verify if the forwarding destination is a legitimate business email or an external/suspicious address
3. Check if the user who created the rule has legitimate business justification for email forwarding
4. Review recent authentication logs for the affected user account for signs of compromise
5. Examine the timing of the rule creation - if created outside business hours or immediately after login, investigate further
6. Check for other suspicious activities by the same user account around the same timeframe
7. If malicious, disable the forwarding rule and reset user credentials
', '["https://docs.microsoft.com/en-us/microsoft-365/compliance/auditing-troubleshooting-scenarios","https://attack.mitre.org/techniques/T1114/001/"]', 'oneOf("action", ["NewInboxRule", "Set-InboxRule", "UpdateInboxRules"]) &&
equals("actionResult", "Succeeded") &&
(contains("log.Parameters", "ForwardTo") || contains("log.Parameters", "ForwardAsAttachmentTo") || contains("log.Parameters", "RedirectTo"))
', '2026-01-28 22:54:35.351689', true, true, 'origin', '["origin.user","origin.ip"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1562, 'Multi-Geo Data Violations', 3, 2, 1, 'Exfiltration', 'T1030 - Data Transfer Size Limits', 'Detects violations of multi-geo data residency policies including unauthorized data movements between regions, cross-geo access violations, or attempts to bypass geo-restrictions. This rule identifies activities such as site geo moves, cross-geo file operations, and data location parameter modifications that may indicate unauthorized data transfer or policy violations.

Next Steps:
1. Review the user''s authorization level for multi-geo operations
2. Verify if the geo movement was approved through proper change management
3. Check data classification and residency requirements for affected content
4. Examine the source and destination locations for compliance violations
5. Review audit logs for related suspicious activities by the same user
6. Validate that the operation aligns with organizational data governance policies
', '["https://learn.microsoft.com/en-us/microsoft-365/enterprise/microsoft-365-multi-geo","https://attack.mitre.org/techniques/T1030/"]', '(oneOf("action", ["SiteGeoMoveScheduled", "SiteGeoMoveCompleted", "SiteGeoMoveCancelled", "AllowedDataLocationAdded", "GeoQuotaAllocated", "MigrationJobCompleted"]) ||
(equals("log.Workload", "OneDrive") && contains("log.ItemName", "cross-geo")) ||
(!equals("log.SourceFileName", "") && !equals("log.DestinationFileName", "") && contains("log.SourceRelativeUrl", "geo") && contains("log.DestinationRelativeUrl", "geo")) ||
(contains("log.Parameters", "DataLocation") || contains("log.Parameters", "PreferredDataLocation"))) &&
!equals("origin.user", "") &&
equals("actionResult", "Succeeded") &&
!equals("origin.ip", "")
', '2026-01-28 22:54:36.478882', true, true, 'origin', '["origin.user","origin.ip"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1563, 'Office 365 OAuth Application Anomalous Activity', 3, 3, 2, 'Credential Access', 'T1528 - Steal Application Access Token', 'Detects anomalous OAuth application activities including suspicious consent patterns, high-privilege permission grants, or rapid consent events from a single user which may indicate compromised accounts or malicious OAuth apps. This rule identifies when users consent to OAuth applications with high-risk permissions such as Mail access, Files access, or User.ReadWrite permissions, followed by multiple consent events from the same IP address.

Next Steps:
- Review the OAuth application details and permissions granted
- Verify if the user intentionally consented to the application
- Check for any suspicious activities from the user account after consent
- Investigate the OAuth application''s reputation and publisher
- Consider revoking application permissions if deemed malicious
- Review other users who may have consented to the same application
- Implement conditional access policies to restrict high-risk OAuth consents
', '["https://office365itpros.com/2023/12/15/oauth-apps-security/","https://attack.mitre.org/techniques/T1528/"]', 'equals("action", "Consent to application") && equals("actionResult", "Success") && (contains("log.Scope", "Mail.") || contains("log.Scope", "Files.") || contains("log.Scope", "User.ReadWrite") || contains("log.Scope", ".All")) && exists("origin.ip")', '2026-01-28 22:54:37.608661', true, true, 'origin', '["origin.ip","log.appAccessContextClientAppId"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"action.keyword","operator":"filter_term","value":"Consent to application"}],"or":null,"within":"now-1h","count":5}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1564, 'OneDrive Mass File Access Detected', 3, 1, 2, 'Collection', 'T1530 - Data from Cloud Storage Object', 'Detects when a user accesses an abnormally high number of files in OneDrive within a short time period (200+ files in 30 minutes), which could indicate automated data collection, reconnaissance, or preparation for data exfiltration. This behavior is often associated with insider threats, compromised accounts, or malicious scripts designed to harvest sensitive data from cloud storage.

Next Steps:
1. Immediately verify the legitimacy of the user account and check for signs of compromise
2. Review the specific files accessed to determine sensitivity and business impact
3. Check for concurrent suspicious activities like bulk downloads or external sharing
4. Examine authentication logs for unusual login patterns or locations
5. Verify if the access pattern aligns with the user''s normal work responsibilities
6. Consider temporarily restricting the user''s OneDrive access pending investigation
7. Review OneDrive sharing permissions and external access configurations
8. Check for any automated tools or scripts that might be accessing the files
9. Correlate with other security events from the same user or IP address
10. Document findings and escalate to incident response team if malicious activity is confirmed
', '["https://o365reports.com/2024/01/30/audit-file-access-in-sharepoint-online-using-powershell/","https://attack.mitre.org/techniques/T1530/"]', 'oneOf("action", ["FileAccessed", "FileAccessedExtended", "FilePreviewed"]) && !equals("origin.user", "") && equals("log.Workload", "OneDrive")', '2026-01-28 22:54:38.734195', true, true, 'origin', '["origin.user","origin.ip"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"FileAccessed"},{"field":"log.Workload.keyword","operator":"filter_term","value":"OneDrive"}],"or":[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"FileAccessedExtended"},{"field":"log.Workload.keyword","operator":"filter_term","value":"OneDrive"}],"or":null,"within":"now-30m","count":67},{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"FilePreviewed"},{"field":"log.Workload.keyword","operator":"filter_term","value":"OneDrive"}],"or":null,"within":"now-30m","count":67}],"within":"now-30m","count":67}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1565, 'Power Apps Data Connector Suspicious Activity', 3, 2, 1, 'Collection', 'T1530 - Data from Cloud Storage Object', 'Detects creation or modification of Power Apps data connectors that could lead to unauthorized data access or exfiltration from corporate data sources. This rule identifies suspicious Power Apps activities including app creation/modification, data connection changes, and data export/import operations that may indicate an attempt to access or exfiltrate sensitive organizational data.

Next Steps:
1. Review the specific Power Apps activity and determine if it was authorized
2. Verify the user''s role and permissions for Power Apps and data connectors
3. Examine the data sources being connected to and assess their sensitivity
4. Check for any recent data exports or unusual data access patterns
5. Review the user''s recent activity for other suspicious behavior
6. Validate the legitimacy of any new apps or data connections created
7. Consider implementing additional monitoring for the affected user account
8. If unauthorized, immediately revoke access and investigate potential data compromise
', '["https://docs.microsoft.com/en-us/power-platform/admin/audit-data-user-activity","https://attack.mitre.org/techniques/T1530/"]', 'oneOf("action", ["CreateApp", "EditApp", "DeleteApp", "ShareApp", "UnshareApp", "CreateDataConnection", "UpdateDataConnection", "DeleteDataConnection", "ExportData", "ImportData"]) &&
equals("actionResult", "Success") &&
equals("log.Workload", "PowerApps")
', '2026-01-28 22:54:39.898236', true, true, 'origin', '["origin.user","origin.ip"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1566, 'Suspicious Power Automate Flow Activity', 3, 3, 2, 'Collection', 'Automated Collection', 'Detects creation or modification of Power Automate flows that could be used for automated data exfiltration or unauthorized process automation. This rule identifies suspicious Power Automate activity including flow creation, modification, sharing, and connection management that may indicate malicious automation attempts.

Next Steps:
1. Investigate the user account performing the Power Automate actions - verify if this is legitimate business activity
2. Review the specific flows created or modified - analyze their triggers, actions, and data access patterns
3. Check if the flows are accessing sensitive data sources (SharePoint, OneDrive, databases, external APIs)
4. Examine flow sharing patterns - verify if flows are being shared with unauthorized users or external accounts
5. Review connection creation/deletion activities - check for connections to suspicious external services
6. Correlate with other Office 365 audit logs to identify broader attack patterns
7. If malicious, immediately disable the flows and revoke user permissions
8. Consider implementing Power Platform DLP policies to prevent future abuse
', '["https://docs.microsoft.com/en-us/power-platform/admin/audit-data-user-activity","https://attack.mitre.org/techniques/T1119/"]', 'oneOf("action", ["CreateFlow", "EditFlow", "DeleteFlow", "EnableFlow", "DisableFlow", "ShareFlow", "UnshareFlow", "CreateConnection", "DeleteConnection"]) &&
equals("actionResult", "Success") && equals("log.Workload", "PowerAutomate")
', '2026-01-28 22:54:41.063492', true, true, 'origin', '["origin.user","action"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1567, 'Office 365 Safe Attachment Policy Violation', 3, 3, 2, 'Initial Access', 'T1566.001 - Phishing: Spearphishing Attachment', 'Detects violations of Safe Attachment policies including malicious files blocked during detonation, policy modifications that reduce protection, or attempts to bypass attachment scanning that could lead to malware delivery. This rule identifies when Safe Attachment policies are modified, disabled, or when malicious attachments are detected and blocked by Office 365 security controls.

Next Steps:
1. Review the specific Safe Attachment policy change or violation that occurred
2. Verify if the policy modification was authorized and legitimate
3. Check if any malicious attachments were successfully blocked or if any bypassed controls
4. Investigate the user account that performed the action for signs of compromise
5. Review email logs for any related suspicious email activity
6. Validate current Safe Attachment policy configurations are properly restrictive
7. Consider implementing additional email security controls if gaps are identified
', '["https://learn.microsoft.com/en-us/defender-office-365/safe-attachments-about","https://attack.mitre.org/techniques/T1566/001/"]', '(contains("action", "SafeAttachment") || equals("action", "Set-SafeAttachmentPolicy") || equals("action", "Remove-SafeAttachmentPolicy") || equals("action", "Disable-SafeAttachmentRule") || contains("action", "MalwareDetected") || contains("action", "AttachmentBlocked") || contains("action", "DetonationBlock")) && equals("actionResult", "Succeeded")', '2026-01-28 22:54:42.191549', true, true, 'origin', '["origin.user","category"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1568, 'Safe Links Click Pattern Anomaly', 2, 3, 1, 'Execution', 'T1204.001 - User Execution: Malicious Link', 'Detects unusual patterns in Safe Links click behavior that may indicate phishing attempts or malicious URL access. This rule monitors for multiple clicks on suspicious URLs from the same user within a short timeframe, which could indicate either a persistent phishing campaign or user behavior that poses security risks.

Next Steps:
1. Review the user''s activity timeline to identify the source of the malicious links
2. Check email security logs for the original messages containsing the blocked URLs
3. Verify if other users received similar phishing emails
4. Assess whether the user''s account has been compromised
5. Implement additional security awareness training for the affected user
6. Consider blocking the malicious domains at the network level
', '["https://learn.microsoft.com/en-us/defender-office-365/safe-links-about","https://attack.mitre.org/techniques/T1204/001/"]', 'equals("action", "ClickedSafeLink") && equals("actionResult", "Blocked") && !equals("origin.user", "")', '2026-01-28 22:54:43.272229', true, true, 'origin', '["origin.user"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"ClickedSafeLink"}],"or":null,"within":"now-30m","count":5}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1569, 'SharePoint Mass File Download Detected', 3, 1, 2, 'Collection', 'T1213 - Data from Information Repositories', 'Detects when a user downloads an unusually large number of files from SharePoint or OneDrive within a short time period, which could indicate data exfiltration or insider threat activity. This rule triggers when a user downloads 100 or more files within 30 minutes.

**Next Steps:**
1. Review the user''s recent activity patterns and determine if the download volume is consistent with their role and responsibilities
2. Check if the user has legitimate business justification for downloading large amounts of data
3. Examine the types of files downloaded and their sensitivity levels
4. Verify the user''s location and device used for the downloads
5. Look for any concurrent suspicious activities such as unusual login times or access from new locations
6. Contact the user''s manager to validate the business need for the file downloads
7. Consider implementing additional monitoring or restrictions if the activity appears suspicious
8. Review data loss prevention (DLP) policies and ensure they are properly configured
', '["https://www.sharepointdiary.com/2020/10/how-to-track-document-downloads-using-audit-log-in-sharepoint-online.html","https://attack.mitre.org/techniques/T1213/"]', 'equals("action", "FileDownloaded") && !equals("origin.user", "") && oneOf("log.Workload", ["SharePoint", "OneDrive"])', '2026-01-28 22:54:44.403294', true, true, 'origin', '["origin.user","origin.ip"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"action.keyword","operator":"filter_term","value":"FileDownloaded"}],"or":null,"within":"now-30m","count":100}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1570, 'Suspicious Email Forwarding Rule Created', 3, 2, 1, 'Collection', 'T1114.003 - Email Collection: Email Forwarding Rule', 'Detects creation or modification of inbox rules that forward emails to external domains, which is a common technique used by attackers to exfiltrate emails and maintain persistence after compromising an account. Attackers often create these rules to automatically forward sensitive emails to external addresses under their control.

Next Steps:
1. Verify if the user creating the rule is legitimate and authorized
2. Check the destination email address in the forwarding rule for external domains
3. Review recent authentication logs for the affected user account
4. Examine other administrative actions performed by this user recently
5. Check for any other suspicious inbox rules created by the same user
6. Validate if the forwarding address belongs to a legitimate business contact
7. Consider temporarily disabling the forwarding rule pending investigation
8. Review email logs to see what emails may have already been forwarded
', '["https://redcanary.com/blog/threat-detection/email-forwarding-rules/","https://attack.mitre.org/techniques/T1114/003/"]', 'oneOf("action", ["New-InboxRule", "Set-InboxRule"]) && !equals("origin.user", "") && (contains("log.Parameters", "ForwardTo") || contains("log.Parameters", "RedirectTo") || contains("log.Parameters", "ForwardAsAttachmentTo"))', '2026-01-28 22:54:45.534003', true, true, 'origin', '["origin.user"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1571, 'Suspicious Teams Message Export Activity', 3, 2, 1, 'Collection', 'T1114 - Email Collection', 'Detects when Teams messages are exported or accessed in bulk through API calls, which could indicate an attempt to exfiltrate chat history and shared files from Microsoft Teams.

Next Steps:
1. Review the user''s recent Teams activity and authentication events
2. Check if the export activity was authorized or part of legitimate business operations
3. Examine the volume and scope of data accessed or exported
4. Verify the user''s current permissions and access levels
5. Look for any concurrent suspicious activities from the same user account
6. Contact the user to confirm if they initiated these export operations
7. Review any third-party applications that may have access to Teams data
', '["https://learn.microsoft.com/en-us/purview/audit-teams-audit-log-events","https://attack.mitre.org/techniques/T1114/"]', 'oneOf("action", ["MessagesListed", "MessagesExported", "RecordingExported", "TranscriptsExported"]) && !equals("origin.user", "") && equals("log.Workload", "MicrosoftTeams")', '2026-01-28 22:54:46.573120', true, true, 'origin', '["origin.user","log.appAccessContextClientAppId"]', '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{.origin.user}}"},{"field":"log.Workload.keyword","operator":"filter_term","value":"MicrosoftTeams"}],"or":null,"within":"now-1h","count":20}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1572, 'Threat Intelligence Alert Pattern', 3, 3, 2, 'Initial Access', 'Phishing', 'Detects threat intelligence alerts from Office 365 security services indicating known malicious activities, threat actor patterns, or indicators of compromise matching threat intelligence feeds. These alerts are generated when Microsoft''s threat intelligence systems identify suspicious activities, malicious file attachments, URLs, or communication patterns associated with known threat actors.

Next Steps:
1. Immediately review the alert details and associated threat intelligence indicators
2. Investigate the affected user account for any compromise indicators
3. Check email activity and file access patterns for the affected user
4. Verify if any malicious files were opened or downloaded
5. Review authentication logs for unusual sign-in patterns
6. Examine network connections and data transfer activities
7. Check for lateral movement attempts within the organization
8. Implement containsment measures if compromise is confirmed
9. Update security policies and user training based on attack vectors
10. Report findings to incident response team and document lessons learned
', '["https://learn.microsoft.com/en-us/microsoft-365/security/defender/threat-analytics","https://attack.mitre.org/techniques/T1566/"]', 'equals("action", "ThreatIntelligenceAlertTriggered") || (equals("log.AlertType", "ThreatIntelligence") && !equals("actionResult", "Success"))', '2026-01-28 22:54:47.664261', true, true, 'origin', '["origin.user","target.user"]', '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1573, 'Microsoft 365 New Inbox Rule Created', 3, 2, 1, 'Email Collection', 'Collection', 'Credential Access consists of techniques for stealing credentials like account names and passwords. Techniques used to get credentials include keylogging or credential dumping. Using legitimate credentials can give adversaries access to systems, make them harder to detect, and provide the opportunity to create more accounts to help achieve their goals.<br> Identifies when a new Inbox rule is created in Microsoft 365. Inbox rules process messages in the Inbox based on conditions and take actions, such as moving a message to a specified folder or deleting a message. Adequate permissions are required on the mailbox to create an Inbox rule.', '["https://docs.microsoft.com/en-us/microsoft-365/security/office-365-security/responding-to-a-compromised-email-account?view=o365-worldwide","https://docs.microsoft.com/en-us/powershell/module/exchange/new-inboxrule?view=exchange-ps","https://docs.microsoft.com/en-us/microsoft-365/security/office-365-security/detect-and-remediate-outlook-rules-forms-attack?view=o365-worldwide","https://attack.mitre.org/techniques/T1114/","https://attack.mitre.org/techniques/T1114/003/","https://attack.mitre.org/tactics/TA0009/"]', 'equals("log.workLoad", "Exchange") && equals("action", "New-InboxRule") && oneOf("actionResult", ["Success","Succeeded","PartiallySucceeded","True"])', '2026-01-28 22:56:09.186372', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1574, 'Attempts to Brute Force a Microsoft 365 User Account', 3, 3, 3, 'Brute Force', 'Credential Access', 'Credential Access consists of techniques for stealing credentials like account names and passwords. Techniques used to get credentials include keylogging or credential dumping. Using legitimate credentials can give adversaries access to systems, make them harder to detect, and provide the opportunity to create more accounts to help achieve their goals.<br> Identifies attempts to brute force a Microsoft 365 user account. An adversary may attempt a brute force attack to obtain unauthorized access to user accounts.', '["https://attack.mitre.org/techniques/T1110/","https://attack.mitre.org/tactics/TA0006/"]', 'oneOf("log.workLoad", ["Exchange", "AzureActiveDirectory"]) && oneOf("action", ["UserLoginFailed", "PasswordLogonInitialAuthUsingPassword"]) && oneOf("actionResult", ["Failed", "False"]) && exists("origin.user")', '2026-01-28 22:56:10.186443', true, true, 'origin', null, '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.user.keyword","operator":"filter_term","value":"{{origin.user}}"}],"or":null,"within":"now-60s","count":5}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1575, 'Potential Password Spraying of Microsoft 365 User Accounts', 3, 2, 2, 'Brute Force', 'Credential Access', 'Credential Access consists of techniques for stealing credentials like account names and passwords. Techniques used to get credentials include keylogging or credential dumping. Using legitimate credentials can give adversaries access to systems, make them harder to detect, and provide the opportunity to create more accounts to help achieve their goals.<br> Identifies a high number (25) of failed Microsoft 365 user authentication attempts from a single IP address within 30 minutes, which could be indicative of a password spraying attack. An adversary may attempt a password spraying attack to obtain unauthorized access to user accounts.', '["https://attack.mitre.org/techniques/T1110/","https://attack.mitre.org/tactics/TA0006/"]', 'oneOf("log.workLoad", ["Exchange", "AzureActiveDirectory"]) && oneOf("action", ["UserLoginFailed", "PasswordLogonInitialAuthUsingPassword"]) && oneOf("actionResult", ["Failed", "False"]) && exists("origin.ip")', '2026-01-28 22:56:11.274900', true, true, 'origin', null, '[{"indexPattern":"v11-log-o365-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{origin.ip}}"}],"or":null,"within":"now-60s","count":5}]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1576, 'O365 Excessive Single Sign-On Logon Errors', 2, 3, 2, 'Brute Force', 'Credential Access', 'Credential Access consists of techniques for stealing credentials like account names and passwords. Techniques used to get credentials include keylogging or credential dumping. Using legitimate credentials can give adversaries access to systems, make them harder to detect, and provide the opportunity to create more accounts to help achieve their goals.<br> Identifies accounts with a high number of single sign-on (SSO) logon errors. Excessive logon errors may indicate an attempt to brute force a password or SSO token.', '["https://attack.mitre.org/techniques/T1110/","https://attack.mitre.org/tactics/TA0006/"]', 'equals("log.workLoad", "AzureActiveDirectory") && equals("log.LogonError", "SsoArtifactInvalidOrExpired")', '2026-01-28 22:56:12.281662', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1577, 'Microsoft 365 Exchange DLP Policy Removed', 3, 3, 2, 'Impair Defenses', 'Defense Evasion', 'Defense Evasion consists of techniques that adversaries use to avoid detection throughout their compromise. Techniques used for defense evasion include uninstalling/disabling security software or obfuscating/encrypting data and scripts. Adversaries also leverage and abuse trusted processes to hide and masquerade their malware. Other tactics’ techniques are cross-listed here when those techniques include the added benefit of subverting defenses.<br> Identifies when a Data Loss Prevention (DLP) policy is removed in Microsoft 365. An adversary may remove a DLP policy to evade existing DLP monitoring.', '["https://docs.microsoft.com/en-us/microsoft-365/compliance/data-loss-prevention-policies?view=o365-worldwide","https://attack.mitre.org/techniques/T1562/","https://attack.mitre.org/tactics/TA0005/"]', 'equals("log.workLoad", "Exchange") && equals("action", "Remove-DlpPolicy") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:13.342286', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1578, 'Microsoft 365 Exchange Malware Filter Policy Deletion', 3, 3, 2, 'Impair Defenses', 'Defense Evasion', 'Defense Evasion consists of techniques that adversaries use to avoid detection throughout their compromise. Techniques used for defense evasion include uninstalling/disabling security software or obfuscating/encrypting data and scripts. Adversaries also leverage and abuse trusted processes to hide and masquerade their malware. Other tactics’ techniques are cross-listed here when those techniques include the added benefit of subverting defenses.<br> Identifies when a malware filter policy has been deleted in Microsoft 365. A malware filter policy is used to alert administrators that an internal user sent a message that containsed malware. This may indicate an account or machine compromise that would need to be investigated. Deletion of a malware filter policy may be done to evade detection.', '["https://attack.mitre.org/techniques/T1562/","https://attack.mitre.org/tactics/TA0005/"]', 'equals("log.workLoad", "Exchange") && equals("action", "Remove-MalwareFilterPolicy") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:14.461010', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1579, 'Microsoft 365 Exchange Safe Attachment Rule Disabled', 3, 3, 2, 'Impair Defenses', 'Defense Evasion', 'Defense Evasion consists of techniques that adversaries use to avoid detection throughout their compromise. Techniques used for defense evasion include uninstalling/disabling security software or obfuscating/encrypting data and scripts. Adversaries also leverage and abuse trusted processes to hide and masquerade their malware. Other tactics’ techniques are cross-listed here when those techniques include the added benefit of subverting defenses.<br> Identifies when a safe attachment rule is disabled in Microsoft 365. Safe attachment rules can extend malware protections to include routing all messages and attachments without a known malware signature to a special hypervisor environment. An adversary or insider threat may disable a safe attachment rule to exfiltrate data or evade defenses.', '["https://attack.mitre.org/techniques/T1562/","https://attack.mitre.org/tactics/TA0005/"]', 'equals("log.workLoad", "Exchange") && equals("action", "Disable-SafeAttachmentRule") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:15.632685', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1580, 'Microsoft 365 Detection of a connection to a .onion domain', 3, 2, 2, 'Software', 'Command and Control', 'Tor is a software suite and network that provides increased anonymity on the Internet.  It creates a multi-hop proxy network and utilizes multilayer encryption to  protect both the message and routing information. Tor utilizes Onion Routing,  in which messages are encrypted with multiple layers of encryption; at each step in the proxy network,  the topmost layer is decrypted and the contents forwarded on to the next node until it reaches its destination.', '["https://attack.mitre.org/software/S0183/","https://attack.mitre.org/software/"]', 'contains("log.siteUrl", ".onion")', '2026-01-28 22:56:16.800635', true, true, 'target', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1581, 'Microsoft 365 Exchange Malware Filter Rule Modification', 3, 3, 2, 'Impair Defenses', 'Defense Evasion', 'Defense Evasion consists of techniques that adversaries use to avoid detection throughout their compromise. Techniques used for defense evasion include uninstalling/disabling security software or obfuscating/encrypting data and scripts. Adversaries also leverage and abuse trusted processes to hide and masquerade their malware. Other tactics’ techniques are cross-listed here when those techniques include the added benefit of subverting defenses.<br> Identifies when a malware filter rule has been deleted or disabled in Microsoft 365. An adversary or insider threat may want to modify a malware filter rule to evade detection.', '["https://attack.mitre.org/techniques/T1562/","https://attack.mitre.org/tactics/TA0005/"]', 'equals("log.workLoad", "Exchange") && oneOf("action", ["Remove-MalwareFilterRule","Disable-MalwareFilterRule"]) && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:17.977905', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1582, 'Microsoft 365 Exchange Transport Rule Creation', 3, 2, 1, 'Transfer Data to Cloud Account', 'Exfiltration', 'Exfiltration consists of techniques that adversaries may use to steal data from your network. Once they’ve collected data, adversaries often package it to avoid detection while removing it. This can include compression and encryption. Techniques for getting data out of a target network typically include transferring it over their command and control channel or an alternate channel and may also include putting size limits on the transmission.<br> Identifies a transport rule creation in Microsoft 365. Exchange Online mail transport rules should be set to not forward email to domains outside of your organization as a best practice. An adversary may create transport rules to exfiltrate data.', '["https://docs.microsoft.com/en-us/exchange/security-and-compliance/mail-flow-rules/mail-flow-rules","https://attack.mitre.org/techniques/T1537/","https://attack.mitre.org/tactics/TA0010/"]', 'equals("log.workLoad", "Exchange") && equals("action", "New-TransportRule") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:19.145972', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1583, 'Microsoft 365 Exchange Transport Rule Modification', 3, 2, 1, 'Transfer Data to Cloud Account', 'Exfiltration', 'Exfiltration consists of techniques that adversaries may use to steal data from your network. Once they’ve collected data, adversaries often package it to avoid detection while removing it. This can include compression and encryption. Techniques for getting data out of a target network typically include transferring it over their command and control channel or an alternate channel and may also include putting size limits on the transmission.<br> Identifies when a transport rule has been disabled or deleted in Microsoft 365. Mail flow rules (also known as transport rules) are used to identify and take action on messages that flow through your organization. An adversary or insider threat may modify a transport rule to exfiltrate data or evade defenses.', '["https://docs.microsoft.com/en-us/exchange/security-and-compliance/mail-flow-rules/mail-flow-rules","https://attack.mitre.org/techniques/T1537/","https://attack.mitre.org/tactics/TA0010/"]', 'equals("log.Workload", "Exchange") && oneOf("action", ["Remove-TransportRule","Disable-TransportRule"]) && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:20.319715', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1584, 'Microsoft 365 Exchange Anti-Phish Policy Deletion', 2, 3, 2, 'Phishing', 'Initial Access', 'Adversaries may send phishing messages to gain access to victim systems. All forms of phishing are electronically delivered social engineering. Phishing can be targeted, known as spearphishing. In spearphishing, a specific individual, company, or industry will be targeted by the adversary. More generally, adversaries can conduct non-targeted phishing, such as in mass malware spam campaigns.<br> Adversaries may send victims emails containsing malicious attachments or links, typically to execute malicious code on victim systems. Phishing may also be conducted via third-party services, like social media platforms. Phishing may also involve social engineering techniques, such as posing as a trusted source.<br> Identifies the deletion of an anti-phishing policy in Microsoft 365. By default, Microsoft 365 includes built-in features that help protect users from phishing attacks. Anti-phishing polices increase this protection by refining settings to better detect and prevent attacks.', '["https://docs.microsoft.com/en-us/microsoft-365/security/office-365-security/set-up-anti-phishing-policies?view=o365-worldwide","https://attack.mitre.org/techniques/T1566/","https://attack.mitre.org/tactics/TA0001/"]', 'equals("log.Workload", "Exchange") && equals("action", "Remove-AntiPhishPolicy") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:21.363846', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1585, 'Microsoft 365 Exchange Anti-Phish Rule Modification', 2, 3, 2, 'Phishing', 'Initial Access', 'Adversaries may send phishing messages to gain access to victim systems. All forms of phishing are electronically delivered social engineering. Phishing can be targeted, known as spearphishing. In spearphishing, a specific individual, company, or industry will be targeted by the adversary. More generally, adversaries can conduct non-targeted phishing, such as in mass malware spam campaigns.<br> Adversaries may send victims emails containsing malicious attachments or links, typically to execute malicious code on victim systems. Phishing may also be conducted via third-party services, like social media platforms. Phishing may also involve social engineering techniques, such as posing as a trusted source.<br> Identifies the modification of an anti-phishing rule in Microsoft 365. By default, Microsoft 365 includes built-in features that help protect users from phishing attacks. Anti-phishing rules increase this protection by refining settings to better detect and prevent attacks.', '["https://attack.mitre.org/techniques/T1566/","https://attack.mitre.org/tactics/TA0001/"]', 'equals("log.Workload", "Exchange") && oneOf("action", ["Remove-AntiPhishRule","Disable-AntiPhishRule"]) && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:22.461878', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1586, 'Microsoft 365 Exchange Safe Link Policy Disabled', 3, 2, 2, 'Phishing', 'Initial Access', 'When it comes to protecting its users, Microsoft takes the threat of phishing seriously. Spoofing is a common technique that''s  used by attackers. Spoofed messages appear to originate from someone or somewhere other than the actual source. This technique is often  used in phishing campaigns that are designed to obtain user credentials. The anti-spoofing technology in EOP specifically examines forgery of the From header in the message body (used to display the message sender in email clients). When EOP has high confidence that the From header is forged, the message is identified as spoofed.<br> Identifies when a Safe Link policy is disabled in Microsoft 365. Safe Link policies for Office applications extend phishing protection to documents that contains hyperlinks, even after they have been delivered to a user.', '["https://docs.microsoft.com/en-us/microsoft-365/security/office-365-security/atp-safe-links?view=o365-worldwide","https://attack.mitre.org/techniques/T1566/"]', 'equals("log.Workload", "Exchange") && equals("action", "Disable-SafeLinksRule") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:23.594850', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1587, 'Microsoft 365 Teams Custom Application Interaction Allowed', 2, 2, 2, 'Teams Configuration Management', 'Maintain access', 'Identifies when custom applications are allowed in Microsoft Teams. If an organization requires applications other than those available in the Teams app store, custom applications can be developed as packages and uploaded. An adversary may abuse this behavior to establish persistence in an environment.', '["https://docs.microsoft.com/en-us/microsoftteams/platform/concepts/deploy-and-publish/apps-upload"]', 'equals("log.Workload", "Exchange") && equals("action", "TeamsTenantSettingChanged") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:24.760455', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1588, 'Microsoft 365 Exchange DKIM Signing Configuration Disabled', 3, 3, 1, 'Phishing', 'Spoofing', 'When it comes to protecting its users, Microsoft takes the threat of phishing seriously. Spoofing is a common technique that''s  used by attackers. Spoofed messages appear to originate from someone or somewhere other than the actual source. This technique is often  used in phishing campaigns that are designed to obtain user credentials. The anti-spoofing technology in EOP specifically examines forgery of the From header in the message body (used to display the message sender in email clients). When EOP has high confidence that the From header is forged, the message is identified as spoofed.<br> Identifies when a DomainKeys Identified Mail (DKIM) signing configuration is disabled in Microsoft 365. With DKIM in Microsoft 365, messages that are sent from Exchange Online will be cryptographically signed. This will allow the receiving email system to validate that the messages were generated by a server that the organization authorized and not being spoofed.', '["https://docs.microsoft.com/en-us/microsoft-365/security/office-365-security/anti-spoofing-protection?view=o365-worldwide"]', 'equals("log.Workload", "Exchange") && equals("action", "Set-DkimSigningConfig") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:25.849264', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1589, 'Microsoft 365 Exchange Management Group Role Assignment', 2, 3, 2, 'Account Manipulation', 'Persistence', 'Adversaries may manipulate accounts to maintain access to victim systems. Account manipulation may consist of any action that preserves adversary access to a compromised account, such as modifying credentials or permission groups. These actions could also include account activity designed to subvert security policies,  such as performing iterative password updates to bypass password duration policies and preserve the life of compromised credentials. In order to create or  manipulate accounts, the adversary must already have sufficient permissions on systems or the domain.<br> Identifies when a new role is assigned to a management group in Microsoft 365. An adversary may attempt to add a role in order to maintain persistence in an environment.', '["https://docs.microsoft.com/en-us/microsoft-365/admin/add-users/about-admin-roles?view=o365-worldwide","https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1098"]', 'equals("log.Workload", "Exchange") && equals("action", "New-ManagementRoleAssignment") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:27.027346', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1590, 'Microsoft 365 Teams External Access Enabled', 3, 2, 2, 'Account Manipulation', 'Persistence', 'Adversaries may manipulate accounts to maintain access to victim systems. Account manipulation may consist of any action that preserves adversary access to a compromised account, such as modifying credentials or permission groups. These actions could also include account activity designed to subvert security policies,  such as performing iterative password updates to bypass password duration policies and preserve the life of compromised credentials. In order to create or  manipulate accounts, the adversary must already have sufficient permissions on systems or the domain.<br> Identifies when external access is enabled in Microsoft Teams. External access lets Teams and Skype for Business users communicate with other users that are outside their organization. An adversary may enable external access or add an allowed domain to exfiltrate data or maintain persistence in an environment.', '["https://docs.microsoft.com/en-us/microsoftteams/manage-external-access","https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1098"]', 'oneOf("log.Workload", ["SkypeForBusiness","MicrosoftTeams"]) && equals("action", "Set-CsTenantFederationConfiguration") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:28.190516', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1591, 'Microsoft 365 Teams Guest Access Enabled', 3, 2, 2, 'Account Manipulation', 'Persistence', 'Adversaries may manipulate accounts to maintain access to victim systems. Account manipulation may consist of any action that preserves adversary access to a  compromised account, such as modifying credentials or permission groups. These actions could also include account activity designed to subvert security policies,  such as performing iterative password updates to bypass password duration policies and preserve the life of compromised credentials. In order to create or  manipulate accounts, the adversary must already have sufficient permissions on systems or the domain.<br> Identifies when guest access is enabled in Microsoft Teams. Guest access in Teams allows people outside the organization to access teams and channels. An adversary may enable guest access to maintain persistence in an environment. <br> The Microsoft 365 Fleet integration, Filebeat module, or similarly structured data is required to be compatible with this rule.', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1098"]', 'oneOf("log.Workload", ["SkypeForBusiness","MicrosoftTeams"]) && equals("action", "New-ManagementRoleAssignment") && oneOf("actionResult", ["Success","PartiallySucceeded","True"])', '2026-01-28 22:56:29.355607', true, true, 'origin', null, '[]', null);
insert into public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) values (1592, 'Microsoft 365 Possible Successful Password Guessing detected', 3, 2, 2, 'Credential Access', 'Brute Force: Password Guessing', 'Adversaries with no prior knowledge of legitimate credentials within the system or environment  may guess passwords to attempt access to accounts. Without knowledge of the password for an account,  an adversary may opt to systematically guess the password using a repetitive or iterative mechanism.  An adversary may guess login credentials without prior knowledge of system or environment passwords  during an operation by using a list of common passwords. Password guessing may or may not take into  account the target''s policies on password complexity or use policies that may lock accounts out after  a number of failed attempts.', '["https://attack.mitre.org/tactics/TA0006","https://attack.mitre.org/techniques/T1110/001/"]', 'oneOf("log.Workload", ["Exchange","AzureActiveDirectory"]) && equals("log.Operation", "UserLoginFailed") && oneOf("log.ResultStatus", ["Failed","False"]) && exists("log.UserId") && exists("log.ClientIp")', '2026-01-28 22:56:30.452230', true, true, 'origin', null, '[{"indexPattern":"v11-log-o365-*","with":[{"field":"log.Workload.keyword","operator":"filter_term","value":"Exchange, AzureActiveDirectory"},{"field":"log.Operation.keyword","operator":"filter_term","value":"UserLoginFailed"},{"field":"log.ResultStatus.keyword","operator":"filter_term","value":"Failed,False"},{"field":"log.UserId.keyword","operator":"filter_term","value":"{{.log.UserId}}"},{"field":"log.ClientIp.keyword","operator":"filter_term","value":"{{.log.ClientIp}}"}],"or":null,"within":"now-1m","count":10}]', null);


insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1546, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1547, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1548, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1549, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1550, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1551, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1552, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1553, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1554, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1555, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1556, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1557, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1558, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1559, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1560, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1561, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1562, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1563, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1564, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1565, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1566, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1567, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1568, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1569, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1570, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1571, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1572, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1573, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1574, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1575, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1576, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1577, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1578, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1579, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1580, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1581, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1582, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1583, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1584, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1585, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1586, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1587, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1588, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1589, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1590, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1591, 4, null);
insert into public.utm_group_rules_data_type (rule_id, data_type_id, last_update) values (1592, 4, null);
