INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1186, 'AWS Management Console Brute Force of Root User Identity', 3, 2, 1, 'Credential Access', 'T1110 - Brute Force', 'Identifies a high number of failed authentication attempts to the AWS management console for the Root user identity. An adversary may attempt to brute force the password for the Root user identity, as it has complete access to all services and resources for the AWS account', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1110/","https://docs.aws.amazon.com/IAM/latest/UserGuide/id_root-user.html"]', e'equals("log.eventSource", "signin.amazonaws.com") &&
equals("log.eventName", "ConsoleLogin") &&
equals("log.userIdentityType", "root") &&
(exists("log.errorCode") || exists("log.errorMessage"))
', '2026-03-02 19:28:13.989146', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"origin.ip","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-15m","count":5}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1187, 'AWS IAM Brute Force of Assume Role Policy', 3, 2, 1, 'Credential Access', 'T1110 - Brute Force', 'Identifies a high number of failed attempts to assume an AWS Identity and Access Management (IAM) role. IAM roles are used to delegate access to users or services. An adversary may attempt to enumerate IAM  roles in order to determine if a role exists before attempting to assume or hijack the discovered role', '["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1110/","https://www.praetorian.com/blog/aws-iam-assume-role-vulnerabilities","https://rhinosecuritylabs.com/aws/assume-worst-aws-assume-role-enumeration/"]', e'equals("log.eventSource", "iam.amazonaws.com") &&
equals("log.eventName", "UpdateAssumeRolePolicy") &&
equals("log.errorCode", "MalformedPolicyDocumentException")
', '2026-03-02 19:28:15.419013', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"origin.user","operator":"filter_term","value":"{{.origin.user}}"}],"or":null,"within":"now-15m","count":5}]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1188, 'AWS Root Login Without MFA', 3, 2, 2, 'Initial Access', 'T1078 - Valid Accounts', 'Identifies attempts to login to AWS as the root user without using multi-factor authentication (MFA). Amazon AWS best practices indicate that the root user should be protected by MFA', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1078/","https://docs.aws.amazon.com/IAM/latest/UserGuide/id_root-user.html"]', e'equals("log.eventSource", "signin.amazonaws.com") &&
equals("log.eventName", "ConsoleLogin") &&
equals("log.userIdentityType", "root") &&
equals("log.additionalEventData.MFAUsed", "no")
', '2026-03-02 19:28:16.726404', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1242, 'EC2 Instance Metadata Abuse', 3, 2, 1, 'Credential Access', 'T1552.005 - Unsecured Credentials: Cloud Instance Metadata API', e'Detects potential abuse of EC2 instance metadata service (IMDS) which could indicate SSRF exploitation or credential theft. Monitors for unusual API calls using credentials with IMDSv1 role delivery or suspicious patterns of EC2 metadata access.

Next Steps:
1. Identify the EC2 instance and application making the metadata requests
2. Check if the instance has IMDSv2 enforced (HttpTokens set to "required")
3. Review the instance\'s IAM role permissions and recent API activity
4. Investigate any web applications running on the instance for SSRF vulnerabilities
5. Check CloudTrail logs for unusual API calls using the instance profile credentials
6. If unauthorized access is confirmed, rotate the instance profile credentials and enforce IMDSv2
', '["https://hackingthe.cloud/aws/exploitation/ec2-metadata-ssrf/","https://attack.mitre.org/techniques/T1552/005/"]', e'(equals("log.eventSource", "ec2.amazonaws.com") &&
 equals("log.eventName", "ModifyInstanceMetadataOptions") &&
 equals("log.errorCode", "") &&
 contains("log.requestParameters", "httpTokens\\":\\"optional")) ||
(exists("log.requestParameters") &&
 contains("log.requestParameters", "ec2:RoleDelivery\\":\\"1.0") &&
 equals("log.errorCode", ""))
', '2026-03-02 19:29:28.778287', true, true, 'origin', null, '[]', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentityAccountId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1243, 'AWS EC2 Cryptomining Instance Launch Detection', 1, 2, 3, 'Resource Hijacking', 'T1496 - Resource Hijacking', e'Detects EC2 instance launches using GPU or high-compute instance types commonly associated with cryptomining operations. Attackers with compromised AWS credentials frequently launch expensive GPU instances (p3, p4, g4, g5) or large compute-optimized instances for cryptocurrency mining.

