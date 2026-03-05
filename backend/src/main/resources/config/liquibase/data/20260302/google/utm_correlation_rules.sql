INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1305, 'Google Cloud Service Account Key Creation Spike', 3, 3, 2, 'Credential Access', 'Account Manipulation', e'Detects spikes in service account key creation which could indicate credential harvesting or preparation for unauthorized access. Service account keys provide long-term credentials that can be used to authenticate as the service account. Multiple key creations by the same user within a short timeframe may indicate malicious activity or preparation for privilege escalation attacks.

Next Steps:
1. Investigate the user account creating multiple service account keys
2. Review the service accounts for which keys were created and their permissions
3. Check if the key creation was authorized and follows organizational policies
4. Examine subsequent activities performed using these service account credentials
5. Verify if the keys were created from expected IP addresses and locations
6. Review access patterns and identify any unusual resource access or API calls
7. Consider rotating or disabling the created keys if unauthorized activity is confirmed
', '["https://cloud.google.com/iam/docs/audit-logging/examples-service-accounts","https://attack.mitre.org/techniques/T1098/001/"]', e'equals("log.protoPayload.methodName", "google.iam.admin.v1.CreateServiceAccountKey") &&
equals("log.protoPayload.serviceName", "iam.googleapis.com")
', '2026-03-02 22:50:02.585329', true, true, 'origin', null, '[{"indexPattern":"v11-log-google-*","with":[{"field":"log.protoPayload.authenticationInfo.principalEmail","operator":"filter_term","value":"{{.log.protoPayload.authenticationInfo.principalEmail}}"}],"or":null,"within":"now-1h","count":5}]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","lastEvent.log.protoPayload.methodName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1306, 'GCP Secret Manager Bulk Access Detection', 3, 1, 1, 'Credential Access', 'T1552 - Unsecured Credentials', e'Detects bulk access to GCP Secret Manager secrets which may indicate credential harvesting. Attackers who gain access to a GCP project may enumerate and retrieve all stored secrets to obtain API keys, database credentials, and other sensitive data.

Next Steps:
1. Review which secrets were accessed and their sensitivity classification
2. Verify the identity accessing the secrets has legitimate need
3. Check the access pattern for unusual timing or volume
4. Review the caller\'s IP address and user agent for anomalies
5. Determine if the accessed secrets have been used from unauthorized locations
6. If unauthorized, rotate all accessed secrets immediately
7. Review Secret Manager IAM bindings and apply least privilege
8. Enable VPC Service Controls to restrict secret access
', '["https://cloud.google.com/secret-manager/docs/audit-logging","https://attack.mitre.org/techniques/T1552/"]', e'contains("log.protoPayload.serviceName", "secretmanager.googleapis.com") &&
contains("log.protoPayload.methodName", "AccessSecretVersion")
', '2026-03-02 22:50:03.849969', true, true, 'origin', null, '[{"indexPattern":"v11-log-google-*","with":[{"field":"log.protoPayload.authenticationInfo.principalEmail","operator":"filter_term","value":"{{.log.protoPayload.authenticationInfo.principalEmail}}"},{"field":"log.protoPayload.methodName","operator":"filter_term","value":"AccessSecretVersion"}],"or":null,"within":"now-15m","count":5}]', '["lastEvent.log.protoPayload.methodName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1307, 'GCP Service Account Impersonation Detection', 3, 3, 1, 'Credential Access', 'T1550.001 - Use Alternate Authentication Material: Application Access Token', e'Detects service account impersonation through token generation APIs including GenerateAccessToken, GenerateIdToken, and SignBlob. Attackers may impersonate service accounts to escalate privileges or access resources the service account has been granted.

Next Steps:
1. Verify the identity performing the impersonation is authorized
2. Check the target service account and its IAM bindings
3. Review the permissions available through the impersonated service account
4. Examine the API calls made using the generated token
5. Verify if the impersonation is part of a legitimate workload chain
6. If unauthorized, remove the iam.serviceAccountTokenCreator role from the caller
7. Review the service account\'s access patterns for anomalies
8. Implement Organization Policy constraints to limit service account impersonation
', '["https://cloud.google.com/iam/docs/create-short-lived-credentials-direct","https://attack.mitre.org/techniques/T1550/001/"]', e'(contains("log.protoPayload.methodName", "GenerateAccessToken") ||
 contains("log.protoPayload.methodName", "GenerateIdToken") ||
 contains("log.protoPayload.methodName", "SignBlob") ||
 contains("log.protoPayload.methodName", "SignJwt")) &&
contains("log.protoPayload.serviceName", "iamcredentials.googleapis.com")
', '2026-03-02 22:50:04.982229', true, true, 'origin', null, '[{"indexPattern":"v11-log-google-*","with":[{"field":"log.protoPayload.authenticationInfo.principalEmail","operator":"filter_term","value":"{{.log.protoPayload.authenticationInfo.principalEmail}}"}],"or":null,"within":"now-30m","count":10}]', '["lastEvent.log.protoPayload.methodName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1308, 'GCP probable Password Guessing', 3, 3, 2, 'Credential Access', 'T1110.001 - Brute Force: Password Guessing', 'Adversaries with no prior knowledge of legitimate credentials within the system or environment may guess passwords to attempt access to accounts. Without knowledge of the password for an account, an adversary may opt to systematically guess the password using a repetitive or iterative mechanism. An adversary may guess login credentials without prior knowledge of system or environment passwords during an operation by using a list of common passwords. Password guessing may or may not take into account the target''s policies on password complexity or use policies that may lock accounts out after a number of failed attempts.', '["https://attack.mitre.org/tactics/TA0006","https://attack.mitre.org/techniques/T1110/001/"]', e'equals("log.protoPayload.methodName", "google.login.LoginService.loginFailure") && exists("log.protoPayload.authenticationInfo.principalEmail")
', '2026-03-02 22:50:06.133124', true, true, 'origin', null, '[{"indexPattern":"v11-log-google-*","with":[{"field":"log.protoPayload.methodName","operator":"filter_term","value":"google.login.LoginService.loginFailure"},{"field":"log.protoPayload.authenticationInfo.principalEmail","operator":"filter_term","value":"{{.log.protoPayload.authenticationInfo.principalEmail}}"}],"or":null,"within":"now-5m","count":5}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1309, 'GCP Custom Role with Overly Permissive Permissions', 3, 3, 1, 'Privilege Escalation', 'T1098 - Account Manipulation', e'Detects creation or modification of GCP custom IAM roles which may include overly permissive permissions for privilege escalation. Attackers may create custom roles with broad permissions like iam.serviceAccountKeys.create, iam.serviceAccounts.actAs, or compute.instances.setMetadata to escalate privileges.

Next Steps:
1. Review the custom role definition and its included permissions
2. Verify the role follows least privilege principles
3. Check for high-risk permissions like iam.* or resourcemanager.*
4. Review the identity creating the role and verify authorization
5. Check which users or service accounts are bound to the role
6. If overly permissive, modify the role to include only necessary permissions
7. Implement Organization Policy to restrict custom role creation
8. Use IAM Recommender to identify and reduce excess permissions
', '["https://cloud.google.com/iam/docs/creating-custom-roles","https://attack.mitre.org/techniques/T1098/"]', e'contains("log.protoPayload.serviceName", "iam.googleapis.com") &&
(contains("log.protoPayload.methodName", "CreateRole") ||
 contains("log.protoPayload.methodName", "UpdateRole"))
', '2026-03-02 22:50:07.344879', true, true, 'origin', null, '[{"indexPattern":"v11-log-google-*","with":[{"field":"log.protoPayload.authenticationInfo.principalEmail","operator":"filter_term","value":"{{.log.protoPayload.authenticationInfo.principalEmail}}"}],"or":null,"within":"now-1h","count":2}]', '["lastEvent.log.protoPayload.methodName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1310, 'GCP BigQuery Data Exfiltration Detection', 3, 1, 1, 'Data Exfiltration', 'T1567 - Exfiltration Over Web Service', e'Detects BigQuery operations that may indicate data exfiltration including large data exports, table copies to external projects, and extract jobs writing to external storage. Attackers may use BigQuery to query and export large datasets from compromised projects.

Next Steps:
1. Review the BigQuery job details including source and destination datasets
2. Check the data volume being exported or copied
3. Verify the destination project or storage bucket is legitimate
4. Review the identity performing the operation and verify authorization
5. Check if the query accesses sensitive tables or datasets
6. If unauthorized, cancel running jobs and revoke the identity\'s BigQuery permissions
7. Implement VPC Service Controls to restrict data export
8. Enable BigQuery authorized views to restrict data access
', '["https://cloud.google.com/bigquery/docs/audit-logging","https://attack.mitre.org/techniques/T1567/"]', e'contains("log.protoPayload.serviceName", "bigquery.googleapis.com") &&
(contains("log.protoPayload.methodName", "jobservice.insert") ||
 contains("log.protoPayload.methodName", "tableservice.exportdata") ||
 contains("log.protoPayload.methodName", "datasets.copy"))
', '2026-03-02 22:50:08.569092', true, true, 'origin', null, '[{"indexPattern":"v11-log-google-*","with":[{"field":"log.protoPayload.authenticationInfo.principalEmail","operator":"filter_term","value":"{{.log.protoPayload.authenticationInfo.principalEmail}}"},{"field":"log.protoPayload.serviceName","operator":"filter_term","value":"bigquery.googleapis.com"}],"or":null,"within":"now-30m","count":10}]', '["lastEvent.log.protoPayload.methodName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1311, 'GCP 2-step verification disabled', 1, 2, 3, 'Defense Evasion', 'T1562 - Impair Defenses', 'Google Cloud has detected that 2-step verification was disabled for the organization or a user', '["https://attack.mitre.org/tactics/TA0005","https://attack.mitre.org/techniques/T1562/"]', e'equals("log.protoPayload.methodName", "google.login.LoginService.2svDisable")
', '2026-03-02 22:50:09.810631', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1312, 'GCP Workload Identity Federation Abuse', 3, 3, 1, 'Credential Access', 'T1078 - Valid Accounts', e'Detects creation or modification of workload identity pools and providers that enable external identities to access GCP resources. Attackers may create workload identity configurations to grant access to external attacker-controlled identity providers for persistent cloud access.

Next Steps:
1. Review the workload identity pool and provider configuration
2. Verify the external identity provider is trusted and authorized
3. Check the attribute mappings and conditions for overly permissive access
4. Review which service accounts are bound to the workload identity pool
5. Verify the change was authorized through security change management
6. If unauthorized, delete the workload identity pool and revoke associated permissions
7. Audit all existing workload identity configurations for unauthorized providers
8. Implement Organization Policy to restrict workload identity pool creation
', '["https://cloud.google.com/iam/docs/workload-identity-federation","https://attack.mitre.org/techniques/T1078/"]', e'contains("log.protoPayload.serviceName", "iam.googleapis.com") &&
(contains("log.protoPayload.methodName", "CreateWorkloadIdentityPool") ||
 contains("log.protoPayload.methodName", "CreateWorkloadIdentityPoolProvider") ||
 contains("log.protoPayload.methodName", "UpdateWorkloadIdentityPool") ||
 contains("log.protoPayload.methodName", "UpdateWorkloadIdentityPoolProvider"))
', '2026-03-02 22:50:11.101458', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.methodName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1313, 'GCP suspicious programmatic login', 1, 2, 3, 'Credential Access', 'T1110 - Brute Force', 'Google Cloud has detected a suspicious programmatic login. Programmatic login can be use to perform brute force attack.', '["https://attack.mitre.org/tactics/TA0006","https://attack.mitre.org/techniques/T1110"]', e'equals("log.protoPayload.methodName", "google.login.LoginService.suspiciousProgrammaticLogin")
', '2026-03-02 22:50:12.324932', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1314, 'Google Workspace MFA Enforcement Disabled', 3, 3, 1, 'Defense Evasion', 'T1556 - Modify Authentication Process', e'Detects when MFA enforcement is disabled in Google Workspace. Disabling MFA removes a critical security control and enables credential-based attacks against all users in the organization.

Next Steps:
1. Immediately verify if the MFA policy change was authorized
2. Identify the admin who made the change and their authorization
3. Check for brute force or credential stuffing attempts following the change
4. Re-enable MFA enforcement immediately if unauthorized
5. Review all sign-ins that occurred while MFA was disabled
6. Check for other security policy changes from the same admin
7. Audit admin roles and consider implementing super admin 2SV enforcement
', '["https://support.google.com/a/answer/9176657","https://attack.mitre.org/techniques/T1556/"]', e'contains("log.protoPayload.methodName", "ENFORCE_STRONG_AUTHENTICATION") ||
(contains("log.protoPayload.serviceName", "admin.googleapis.com") && contains("log.protoPayload.methodName", "2sv") && contains("log.protoPayload.request", "disable"))
', '2026-03-02 22:50:13.478141', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","lastEvent.log.protoPayload.methodName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1315, 'GCP suspicious login from less secure app', 1, 2, 3, 'Initial Access', 'T1190 - Exploit Public-Facing Application', 'Less secure apps (LSAs) are non-Google apps that can access your Google account with only a username and password.  They make your account more vulnerable to hijacking attempts.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1190"]', e'equals("log.protoPayload.methodName", "google.login.LoginService.suspiciousLoginLessSecureApp")
', '2026-03-02 22:50:14.785489', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1316, 'GCP suspicious login blocked', 1, 2, 3, 'Initial Access', 'T1078 - Valid Accounts', 'A suspicious login to a user''s account was detected and blocked by Google Cloud.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1078"]', e'equals("log.protoPayload.methodName", "google.login.LoginService.suspiciousLogin")
', '2026-03-02 22:50:16.081535', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1317, 'GCP Cloud Storage Data Exfiltration', 3, 1, 1, 'Data Exfiltration', 'T1530 - Data from Cloud Storage Object', e'Detects GCP Cloud Storage operations indicating potential data exfiltration including making buckets publicly accessible, modifying IAM policies to grant allUsers access, or bulk object downloads. These actions may indicate an attacker attempting to exfiltrate data from cloud storage.

Next Steps:
1. Review the affected bucket and its data classification
2. Check if the bucket was made publicly accessible
3. Verify the identity making the change has authorization
4. Review the IAM policy changes for allUsers or allAuthenticatedUsers bindings
5. Check for bulk GetObject operations following the policy change
6. If unauthorized, revert the bucket IAM policy and enable uniform bucket-level access
7. Review VPC Service Controls for the project
8. Enable Cloud Storage audit logging for data access events
', '["https://cloud.google.com/storage/docs/access-control","https://attack.mitre.org/techniques/T1530/"]', e'contains("log.protoPayload.serviceName", "storage.googleapis.com") &&
(contains("log.protoPayload.methodName", "storage.setIamPermissions") ||
 contains("log.protoPayload.methodName", "storage.buckets.update") ||
 contains("log.protoPayload.methodName", "storage.objects.update")) &&
(contains("log.protoPayload.request.policy.bindings", "allUsers") ||
 contains("log.protoPayload.request.policy.bindings", "allAuthenticatedUsers") ||
 contains("log.protoPayload.request.acl", "allUsers"))
', '2026-03-02 22:50:17.435194', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.resourceName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1318, 'GCP Project Manipulation and Shadow Projects', 2, 3, 3, 'Account Manipulation', 'T1578 - Modify Cloud Compute Infrastructure', e'Detects GCP project creation, deletion, and undelete operations that could indicate shadow project creation for persistence or project deletion for impact. Attackers may create new projects outside organizational controls to host malicious workloads.

Next Steps:
1. Verify the project creation or deletion was authorized
2. Check if the new project is within the expected folder hierarchy
3. Review the project\'s billing account association
4. Examine IAM bindings on the new project for overly permissive access
5. Check if Organization Policies are applied to the new project
6. If unauthorized, shut down the project and investigate the creating identity
7. Implement Organization Policy constraints for project creation
8. Enable alerts for projects created outside approved folders
', '["https://cloud.google.com/resource-manager/docs/creating-managing-projects","https://attack.mitre.org/techniques/T1578/"]', e'contains("log.protoPayload.serviceName", "cloudresourcemanager.googleapis.com") &&
(contains("log.protoPayload.methodName", "CreateProject") ||
 contains("log.protoPayload.methodName", "DeleteProject") ||
 contains("log.protoPayload.methodName", "UndeleteProject"))
', '2026-03-02 22:50:18.878275', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.methodName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1319, 'GCP probable Privilege Escalation, Kubernetes role bindings created or patched', 1, 2, 3, 'Privilege Escalation', 'T1548 - Abuse Elevation Control Mechanism', 'Privilege Escalation consists of techniques that adversaries use to gain higher-level permissions on a system or network.  Adversaries can often enter and explore a network with unprivileged access but require elevated permissions to follow through  on their objectives. Common approaches are to take advantage of system weaknesses, misconfigurations, and vulnerabilities.  Identifies the creation or patching of potentially malicious role bindings. Users can use role bindings and cluster role  bindings to assign roles to Kubernetes subjects (users, groups, or service accounts).', '["https://cloud.google.com/kubernetes-engine/docs/how-to/role-based-access-control","https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1548"]', e'contains("log.protoPayload.methodName", ".rbac") &&
    regexMatch("log.protoPayload.methodName", \'((.+)\\\\.)?(cluster)?rolebinding(s)?\\\\.(create|patch)$\') &&
    !equals("log.protoPayload.authenticationInfo.principalEmail", "system:addon-manager")
', '2026-03-02 22:50:20.237466', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1320, 'GCP Network Packet Capture Configuration', 3, 1, 1, 'Credential Access', 'T1040 - Network Sniffing', e'Detects creation or modification of Packet Mirroring configurations in GCP. Attackers use packet mirroring to capture network traffic for credential theft, data exfiltration, or reconnaissance.

Next Steps:
1. Verify the packet mirroring configuration was authorized for legitimate purposes
2. Review the mirrored network scope (which subnets, instances, protocols)
3. Check the collector destination for the mirrored traffic
4. Identify the user who created the configuration
5. If unauthorized, delete the packet mirroring policy immediately
6. Review the mirrored traffic destination for data exfiltration
7. Check for captured credentials or sensitive data
', '["https://cloud.google.com/vpc/docs/packet-mirroring","https://attack.mitre.org/techniques/T1040/"]', e'contains("log.protoPayload.methodName", "PacketMirrorings") &&
(contains("log.protoPayload.methodName", "insert") || contains("log.protoPayload.methodName", "patch") || contains("log.protoPayload.methodName", "create"))
', '2026-03-02 22:50:21.673547', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","lastEvent.log.protoPayload.resourceName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1321, 'GKE Kubernetes Admission Webhook Modified', 3, 3, 2, 'Persistence', 'T1078.004 - Valid Accounts: Cloud Accounts', e'Detects creation or modification of admission webhook configurations in Google Kubernetes Engine. Attackers use malicious admission controllers to inject sidecar containers, modify workload specs, or intercept secrets.

Next Steps:
1. Review the webhook configuration and its target service endpoint
2. Verify the webhook was deployed as part of a legitimate application
3. Check the namespace selector and object rules for the webhook
4. Examine what Kubernetes resources the webhook intercepts
5. If unauthorized, delete the webhook and audit all recent workload deployments
6. Review cluster RBAC for webhook management permissions
', '["https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/","https://attack.mitre.org/techniques/T1078/004/"]', e'contains("log.protoPayload.methodName", "admissionregistration.k8s.io") &&
(contains("log.protoPayload.methodName", "mutatingwebhookconfigurations") || contains("log.protoPayload.methodName", "validatingwebhookconfigurations")) &&
(contains("log.protoPayload.methodName", "create") || contains("log.protoPayload.methodName", "update") || contains("log.protoPayload.methodName", "patch"))
', '2026-03-02 22:50:23.085343', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","lastEvent.log.protoPayload.resourceName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1322, 'GCP probable Impact, Storage Bucket Deleted', 1, 2, 3, 'Impact', 'T1485 - Data Destruction', 'Impact consists of techniques that adversaries use to disrupt availability or compromise integrity by manipulating business  and operational processes. Techniques used for impact can include destroying or tampering with data. In some cases, business  processes can look fine, but may have been altered to benefit the adversaries goals. These techniques might be used by  adversaries to follow through on their end goal or to provide cover for a confidentiality breach.   Identifies when a Google Cloud Platform (GCP) storage bucket is deleted. An adversary may delete a storage bucket in  order to disrupt their target''s business operations.', '["https://cloud.google.com/logging/docs/buckets","https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1485/"]', e'regexMatch("log.protoPayload.methodName", "(.+)\\\\.bucket(s)?\\\\.delete")
', '2026-03-02 22:50:24.421324', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1323, 'GCP KMS Key Destruction or Disabling', 1, 3, 3, 'Impact', 'T1552 - Unsecured Credentials', e'Detects destruction or disabling of Cloud KMS key versions which could render encrypted data unrecoverable. Attackers may destroy encryption keys as part of a destructive attack to prevent data recovery or to disrupt operations dependent on encrypted resources.

Next Steps:
1. Immediately verify if the KMS key operation was authorized
2. Identify which resources are encrypted with the affected key
3. Check if the key version is in the scheduled destruction period and can be restored
4. Review the identity performing the operation and verify authorization
5. Assess the business impact of the key becoming unavailable
6. If unauthorized, restore the key version immediately during the destruction grace period
7. Implement IAM conditions to restrict KMS key destruction permissions
8. Enable Cloud KMS key rotation policies and cross-region key replication
', '["https://cloud.google.com/kms/docs/destroy-restore","https://attack.mitre.org/techniques/T1552/"]', e'contains("log.protoPayload.serviceName", "cloudkms.googleapis.com") &&
(contains("log.protoPayload.methodName", "DestroyCryptoKeyVersion") ||
 contains("log.protoPayload.methodName", "DisableCryptoKeyVersion") ||
 contains("log.protoPayload.methodName", "UpdateCryptoKeyPrimaryVersion"))
', '2026-03-02 22:50:25.781701', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.resourceName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1324, 'GCP probable Government-backed attack', 3, 3, 2, 'Collection', 'T1560 - Archive Collected Data', 'A user''s account might have been targeted by government-backed attack.  Government-backed attackers are trying to access the account of one of your users.  An attack happens to less than 0.1% of all Google Account users. There''s a chance the alert is  a false alarm. However, we believe we detected activities that government-backed attackers use  to try to steal a password or other personal information. Such activity includes the user receiving  an email containing a harmful attachment, links to malicious software downloads, or links to fake  websites that are designed to access passwords.', '["https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1560"]', e'contains("log.protoPayload.methodName", "google.login.LoginService.govAttackWarning")
', '2026-03-02 22:50:27.049223', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1325, 'GCP probable Exfiltration, Logging Sink Modification', 3, 2, 2, 'Exfiltration', 'T1537 - Transfer Data to Cloud Account', 'Exfiltration consists of techniques that adversaries may use to steal data from your network. Once they''ve collected data,  adversaries often package it to avoid detection while removing it. This can include compression and encryption. Techniques for  getting data out of a target network typically include transferring it over their command and control channel or an alternate  channel and may also include putting size limits on the transmission.   Identifies a modification to a Logging sink in Google Cloud Platform (GCP). Logging compares the log entry to the sinks  in that resource. Each sink whose filter matches the log entry writes a copy of the log entry to the sink''s export  destination. An adversary may update a Logging sink to exfiltrate logs to a different export destination.', '["https://cloud.google.com/logging/docs/export#how_sinks_work","https://cloud.google.com/logging/docs/reference/v2/rest/v2/projects.sinks#LogSink","https://attack.mitre.org/techniques/T1537/","https://attack.mitre.org/tactics/TA0010/"]', e'regexMatch("log.protoPayload.methodName", "((.+)?sink(s)?\\\\.update|(.+)?v(\\\\w+)\\\\.ConfigServiceV(\\\\w+)\\\\.UpdateSink)")
', '2026-03-02 22:50:28.383564', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1326, 'GCP Domain-Wide API Access Granted', 3, 3, 2, 'Privilege Escalation', 'T1098 - Account Manipulation', e'Detects when domain-wide delegation is granted to a service account in Google Workspace. This allows the service account to impersonate any user in the domain and access their data, making it a high-impact privilege escalation vector.

Next Steps:
1. Verify the domain-wide delegation was authorized by a domain administrator
2. Review the OAuth scopes granted to the service account
3. Check the service account\'s usage history and associated project
4. Verify the scopes follow the principle of least privilege
5. If unauthorized, revoke the delegation immediately
6. Audit all API calls made by the service account since the delegation was granted
7. Review Google Workspace admin logs for related changes
', '["https://cloud.google.com/iam/docs/using-iam-securely","https://attack.mitre.org/techniques/T1098/"]', e'contains("log.protoPayload.methodName", "AUTHORIZE_API_CLIENT_ACCESS") ||
(contains("log.protoPayload.serviceName", "admin.googleapis.com") && contains("log.protoPayload.methodName", "GrantClientAccess"))
', '2026-03-02 22:50:29.709668', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","lastEvent.log.protoPayload.resourceName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1327, 'GCP DLP Re-Identification API Call', 3, 2, 0, 'Collection', 'T1565 - Data Manipulation', e'Detects calls to the DLP re-identification API which reverses data de-identification. This is a sensitive operation that could expose previously protected PII, financial data, or health records. Unauthorized use indicates potential data exfiltration attempts.

Next Steps:
1. Verify the re-identification request was authorized for the specific use case
2. Review the data being re-identified and its sensitivity classification
3. Check the user identity and whether they have legitimate access to this data
4. Review the destination of the re-identified data
5. If unauthorized, revoke access and investigate potential data exposure
6. Review DLP API permissions and restrict re-identification access
', '["https://cloud.google.com/dlp/docs/reference/rest/v2/projects.content/reidentify","https://attack.mitre.org/techniques/T1565/"]', e'contains("log.protoPayload.methodName", "ReidentifyContent") ||
contains("log.protoPayload.methodName", "reidentify")
', '2026-03-02 22:50:31.154034', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","lastEvent.log.protoPayload.methodName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1328, 'GCP probable Defense Evasion, Logging Sink Deletion', 1, 2, 3, 'Defense Evasion', 'T1562 - Impair Defenses', 'Defense Evasion consists of techniques that adversaries use to avoid detection throughout their compromise.  Techniques used for defense evasion include uninstalling/disabling security software or  obfuscating/encrypting data and scripts. Adversaries also leverage and abuse trusted processes  to hide and masquerade their malware. Other tactics are cross-listed here when those techniques  include the added benefit of subverting defenses.   Identifies a Logging sink deletion in Google Cloud Platform (GCP). Every time a log entry arrives, Logging  compares the log entry to the sinks in that resource. Each sink whose filter matches the log entry writes a  copy of the log entry to the sink''s export destination. An adversary may delete a Logging sink to evade detection.', '["https://cloud.google.com/logging/docs/export","https://attack.mitre.org/techniques/T1562/","https://attack.mitre.org/tactics/TA0005/"]', e'regexMatch("log.protoPayload.methodName", "((.+)?sink(s)?\\\\.delete|(.+)?v(\\\\w+)\\\\.ConfigServiceV(\\\\w+)\\\\.DeleteSink)")
', '2026-03-02 22:50:32.509879', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1329, 'GCP Cryptomining Instance Launch Detection', 1, 2, 3, 'Resource Hijacking', 'T1496 - Resource Hijacking', e'Detects creation of GPU-accelerated or high-compute GCP instances commonly used for cryptomining. Attackers with compromised GCP credentials frequently launch expensive GPU instances (a2, g2) or compute-optimized instances in unusual regions for cryptocurrency mining operations.

Next Steps:
1. Verify the identity launching the instance and confirm business justification
2. Check if GPU instances are normally used in this project
3. Review the instance\'s machine type and attached GPU accelerators
4. Examine the instance image for known mining software
5. Check billing dashboards for unexpected cost increases
6. If unauthorized, stop and delete the instance immediately
7. Rotate compromised credentials and review IAM bindings
8. Implement Organization Policy constraints to restrict GPU instance creation
', '["https://cloud.google.com/compute/docs/machine-types","https://attack.mitre.org/techniques/T1496/"]', e'contains("log.protoPayload.methodName", "compute.instances.insert") &&
(contains("log.protoPayload.request.machineType", "a2-") ||
 contains("log.protoPayload.request.machineType", "g2-") ||
 contains("log.protoPayload.request.machineType", "n1-highmem-96") ||
 contains("log.protoPayload.request.machineType", "c2d-highcpu") ||
 contains("log.protoPayload.request.guestAccelerators", "nvidia"))
', '2026-03-02 22:50:33.819638', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.resourceName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1330, 'GCP Cloud Function and Cloud Run Abuse', 2, 2, 1, 'Persistence', 'T1059 - Command and Scripting Interpreter', e'Detects creation or modification of Cloud Functions and Cloud Run services which can be used for persistence, backdoor access, or command execution. Attackers may deploy serverless functions with high-privilege service accounts to maintain access or exfiltrate data.

Next Steps:
1. Review the function or service code for malicious content
2. Check the associated service account and its permissions
3. Verify the deployer identity has authorization
4. Review the function trigger configuration (HTTP, Pub/Sub, etc.)
5. Check if the function allows unauthenticated invocations
6. If unauthorized, delete the function and revoke the service account\'s permissions
7. Review invocation logs for the function
8. Implement Organization Policy to restrict Cloud Function deployment
', '["https://cloud.google.com/functions/docs/securing","https://attack.mitre.org/techniques/T1059/"]', e'((contains("log.protoPayload.serviceName", "cloudfunctions.googleapis.com") &&
  (contains("log.protoPayload.methodName", "CreateFunction") ||
   contains("log.protoPayload.methodName", "UpdateFunction"))) ||
 (contains("log.protoPayload.serviceName", "run.googleapis.com") &&
  (contains("log.protoPayload.methodName", "CreateService") ||
   contains("log.protoPayload.methodName", "ReplaceService"))))
', '2026-03-02 22:50:35.261306', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.methodName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1331, 'GCP Break-Glass Container Workload Deployed', 3, 3, 2, 'Defense Evasion', 'T1548 - Abuse Elevation Control Mechanism', e'Detects deployment of container workloads using the break-glass mechanism to bypass Binary Authorization policy. While legitimate in emergency scenarios, this bypasses security controls and can be abused to deploy malicious or untrusted container images.

Next Steps:
1. Verify the break-glass deployment was authorized and documented
2. Review the container image that was deployed
3. Check the user identity and their authorization level
4. Validate the business justification for the emergency bypass
5. Ensure Binary Authorization policies are restored after the emergency
6. Scan the deployed container for vulnerabilities and malware
7. Review cluster activity following the deployment
', '["https://cloud.google.com/binary-authorization/docs/using-breakglass","https://attack.mitre.org/techniques/T1548/"]', e'(equals("log.protoPayload.serviceName", "binaryauthorization.googleapis.com") &&
contains("log.protoPayload.response", "breakglass")) ||
(contains("log.protoPayload.methodName", "container.clusters") &&
contains("log.protoPayload.request", "breakglass"))
', '2026-03-02 22:50:36.569229', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","lastEvent.log.protoPayload.resourceName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1332, 'GCP Audit Log Disabling or Tampering', 3, 3, 2, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', e'Detects attempts to disable GCP audit logging including sink deletion, log exclusion filter creation, and audit configuration changes. Attackers may manipulate logging infrastructure to hide their activities from security monitoring.

Next Steps:
1. Immediately verify if the logging change was authorized
2. Review the specific sink or exclusion filter that was modified
3. Check the identity making the change and verify authorization
4. Assess what log types are no longer being collected
5. Restore logging configuration and ensure all critical logs are captured
6. Review activities that may have been hidden during the logging gap
7. Implement Organization Policy to prevent log sink deletion
8. Set up alerting on any changes to logging infrastructure
', '["https://cloud.google.com/logging/docs/audit","https://attack.mitre.org/techniques/T1562/008/"]', e'(contains("log.protoPayload.methodName", "DeleteSink") ||
 contains("log.protoPayload.methodName", "UpdateSink") ||
 contains("log.protoPayload.methodName", "CreateExclusion") ||
 contains("log.protoPayload.methodName", "UpdateExclusion") ||
 contains("log.protoPayload.methodName", "DeleteLog") ||
 contains("log.protoPayload.methodName", "SetIamPolicy")) &&
(contains("log.protoPayload.serviceName", "logging.googleapis.com") ||
 contains("log.resource.type", "logging_sink") ||
 contains("log.resource.type", "logging_exclusion"))
', '2026-03-02 22:50:37.755334', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.methodName","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1333, 'GCP account is probably used for spamming', 1, 2, 3, 'Initial Access', 'T1566 - Phishing', 'A user''s account was disabled because Google has become aware that it was used to engage in spamming. Usually, spamming is used to perform other attacks like phishing or spread malware.', '["https://attack.mitre.org/tactics/TA0001","https://attack.mitre.org/techniques/T1566/"]', e'equals("log.protoPayload.methodName", "google.login.LoginService.accountDisabledSpamming") ||
equals("log.protoPayload.methodName", "google.login.LoginService.accountDisabledSpammingThroughRelay")
', '2026-03-02 22:50:38.930239', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1334, 'GCP detected account with password leak', 3, 3, 2, 'Initial Access', 'T1078 - Valid Accounts', 'A user''s account was disabled because a password leak was detected by google.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1078"]', e'equals("log.protoPayload.methodName", "google.login.LoginService.accountDisabledPasswordLeak")
', '2026-03-02 22:50:40.003674', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1335, 'GCP probable hijacked account', 3, 3, 2, 'Collection', 'T1560 - Archive Collected Data', 'A user''s account was disabled because Google has detected a suspicious activity  indicating it might have been compromised. Hijacked account can be used to perform other attacks like data collection and exfiltration', '["https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1560"]', e'equals("log.protoPayload.methodName", "google.login.LoginService.accountDisabledHijacked")
', '2026-03-02 22:50:41.199708', true, true, 'target', null, '[]', '["target.ip","target.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1336, 'Cloud Identity Suspicious Sign-ins Detection', 3, 2, 1, 'Initial Access', 'T1078 - Valid Accounts', e'Detects suspicious sign-in attempts to Google Cloud Identity, including logins from unfamiliar locations, unusual IP addresses, or after multiple failed attempts. These could indicate compromised credentials or unauthorized access attempts.

Next Steps:
1. Verify the legitimacy of the login attempt with the user
2. Check if the IP address is from a known malicious source
3. Review recent account activity for signs of compromise
4. Consider implementing additional MFA if not already enabled
5. If confirmed malicious, reset user credentials immediately
6. Review access logs for any unauthorized activities
', '["https://support.google.com/cloudidentity/answer/4580120?hl=en","https://cloud.google.com/blog/products/identity-security/logs-based-security-alerting-in-google-cloud","https://attack.mitre.org/techniques/T1078/"]', e'equals("log.protoPayload.serviceName", "login.googleapis.com") &&
(
  equals("log.protoPayload.metadata.event.type", "Suspicious Login") ||
  (equals("log.protoPayload.metadata.event.type", "login") && equals("log.protoPayload.metadata.event.parameter.is_suspicious", true)) ||
  equals("log.protoPayload.metadata.event.parameter.is_suspicious", true)
)
', '2026-03-02 22:50:42.596769', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1337, 'Binary Authorization Bypass Detection', 3, 3, 2, 'Defense Evasion', 'T1553 - Subvert Trust Controls', e'Detects attempts to bypass Binary Authorization controls including use of breakglass deployments, policy violations, and unauthorized container deployments. These events could indicate attempts to deploy untrusted or malicious container images.

Next Steps:
1. Verify the legitimacy of the breakglass deployment or policy bypass
2. Review the container image source and verify its authenticity
3. Check if the user had proper authorization for emergency deployments
4. Examine the deployment context and business justification
5. Validate that security policies are restored after emergency deployment
6. Monitor for any subsequent suspicious activity from deployed containers
', '["https://cloud.google.com/binary-authorization/docs/audit-logging","https://cloud.google.com/binary-authorization/docs/run/using-breakglass-cloud-run","https://attack.mitre.org/techniques/T1553/"]', e'(
  equals("log.protoPayload.serviceName", "binaryauthorization.googleapis.com") &&
  (
    contains("log.logName", "cloudaudit.googleapis.com/system_event") &&
    (contains("log.protoPayload.response.details", "breakglass") || equals("log.jsonPayload.breakglass", true))
  )
) ||
(
  equals("log.resourceType", "cloud_run_revision") &&
  contains("log.logName", "cloudaudit.googleapis.com/system_event") &&
  (
    contains("log.protoPayload.response.status.conditions", "ContainerImageUnauthorized") ||
    equals("log.jsonPayload.policyViolation", true) ||
    equals("log.protoPayload.metadata.dryRun", true)
  )
)
', '2026-03-02 22:50:43.788032', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.authenticationInfo.principalEmail","lastEvent.log.protoPayload.resourceName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1338, 'Anthos Security Policy Violations', 3, 3, 2, 'Security Control Bypass', 'T1562 - Impair Defenses', e'Detects security-related events in Google Anthos environments including policy violations, authentication failures, and suspicious container activities. Monitors Anthos Service Mesh, Config Management, and Policy Controller events.

Next Steps:
- Review the specific policy violation details in the event logs
- Verify if the violation was authorized or represents a legitimate security concern
- Check the source IP and user account associated with the violation
- Examine recent configuration changes to Anthos security policies
- Validate that security controls are properly configured and enforced
- Consider implementing additional monitoring for the affected resources
', '["https://cloud.google.com/anthos/docs/concepts/overview","https://attack.mitre.org/techniques/T1562/"]', e'(
  oneOf("log.protoPayload.serviceName", ["anthos.googleapis.com", "anthospolicycontroller.googleapis.com", "anthosservicemesh.googleapis.com"]) ||
  oneOf("log.resourceType", ["k8s_cluster", "gke_cluster"])
) &&
(
  contains("log.protoPayload.methodName", "Policy") ||
  oneOf("log.jsonPayload.type", ["admission.k8s.io/violation", "policy.violation", "security.alert"]) ||
  oneOf("log.severity", ["ERROR", "WARNING"])
) &&
(
  equals("log.protoPayload.response.status", "PERMISSION_DENIED") ||
  contains("log.protoPayload.status.message", "violation") ||
  contains("log.protoPayload.status.message", "denied") ||
  contains("log.jsonPayload.details", "policy")
)
', '2026-03-02 22:50:45.089959', true, true, 'origin', null, '[]', '["lastEvent.log.protoPayload.resourceName","lastEvent.log.resource.labels.project_id"]');