Next Steps:
1. Verify the identity launching the instance ({{log.userIdentityArn}}) and confirm authorization
2. Check if the instance type matches legitimate workloads for the account
3. Review the source IP ({{log.sourceIPAddress}}) for suspicious origins
4. Examine the AMI used for the instance launch for known mining software
5. Check billing dashboards for unexpected cost spikes
6. If unauthorized, terminate the instance immediately and rotate compromised credentials
7. Review IAM policies to restrict instance type launches using SCPs
8. Enable AWS Budgets alerts for cost anomaly detection
', '["https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html","https://attack.mitre.org/techniques/T1496/"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
equals("log.eventName", "RunInstances") &&
equals("log.errorCode", "") &&
regexMatch("log.requestParameters.instanceType", "(?i)^(p3|p4d|p4de|p5|g4dn|g4ad|g5|g5g|c5\\\\.18xlarge|c5\\\\.24xlarge|c5a\\\\.24xlarge|c6i\\\\.32xlarge|c7g\\\\.16xlarge)")
', '2026-03-02 19:29:30.086015', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.instanceType"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1189, 'AWS IAM Assume Role Policy Update', 2, 2, 1, 'Initial Access', 'T1078 - Valid Accounts', 'Identifies attempts to modify an AWS IAM Assume Role Policy. An adversary may attempt to modify the AssumeRolePolicy of a misconfigured role in order to gain the privileges of that role', '["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1078/","https://labs.bishopfox.com/tech-blog/5-privesc-attack-vectors-in-aws"]', e'equals("log.eventSource", "iam.amazonaws.com") &&
equals("log.eventName", "UpdateAssumeRolePolicy")
', '2026-03-02 19:28:18.139259', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1190, 'AWS Route 53 Domain Transferred to Another Account', 3, 3, 3, 'Persistence', 'T1098 - Account Manipulation', 'Identifies when a request has been made to transfer a Route 53 domain to another AWS account', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1098/","https://attack.mitre.org/tactics/TA0006/","https://docs.aws.amazon.com/Route53/latest/APIReference/API_Operations_Amazon_Route_53.html"]', e'equals("log.eventSource", "route53.amazonaws.com") &&
equals("log.eventName", "TransferDomainToAnotherAwsAccount")
', '2026-03-02 19:28:19.570552', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1191, 'AWS Route 53 Domain Transfer Lock Disabled', 3, 2, 2, 'Persistence', 'T1098 - Account Manipulation', 'Identifies when a transfer lock was removed from a Route 53 domain. It is recommended to refrain from performing this action unless intending to transfer the domain to a different registrar', '["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1098/","https://attack.mitre.org/tactics/TA0006/","https://docs.aws.amazon.com/Route53/latest/APIReference/API_Operations_Amazon_Route_53.html","https://docs.aws.amazon.com/Route53/latest/APIReference/API_domains_DisableDomainTransferLock.html"]', e'equals("log.eventSource", "route53.amazonaws.com") &&
equals("log.eventName", "DisableDomainTransferLock")
', '2026-03-02 19:28:20.877451', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1192, 'AWS Execution via System Manager', 2, 1, 1, 'Initial Access', 'T1566 - Phishing', 'Identifies the execution of commands and scripts via System Manager. Execution methods such as RunShellScript, RunPowerShellScript, and alike can be abused by an authenticated attacker to install a backdoor or to interact with a compromised instance via reverse-shell using system only commands<br><strong>Potential false positives</strong><br>Verify whether the user identity, user agent, and/or hostname should be making changes in your environment. Suspicious commands from unfamiliar users or hosts should be investigated. If known behavior is causing false positives, it can be exempted from the rule.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1566/","https://docs.aws.amazon.com/systems-manager/latest/userguide/ssm-plugins.html"]', e'equals("log.eventSource", "ssm.amazonaws.com") &&
equals("log.eventName", "SendCommand")
', '2026-03-02 19:28:22.195231', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1193, 'AWS IAM Password Recovery Requested', 2, 1, 0, 'Initial Access', 'T1078 - Valid Accounts', 'Identifies AWS IAM password recovery requests. An adversary may attempt to gain unauthorized AWS access by abusing password recovery mechanisms.<br><strong>Potential false positives</strong><br>Verify whether the user identity, user agent, and/or hostname should be requesting changes in your environment. Password reset attempts from unfamiliar users should be investigated. If known behavior is causing false positives, it can be exempted from the rule.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1078/","https://www.cadosecurity.com/2020/06/11/an-ongoing-aws-phishing-campaign/"]', e'equals("log.eventSource", "signin.amazonaws.com") &&
equals("log.eventName", "PasswordRecoveryRequested")
', '2026-03-02 19:28:23.414873', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1194, 'AWS Management Console Root Login', 3, 3, 3, 'Initial Access', 'T1078 - Valid Accounts', 'Identifies a successful login to the AWS Management Console by the Root user.<br>Adversaries may obtain and abuse credentials of a cloud account as a means of gaining Initial Access, Persistence, Privilege Escalation, or Defense Evasion.<br>Compromised credentials for cloud accounts can be used to harvest sensitive data from online storage accounts and databases.<br><strong>Potential false positives</strong><br>It’s strongly recommended that the root user is not used for everyday tasks, including the administrative ones. Verify whether the IP address, location, and/or hostname should be logging in as root in your environment. Unfamiliar root logins should be investigated immediately. If known behavior is causing false positives, it can be exempted from the rule.', '["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1078/","https://docs.aws.amazon.com/IAM/latest/UserGuide/id_root-user.html"]', e'equals("log.eventSource", "signin.amazonaws.com") &&
equals("log.eventName", "ConsoleLogin") &&
equals("log.userIdentityType", "root")
', '2026-03-02 19:28:24.682561', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1195, 'AWS IAM Deactivation of MFA Device', 3, 2, 2, 'Impact', 'T1531 - Account Access Removal', 'Identifies the deactivation of a specified multi-factor authentication (MFA) device and removes it from association with the user name for which it was originally enabled. In AWS Identity and Access Management (IAM), a device must be deactivated before it can be deleted', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1531/","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/iam/deactivate-mfa-device.html","https://docs.aws.amazon.com/IAM/latest/APIReference/API_DeactivateMFADevice.html"]', e'equals("log.eventSource", "iam.amazonaws.com") &&
oneOf("log.eventName", ["DeactivateMFADevice", "DeleteVirtualMFADevice"])
', '2026-03-02 19:28:26.077268', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1196, 'AWS EC2 Encryption Disabled', 3, 2, 2, 'Impact', 'T1565 - Data Manipulation', 'Identifies disabling of Amazon Elastic Block Store (EBS) encryption by default in the current region. Disabling encryption by default does not change the encryption status of your existing volumes', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1565/","https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/ec2/disable-ebs-encryption-by-default.html","https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DisableEbsEncryptionByDefault.html"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
equals("log.eventName", "DisableEbsEncryptionByDefault")
', '2026-03-02 19:28:27.480574', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1197, 'AWS CloudWatch Log Stream Deletion', 3, 2, 2, 'Impact', 'T1485 - Data Destruction', 'Identifies the deletion of an AWS CloudWatch log stream, which permanently deletes all associated archived log events with the stream', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1485/","https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/logs/delete-log-stream.html","https://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_DeleteLogStream.html"]', e'equals("log.eventSource", "logs.amazonaws.com") &&
equals("log.eventName", "DeleteLogStream")
', '2026-03-02 19:28:28.742279', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1198, 'AWS RDS Cluster Deletion', 3, 2, 2, 'Impact', 'T1485 - Data Destruction', 'Identifies the deletion of an Amazon Relational Database Service (RDS) Aurora database cluster or global database cluster', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1485/","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/rds/delete-db-cluster.html","https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DeleteDBCluster.html","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/rds/delete-global-cluster.html","https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_DeleteGlobalCluster.html"]', e'equals("log.eventSource", "rds.amazonaws.com") &&
oneOf("log.eventName", ["DeleteDBCluster", "DeleteGlobalCluster"])
', '2026-03-02 19:28:30.142155', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1199, 'AWS CloudWatch Log Group Deletion', 3, 2, 2, 'Impact', 'T1485 - Data Destruction', 'Identifies the deletion of a specified AWS CloudWatch log group. When a log group is deleted, all the archived log events associated with the log group are also permanently deleted', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1485/","https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/logs/delete-log-group.html","https://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_DeleteLogGroup.html"]', e'equals("log.eventSource", "logs.amazonaws.com") &&
equals("log.eventName", "DeleteLogGroup")
', '2026-03-02 19:28:32.197934', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1200, 'AWS CloudTrail Log Updated', 2, 2, 1, 'Impact', 'T1565 - Data Manipulation', 'Identifies an update to an AWS log trail setting that specifies the delivery of log files', '["https://attack.mitre.org/tactics/TA0040/","https://attack.mitre.org/techniques/T1565/","https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1530/","https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_UpdateTrail.html","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/cloudtrail/update-trail.html"]', e'equals("log.eventSource", "cloudtrail.amazonaws.com") &&
equals("log.eventName", "UpdateTrail")
', '2026-03-02 19:28:33.440966', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1201, 'AWS RDS Snapshot Export', 3, 2, 2, 'Exfiltration', 'T1537 - Transfer Data to Cloud Account', 'Identifies the export of an Amazon Relational Database Service (RDS) Aurora database snapshot', '["https://attack.mitre.org/tactics/TA0010/","https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_StartExportTask.html"]', e'equals("log.eventSource", "rds.amazonaws.com") &&
equals("log.eventName", "StartExportTask")
', '2026-03-02 19:28:34.650471', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1202, 'AWS EC2 VM Export Failure', 3, 2, 2, 'Exfiltration', 'T1537 - Transfer Data to Cloud Account', 'Identifies an attempt to export an AWS EC2 instance. A virtual machine (VM) export may indicate an attempt to extract or exfiltrate information', '["https://attack.mitre.org/techniques/T1537/","https://attack.mitre.org/tactics/TA0010/","https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1005/","https://docs.aws.amazon.com/vm-import/latest/userguide/vmexport.html#export-instance"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
equals("log.eventName", "CreateInstanceExportTask")
', '2026-03-02 19:28:35.962704', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1203, 'AWS EC2 Snapshot Activity', 3, 2, 2, 'Exfiltration', 'T1537 - Transfer Data to Cloud Account', 'An attempt was made to modify AWS EC2 snapshot attributes. Snapshots are sometimes shared by threat actors in order to exfiltrate bulk data from an EC2 fleet. If the permissions were modified, verify the snapshot was not shared with an unauthorized or unexpected AWS account', '["https://attack.mitre.org/tactics/TA0010/","https://attack.mitre.org/techniques/T1537/","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/ec2/modify-snapshot-attribute.html","https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySnapshotAttribute.html"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
equals("log.eventName", "ModifySnapshotAttribute")
', '2026-03-02 19:28:37.214013', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1204, 'AWS EC2 Full Network Packet Capture Detected', 3, 2, 2, 'Exfiltration', 'T1020 - Automated Exfiltration', 'Identifies potential Traffic Mirroring in an Amazon Elastic Compute Cloud (EC2) instance. Traffic Mirroring is an Amazon VPC feature that you can use to copy network traffic from an elastic network interface. This feature can potentially be abused to exfiltrate sensitive data from unencrypted internal traffic', '["https://attack.mitre.org/tactics/TA0010/","https://attack.mitre.org/techniques/T1020/","https://attack.mitre.org/tactics/TA0009/","https://attack.mitre.org/techniques/T1074/","https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_TrafficMirrorFilter.html","https://github.com/easttimor/aws-incident-response"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
(equals("log.eventName", "CreateTrafficMirrorFilter") ||
equals("log.eventName", "CreateTrafficMirrorFilterRule") ||
equals("log.eventName", "CreateTrafficMirrorSession") ||
equals("log.eventName", "CreateTrafficMirrorTarget"))
', '2026-03-02 19:28:38.332137', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1205, 'AWS WAF Rule or Rule Group Deletion', 3, 2, 2, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies the deletion of a specified AWS Web Application Firewall (WAF) rule or rule group', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/waf/delete-rule-group.html","https://docs.aws.amazon.com/waf/latest/APIReference/API_waf_DeleteRuleGroup.html"]', e'oneOf("log.eventSource", ["waf.amazonaws.com", "waf-regional.amazonaws.com", "wafv2.amazonaws.com"]) &&
(equals("log.eventName", "DeleteRule") || equals("log.eventName", "DeleteRuleGroup"))
', '2026-03-02 19:28:39.542463', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1206, 'AWS WAF Access Control List Deletion', 3, 2, 2, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies the deletion of a specified AWS Web Application Firewall (WAF) access control list', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/waf-regional/delete-web-acl.html","https://docs.aws.amazon.com/waf/latest/APIReference/API_wafRegional_DeleteWebACL.html"]', e'oneOf("log.eventSource", ["waf.amazonaws.com", "waf-regional.amazonaws.com", "wafv2.amazonaws.com"]) &&
equals("log.eventName", "DeleteWebACL")
', '2026-03-02 19:28:40.676263', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1207, 'AWS S3 Bucket Configuration Deletion', 3, 2, 2, 'Defense Evasion', 'T1070 - Indicator Removal', 'Identifies the deletion of various Amazon Simple Storage Service (S3) bucket configuration components', '["https://attack.mitre.org/techniques/T1070/","https://attack.mitre.org/tactics/TA0005/","https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketPolicy.html","https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketReplication.html","https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketCors.html","https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketEncryption.html","https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketLifecycle.html"]', e'equals("log.eventSource", "s3.amazonaws.com") &&
oneOf("log.eventName", ["DeleteBucketPolicy", "DeleteBucketReplication",
"DeleteBucketCors", "DeleteBucketEncryption", "DeleteBucketLifecycle"])
', '2026-03-02 19:28:41.939392', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1208, 'AWS GuardDuty Detector Deletion', 3, 2, 2, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies the deletion of an Amazon GuardDuty detector. Upon deletion, GuardDuty stops monitoring the environment and all existing findings are lost', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/guardduty/delete-detector.html","https://docs.aws.amazon.com/guardduty/latest/APIReference/API_DeleteDetector.html"]', e'equals("log.eventSource", "guardduty.amazonaws.com") &&
equals("log.eventName", "DeleteDetector")
', '2026-03-02 19:28:43.252126', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1209, 'AWS EC2 Flow Log Deletion', 3, 2, 2, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies the deletion of one or more flow logs in AWS Elastic Compute Cloud (EC2). An adversary may delete flow logs in an attempt to evade defenses', '["https://awscli.amazonaws.com/v2/documentation/api/latest/reference/ec2/delete-flow-logs.html","https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DeleteFlowLogs.html","https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
equals("log.eventName", "DeleteFlowLogs")
', '2026-03-02 19:28:44.521354', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1210, 'AWS Configuration Recorder Stopped', 3, 2, 2, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies an AWS configuration change to stop recording a designated set of resources', '["https://awscli.amazonaws.com/v2/documentation/api/latest/reference/configservice/stop-configuration-recorder.html","https://docs.aws.amazon.com/config/latest/APIReference/API_StopConfigurationRecorder.html","https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/"]', e'equals("log.eventSource", "config.amazonaws.com") &&
equals("log.eventName", "StopConfigurationRecorder")
', '2026-03-02 19:28:45.916524', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1211, 'AWS Config Service Tampering', 3, 2, 2, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies attempts to delete an AWS Config Service resource. An adversary may tamper with Config services in order to reduce visibility into the security posture of an account and / or its workload instances', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/","https://docs.aws.amazon.com/config/latest/developerguide/how-does-config-work.html","https://docs.aws.amazon.com/config/latest/APIReference/API_Operations.html"]', e'equals("log.eventSource", "config.amazonaws.com") &&
oneOf("log.eventName", ["DeleteConfigRule", "DeleteOrganizationConfigRule",
"DeleteConfigurationAggregator", "DeleteConfigurationRecorder",
"DeleteConformancePack", "DeleteOrganizationConformancePack",
"DeleteDeliveryChannel", "DeleteRemediationConfiguration",
"DeleteRetentionConfiguration"])
', '2026-03-02 19:28:47.262856', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1212, 'AWS CloudTrail Log Suspended', 3, 2, 2, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies suspending the recording of AWS API calls and log file delivery for the specified trail. An adversary may suspend trails in an attempt to evade defenses', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/","https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_StopLogging.html","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/cloudtrail/stop-logging.html"]', e'equals("log.eventSource", "cloudtrail.amazonaws.com") &&
equals("log.eventName", "StopLogging")
', '2026-03-02 19:28:48.579210', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1213, 'AWS CloudTrail Log Deleted', 2, 3, 2, 'Defense Evasion', 'T1562 - Impair Defenses', 'Identifies the deletion of an AWS log trail. An adversary may delete trails in an attempt to evade defenses', '["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1562/","https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_DeleteTrail.html","https://awscli.amazonaws.com/v2/documentation/api/latest/reference/cloudtrail/delete-trail.html"]', e'equals("log.eventSource", "cloudtrail.amazonaws.com") &&
equals("log.eventName", "DeleteTrail")
', '2026-03-02 19:28:49.886394', true, true, 'origin', null, '[]', '["adversary.ip","adversary.user"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1214, 'AWS VPC Flow Log Anomalies', 3, 2, 2, 'Discovery', 'T1046 - Network Service Discovery', e'Detects anomalies in VPC Flow Logs configuration that could indicate attempts to hide malicious network activity. Monitors for deletion or modification of flow log configurations.

Next Steps:
1. Verify if the flow log changes were authorized by reviewing the AWS CloudTrail logs for the userIdentityArn
2. Check if the source IP address belongs to known administrative systems or jump boxes
3. Review other activities from the same source IP and user identity in the past 24-48 hours
4. Examine the affected VPC and its resources to understand the impact of disabled flow logging
5. If unauthorized, immediately re-enable flow logs and investigate what network activity may have occurred while logging was disabled
6. Review IAM permissions for the user/role that made these changes to ensure least privilege
7. Consider implementing preventive controls using AWS Config rules or SCPs to prevent unauthorized flow log modifications
', '["https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs.html","https://attack.mitre.org/techniques/T1046/"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
oneOf("log.eventName", ["DeleteFlowLogs", "CreateFlowLogs", "ModifyFlowLogsAttribute"]) &&
exists("log.sourceIPAddress") &&
equals("log.errorCode", "") &&
(
  equals("log.eventName", "DeleteFlowLogs") ||
  (equals("log.eventName", "CreateFlowLogs") && contains("log.requestParameters.deliverLogsStatus", "FAILED")) ||
  (equals("log.eventName", "ModifyFlowLogsAttribute") && equals("log.requestParameters.deliverLogsStatus", "INACTIVE"))
)
', '2026-03-02 19:28:51.243729', true, true, 'origin', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentity.arn","lastEvent.log.awsRegion"]', '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.sourceIPAddress","operator":"filter_term","value":"{{.log.sourceIPAddress}}"},{"field":"log.eventName","operator":"filter_term","value":"DeleteFlowLogs"}],"or":null,"within":"now-24h","count":2}]', null);
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1215, 'AWS Unusual API Call Patterns', 2, 1, 1, 'Execution', 'T1106 - Execution through API', e'Detects unusual API call patterns in AWS that may indicate unauthorized access or reconnaissance activities. This rule triggers when multiple sensitive API calls are made from the same source IP within a short time window, suggesting potential enumeration or discovery activities by attackers.

Next Steps:
- Review the source IP address and determine if it\'s authorized for AWS API access
- Check if the user identity associated with these calls is legitimate and expected
- Examine the specific API calls made to understand the reconnaissance pattern and scope
- Review CloudTrail logs for the full session to identify any successful exploitation attempts
- Check if any resources were modified or accessed following the reconnaissance activities
- Verify if the API calls originated from expected geographical locations
- Consider blocking the source IP if unauthorized activity is confirmed
- Implement additional monitoring for the affected account and resources
', '["https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-record-contents.html","https://attack.mitre.org/techniques/T1106/"]', e'exists("log.eventSource") &&
exists("log.sourceIPAddress") &&
exists("log.eventName") &&
(
  oneOf("log.eventName", ["DescribeSecurityGroups", "DescribeNetworkAcls", "DescribeVpcs", "DescribeSubnets", "DescribeRouteTables", "DescribeInstances", "DescribeSnapshots", "DescribeVolumes", "DescribeImages", "DescribeKeyPairs", "ListBuckets", "GetBucketAcl", "GetBucketPolicy", "ListAccessKeys", "ListUsers", "ListRoles", "ListPolicies", "GetAccountAuthorizationDetails", "GenerateCredentialReport", "GetCredentialReport"])
) &&
equals("log.errorCode", "")
', '2026-03-02 19:28:52.646279', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.sourceIPAddress","operator":"filter_term","value":"{{.log.sourceIPAddress}}"},{"field":"log.eventName","operator":"filter_match","value":"Describe List Get Generate"}],"or":null,"within":"now-10m","count":50}]', '["lastEvent.log.sourceIPAddress"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1216, 'AWS STS Token Abuse Detection', 3, 3, 1, 'Privilege Escalation', 'T1078.004 - Cloud Accounts', e'Detects potential abuse of AWS STS AssumeRole operations. This rule identifies when roles are assumed from unusual IP addresses or when there are multiple role assumptions in a short time period, which could indicate lateral movement or privilege escalation. The rule specifically flags AssumeRole operations performed without MFA authentication.

Next Steps:
1. Verify if the source IP address belongs to your organization\'s known IP ranges
2. Check if the assumed role is appropriate for the user or service that initiated the request
3. Review the user identity and verify if this is expected behavior for this account
4. Examine CloudTrail logs for other suspicious activities from the same source IP or user identity
5. If unauthorized, immediately revoke the temporary credentials and review IAM policies
6. Consider implementing MFA requirements for sensitive role assumptions
7. Review and potentially restrict the trust policy for the assumed role
', '["https://docs.aws.amazon.com/IAM/latest/UserGuide/cloudtrail-integration.html","https://attack.mitre.org/techniques/T1078/004/","https://www.elastic.co/security-labs/exploring-aws-sts-assumeroot"]', e'equals("log.eventSource", "sts.amazonaws.com") &&
equals("log.eventName", "AssumeRole") &&
equals("log.errorCode", "") &&
!equals("log.userIdentitySessionContextAttributesMfaAuthenticated", "true") &&
!equals("log.userIdentityType", "AWSService") &&
!contains("log.userAgent", "aws-sdk") &&
!contains("log.userAgent", "Botocore")
', '2026-03-02 19:28:53.957418', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.sourceIPAddress","operator":"filter_term","value":"{{.log.sourceIPAddress}}"}],"or":null,"within":"now-15m","count":20}]', '["lastEvent.log.userIdentityArn","adversary.ip"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1217, 'AWS Systems Manager Session Abuse', 3, 3, 2, 'Lateral Movement', 'T1021 - Remote Services', e'Detects suspicious use of AWS Systems Manager (SSM) for remote access including StartSession, SendCommand, and SendSSHPublicKey. Attackers use SSM to establish interactive sessions or execute commands on EC2 instances without requiring direct SSH access or security group changes.

Next Steps:
1. Verify the IAM principal initiating the SSM session is authorized for remote access
2. Review the target instance(s) and confirm legitimate operational need
3. Check the commands sent via SendCommand for suspicious payloads
4. Review the source IP address for unusual origins
5. Examine the timing of the session for off-hours access
6. If unauthorized, terminate active sessions and review instance for compromise
7. Implement SSM session logging to S3 and CloudWatch for audit trails
8. Restrict SSM access using IAM policies with condition keys
', '["https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager.html","https://attack.mitre.org/techniques/T1021/"]', e'equals("log.eventSource", "ssm.amazonaws.com") &&
oneOf("log.eventName", ["StartSession", "ResumeSession", "SendCommand", "StartAutomationExecution"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:28:55.353960', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventSource","operator":"filter_term","value":"ssm.amazonaws.com"}],"or":null,"within":"now-30m","count":5}]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1218, 'AWS Security Group Modifications', 2, 3, 2, 'Defense Evasion', 'T1562.007 - Impair Defenses: Disable or Modify Cloud Firewall', e'Detects modifications to AWS security groups that could weaken network security posture. Monitors for changes that add permissive rules or remove restrictive rules, particularly those allowing unrestricted access (0.0.0.0/0 or ::/0).

Next Steps:
1. Review the security group change details in CloudTrail logs
2. Verify if the change was authorized and follows security policies
3. Check the user/role that made the modification
4. Assess if the new rules expose sensitive resources
5. If unauthorized, immediately revert the changes
6. Review other recent activities from the same source IP or user
7. Consider implementing preventive controls via AWS Config rules or SCPs
', '["https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-record-contents.html","https://attack.mitre.org/techniques/T1562/007/"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
oneOf("log.eventName", ["AuthorizeSecurityGroupIngress", "AuthorizeSecurityGroupEgress", "RevokeSecurityGroupIngress", "RevokeSecurityGroupEgress", "CreateSecurityGroup", "DeleteSecurityGroup", "ModifySecurityGroupRules"]) &&
exists("log.sourceIPAddress") &&
equals("log.errorCode", "") &&
(
  contains("log.requestParameters.ipPermissions.ipProtocol", "-1") ||
  contains("log.requestParameters.ipPermissions.cidrIp", "0.0.0.0/0") ||
  contains("log.requestParameters.ipPermissions.ipv6CidrIp", "::/0") ||
  contains("log.requestParameters.ipRanges.cidrIp", "0.0.0.0/0") ||
  contains("log.requestParameters.ipv6Ranges.cidrIpv6", "::/0")
)
', '2026-03-02 19:28:56.662706', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.sourceIPAddress","operator":"filter_term","value":"{{.log.sourceIPAddress}}"},{"field":"log.eventSource","operator":"filter_term","value":"ec2.amazonaws.com"}],"or":null,"within":"now-30m","count":3}]', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentity.arn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1219, 'AWS Secrets Manager Suspicious Access Pattern', 3, 2, 1, 'Credential Access', 'T1552.004 - Private Keys', e'Detects unusual access patterns to AWS Secrets Manager that could indicate credential theft or unauthorized access attempts. This rule monitors for multiple GetSecretValue or BatchGetSecretValue operations from the same source within a short time window, which may indicate an attacker attempting to harvest credentials.

Next Steps:
1. Review the CloudTrail logs for the identified user/role to understand which secrets were accessed
2. Check if the accessing identity has legitimate business need for these secrets
3. Verify if the access pattern matches normal usage for this identity
4. Review the source IP addresses and locations for anomalies
5. If unauthorized, immediately rotate the accessed secrets and review IAM permissions
6. Check for any subsequent API calls using potentially compromised credentials
', '["https://docs.aws.amazon.com/secretsmanager/latest/userguide/monitoring-cloudtrail.html","https://attack.mitre.org/techniques/T1552/004/"]', 'equals("log.eventSource", "secretsmanager.amazonaws.com") && (equals("log.eventName", "GetSecretValue") || equals("log.eventName", "BatchGetSecretValue")) && equals("log.errorCode", "")', '2026-03-02 19:28:58.057840', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventName","operator":"filter_term","value":"GetSecretValue"}],"or":null,"within":"now-10m","count":10},{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventName","operator":"filter_term","value":"BatchGetSecretValue"}],"or":null,"within":"now-10m","count":5}]', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentityArn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1220, 'AWS S3 Bulk Data Exfiltration Detected', 2, 1, 1, 'Data Exfiltration', 'T1530 - Data from Cloud Storage Object', e'Detects bulk GetObject operations from S3 buckets indicating potential data exfiltration. When an attacker gains access to AWS credentials, they may attempt to download large amounts of data from S3 buckets in a short period.

Next Steps:
1. Identify the S3 bucket(s) targeted and classify the data sensitivity
2. Review the IAM principal performing the downloads and verify authorization
3. Check the source IP address for known threat indicators
4. Examine the volume and types of objects downloaded
5. Verify if this matches any legitimate data processing or backup patterns
6. If unauthorized, revoke the credentials used and block the source IP
7. Enable S3 server access logging for affected buckets
8. Review S3 bucket policies and tighten access controls
', '["https://docs.aws.amazon.com/AmazonS3/latest/userguide/cloudtrail-logging.html","https://attack.mitre.org/techniques/T1530/"]', e'equals("log.eventSource", "s3.amazonaws.com") &&
equals("log.eventName", "GetObject") &&
equals("log.errorCode", "")
', '2026-03-02 19:28:59.249633', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventName","operator":"filter_term","value":"GetObject"}],"or":null,"within":"now-15m","count":100}]', '["adversary.user","lastEvent.log.requestParameters.bucketName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1221, 'AWS Route 53 DNS Hijacking Attempt', 3, 3, 3, 'Initial Access', 'T1584.002 - Compromise Infrastructure: DNS Server', e'Detects potential DNS hijacking attempts through unauthorized changes to Route 53 DNS records. This rule monitors for ChangeResourceRecordSets operations that could indicate an attacker modifying DNS entries to redirect traffic.

Next Steps:
1. Review the CloudTrail logs to identify what DNS records were modified and the specific changes made
2. Verify if the user identity making the changes is authorized to modify Route 53 records
3. Check if the source IP address is from a known and trusted location
4. Review the modified DNS records to ensure they point to legitimate resources
5. If unauthorized, immediately revert the DNS changes and rotate the compromised credentials
6. Enable MFA for all users with Route 53 permissions
7. Consider implementing AWS Config rules to monitor Route 53 changes
8. Review and restrict IAM policies for Route 53 access
', '["https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/logging-using-cloudtrail.html","https://attack.mitre.org/techniques/T1584/002/"]', 'equals("log.eventSource", "route53.amazonaws.com") && equals("log.eventName", "ChangeResourceRecordSets") && equals("log.errorCode", "")', '2026-03-02 19:29:00.592800', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventName","operator":"filter_term","value":"ChangeResourceRecordSets"}],"or":null,"within":"now-30m","count":10}]', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentityArn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1222, 'AWS Mass Resource Deletion', 1, 3, 3, 'Impact', 'T1485 - Data Destruction', e'Detects mass deletion of AWS resources which could indicate destructive attack or insider threat. Monitors for multiple delete operations across various AWS services within a 10-minute window.

Next Steps:
1. Immediately identify the user/role performing the deletions through the userIdentity.arn field
2. Review CloudTrail logs to determine the scope and specific resources being deleted
3. Contact the user to verify if these actions are authorized
4. If unauthorized, immediately revoke the user\'s AWS credentials and permissions
5. Enable MFA delete protection on critical resources if not already enabled
6. Consider implementing SCPs (Service Control Policies) to prevent mass deletions
7. Review and restore deleted resources from backups if necessary
8. Document the incident and update IAM policies to prevent future occurrences
', '["https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-concepts.html","https://attack.mitre.org/techniques/T1485/"]', '(contains("log.eventName", "Delete") || contains("log.eventName", "Terminate") || contains("log.eventName", "Remove")) && equals("log.errorCode", "") && !equals("log.eventSource", "s3.amazonaws.com")', '2026-03-02 19:29:01.863432', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentity.arn","operator":"filter_term","value":"{{.log.userIdentity.arn}}"}],"or":null,"within":"now-10m","count":15}]', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentity.arn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1223, 'Lambda Function Privilege Escalation', 3, 3, 2, 'Privilege Escalation', 'T1548 - Abuse Elevation Control Mechanism', e'Detects potential privilege escalation through Lambda functions when IAM policies are attached to roles or users that can be exploited. This may indicate an attacker attempting to escalate privileges by attaching administrative policies to Lambda execution roles.

Next Steps:
1. Verify if the policy attachment was authorized and follows change management procedures
2. Review the attached policy permissions, especially if AdministratorAccess or IAMFullAccess policies were attached
3. Check the Lambda function\'s code and recent invocations for suspicious activity
4. Review CloudTrail logs for other IAM changes by the same user/role
5. Validate if the Lambda function legitimately requires the elevated permissions
6. Consider revoking the policy attachment if unauthorized and investigate the source of the change
7. Check for any unusual Lambda function executions following the policy attachment
8. Review the user/role history for previous privilege escalation attempts
', '["https://bishopfox.com/blog/privilege-escalation-in-aws","https://attack.mitre.org/techniques/T1548/"]', e'equals("log.eventSource", "iam.amazonaws.com") &&
oneOf("log.eventName", ["AttachRolePolicy", "AttachUserPolicy"]) &&
equals("log.errorCode", "") &&
(contains("log.requestParameters.policyArn", "AdministratorAccess") || contains("log.requestParameters.policyArn", "IAMFullAccess"))
', '2026-03-02 19:29:03.262145', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentity.arn","operator":"filter_term","value":"{{.log.userIdentity.arn}}"},{"field":"log.eventName","operator":"filter_term","value":"AttachRolePolicy"}],"or":null,"within":"now-1h","count":2}]', '["lastEvent.log.requestParameters.roleArn","lastEvent.log.userIdentity.arn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1224, 'AWS IAM Privilege Escalation Path Detection', 3, 3, 2, 'Privilege Escalation', 'T1098 - Account Manipulation', e'Detects IAM actions commonly chained for privilege escalation including PassRole, CreatePolicyVersion, AttachUserPolicy, and AssumeRole. Attackers exploit these API calls to elevate their permissions within an AWS environment by creating new policy versions with elevated privileges or assuming roles with higher access.

Next Steps:
1. Review the IAM actions performed and determine if they constitute a privilege escalation chain
2. Check the IAM principal and verify their authorized permission level
3. Examine the policy document or role being targeted for overly permissive access
4. Verify through change management if these IAM changes were approved
5. If unauthorized, immediately revert the IAM changes and rotate credentials
6. Review IAM Access Analyzer findings for excessive permissions
7. Implement permission boundaries to limit privilege escalation paths
8. Enable AWS CloudTrail Insights for anomaly detection on IAM operations
', '["https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_manage.html","https://attack.mitre.org/techniques/T1098/","https://rhinosecuritylabs.com/aws/aws-privilege-escalation-methods-mitigation/"]', e'equals("log.eventSource", "iam.amazonaws.com") &&
oneOf("log.eventName", ["CreatePolicyVersion", "SetDefaultPolicyVersion", "AttachUserPolicy", "AttachGroupPolicy", "AttachRolePolicy", "PutUserPolicy", "PutGroupPolicy", "PutRolePolicy", "AddUserToGroup"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:04.569454', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventSource","operator":"filter_term","value":"iam.amazonaws.com"}],"or":null,"within":"now-30m","count":3}]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1225, 'AWS IAM Backdoor Creation Attempts', 3, 3, 2, 'Privilege Escalation', 'T1136.003 - Create Account: Cloud Account', e'Detects potential IAM backdoor creation attempts through suspicious IAM user creation, access key generation, or policy attachment activities that could provide persistent access.

Next Steps:
1. Review the source IP address and user identity performing these actions
2. Verify if the IAM user creation and policy attachments are authorized
3. Check CloudTrail logs for the complete sequence of IAM actions from this source
4. Review the policies attached to determine if they grant excessive permissions
5. If unauthorized, immediately disable the created user and revoke access keys
6. Investigate other activities from the same source IP or user identity
', '["https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_manage.html","https://attack.mitre.org/techniques/T1136/003/"]', 'equals("log.eventSource", "iam.amazonaws.com") && oneOf("log.eventName", ["CreateUser", "CreateAccessKey", "AttachUserPolicy", "PutUserPolicy", "CreateLoginProfile"]) && equals("log.errorCode", "")', '2026-03-02 19:29:05.887758', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.sourceIPAddress","operator":"filter_term","value":"{{.log.sourceIPAddress}}"},{"field":"log.eventSource","operator":"filter_term","value":"iam.amazonaws.com"}],"or":null,"within":"now-30m","count":3}]', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentity.arn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1226, 'AWS Cross-Account Access Anomalies', 3, 3, 1, 'Unauthorized Access', 'T1550.001 - Use Alternate Authentication Material: Application Access Token', e'Detects anomalous cross-account access patterns in AWS that may indicate account compromise or privilege escalation. Monitors for AssumeRole activities across different accounts where the assumed role ARN does not match the originating account ID, potentially indicating unauthorized cross-account access.

Next Steps:
1. Verify if the cross-account access was authorized by checking AWS IAM policies and trust relationships
2. Review the source IP address to determine if it matches known corporate IP ranges or expected locations
3. Check the assumed role permissions to understand what access was granted
4. Look for any subsequent API calls made using the assumed role credentials
5. Contact the owner of the originating account to verify if the activity was legitimate
6. If unauthorized, immediately revoke the assumed role session and review all IAM trust policies
', '["https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-user-identity.html","https://attack.mitre.org/techniques/T1550/001/"]', e'equals("log.eventSource", "sts.amazonaws.com") &&
equals("log.eventName", "AssumeRole") &&
exists("log.userIdentityAccountId") &&
exists("log.responseElementsAssumedRoleUserArn") &&
exists("log.sourceIPAddress") &&
equals("log.errorCode", "") &&
!contains(safe(log.responseElementsAssumedRoleUserArn, ""), safe(log.userIdentityAccountId, ""))
', '2026-03-02 19:29:07.145599', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.sourceIPAddress","operator":"filter_term","value":"{{.log.sourceIPAddress}}"},{"field":"log.eventName","operator":"filter_term","value":"AssumeRole"}],"or":null,"within":"now-15m","count":15}]', '["lastEvent.log.responseElementsAssumedRoleUserArn","lastEvent.log.sourceIPAddress","lastEvent.log.userIdentityAccountId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1227, 'AWS Console Login Impossible Travel Detection', 3, 2, 1, 'Credential Access', 'T1078.004 - Valid Accounts: Cloud Accounts', e'Detects AWS Console logins from different geographic locations within a short time window, indicating potential credential compromise. This correlation identifies when the same user authenticates from different countries within 30 minutes, which would be physically impossible.

Next Steps:
1. Review the login locations and IP addresses for the affected user
2. Check if the user employs a VPN or proxy service that could explain different geolocations
3. Verify with the user whether both login sessions are legitimate
4. Review the actions performed in each session for suspicious activity
5. If unauthorized, immediately disable the user account and rotate credentials
6. Enable MFA if not already configured for the affected account
7. Review CloudTrail logs for any API calls made during the suspicious session
8. Check for concurrent sessions from the compromised account
', '["https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-event-reference-user-identity.html","https://attack.mitre.org/techniques/T1078/004/"]', e'equals("log.eventName", "ConsoleLogin") &&
equals("log.responseElements.ConsoleLogin", "Success") &&
exists("origin.geolocation.countryCode")
', '2026-03-02 19:29:08.633139', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventName","operator":"filter_term","value":"ConsoleLogin"},{"field":"origin.geolocation.countryCode","operator":"must_not_term","value":"{{.origin.geolocation.countryCode}}"}],"or":null,"within":"now-30m","count":1}]', '["adversary.user","adversary.geolocation.countryCode"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1228, 'AWS CloudFormation Stack Deletion', 1, 3, 3, 'Impact', 'T1485 - Data Destruction', e'Detects deletion of CloudFormation stacks which could indicate destructive actions by an attacker or unauthorized infrastructure changes. This rule monitors for DeleteStack operations that could result in loss of critical infrastructure. Multiple stack deletions from the same user within a short time frame may indicate malicious activity.

Next Steps:
1. Verify if the stack deletion was authorized and part of planned maintenance
2. Check the user identity who performed the deletion and validate if they should have this permission
3. Review CloudTrail logs for other destructive actions by the same user
4. Examine what resources were contained in the deleted stack
5. If unauthorized, immediately revoke the user\'s permissions and investigate for other compromise indicators
6. Consider enabling stack termination protection on critical stacks
7. Review and implement least privilege access policies for CloudFormation operations
', '["https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/logging-cloudformation-api-calls.html","https://attack.mitre.org/techniques/T1485/"]', 'equals("log.eventSource", "cloudformation.amazonaws.com") && equals("log.eventName", "DeleteStack") && equals("log.errorCode", "")', '2026-03-02 19:29:09.950883', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventSource","operator":"filter_term","value":"cloudformation.amazonaws.com"}],"or":null,"within":"now-30m","count":5}]', '["lastEvent.log.userIdentityAccountId","lastEvent.log.userIdentityArn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1229, 'AWS SSO Suspicious Activities', 3, 3, 1, 'Credential Access, Defense Evasion, Persistence', 'T1556.006 - Modify Authentication Process: Multi-Factor Authentication', e'Detects suspicious AWS SSO activities including multiple failed login attempts, unusual permission set assignments, or SSO configuration changes that could indicate an attempt to compromise single sign-on authentication mechanisms.

Next Steps:
1. Review the source IP address and user identity associated with the suspicious SSO activity
2. Check if the user\'s credentials may have been compromised
3. Verify if the permission set changes or role assumptions were authorized
4. Review CloudTrail logs for other suspicious activities from the same IP or user
5. If unauthorized, immediately revoke the affected SSO sessions and reset user credentials
6. Consider implementing additional MFA requirements for SSO administrative actions
', '["https://docs.aws.amazon.com/singlesignon/latest/userguide/what-is.html","https://attack.mitre.org/techniques/T1556/006/"]', e'equals("log.eventSource", "sso.amazonaws.com") &&
(
  oneOf("log.eventName", ["AssumeRoleWithSAML", "CreatePermissionSet", "AttachManagedPolicyToPermissionSet", "DeletePermissionSet", "PutInlinePolicyToPermissionSet"]) ||
  exists("log.errorCode")
)
', '2026-03-02 19:29:11.211276', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.sourceIPAddress","operator":"filter_term","value":"{{.log.sourceIPAddress}}"},{"field":"log.eventSource","operator":"filter_term","value":"sso.amazonaws.com"}],"or":null,"within":"now-30m","count":10}]', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentity.principalId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1230, 'AWS SSM SendCommand Remote Execution', 2, 3, 2, 'Execution', 'T1059 - Command and Scripting Interpreter', e'Detects AWS Systems Manager SendCommand API calls which allow remote command execution on EC2 instances. Attackers with sufficient IAM permissions abuse SSM to execute commands across multiple instances without SSH/RDP access.

Next Steps:
1. Review the command document and parameters sent to the instances
2. Verify the user identity has legitimate need for remote command execution
3. Check the target instances and whether they should be managed via SSM
4. Review the command output for suspicious activity
5. Correlate with instance-level logs for the executed commands
6. If unauthorized, cancel pending commands and investigate the user
', '["https://docs.aws.amazon.com/systems-manager/latest/userguide/run-command.html","https://attack.mitre.org/techniques/T1059/"]', e'equals("log.eventSource", "ssm.amazonaws.com") &&
oneOf("log.eventName", ["SendCommand", "StartSession"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:12.747199', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventSource","operator":"filter_term","value":"ssm.amazonaws.com"}],"or":null,"within":"now-30m","count":5}]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1231, 'AWS SecurityHub Finding Evasion', 2, 3, 2, 'Defense Evasion', 'T1562 - Impair Defenses', e'Detects attempts to suppress or manipulate AWS SecurityHub findings through BatchUpdateFindings, DeleteInsight, or UpdateFindings operations. Attackers use these to hide evidence of their activities from security monitoring.

Next Steps:
1. Review what findings were modified or suppressed and their severity
2. Check the user identity performing these actions for legitimacy
3. Investigate whether this was part of legitimate finding management workflow
4. Review the original findings that were suppressed for security relevance
5. Check for other defense evasion activities from the same user or IP
6. Restore suppressed findings if the action was unauthorized
7. Review SecurityHub aggregation and automation rules for tampering
', '["https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-findings.html","https://attack.mitre.org/techniques/T1562/"]', e'equals("log.eventSource", "securityhub.amazonaws.com") &&
oneOf("log.eventName", ["BatchUpdateFindings", "DeleteInsight", "UpdateFindings", "DeleteActionTarget"]) &&
equals("log.errorCode", "") &&
!equals("log.userIdentityType", "AWSService")
', '2026-03-02 19:29:13.879754', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityArn","operator":"filter_term","value":"{{.log.userIdentityArn}}"},{"field":"log.eventSource","operator":"filter_term","value":"securityhub.amazonaws.com"}],"or":null,"within":"now-30m","count":5}]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1232, 'AWS Golden SAML Attack Detected', 3, 3, 3, 'Credential Access', 'T1556 - Modify Authentication Process', e'Detects potential Golden SAML attacks by monitoring for AssumeRoleWithSAML calls combined with UpdateSAMLProvider or CreateSAMLProvider operations. Golden SAML was the technique used in the SolarWinds attack, allowing attackers to forge SAML tokens and assume any role in the AWS account.

Next Steps:
1. Immediately investigate the SAML provider configuration changes and the AssumeRoleWithSAML calls
2. Verify the SAML provider metadata document has not been tampered with
3. Check if the identity provider federation trust is legitimate
4. Review the assumed role ARN and session name for suspicious values
5. Correlate with identity provider logs to verify the SAML assertion was legitimately issued
6. Check for lateral movement or privilege escalation after the role assumption
7. If unauthorized, rotate the SAML provider trust and revoke all active sessions
', '["https://www.cyberark.com/resources/threat-research-blog/golden-saml-newly-discovered-attack-technique-forges-authentication-to-cloud-apps","https://attack.mitre.org/techniques/T1556/"]', e'equals("log.eventSource", "sts.amazonaws.com") &&
equals("log.eventName", "AssumeRoleWithSAML") &&
equals("log.errorCode", "")
', '2026-03-02 19:29:15.012526', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.userIdentityAccountId","operator":"filter_term","value":"{{.log.userIdentityAccountId}}"}],"or":[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.eventName","operator":"filter_term","value":"UpdateSAMLProvider"}],"or":null,"within":"now-24h","count":1},{"indexPattern":"v11-log-aws-*","with":[{"field":"log.eventName","operator":"filter_term","value":"CreateSAMLProvider"}],"or":null,"within":"now-24h","count":1}],"within":"now-24h","count":1}]', '["adversary.user","lastEvent.log.requestParametersRoleArn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1233, 'AWS ECS Task Credential Endpoint Query', 3, 2, 1, 'Credential Access', 'T1552.005 - Unsecured Credentials: Cloud Instance Metadata API', e'Detects queries to the ECS task credential endpoint that may indicate container credential theft. Attackers who compromise a container can query the credential endpoint to steal IAM role credentials attached to the ECS task.

Next Steps:
1. Identify the ECS task and cluster involved
2. Check if the API calls were from expected task processes
3. Review the IAM role attached to the task for excessive permissions
4. Check for lateral movement using the stolen credentials
5. Investigate the container image and running processes for compromise
6. If unauthorized, rotate the task role credentials and investigate the container
', '["https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html","https://attack.mitre.org/techniques/T1552/005/"]', e'equals("log.eventSource", "ecs.amazonaws.com") &&
oneOf("log.eventName", ["DescribeTaskDefinition", "RunTask", "StartTask"]) &&
equals("log.errorCode", "") &&
equals("log.userIdentityType", "AssumedRole") &&
contains("log.userIdentityArn", ":assumed-role/")
', '2026-03-02 19:29:16.411056', true, true, 'origin', null, '[{"indexPattern":"v11-log-aws-*","with":[{"field":"log.sourceIPAddress","operator":"filter_term","value":"{{.log.sourceIPAddress}}"},{"field":"log.eventSource","operator":"filter_term","value":"ecs.amazonaws.com"}],"or":null,"within":"now-30m","count":5}]', '["adversary.user","lastEvent.log.sourceIPAddress"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1234, 'AWS WAF and Shield Rule Modifications', 2, 3, 3, 'Defense Evasion', 'T1562.001 - Impair Defenses: Disable or Modify Tools', e'Detects modifications or deletions of AWS WAF web ACLs and rules which could be used to remove web application protections before launching an attack. Attackers may disable WAF rules to allow malicious traffic or remove rate limiting protections.

Next Steps:
1. Verify if the WAF modification was authorized through change management
2. Review the specific rules or web ACL that were modified or deleted
3. Check the IAM principal and source IP performing the change
4. Assess whether any web applications are now exposed without WAF protection
5. Review recent web application logs for attack attempts following the WAF change
6. If unauthorized, restore the WAF configuration from backup or AWS Config
7. Implement IAM policies to restrict WAF modification access
8. Enable AWS Config rules to monitor WAF configuration compliance
', '["https://docs.aws.amazon.com/waf/latest/developerguide/waf-chapter.html","https://attack.mitre.org/techniques/T1562/001/"]', e'(equals("log.eventSource", "wafv2.amazonaws.com") || equals("log.eventSource", "waf.amazonaws.com")) &&
oneOf("log.eventName", ["DeleteWebACL", "DeleteRule", "DeleteRuleGroup", "UpdateWebACL", "DeleteIPSet", "DeleteRegexPatternSet", "DisassociateWebACL"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:17.677995', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1235, 'AWS S3 Bucket Public Exposure', 3, 2, 1, 'Collection', 'T1530 - Data from Cloud Storage Object', e'Detects S3 bucket configuration changes that expose buckets to public access, including ACL modifications and public access block removal. This can lead to unauthorized data exposure and potential data breach.

Next Steps:
1. Immediately review the affected S3 bucket permissions and revert unauthorized changes
2. Check CloudTrail logs for the source of the configuration change
3. Verify if the change was authorized by examining change management tickets
4. Review bucket contents to determine sensitivity of potentially exposed data
5. Enable S3 Block Public Access at the account level to prevent future exposures
6. Implement bucket policies that restrict public access
7. Enable S3 access logging for the affected bucket
8. Consider implementing AWS Config rules to monitor S3 bucket permissions
', '["https://docs.aws.amazon.com/AmazonS3/latest/userguide/access-control-block-public-access.html","https://attack.mitre.org/techniques/T1530/"]', e'equals("log.eventSource", "s3.amazonaws.com") &&
(equals("log.eventName", "PutBucketAcl") ||
 equals("log.eventName", "PutBucketPublicAccessBlock") ||
 equals("log.eventName", "DeleteBucketPublicAccessBlock") ||
 equals("log.eventName", "PutObjectAcl")) &&
equals("log.errorCode", "") &&
(oneOf("log.requestParameters.x-amz-acl", ["public-read", "public-read-write"]) ||
 contains("log.requestParameters.acl", "AllUsers") ||
 equals("log.eventName", "DeleteBucketPublicAccessBlock"))
', '2026-03-02 19:29:19.074343', true, true, 'origin', null, '[]', '["lastEvent.log.requestParameters.bucketName","lastEvent.log.userIdentity.accessKeyId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1236, 'AWS Root Account Usage Without MFA', 3, 3, 2, 'Defense Evasion, Persistence, Privilege Escalation, Initial Access', 'T1078.004 - Valid Accounts: Cloud Accounts', e'Detects usage of AWS root account without Multi-Factor Authentication (MFA). Root account usage should be avoided for daily operations and must always use MFA when necessary. This is a critical security violation as the root account has unrestricted access to all AWS resources.

Next Steps:
1. Immediately verify if this root account activity is authorized
2. Check the source IP address and user agent for suspicious patterns
3. Review all actions performed by the root account during this session
4. Enable MFA for the root account if not already enabled
5. Investigate why MFA was not used (possible bypass or configuration issue)
6. Consider implementing SCPs to restrict root account usage
7. Audit all recent root account activities for potential compromise
8. Review AWS CloudTrail logs for complete session details
9. Check for any privilege escalation attempts or unusual API calls
10. Implement additional monitoring for root account activities
', '["https://docs.aws.amazon.com/IAM/latest/UserGuide/id_root-user.html","https://attack.mitre.org/techniques/T1078/004/"]', e'equals("log.userIdentityType", "Root") &&
!equals("log.userIdentitySessionContextAttributesMfaAuthenticated", "true") &&
equals("log.errorCode", "")
', '2026-03-02 19:29:20.389256', true, true, 'origin', null, '[]', '["lastEvent.log.sourceIPAddress","lastEvent.log.userIdentityAccountId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1237, 'AWS RDS Snapshot Sharing for Data Exfiltration', 3, 2, 1, 'Data Exfiltration', 'T1537 - Transfer Data to Cloud Account', e'Detects RDS snapshot sharing modifications that could be used for data exfiltration. Attackers may share database snapshots with external AWS accounts to exfiltrate sensitive data, or make snapshots public to download them from an attacker-controlled account.

Next Steps:
1. Identify which RDS snapshot was shared and its data sensitivity
2. Verify the target AWS account ID the snapshot was shared with
3. Check if the snapshot was made public (shared with all AWS accounts)
4. Verify the change was authorized through data governance procedures
5. If unauthorized, immediately remove the snapshot sharing and rotate database credentials
6. Review the database contents for sensitive or regulated data
7. Check for other snapshot operations by the same principal
8. Implement preventive SCPs to restrict snapshot sharing
', '["https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ShareSnapshot.html","https://attack.mitre.org/techniques/T1537/"]', e'equals("log.eventSource", "rds.amazonaws.com") &&
oneOf("log.eventName", ["ModifyDBSnapshotAttribute", "ModifyDBClusterSnapshotAttribute"]) &&
equals("log.errorCode", "") &&
(contains("log.requestParameters", "attributeName\\":\\"restore") ||
 contains("log.requestParameters", "all"))
', '2026-03-02 19:29:21.659417', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.dBSnapshotIdentifier"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1238, 'RDS Security Group Changes', 3, 2, 2, 'Defense Evasion', 'T1562.007 - Impair Defenses: Disable or Modify Cloud Firewall', e'Detects modifications to RDS database security groups that could expose databases to unauthorized access. This includes adding new ingress rules or modifying existing security group configurations that may compromise database security.

Next Steps:
1. Review the security group changes to determine if they were authorized
2. Check if the changes expose RDS instances to public internet (0.0.0.0/0)
3. Verify the identity of the user who made the changes and confirm authorization
4. Review CloudTrail logs for additional suspicious activities by the same user
5. If unauthorized, immediately revert the security group changes
6. Consider implementing preventive controls using AWS Config rules or SCPs
7. Document the incident and update change management procedures if needed
', '["https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/logging-using-cloudtrail.html","https://attack.mitre.org/techniques/T1562/007/"]', e'equals("log.eventSource", "rds.amazonaws.com") &&
oneOf("log.eventName", ["AuthorizeDBSecurityGroupIngress", "CreateDBSecurityGroup", "DeleteDBSecurityGroup", "RevokeDBSecurityGroupIngress"]) &&
equals("log.errorCode", "") &&
contains("log.requestParameters", "0.0.0.0/0")
', '2026-03-02 19:29:23.141256', true, true, 'origin', null, '[]', '["lastEvent.log.eventName","lastEvent.log.userIdentityArn"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1239, 'AWS Lambda Function URL Backdoor Creation', 3, 3, 1, 'Persistence', 'T1059 - Command and Scripting Interpreter', e'Detects creation of Lambda function URLs which provide direct HTTPS endpoints to Lambda functions. Attackers may create function URL configurations as backdoor endpoints for persistent access, command execution, or data exfiltration without requiring API Gateway.

Next Steps:
1. Review the Lambda function code for malicious payloads or backdoor functionality
2. Check the function URL authentication type (AWS_IAM vs NONE)
3. Verify the IAM principal creating the function URL has authorization
4. Review the Lambda function\'s execution role for excessive permissions
5. Check if CORS settings allow cross-origin requests from unauthorized domains
6. If unauthorized, delete the function URL configuration and the Lambda function
7. Review CloudTrail logs for the function\'s invocation history
8. Implement SCPs to restrict Lambda function URL creation to authorized roles
', '["https://docs.aws.amazon.com/lambda/latest/dg/lambda-urls.html","https://attack.mitre.org/techniques/T1059/"]', e'equals("log.eventSource", "lambda.amazonaws.com") &&
oneOf("log.eventName", ["CreateFunctionUrlConfig", "UpdateFunctionUrlConfig"]) &&
equals("log.errorCode", "") &&
contains("log.requestParameters", "\\"authType\\":\\"NONE\\"")
', '2026-03-02 19:29:24.488238', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.functionName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1240, 'AWS GuardDuty High-Severity Finding', 3, 3, 2, 'Defense Evasion, Persistence, Privilege Escalation, Initial Access', 'T1078 - Valid Accounts', e'Detects high-severity findings from AWS GuardDuty indicating potential security threats such as malicious activity, unauthorized access, or compromised instances. GuardDuty analyzes CloudTrail events, VPC Flow Logs, and DNS logs to identify threats.

Next Steps:
1. Review the specific GuardDuty finding details including the threat type and affected resources
2. Check the affected AWS account and region for any unauthorized changes
3. Investigate the source IP addresses and user activities associated with the finding
4. Review CloudTrail logs for suspicious API calls around the time of the finding
5. If compromise is confirmed, rotate credentials and review IAM permissions
6. Enable AWS CloudTrail logging if not already enabled
7. Consider implementing AWS Security Hub for centralized security findings
', '["https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings.html","https://attack.mitre.org/techniques/T1078/"]', 'equals("log.eventSource", "guardduty.amazonaws.com") && greaterOrEqual("log.severity", 7)', '2026-03-02 19:29:25.890455', true, true, 'origin', null, '[]', '["lastEvent.log.accountId","lastEvent.log.region","lastEvent.log.type"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1241, 'AWS ECS/EKS Container Abuse Detection', 3, 3, 2, 'Execution', 'T1610 - Deploy Container', e'Detects suspicious container operations in AWS ECS and EKS including registration of task definitions with privileged mode, host network access, or sensitive volume mounts. Attackers may deploy malicious containers to execute code, escalate privileges, or access host resources.

Next Steps:
1. Review the task definition or pod specification for privileged settings
2. Check if the container image is from a trusted registry
3. Verify the IAM principal registering the task has authorization
4. Examine container environment variables for embedded credentials
5. Review network mode settings for host network access
6. Check volume mounts for access to sensitive host paths
7. If unauthorized, deregister the task definition and terminate running tasks
8. Implement OPA/Gatekeeper policies to prevent privileged containers
', '["https://docs.aws.amazon.com/AmazonECS/latest/developerguide/security.html","https://attack.mitre.org/techniques/T1610/"]', e'(equals("log.eventSource", "ecs.amazonaws.com") &&
 equals("log.eventName", "RegisterTaskDefinition") &&
 equals("log.errorCode", "") &&
 (contains("log.requestParameters", "\\"privileged\\":true") ||
  contains("log.requestParameters", "\\"networkMode\\":\\"host\\"") ||
  contains("log.requestParameters", "\\"pidMode\\":\\"host\\"") ||
  contains("log.requestParameters", "/var/run/docker.sock") ||
  contains("log.requestParameters", "/etc/shadow") ||
  contains("log.requestParameters", "/root")))
', '2026-03-02 19:29:27.249861', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1244, 'EBS Snapshot Sharing Violations', 3, 2, 1, 'Exfiltration', 'T1537 - Transfer Data to Cloud Account', e'Detects potential data exfiltration through EBS snapshot sharing. Monitors for CreateSnapshot followed by ModifySnapshotAttribute events that could indicate an attacker creating snapshots and sharing them with external accounts or making them public.

Next Steps:
1. Identify the affected snapshot ID and volume ID from the event details
2. Review the account or group the snapshot was shared with in requestParameters
3. Verify if the sharing was authorized by checking with the account owner
4. If unauthorized, immediately revoke the snapshot sharing permissions
5. Review CloudTrail logs for other suspicious activities by the same user/role
6. Check if the source volume contains sensitive data
7. Consider implementing preventive controls using IAM policies or SCPs to restrict snapshot sharing
', '["https://securitylabs.datadoghq.com/cloud-security-atlas/attacks/sharing-ebs-snapshot/","https://attack.mitre.org/techniques/T1537/"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
equals("log.eventName", "ModifySnapshotAttribute") &&
equals("log.errorCode", "") &&
(
  contains("log.requestParameters", "CREATE_VOLUME_PERMISSION") ||
  contains("log.requestParameters", "createVolumePermission") ||
  (contains("log.requestParameters", "group") && contains("log.requestParameters", "all"))
)
', '2026-03-02 19:29:31.400560', true, true, 'origin', null, '[]', '["lastEvent.log.responseElements.snapshotId","lastEvent.log.userIdentity.accountId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1245, 'AWS CloudTrail Logging Disabled', 3, 3, 2, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', e'Detects attempts to disable CloudTrail logging which could be used to hide malicious activities and evade detection. CloudTrail provides audit logs of all AWS API calls and is critical for security monitoring and compliance.

Next Steps:
1. Immediately verify if the CloudTrail modification was authorized
2. Check the user identity ({{log.userIdentityArn}}) and source IP ({{log.sourceIPAddress}}) for legitimacy
3. Review CloudTrail configuration to ensure logging is re-enabled for all regions
4. Check for any suspicious API calls made around the time of this event
5. If unauthorized, investigate what activities may have occurred while logging was disabled
6. Consider implementing SCPs or IAM policies to prevent CloudTrail modifications
7. Verify if the trail was configured for multi-region logging
8. Check if compliance requirements have been violated
9. Review AWS Config rules for CloudTrail compliance
10. Enable CloudTrail Insights for anomaly detection
', '["https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-concepts.html","https://attack.mitre.org/techniques/T1562/008/"]', e'equals("log.eventSource", "cloudtrail.amazonaws.com") &&
oneOf("log.eventName", ["StopLogging", "DeleteTrail"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:32.704760', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.name"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1246, 'AWS CloudTrail Event Selector Manipulation', 3, 3, 2, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', e'Detects manipulation of CloudTrail event selectors which can be used to selectively exclude data events from logging. Attackers may use PutEventSelectors to disable logging of S3, Lambda, or DynamoDB data events to hide their data exfiltration or manipulation activities.

Next Steps:
1. Immediately review the CloudTrail event selector configuration changes
2. Verify if the change was authorized through change management
3. Check which data event types were excluded from logging
4. Review the IAM principal ({{log.userIdentityArn}}) making the change
5. Restore event selectors to include all critical data event types
6. Review recent activities that may have been hidden by the selector change
7. Implement SCPs to prevent modification of CloudTrail event selectors
8. Enable AWS Config rules to monitor CloudTrail configuration compliance
', '["https://docs.aws.amazon.com/awscloudtrail/latest/userguide/logging-data-events-with-cloudtrail.html","https://attack.mitre.org/techniques/T1562/008/"]', e'equals("log.eventSource", "cloudtrail.amazonaws.com") &&
oneOf("log.eventName", ["PutEventSelectors", "PutInsightSelectors"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:34.072331', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.trailName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1247, 'AWS TruffleHog Credential Scanning Detected', 3, 1, 0, 'Credential Access', 'T1552 - Unsecured Credentials', e'Detects TruffleHog credential scanning tool activity in AWS CloudTrail logs identified by its characteristic user agent string. TruffleHog is used to scan for exposed secrets and credentials in AWS environments.

Next Steps:
1. Identify the source of the TruffleHog scan (user, role, source IP)
2. Determine if this is an authorized security assessment
3. Review what resources were accessed during the scan
4. If unauthorized, revoke the credentials used and investigate
5. Check for any credentials that may have been discovered and exfiltrated
6. Review IAM policies to limit enumeration permissions
', '["https://github.com/trufflesecurity/trufflehog","https://attack.mitre.org/techniques/T1552/"]', e'contains("log.userAgent", "trufflehog") ||
contains("log.userAgent", "TruffleHog")
', '2026-03-02 19:29:35.286155', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.sourceIPAddress"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1248, 'AWS SSO Identity Provider Configuration Changed', 3, 3, 2, 'Credential Access', 'T1556 - Modify Authentication Process', e'Detects changes to AWS SSO identity provider configuration including directory association and external IdP configuration changes. Modifying the identity provider can enable impersonation of any user in the organization.

Next Steps:
1. Verify the IdP configuration change was part of an approved change request
2. Check the user identity performing the change for proper authorization
3. Review the new IdP configuration for legitimacy
4. Check for suspicious sign-ins following the IdP change
5. Validate the external IdP settings against the organization\'s known configuration
6. If unauthorized, revert the IdP configuration immediately
7. Audit all SSO-based sessions created after the change
', '["https://docs.aws.amazon.com/singlesignon/latest/userguide/manage-your-identity-source.html","https://attack.mitre.org/techniques/T1556/"]', e'equals("log.eventSource", "sso.amazonaws.com") &&
oneOf("log.eventName", ["AssociateDirectory", "DisassociateDirectory", "EnableExternalIdPConfiguration", "DisableExternalIdPConfiguration", "UpdateExternalIdPConfiguration"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:36.728046', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1249, 'AWS EC2 Snapshot Shared with External Account', 3, 2, 1, 'Exfiltration', 'T1537 - Transfer Data to Cloud Account', e'Detects ModifySnapshotAttribute API calls that share EBS snapshots with external AWS accounts. Attackers exfiltrate data by sharing snapshots containing sensitive data with accounts they control.

Next Steps:
1. Identify the shared snapshot and its contents
2. Verify the target account ID is a known and authorized account
3. Check the user identity and source IP for legitimacy
4. Review the snapshot\'s source volume for sensitive data
5. If unauthorized, immediately remove the sharing permission
6. Check for other snapshots shared around the same time
7. Review IAM policies to restrict snapshot sharing
', '["https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-modifying-snapshot-permissions.html","https://attack.mitre.org/techniques/T1537/"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
equals("log.eventName", "ModifySnapshotAttribute") &&
equals("log.errorCode", "") &&
contains("log.requestParameters", "createVolumePermission")
', '2026-03-02 19:29:38.092147', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.snapshotId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1250, 'AWS S3 Bucket Versioning Suspended', 2, 3, 3, 'Impact', 'T1490 - Inhibit System Recovery', e'Detects when S3 bucket versioning is suspended via PutBucketVersioning. Disabling versioning is a cloud ransomware preparation technique that prevents recovery of overwritten or deleted objects from version history.

Next Steps:
1. Verify if the versioning suspension was authorized and part of a change request
2. Identify the affected S3 bucket and its contents sensitivity
3. Check for subsequent DeleteObject or PutObject calls that would overwrite data
4. Review the user identity and source IP for legitimacy
5. Re-enable versioning immediately if unauthorized
6. Check if MFA Delete was configured on the bucket
7. Review S3 bucket lifecycle policies for any recent changes
', '["https://docs.aws.amazon.com/AmazonS3/latest/userguide/Versioning.html","https://attack.mitre.org/techniques/T1490/"]', e'equals("log.eventSource", "s3.amazonaws.com") &&
equals("log.eventName", "PutBucketVersioning") &&
equals("log.errorCode", "") &&
contains("log.requestParameters", "Suspended")
', '2026-03-02 19:29:39.433504', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.bucketName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1251, 'AWS RDS Database Restored as Publicly Accessible', 3, 2, 1, 'Exfiltration', 'T1020 - Automated Exfiltration', e'Detects RestoreDBInstanceFromDBSnapshot or RestoreDBClusterFromSnapshot API calls that may create publicly accessible database instances. Attackers use this to exfiltrate database contents by restoring a snapshot as a publicly accessible instance.

Next Steps:
1. Verify the database restore was authorized and part of an approved workflow
2. Check if the restored instance is publicly accessible
3. Review the security group and subnet configuration of the restored instance
4. Identify the source snapshot and its data sensitivity
5. If publicly accessible and unauthorized, immediately modify the instance to private
6. Check for any connections to the restored instance from external IPs
', '["https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_RestoreFromSnapshot.html","https://attack.mitre.org/techniques/T1020/"]', e'equals("log.eventSource", "rds.amazonaws.com") &&
oneOf("log.eventName", ["RestoreDBInstanceFromDBSnapshot", "RestoreDBClusterFromSnapshot"]) &&
equals("log.errorCode", "") &&
contains("log.requestParameters", "publiclyAccessible")
', '2026-03-02 19:29:40.621788', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1252, 'AWS KMS Key Material Import or Deletion', 3, 3, 3, 'Impact', 'T1486 - Data Encrypted for Impact', e'Detects ImportKeyMaterial or DeleteImportedKeyMaterial operations on AWS KMS. These are extremely rare operations that could indicate cloud ransomware preparation - an attacker importing their own key material to encrypt data and then deleting the imported key material to make decryption impossible without paying ransom.

Next Steps:
1. Immediately verify if this KMS key material operation was authorized
2. Identify the KMS key ID affected and all resources encrypted with it
3. Check the user identity and source IP for legitimacy
4. Review if any data encryption operations followed this event
5. Verify backup key material exists and is securely stored
6. If unauthorized, immediately disable the KMS key and investigate data integrity
7. Check for associated DeleteImportedKeyMaterial calls that would prevent decryption
', '["https://docs.aws.amazon.com/kms/latest/developerguide/importing-keys.html","https://attack.mitre.org/techniques/T1486/"]', e'equals("log.eventSource", "kms.amazonaws.com") &&
oneOf("log.eventName", ["ImportKeyMaterial", "DeleteImportedKeyMaterial"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:41.886072', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1253, 'AWS IAM Login Profile Modified for Another User', 3, 3, 1, 'Persistence', 'T1098 - Account Manipulation', e'Detects UpdateLoginProfile API calls where the modifier is potentially different from the target user. This can indicate an attacker changing another user\'s password to maintain access or enable console login for a compromised programmatic-only account.

Next Steps:
1. Verify the password change was authorized by comparing the caller with the target user
2. Check if the target user reported any password issues
3. Review subsequent console logins for the target user
4. Investigate the source IP and user agent for the modification
5. Check if the target user\'s MFA was also modified
6. If unauthorized, reset the password and rotate all credentials for the target user
', '["https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_passwords_admin-change-user.html","https://attack.mitre.org/techniques/T1098/"]', e'equals("log.eventSource", "iam.amazonaws.com") &&
equals("log.eventName", "UpdateLoginProfile") &&
equals("log.errorCode", "")
', '2026-03-02 19:29:43.067349', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.userName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1254, 'AWS Glue Development Endpoint Privilege Escalation', 3, 3, 2, 'Privilege Escalation', 'T1078.004 - Valid Accounts: Cloud Accounts', e'Detects creation or modification of AWS Glue development endpoints which can be used for privilege escalation. An attacker can pass an IAM role to a Glue dev endpoint and then use it to execute code with that role\'s permissions, potentially escalating privileges.

Next Steps:
1. Verify the Glue development endpoint creation was authorized
2. Check the IAM role passed to the endpoint for excessive permissions
3. Review the user identity and source IP for legitimacy
4. Check for code execution on the development endpoint
5. If unauthorized, delete the endpoint and investigate the passed role
6. Review IAM policies to restrict Glue PassRole permissions
', '["https://rhinosecuritylabs.com/aws/escalating-aws-iam-privileges-undocumented-codestar-api/","https://attack.mitre.org/techniques/T1078/004/"]', e'equals("log.eventSource", "glue.amazonaws.com") &&
oneOf("log.eventName", ["CreateDevEndpoint", "UpdateDevEndpoint"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:44.373009', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1255, 'AWS EC2 Instance Startup Script Modified', 3, 3, 2, 'Execution', 'T1059 - Command and Scripting Interpreter', e'Detects ModifyInstanceAttribute API calls targeting userData, which modifies the EC2 instance startup script. Attackers use this for persistent code execution with root/SYSTEM privileges on instance start.

Next Steps:
1. Identify the instance and review the new userData content (base64 decode it)
2. Check if the modification was part of an authorized deployment
3. Review the user identity and source IP for legitimacy
4. Check if the instance was stopped and restarted after the modification
5. Examine the instance for signs of compromise
6. If unauthorized, remove the userData and investigate the instance
', '["https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html","https://attack.mitre.org/techniques/T1059/"]', e'equals("log.eventSource", "ec2.amazonaws.com") &&
equals("log.eventName", "ModifyInstanceAttribute") &&
equals("log.errorCode", "") &&
contains("log.requestParameters", "userData")
', '2026-03-02 19:29:45.644330', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.requestParameters.instanceId"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1256, 'AWS Console GetSigninToken Abuse', 3, 2, 1, 'Credential Access', 'T1550.001 - Use Alternate Authentication Material: Application Access Token', e'Detects GetSigninToken API calls which allow federated users to get a console sign-in token. Attackers abuse this to pivot from programmatic credentials to console access, potentially bypassing MFA requirements.

Next Steps:
1. Verify the federated sign-in was from an authorized user and application
2. Check the source IP and user agent for legitimacy
3. Review what actions were performed in the console after sign-in
4. Correlate with the original AssumeRole call that created the federated session
5. Check if MFA was properly enforced on the original authentication
6. If unauthorized, revoke the federated session immediately
', '["https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_enable-console-custom-url.html","https://attack.mitre.org/techniques/T1550/001/"]', e'equals("log.eventSource", "signin.amazonaws.com") &&
equals("log.eventName", "GetSigninToken") &&
equals("log.errorCode", "")
', '2026-03-02 19:29:46.741585', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.sourceIPAddress"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1257, 'AWS Config Service Disabled', 2, 3, 2, 'Defense Evasion', 'T1562.008 - Impair Defenses: Disable Cloud Logs', e'Detects attempts to disable AWS Config by stopping the configuration recorder or deleting the delivery channel. AWS Config tracks resource configuration changes and compliance, and disabling it is a common defense evasion technique.

Next Steps:
1. Verify if the AWS Config modification was authorized
2. Check the user identity and source IP for legitimacy
3. Review any infrastructure changes made while Config was disabled
4. Re-enable AWS Config immediately if unauthorized
5. Check for other defense evasion activities from the same source
6. Review CloudTrail for API calls during the gap in Config monitoring
7. Verify compliance requirements are being met
', '["https://docs.aws.amazon.com/config/latest/developerguide/stop-start-recorder.html","https://attack.mitre.org/techniques/T1562/008/"]', e'equals("log.eventSource", "config.amazonaws.com") &&
oneOf("log.eventName", ["DeleteDeliveryChannel", "StopConfigurationRecorder", "DeleteConfigurationRecorder"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:47.949489', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.eventName"]');
INSERT INTO public.utm_correlation_rules (id, rule_name, rule_confidentiality, rule_integrity, rule_availability, rule_category, rule_technique, rule_description, rule_references_def, rule_definition_def, rule_last_update, rule_active, system_owner, rule_adversary, rule_deduplicate_by_def, rule_after_events_def, rule_group_by_def) VALUES (1258, 'AWS Backup Deletion for Impact', 1, 3, 3, 'Impact', 'T1490 - Inhibit System Recovery', e'Detects deletion of AWS Backup vaults, recovery points, or backup plans which could indicate an attacker attempting to prevent recovery after a destructive attack. Removing backups is a common precursor to ransomware or data destruction campaigns.

Next Steps:
1. Immediately verify if the backup deletion was authorized
2. Check the IAM principal and source IP performing the deletion
3. Assess which backup vaults and recovery points were affected
4. Verify if backup vault lock policies were in place and bypassed
5. Check for concurrent destructive activities (instance termination, data deletion)
6. If unauthorized, immediately implement vault lock on remaining backup vaults
7. Restore deleted recovery points from any cross-region or cross-account copies
8. Enable MFA delete on backup vaults and implement SCPs to restrict backup deletion
', '["https://docs.aws.amazon.com/aws-backup/latest/devguide/API_Operations.html","https://attack.mitre.org/techniques/T1490/"]', e'equals("log.eventSource", "backup.amazonaws.com") &&
oneOf("log.eventName", ["DeleteBackupVault", "DeleteRecoveryPoint", "DeleteBackupPlan", "DeleteBackupVaultAccessPolicy", "DeleteBackupVaultLockConfiguration"]) &&
equals("log.errorCode", "")
', '2026-03-02 19:29:49.135879', true, true, 'origin', null, '[]', '["adversary.user","lastEvent.log.eventName"]');
