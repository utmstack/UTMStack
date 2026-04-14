INSERT INTO public.utm_correlation_rules (id,rule_name,rule_confidentiality,rule_integrity,rule_availability,rule_category,rule_technique,rule_description,rule_references_def,rule_definition_def,rule_last_update,rule_active,system_owner,rule_adversary,rule_deduplicate_by_def,rule_after_events_def,rule_group_by_def) VALUES (1356,'Windows audit log was cleared',1,2,3,'Defense Evasion','T1070.001 - Indicator Removal: Clear Windows Event Logs','Detects when the Windows audit log (Security event log) has been cleared. Adversaries may clear event logs to remove evidence of an intrusion.','["https://attack.mitre.org/techniques/T1070/001/"]','equals("log.eventCode", 1102)','2026-04-01 17:58:34.362',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1357,'Windows: Possible Brute Force Attack',2,2,3,'Credential Access','T1110 - Brute Force','This rule is triggered when a pattern of repeated and rapid login attempts from the same IP address or source is detected. These login attempts may target specific user accounts or services in an attempt to crack passwords through automated brute force. The purpose of this rule is to identify possible malicious unauthorized access attempts and prevent a brute force attack against the system.','["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1110/"]','equals("log.eventCode", 4625)','2026-04-01 17:58:36.932',true,true,'origin','["origin.host","target.user"]','[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"{{.log.eventCode}}"},{"field":"target.user.keyword","operator":"filter_term","value":"{{.target.user}}"}],"or":null,"within":"now-5m","count":10}]',NULL), (1358,'Windows: Multiple Logon Failure Followed by Logon Success',2,2,3,'Credential Access','T1110 - Brute Force','This rule is triggered when a sequence of multiple failed login attempts followed immediately by a successful login from the same IP address or source is detected. This unusual sequence of events may indicate a possible unauthorized access attempt using a brute force or password guessing technique. The purpose of this rule is to identify suspicious patterns of login activity and alert you to potential unauthorized access attempts.','["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1110/"]','equals("log.eventCode", 4624)','2026-04-01 17:58:39.500',true,true,'origin','["origin.ip","target.user"]','[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":4625},{"field":"target.user.keyword","operator":"filter_term","value":"{{.target.user}}"}],"or":null,"within":"now-5m","count":10}]',NULL), (1359,'Windows: LSASS Memory Dump Handle Access',2,3,3,'Credential Access','T1003.001 - OS Credential Dumping: LSASS Memory','Identifies handle requests for the Local Security Authority Subsystem Service (LSASS) object access with specific access masks that many tools with a capability to dump memory to disk use (0x1fffff, 0x1010, 0x120089). This rule is tool agnostic as it has been validated against a host of various LSASS dump tools such as SharpDump, Procdump, Mimikatz, Comsvcs etc. It detects this behavior at a low level and does not depend on a specific tool or dump file name.','["https://attack.mitre.org/tactics/TA0006/","https://attack.mitre.org/techniques/T1003/","https://attack.mitre.org/techniques/T1003/001/"]','equals("log.eventCode", 4656) && regexMatch("log.eventDataObjectName", "(:\\Windows\\System32\\lsass.exe|\\Device\\HarddiskVolume[A-Za-z?:\\]([A-Za-z?])?\\Windows\\System32\\lsass.exe)") && !regexMatch("log.eventDataProcessName", "(:\\Program Files\\(.+).exe|:\\Program Files (x86)\\(.+).exe|:\\Windows\\system32\\wbem\\WmiPrvSE.exe|:\\Windows\\System32\\dllhost.exe|:\\Windows\\System32\\svchost.exe|:\\Windows\\System32\\msiexec.exe|:\\ProgramData\\Microsoft\\Windows Defender\\(.+).exe|:\\Windows\\explorer.exe)") && oneOf("log.eventDataAccessMask", ["2097151", "4112", "1040", "1180185", "2031615"])','2026-04-01 17:58:41.906',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1360,'Windows: User logged using Remote Desktop Connection from loopback address, possible exploit over reverse tunneling using stolen credentials',3,2,1,'Credential Access','T1021.001 - Remote Services: Remote Desktop Protocol','Adversaries may use Valid Accounts to log into a computer using the Remote Desktop Protocol (RDP). The adversary may then perform actions as the logged-on user.','["https://attack.mitre.org/techniques/T1021/001/"]','equals("log.eventDataLogonType", "10") && oneOf("origin.ip", ["::1", "127.0.0.1"]) && oneOf("log.eventCode", [528, 540, 673, 4624, 4769])','2026-04-01 17:58:44.018',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1361,'Windows: Printer driver failed to load, possible remote code execution using PrinterNightmare exploit: CVE-2021-34527',3,2,1,'Lateral Movement','T1210 - Exploitation of Remote Services','Adversaries may exploit remote services to gain unauthorized access to internal systems once inside of a network.  Exploitation of a software vulnerability occurs when an adversary takes advantage of a programming error in a program,  service, or within the operating system software or kernel itself to execute adversary-controlled code.  A common goal for post-compromise exploitation of remote services is for lateral movement to enable access to a remote system.','["https://attack.mitre.org/techniques/T1210/"]','equals("log.eventCode", 4663) && regexMatch("log.eventDataObjectName", "\\Windows\\System32\\spool\\drivers\\") && regexMatch("log.eventDataObjectName", "\\.(dll|exe)$") && !contains("log.eventDataProcessName", "spoolsv.exe")','2026-04-01 17:58:46.512',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1362,'Windows: Persistence via PowerShell profile',2,3,1,'Persistence','T1546.013 - Event Triggered Execution: PowerShell Profile','Identifies the creation or modification of a PowerShell profile. PowerShell profile is a script that is executed when PowerShell starts to customize the user environment, which can be abused by attackers to persist in a environment where PowerShell is common.','["https://attack.mitre.org/tactics/TA0003/","https://attack.mitre.org/techniques/T1098/002/"]','equals("log.eventCode", 4663) && regexMatch("log.eventDataObjectName", "(:\\Users\\.*\\Documents\\(WindowsPowerShell|PowerShell)\\.*profile\\.ps1$|:\\Windows\\System32\\WindowsPowerShell\\.*profile\\.ps1$)")','2026-04-01 17:58:49.133',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1363,'Windows: Suspicious PrintSpooler Service Executable File Creation',2,3,1,'Privilege Escalation','T1068 - Exploitation for Privilege Escalation','Detects attempts to exploit privilege escalation vulnerabilities related to the Print Spooler service. For more information refer to the following CVE''s - CVE-2020-1048, CVE-2020-1337 and CVE-2020-1300 and verify that the impacted system is patched','["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1068/"]','contains("log.eventDataProcessName", "spoolsv.exe") && !regexMatch("log.eventDataProcessPath", "^C:\\\\Windows\\\\System32\\\\spoolsv.exe$")','2026-04-01 17:58:51.519',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1364,'Windows: Suspicious Print Spooler SPL File Created',1,3,2,'Privilege Escalation','T1068 - Exploitation for Privilege Escalation','Detects attempts to exploit privilege escalation vulnerabilities related to the Print Spooler service including CVE-2020-1048 and CVE-2020-1337.','["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1068/"]','!regexMatch("log.eventDataProcessName", "(spoolsv.exe|printfilterpipelinesvc.exe|PrintIsolationHost.exe|splwow64.exe|msiexec.exe|poqexec.exe)") && regexMatch("log.eventDataObjectName", ":\\Windows\\System32\\spool\\PRINTERS\\")
','2026-04-01 17:58:54.019',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1365,'Windows: Possible ransomware attack detected. Multiple File Deletion.',1,3,2,'Impact','T1486 - Data Encrypted for Impact','Detects potential ransomware activity by monitoring multiple file write/modification events (Event ID 4663)  with write access masks in user directories within a short timeframe. Modern ransomware typically  encrypts files in-place rather than deleting them, making write access monitoring more effective  than deletion monitoring alone.','["https://attack.mitre.org/tactics/TA0040/"]','equals("log.eventCode", 4663) &&
oneOf("log.eventDataAccessMask", ["2", "4", "6"]) &&
!(regexMatch("log.eventDataProcessName", "(?i).*(trustedinstaller|svchost|wuauclt|msiexec|windows10upgrade|setuphost|tiworker|dism).*")) &&
regexMatch("log.eventDataObjectName", "(?i).*\\\\(users|documents|desktop|downloads|pictures|videos|music)\\\\.*") &&
!(regexMatch("log.eventDataObjectName", "(?i).*(\\\\windows\\\\|\\\\program files|\\\\programdata\\\\|\\\\temp\\\\|\\\\appdata\\\\local\\\\temp|\\\\softwaredistribution\\\\|\\\\winsxs\\\\|\\\\logs\\\\|\\\\prefetch\\\\).*")) &&
!(regexMatch("log.eventDataObjectName", "(?i).*\\.(tmp|log|etl|dmp|pf|evtx|cache|dat|bak)$")) &&
regexMatch("log.eventDataObjectName", "(?i).*\\.(doc[x]?|xls[x]?|ppt[x]?|pdf|txt|jpg|jpeg|png|gif|bmp|mp4|avi|mp3|zip|rar|7z)$")
','2026-04-01 17:58:56.644',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"4663"},{"field":"target.user.keyword","operator":"filter_term","value":"{{.target.user}}"}],"or":null,"within":"now-5m","count":50}]','["origin.ip","target.user"]');INSERT INTO public.utm_correlation_rules (id,rule_name,rule_confidentiality,rule_integrity,rule_availability,rule_category,rule_technique,rule_description,rule_references_def,rule_definition_def,rule_last_update,rule_active,system_owner,rule_adversary,rule_deduplicate_by_def,rule_after_events_def,rule_group_by_def) VALUES (1366,'Windows: Possible ransomware attack detected. Ransomware Note Creation.',3,3,2,'Impact','T1486 - Data Encrypted for Impact','Ransomware, is a type of malware that prevents users from accessing their system or  personal files and requires payment of a ransom in order to gain access to them again. Identifies  ransomware attempts. A known ransomware note file has been detected, potentially indicating an active ransomware infection.','["https://attack.mitre.org/tactics/TA0040/"]','equals("log.eventCode", 4663) && regexMatch("log.EventDataFileName", "(README_TO_RESTORE_FILES|INSTRUCTION_TO_GET_FILES_BACK|HOW_TO_DECRYPT_FILES|DECRYPT_INSTRUCTION|RECOVER_INSTRUCTION|RESTORE_FILES|READ_ME_NOW|YOUR_FILES_ARE_ENCRYPTED|IMPORTANT_INSTRUCTIONS|NOTICE|DECRYPT_YOUR_FILES|HOW_TO_RESTORE_FILES|HELP_DECRYPT|RECOVERY_FILE|RECOVER-FILES|INSTRUCTION)\\.(txt|html|php)$")','2026-04-01 17:58:59.193',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"{{.log.eventCode}}"},{"field":"log.EventDataFileName.keyword","operator":"filter_term","value":"{{.log.EventDataFileName}}"}],"or":null,"within":"now-60s","count":5}]','["origin.ip","target.user"]'), (1367,'Windows: Possible ransomware attack detected. Unusual File Extensions.',3,3,3,'Impact','T1486 - Data Encrypted for Impact','Ransomware, is a type of malware that prevents users from accessing their system or  personal files and requires payment of a ransom in order to gain access to them again. Identifies  ransomware attempts. Files with unusual file extensions have been detected, potentially indicating encrypted files created by ransomware.','["https://attack.mitre.org/tactics/TA0040/"]','equals("log.eventCode", 4663) && regexMatch("log.eventDataFileName", "\\.(7z\\.encrypted|aaa|abcd|abtc|acc|aes256|aes_ni|aes256ctr|aes256encrypted|aes_gcm|ajp|alcatraz_locked|alfa|amnesia[1-9]?|amsi|apocalypse666|armadillo|arrow|asasin|asi|atom128|auditor|aurora|autoit|avastvirusinfo|avenger|av666|azero|barak|barrax|bart|beef|beetle|bip|bit|bitcoin|bl2r|blackblock|blackmail|blast|blind|bmw|boot|braincrypt[3-8]?|broken|btcware|budak|bull|buydecryptor|cakl|calipso|calum|carote|cats|cbf|cccmn|ccrypt|cerber[3-9]?|chifrator|chimera|ciphered|crypted|clover|cobra|codnat3|combo|comrade|conficker|coot|cpt|crash|crimson|cry|crypt\\d{3,4}|crypt38|crypt72|crypt888|cryptinfinite|cryptolocker|cryptowall|cryptxxx|crypz|csfr|csone|ctb[1-4]|ctb-locker|ctrsa|cryptowin|cube|dcrpt|ddtf|dr2|dragon|dried|druid|ducky|ecrpt|eeta|etr|ee|f[0-9]{3,4}|flock|grt|grt[0-9]+|grtlock|gwz|h3ll|hades|hakunamatata|hallucinating|happy|harmful|harrow|havoc|headdesk|helpdecrypt|hermes|hidden|hideous|hijack|hilda|hitler|hjg|hmpl|hrosas|hsdf|hushed|hwrm|ihsdj|ikarus|ikasir|ikayed|ill|imbtc|img|encrypted|improved|indrik|injected|innocent|insane|interesante|jungle|kaos|karl|katana|kimcilware|kin|kiratos|kiss|kjh|locked|locky|lokf|losers|lukitus|m3g4c0rtx|m4n1fest0|m4s4g3|maas|madmax|mafi|magic|maktub|malware|manamecrypt|mandelbrot|manic|matrix|max|md5|medusa|mega|melme|merry|mesmerize|metropolitan|mikey|mikibackup|milarepa.lotos@aol.com|mirror|mmnn|mole|monro|mosk|muslat|n1n1n1|nabr|napoleon|narrow|nasoh|nataniel|neitrino|neras|nlah|nosu|novasof|nozelesn|nuclear|nwa|nymaim|obelisk|off|offwhite|ogdo|omega|omerta|onion|ooss|opencode|openme|opqz|osiris|otx|p3rf0rm4|pabluk|pack14|packagetrackr@india.com|packrat|pahd|panda|pandemic|pandora|pansy|paradise|paris|paym3|paymer|payms|pcap|pclock|peet|pelikan|penis|petya|pewcrypt|phoenix|photominr|phobos|phps|pirated|pluto|po1|point|poop|potato|pr0tect|preppy|princesa|princess|prosper|prosperity|prq|pshy|pumas|pumax|pure|purple|purpler|pwnd|pysa|q9q9|qbtex|qiuu|qkkd|qscx|qtyop|quimera|r2d2|r5a|rabit|radman|raid10|rainbow|rakhni|rambo|ramses|rat|rcrypted|react|reactor|realtek|reaper|redlion|redmat|redrum|rekt|remk|removal|remsec|remy|renaming|revenge|rezuc|rhino|ribd|rich|rip|rire|rizonesoft@protonmail.ch|rk|rmdir|robinhood|rocke|rogue|roldat|rolin|ronzware|rosenquist|rotten|roza|rpcminer|rsalive|rumba|run|rxx|uak|udjvu|unlock92|unlckr|upd|urcp|usam|usbc|v8|vag|vandt|varasto|vault|vauw|vb|ve|vendetta|venom|veracrypt|versiegelt|veton|vhd|vindows|violate|virus|vivin|vk_677|vma|vmx|volcano|vorasto|vorphal|vos|vscrypt|vxl|w4b|wakanda|wannacash|wannacry|wanted|war|wasted|wcry|weapologize|webmafia|weird|weui|whatthefuck|whistler|white|whitenoise|whiterabbit|whorus|why!decryptor|wicked|wildfire|windows10|windows7|windows8|windowsupdate|winlock|wipe|wisconsin|wizard|wlu|woolger|worm|wormfubuki|wow|wpencrypt|wq!decryptor|wrui|wtg|x1881|x3m|xampug|xdata|xencrypt|xfiles|xhelper|xlr|xman|xmd|xmd|xtbl|xtbl|xtr|xtt|xtz|xyz|yakuza|yatron|ybn|year|yellow|yheq|yty|yuke|yxo|yyto|z3|zatrov|zax|zbot|zbt|zbt|zeppelin|zerber|zet|zet|zfj|zfj|zimbra|zip|zix|zlz|zobm|zoh|zorro|zphs)$")','2026-04-01 17:59:01.291',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventCode","operator":"filter_term","value":"4663"},{"field":"target.user.keyword","operator":"filter_term","value":"{{.target.user}}"}],"or":null,"within":"now-5m","count":20}]','["origin.ip","target.user"]'), (1368,'Windows: Remote File Download via Desktopimgdownldr Utility',2,3,1,'Command and Control','T1105 - Ingress Tool Transfer','Identifies the desktopimgdownldr utility being used to download a remote file. An adversary may use desktopimgdownldr to download arbitrary files as an alternative to certutil.','["https://attack.mitre.org/tactics/TA0011/","https://attack.mitre.org/techniques/T1105/"]','equals("log.eventCode", 4663) && contains("log.eventDataProcessName", "desktopimgdownldr.exe") && !regexMatch("log.eventDataObjectName", "(:\\Windows\\Web\\|:\\ProgramData\\Microsoft\\Windows\\SystemData\\)") && regexMatch("log.eventDataObjectName", "\\.(exe|dll|bat|ps1|zip|rar|7z)$")','2026-04-01 17:59:03.732',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1369,'Windows: New Windows Service Created to start from windows root path. Suspicious event as the binary may have been dropped using Windows Admin Shares',1,2,3,'Execution','T1021.002 - Remote Services: SMB/Windows Admin Shares','Adversaries may use Valid Accounts to interact with a remote network share using Server Message Block (SMB).  The adversary may then perform actions as the logged-on user.','["https://attack.mitre.org/techniques/T1021/002/"]','regexMatch("log.eventDataImagePath", "(^%systemroot%\\(.+)\\(.+).exe)") && equals("log.eventCode", 7045) && oneOf("log.channel", ["system", "System"])','2026-04-01 17:59:06.525',true,true,'target',NULL,'[]','["origin.host","target.user"]'), (1370,'Windows: Suspicious Managed Code Hosting Process',3,3,2,'Defense Evasion','T1055 - Process Injection','Identifies a suspicious managed code hosting process which could indicate code injection or other form of suspicious code execution.','["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1055/"]','regexMatch("log.eventDataProcessName", "(wscript.exe|cscript.exe|mshta.exe|wmic.exe|regsvr32.exe|svchost.exe|dllhost.exe|cmstp.exe)")','2026-04-01 17:59:08.779',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"log.eventDataProcessName.keyword","operator":"filter_term","value":"{{.log.eventDataProcessName}}"},{"field":"log.eventDataProcessID","operator":"filter_term","value":"{{.log.eventDataProcessID}}"}],"or":null,"within":"now-5m","count":3}]','["origin.ip","target.user"]'), (1371,'Windows: UAC Bypass Attempt via Privileged IFileOperation COM Interface',1,3,2,'Privilege Escalation','T1548.002 - Abuse Elevation Control Mechanism: Bypass User Account Control','Identifies attempts to bypass User Account Control (UAC) via DLL side-loading. Attackers may attempt to bypass UAC to stealthily execute code with elevated permissions.','["https://attack.mitre.org/tactics/TA0004/","https://attack.mitre.org/techniques/T1548/002/"]','regexMatch("log.eventDataFileName", "(wow64log.dll|comctl32.dll|DismCore.dll|OskSupport.dll|duser.dll|Accessibility.ni.dll)") && regexMatch("log.eventDataProcessName", "(dllhost.exe|eventvwr.exe|computerdefaults.exe|sdclt.exe|fodhelper.exe|wsreset.exe|slui.exe)") && !regexMatch("log.eventDataFileName", "(C:\\\\Windows\\\\SoftwareDistribution\\\\|C:\\\\Windows\\\\WinSxS\\\\)")','2026-04-01 17:59:11.347',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1372,'Windows: Unusual File Modification by dns.exe',1,3,2,'Lateral Movement','T1210 - Exploitation of Remote Services','Identifies an unexpected file being modified by dns.exe, the process responsible for Windows DNS Server services, which may indicate activity related to remote code execution or other forms of exploitation.','["https://attack.mitre.org/tactics/TA0001/","https://attack.mitre.org/techniques/T1133/"]','equals("log.eventCode", 4663) && contains("log.eventDataProcessName", "dns.exe") && !regexMatch("log.eventDataFileName", "(dns\\.log$|\\.dns$|\\Windows\\System32\\dns\\)") && regexMatch("log.eventDataFileName", "\\.(dll|exe|bat|ps1|vbs)$")','2026-04-01 17:59:13.312',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1373,'Windows: Unusual Network Connection via DllHost or via RunDLL32',2,3,2,'Defense Evasion','T1218 - System Binary Proxy Execution','Identifies unusual instances of dllhost.exe making outbound network connections. This may indicate adversarial Command and Control activity. Identifies unusual instances of rundll32.exe making outbound network connections. This may indicate adversarial Command and Control activity.','["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1218/"]','regexMatch("log.eventDataProcessName", "(dllhost.exe|rundll32.exe)") && !inCIDR("origin.ip", "10.0.0.0/8") && !inCIDR("origin.ip", "172.16.0.0/12") && !inCIDR("origin.ip", "192.168.0.0/16") && !inCIDR("origin.ip", "127.0.0.0/8") && !inCIDR("origin.ip", "169.254.0.0/16") && !inCIDR("origin.ip", "198.18.0.0/15") && !inCIDR("origin.ip", "224.0.0.0/4") && !inCIDR("origin.ip", "240.0.0.0/4")','2026-04-01 17:59:16.116',true,true,'target',NULL,'[]','["origin.host","target.user"]'), (1374,'Windows: Unusual Process Network Connection',3,3,2,'Defense Evasion','Trusted Developer Utilities Proxy Execution','Identifies network activity from unexpected system applications. This may indicate adversarial activity as these applications are often leveraged by adversaries to execute code and evade detection.','["https://attack.mitre.org/tactics/TA0005/","https://attack.mitre.org/techniques/T1127/"]','regexMatch("log.eventDataProcessName", "(Microsoft.Workflow.Compiler.exe|bginfo.exe|cdb.exe|cmstp.exe|csi.exe|dnx.exe|fsi.exe|ieexec.exe|iexpress.exe|odbcconf.exe|rcsi.exe|xwizard.exe)")','2026-04-01 17:59:18.862',true,true,'origin',NULL,'[]','["origin.ip","target.user"]'), (1375,'ADFS Authentication Anomalies',3,2,2,'Defense Evasion, Persistence, Privilege Escalation, Initial Access','T1078 - Valid Accounts','Detects anomalous authentication attempts against Active Directory Federation Services (ADFS) including multiple failed attempts that could indicate password spraying or brute force attacks. This rule monitors for authentication failures, token validation failures, and other ADFS security events that may indicate malicious activity.

Next Steps:
1. Review the source IP address and determine if it''s from a known/trusted location
2. Check for patterns of failed authentication attempts across multiple users
3. Examine ADFS audit logs for additional context around the authentication failures
4. Verify if the targeted user accounts are valid and active
5. Consider implementing IP-based blocking if malicious activity is confirmed
6. Review ADFS configuration for security hardening opportunities
7. Correlate with other authentication events across the domain
','["https://learn.microsoft.com/en-us/windows-server/identity/ad-fs/troubleshooting/ad-fs-tshoot-logging","https://attack.mitre.org/techniques/T1078/"]','equals("log.providerName", "AD FS") && (equals("log.eventId", "411") || equals("log.eventId", "342") || equals("log.eventId", "516")) && contains("log.message", "token validation failed")','2026-04-01 17:59:21.713',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-10m","count":10}]','["origin.ip","target.user"]');INSERT INTO public.utm_correlation_rules (id,rule_name,rule_confidentiality,rule_integrity,rule_availability,rule_category,rule_technique,rule_description,rule_references_def,rule_definition_def,rule_last_update,rule_active,system_owner,rule_adversary,rule_deduplicate_by_def,rule_after_events_def,rule_group_by_def) VALUES (1376,'AdminSDHolder Abuse Detection',3,3,2,'Persistence, Privilege Escalation','T1098 - Account Manipulation','Detects modifications to the AdminSDHolder object which can be used for persistence by granting elevated privileges. The SDProp process propagates these permissions to protected groups every 60 minutes, making this a critical security event.

Next Steps:
1. Immediately review the user account that performed the modification
2. Check if the modification was authorized and part of legitimate administrative activities
3. Examine the specific permissions that were changed on the AdminSDHolder object
4. Monitor for privilege escalation activities in the next 60 minutes (SDProp cycle)
5. Review all members of protected groups for unauthorized additions
6. Audit recent administrative activities by the same user account
7. Consider temporarily disabling the user account if unauthorized activity is suspected
','["https://attack.mitre.org/techniques/T1098/","https://adsecurity.org/?p=1906","https://docs.microsoft.com/en-us/windows-server/identity/ad-ds/plan/security-best-practices/appendix-c--protected-accounts-and-groups-in-active-directory"]','oneOf("log.eventCode", ["4662", "5136", "4670"]) &&
equals("log.channel", "Security") &&
(
  contains("log.eventDataObjectName", "CN=AdminSDHolder,CN=System")
) &&
(
  oneOf("log.eventDataOperationType", ["Object Access", "Write Property"]) ||
  oneOf("log.eventDataAccessMask", ["131072", "262144", "524288"])
) &&
!equals("log.eventDataSubjectUserName", "SYSTEM")
','2026-04-01 17:59:24.437',true,true,'origin',NULL,'[]','["lastEvent.log.eventDataObjectName","lastEvent.log.eventDataSubjectUserName"]'), (1377,'AS-REP Roasting Attack Detection',3,2,1,'Credential Access','T1558.004 - Steal or Forge Kerberos Tickets: AS-REP Roasting','Detects AS-REP Roasting attacks targeting accounts with Kerberos pre-authentication disabled.
Attackers request AS-REP messages encrypted with RC4 (0x17) for accounts that do not require
pre-authentication, enabling offline password cracking. This is a companion technique to Kerberoasting
and targets a different set of vulnerable accounts.

Next Steps:
1. Identify accounts with pre-authentication disabled and evaluate business justification
2. Enable Kerberos pre-authentication on all identified accounts
3. Verify the requesting source IP is not a known attack tool
4. Reset passwords for targeted accounts using strong, complex passwords
5. Audit Active Directory for accounts with DONT_REQUIRE_PREAUTH flag
6. Monitor for subsequent credential usage from the requesting IP
','["https://attack.mitre.org/techniques/T1558/004/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4768","https://blog.harmj0y.net/activedirectory/roasting-as-reps/"]','equals("log.eventCode", "4768") &&
equals("log.channel", "Security") &&
equals("log.eventDataTicketEncryptionType", "23") &&
equals("log.eventDataPreAuthType", "0") &&
!regexMatch("target.user", "(?i)\\$$") &&
exists("target.user")
','2026-04-01 17:59:27.203',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-15m","count":3}]','["origin.ip","origin.host"]'), (1378,'Certificate Services Abuse Detection',3,3,1,'Credential Access','T1558 - Steal or Forge Kerberos Tickets','Detects suspicious certificate requests and issuance that could indicate Golden Certificate attacks or unauthorized certificate generation for persistence. This rule monitors Windows Certificate Services events for potentially malicious certificate operations, particularly those involving machine accounts or anonymous logons that could be leveraged for persistence and privilege escalation.

Next Steps:
1. Investigate the certificate request details including the requesting user/machine
2. Verify if the certificate request was legitimate and authorized
3. Check for any recent changes to Certificate Authority policies or templates
4. Review Certificate Authority logs for other suspicious certificate issuance
5. Examine the requesting host for signs of compromise
6. Consider revoking any suspicious certificates issued
7. Validate Certificate Authority security configurations and access controls
','["https://www.splunk.com/en_us/blog/security/breaking-the-chain-defending-against-certificate-services-abuse.html","https://attack.mitre.org/techniques/T1558/"]','(equals("log.eventId", "4886") || equals("log.eventId", "4887")) && equals("log.providerName", "Microsoft-Windows-Security-Auditing") && (contains("log.eventDataSubjectUserName", "$") || equals("log.eventDataSubjectUserName", "ANONYMOUS LOGON"))','2026-04-01 17:59:29.862',true,true,'origin',NULL,'[]','["lastEvent.log.eventDataSubjectUserName","origin.host"]'), (1379,'Golden Ticket Attack Detection',3,3,3,'Credential Access','T1558.001 - Steal or Forge Kerberos Tickets: Golden Ticket','Detects Golden Ticket attacks where adversaries forge Kerberos TGTs using the KRBTGT account
hash, granting unlimited domain access. The rule detects anomalous TGT usage patterns including
TGS requests with unusual encryption types, tickets with abnormally long lifetimes, and Kerberos
authentication from non-domain-controller sources for the KRBTGT service.

Next Steps:
1. Immediately verify if the KRBTGT account password has been compromised
2. Reset the KRBTGT password TWICE to invalidate all existing tickets
3. Identify the source host and investigate for full domain compromise
4. Review all domain admin activity from the suspected timeframe
5. Check for DCSync or NTDS.dit extraction as precursor activities
6. Audit all privileged account access across the domain
7. Consider rebuilding the domain if compromise is confirmed
8. Implement Kerberos armoring and constrained delegation
','["https://attack.mitre.org/techniques/T1558/001/","https://adsecurity.org/?p=1640","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4769"]','(
  equals("log.eventCode", "4769") &&
  equals("log.channel", "Security") &&
  equals("log.eventDataServiceName", "krbtgt") &&
  !equals("log.eventDataStatus", "0") &&
  exists("origin.ip")
) ||
(
  equals("log.eventCode", "4768") &&
  equals("log.channel", "Security") &&
  !oneOf("log.eventDataTicketEncryptionType", ["18", "17"]) &&
  exists("target.user") &&
  !regexMatch("target.user", "(?i)\\$$")
) ||
(
  equals("log.eventCode", "4672") &&
  equals("log.channel", "Security") &&
  contains("log.eventDataPrivilegeList", "SeTcbPrivilege") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|LOCAL SERVICE|NETWORK SERVICE)$") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)\\$$")
)
','2026-04-01 17:59:32.300',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-30m","count":3}]','["origin.host","target.user"]'), (1380,'Kerberoasting Attack Detection',3,2,1,'Credential Access','T1558.003 - Steal or Forge Kerberos Tickets: Kerberoasting','Detects Kerberoasting attacks where adversaries request Kerberos TGS tickets encrypted with RC4 (0x17) for
service accounts in order to crack them offline and obtain plaintext credentials. This is the most common
Active Directory credential theft technique used in real-world compromises. The rule monitors Event ID 4769
(Kerberos Service Ticket Operations) for RC4 encryption requests while excluding machine accounts (ending in $)
and legitimate system services.

Next Steps:
1. Identify the requesting user account and verify if this is authorized security testing
2. Check which service account SPNs were targeted for TGS requests
3. Review if the requesting account has been compromised
4. Audit all service accounts with SPNs for weak passwords
5. Consider implementing AES-only Kerberos encryption policies
6. Rotate passwords for targeted service accounts immediately
7. Enable Group Managed Service Accounts (gMSA) where possible
8. Monitor for follow-up lateral movement using obtained credentials
','["https://attack.mitre.org/techniques/T1558/003/","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4769","https://adsecurity.org/?p=2293"]','equals("log.eventCode", "4769") &&
equals("log.channel", "Security") &&
equals("log.eventDataTicketEncryptionType", "23") &&
!regexMatch("log.eventDataServiceName", "(?i)\\$$") &&
!equals("log.eventDataServiceName", "krbtgt") &&
!oneOf("log.eventDataTicketOptions", ["1082195968", "1082130432", "1082130432"]) &&
exists("log.eventDataServiceName")
','2026-04-01 17:59:34.648',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-15m","count":3}]','["origin.ip","target.user"]'), (1381,'Process Masquerading Detection',2,3,2,'Defense Evasion','T1036.005 - Masquerading: Match Legitimate Name or Location','Detects executables masquerading as legitimate Windows system processes but running from
incorrect locations. For example, svchost.exe should only run from C:\Windows\System32,
and explorer.exe should only run from C:\Windows. Malware commonly uses legitimate process
names to avoid detection by analysts and automated tools.

Next Steps:
1. Identify the actual file path of the masquerading process
2. Compare the file hash against known good versions of the legitimate binary
3. Check the digital signature of the suspicious executable
4. Analyze the executable in a sandbox environment
5. Review the parent process that launched the masquerading binary
6. Kill the suspicious process and quarantine the file
7. Search for other instances of the same file across the environment
','["https://attack.mitre.org/techniques/T1036/005/","https://www.elastic.co/blog/how-hunt-masquerade-ball","https://redcanary.com/threat-detection-report/techniques/masquerading/"]','(equals("log.eventCode", "4688") || equals("log.eventCode", "1")) &&
(
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\csrss\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\Windows\\\\(System32|SysWOW64)\\\\csrss\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\services\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\Windows\\\\System32\\\\services\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\smss\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\Windows\\\\System32\\\\smss\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\wininit\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\Windows\\\\System32\\\\wininit\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\explorer\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\Windows\\\\explorer\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\svchost\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\Windows\\\\(System32|SysWOW64)\\\\svchost\\.exe$")
  ) ||
  (
    regexMatch("log.eventDataProcessName", "(?i)\\\\lsass\\.exe$") &&
    !regexMatch("log.eventDataProcessName", "(?i)^C:\\\\Windows\\\\System32\\\\lsass\\.exe$")
  )
)
','2026-04-01 17:59:37.207',true,true,'origin',NULL,'[]','["origin.host","target.user"]'), (1382,'NTDS.dit Extraction Attempt',3,3,1,'Credential Access','T1003.003 - OS Credential Dumping: NTDS','Detects attempts to access or copy the Active Directory domain database (NTDS.dit) which contains password hashes for all domain users. This is a critical indicator of credential theft attempts and potential domain compromise.

Next Steps:
1. Immediately isolate the affected system to prevent further compromise
2. Review all recent activity from the source host and user account
3. Check for signs of lateral movement from this system
4. Verify integrity of domain controllers and examine recent administrative actions
5. Look for evidence of credential harvesting tools (ntdsutil, vssadmin, mimikatz)
6. Review privileged account usage and consider forcing password resets
7. Examine network traffic for data exfiltration attempts
8. Check backup systems and shadow copies for unauthorized access
9. Coordinate with incident response team for full forensic analysis
','["https://attack.mitre.org/techniques/T1003/003/","https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-10/security/threat-protection/auditing/event-4663"]','oneOf("log.eventCode", ["4663", "4656"]) &&
equals("log.channel", "Security") &&
(
  endsWith("log.eventDataObjectName", "\\ntds.dit") ||
  contains("log.eventDataObjectName", "\\NTDS\\") ||
  endsWith("log.eventDataProcessName", "\\ntdsutil.exe") ||
  endsWith("log.eventDataProcessName", "\\vssadmin.exe")
) &&
!equals("log.eventDataAccessMask", "0")
','2026-04-01 17:59:40.079',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.host","operator":"filter_term","value":"{{.origin.host}}"}],"or":null,"within":"now-30m","count":2}]','["origin.host","target.user"]'), (1383,'NTLM Authentication Downgrade Attack',3,3,1,'Defense Evasion','T1562.001 - Impair Defenses: Disable or Modify Tools','Detects NTLM authentication downgrade attacks via registry modifications to LMCompatibilityLevel,
NtlmMinClientSec, and NtlmMinServerSec. Attackers modify these registry values to weaken NTLM
authentication security, enabling credential interception, relay attacks, and offline cracking
of captured NTLM hashes. Downgrading to LM or NTLMv1 authentication makes credential theft
significantly easier.

Next Steps:
1. Check the new registry value to determine the downgrade severity
2. LMCompatibilityLevel < 3 enables NTLMv1 which is trivially crackable
3. Identify the process and user that made the registry modification
4. Restore the registry values to enforce NTLMv2 (LMCompatibilityLevel = 5)
5. Review Group Policy for conflicting NTLM settings
6. Check for NTLM relay attacks following the downgrade
7. Audit network traffic for NTLMv1 authentication attempts
8. Consider disabling NTLM entirely where possible
','["https://attack.mitre.org/techniques/T1562/001/","https://www.praetorian.com/blog/ntlm-relaying-attacks/","https://docs.microsoft.com/en-us/windows/security/threat-protection/security-policy-settings/network-security-lan-manager-authentication-level"]','equals("log.eventCode", "4657") &&
equals("log.channel", "Security") &&
(
  regexMatch("log.eventDataObjectName", "(?i)\\\\CurrentControlSet\\\\Control\\\\Lsa\\\\LMCompatibilityLevel") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\CurrentControlSet\\\\Control\\\\Lsa\\\\MSV1_0\\\\NtlmMinClientSec") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\CurrentControlSet\\\\Control\\\\Lsa\\\\MSV1_0\\\\NtlmMinServerSec") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\CurrentControlSet\\\\Control\\\\Lsa\\\\MSV1_0\\\\RestrictSendingNTLMTraffic") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\CurrentControlSet\\\\Control\\\\Lsa\\\\MSV1_0\\\\AuditReceivingNTLMTraffic") ||
  regexMatch("log.eventDataObjectName", "(?i)\\\\CurrentControlSet\\\\Services\\\\Netlogon\\\\Parameters\\\\RequireSignOrSeal")
) &&
!regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|TrustedInstaller)$")
','2026-04-01 17:59:42.451',true,true,'origin',NULL,'[]','["origin.host","target.user"]'), (1384,'Pass-the-Hash Attack Detection',3,3,2,'Lateral Movement','T1550.002 - Use Alternate Authentication Material: Pass the Hash','Detects Pass-the-Hash attacks by monitoring for NTLM authentication (Event ID 4624) with
LogonType 9 (NewCredentials) or LogonType 3 (Network) from unusual sources, combined with
the use of Seclogon service. Attackers use stolen NTLM hashes to authenticate without
knowing the plaintext password, commonly through tools like Mimikatz sekurlsa::pth,
Impacket, or CrackMapExec.

Next Steps:
1. Identify the source IP and user account used for the NTLM authentication
2. Verify if the source host should be authenticating with NTLM to this target
3. Check for prior credential dumping activity on the source host
4. Review if the authentication was followed by lateral movement or data access
5. Reset the compromised account password and any related accounts
6. Implement NTLM restrictions via Group Policy where possible
7. Enable Windows Defender Credential Guard to protect NTLM hashes
','["https://attack.mitre.org/techniques/T1550/002/","https://www.sans.org/blog/pass-the-hash-attack-detection/","https://stealthbits.com/blog/how-to-detect-pass-the-hash-attacks/"]','(
  equals("log.eventCode", "4624") &&
  equals("log.channel", "Security") &&
  equals("log.eventDataLogonType", "9") &&
  equals("log.eventDataAuthenticationPackageName", "Negotiate") &&
  !regexMatch("log.eventDataSubjectUserName", "(?i)^(SYSTEM|LOCAL SERVICE|NETWORK SERVICE|ANONYMOUS LOGON|-|\\$)") &&
  exists("target.user") &&
  !regexMatch("target.user", "(?i)\\$$")
) ||
(
  equals("log.eventCode", "4624") &&
  equals("log.channel", "Security") &&
  equals("log.eventDataLogonType", "3") &&
  equals("log.eventDataLmPackageName", "NTLM V1") &&
  exists("origin.ip") &&
  !equals("origin.ip", "-") &&
  !equals("origin.ip", "::1") &&
  !equals("origin.ip", "127.0.0.1")
)
','2026-04-01 17:59:45.498',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.eventCode","operator":"filter_term","value":"4624"}],"or":null,"within":"now-30m","count":3}]','["origin.ip","origin.host","target.user"]'), (1385,'PowerShell Empire Detection',3,3,2,'Execution','T1059.001 - Command and Scripting Interpreter: PowerShell','Detects potential PowerShell Empire framework usage based on characteristic command patterns, obfuscation techniques, and encoded payloads commonly used by this post-exploitation framework. PowerShell Empire is a post-exploitation framework that uses PowerShell and Python agents to maintain persistence and execute commands on compromised systems.

Next Steps:
1. Immediately isolate the affected host to prevent lateral movement
2. Analyze the complete PowerShell script block content for additional IOCs
3. Check for persistence mechanisms (scheduled tasks, registry entries, services)
4. Review network connections from the host for C2 communication
5. Examine process tree and parent processes that spawned PowerShell
6. Search for additional Empire artifacts across the environment
7. Reset credentials for any accounts used on the compromised system
8. Conduct memory analysis to identify injected code or payloads
9. Review recent user activity and file access patterns
10. Update endpoint detection rules based on specific Empire techniques observed
','["https://attack.mitre.org/techniques/T1059/001/","https://www.powershellempire.com/","https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_logging"]','(equals("log.eventCode", "4104") || equals("log.eventId", 4104)) &&
equals("log.providerName", "Microsoft-Windows-PowerShell") &&
(
  contains("log.eventDataScriptBlockText", "System.Management.Automation.AmsiUtils") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)(empire|invoke-empire|invoke-psempire)") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)\\[System\\.Convert\\]::FromBase64String") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)IEX\\s*\\(\\s*New-Object") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)-enc\\s+[A-Za-z0-9+/=]{100,}") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)\\$DoIt\\s*=\\s*@") ||
  regexMatch("log.eventDataScriptBlockText", "(?i)\\[System\\.Text\\.Encoding\\]::Unicode\\.GetString") ||
  contains("log.eventDataScriptBlockText", "Invoke-Shellcode") ||
  contains("log.eventDataScriptBlockText", "Invoke-ReflectivePEInjection") ||
  contains("log.eventDataScriptBlockText", "Invoke-Mimikatz")
)
','2026-04-01 17:59:47.795',true,true,'origin',NULL,'[]','["origin.host","target.user"]');INSERT INTO public.utm_correlation_rules (id,rule_name,rule_confidentiality,rule_integrity,rule_availability,rule_category,rule_technique,rule_description,rule_references_def,rule_definition_def,rule_last_update,rule_active,system_owner,rule_adversary,rule_deduplicate_by_def,rule_after_events_def,rule_group_by_def) VALUES (1386,'RDP Brute Force Attack',3,2,2,'Credential Access','T1110.001 - Brute Force: Password Guessing','Detects multiple failed RDP login attempts from the same source IP address, indicating a potential brute force attack. This rule monitors Windows Event ID 4625 (failed logon) with focus on network logon types (type 3) which are commonly used for RDP connections. The rule triggers when 10 or more failed attempts occur from the same IP within 15 minutes.

Next Steps:
1. Investigate the source IP address for malicious indicators and geolocation
2. Check if the targeted user accounts are legitimate and active
3. Review successful logons from the same IP after failed attempts
4. Implement IP blocking or rate limiting for the source address
5. Enable account lockout policies if not already configured
6. Consider implementing multi-factor authentication for RDP access
7. Review RDP access logs for any successful connections during the attack timeframe
','["https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4625","https://attack.mitre.org/techniques/T1110/001/"]','equals("log.eventCode", "4625") && equals("log.eventDataLogonType", "3") && exists("origin.ip") && !equals("origin.ip", "-") && !equals("origin.ip", "::1") && !equals("origin.ip", "127.0.0.1")','2026-04-01 17:59:50.462',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"},{"field":"log.eventCode","operator":"filter_term","value":"4625"},{"field":"log.eventDataLogonType","operator":"filter_term","value":"3"}],"or":null,"within":"now-15m","count":10}]','["origin.ip","target.host"]'), (1387,'SAM Database Access Attempt',3,3,1,'Credential Access','T1003.002 - OS Credential Dumping: Security Account Manager','Detects attempts to access the Security Account Manager (SAM) database, which contains local user account hashes. This activity may indicate credential dumping attempts by attackers trying to extract password hashes for offline cracking or lateral movement.

Next Steps:
1. Immediately investigate the user account and process that accessed the SAM database
2. Check for any unusual processes running on the affected system
3. Review recent logon events and privilege escalation activities
4. Examine network connections from the affected host for lateral movement
5. Consider isolating the affected system if malicious activity is confirmed
6. Review security policies for SAM database access permissions
7. Check for presence of credential dumping tools or suspicious files
','["https://attack.mitre.org/techniques/T1003/002/","https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-10/security/threat-protection/auditing/event-4661"]','equals("log.eventCode", "4663") &&
equals("log.channel", "Security") &&
(
  endsWith("log.eventDataObjectName", "\\SAM") ||
  endsWith("log.eventDataObjectName", "\\SECURITY") ||
  endsWith("log.eventDataObjectName", "\\SYSTEM")
) &&
oneOf("log.eventDataAccessMask", ["131097", "2032127", "64", "32", "1"])
','2026-04-01 17:59:53.256',true,true,'origin',NULL,'[]','["origin.host","target.user"]'), (1388,'SID History Injection Attempt',3,3,1,'Defense Evasion, Privilege Escalation','T1134.005 - Access Token Manipulation: SID-History Injection','Detects attempts to add SID History to an account, which can be used for privilege escalation. SID History injection allows attackers to inherit permissions from privileged accounts without being members of privileged groups. Both successful (4765) and failed (4766) attempts are monitored.

Next Steps:
1. Immediately investigate the target user account and verify if SID History modification was legitimate
2. Check if the user performing the action has proper administrative privileges for this operation
3. Review the source SID being added to understand what permissions are being inherited
4. Examine recent authentication logs for the target account to identify potential unauthorized access
5. Verify Active Directory configuration and check for signs of domain controller compromise
6. Consider resetting the target account password and removing unauthorized SID History entries
7. Review domain administrator accounts and privileged group memberships for anomalies
','["https://attack.mitre.org/techniques/T1134/005/","https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4765","https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4766","https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventid=4765"]','oneOf("log.eventId", ["4765", "4766"]) &&
equals("log.channel", "Security")
','2026-04-01 17:59:55.635',true,true,'origin',NULL,'[]','["origin.host","target.user"]'), (1389,'Silver Ticket Attack Detection',3,3,2,'Credential Access','T1558.002 - Steal or Forge Kerberos Tickets: Silver Ticket','Detects Silver Ticket attacks where adversaries forge Kerberos TGS tickets using a service
account''s NTLM hash, bypassing the KDC entirely. Unlike Golden Tickets, Silver Tickets target
specific services. The rule detects TGS tickets presented without corresponding TGT requests,
and service access events with anomalous Kerberos authentication patterns.

Next Steps:
1. Identify the targeted service and its associated service account
2. Verify if the service account hash has been compromised via Kerberoasting
3. Reset the targeted service account password immediately
4. Review access logs for the targeted service for unauthorized activity
5. Check for prior Kerberoasting activity targeting the same service SPN
6. Investigate the source host for compromise indicators
7. Implement AES-only encryption for service accounts
8. Enable Kerberos PAC validation on the targeted services
','["https://attack.mitre.org/techniques/T1558/002/","https://adsecurity.org/?p=2011","https://www.sans.org/blog/kerberos-in-the-crosshairs-golden-tickets-silver-tickets-mitm-and-more/"]','equals("log.eventCode", "4769") &&
equals("log.channel", "Security") &&
(
  equals("log.eventDataTicketEncryptionType", "23") &&
  !regexMatch("log.eventDataServiceName", "(?i)(krbtgt|\\$$)") &&
  !oneOf("log.eventDataStatus", ["0", "6"]) &&
  exists("origin.ip")
)
','2026-04-01 17:59:58.404',true,true,'origin',NULL,'[{"indexPattern":"v11-log-wineventlog-*","with":[{"field":"origin.ip.keyword","operator":"filter_term","value":"{{.origin.ip}}"}],"or":null,"within":"now-15m","count":5}]','["origin.ip","origin.host"]'), (1390,'SMBv1 Usage Detection',3,2,2,'Lateral Movement','T1210 - Exploitation of Remote Services','Detects usage of the deprecated and vulnerable SMBv1 protocol which could be exploited for lateral movement or ransomware propagation. SMBv1 is susceptible to numerous security vulnerabilities including EternalBlue and should be disabled in favor of SMBv2/SMBv3.

Next Steps:
1. Immediately investigate the source system using SMBv1 and identify which service or application is still dependent on this protocol
2. Review network traffic logs to determine if this is internal communication or external access attempts
3. Check for any signs of exploitation attempts or successful compromises on the affected system
4. Identify all systems in the environment that may still have SMBv1 enabled
5. Plan migration to SMBv2/SMBv3 and disable SMBv1 on all systems where possible
6. Monitor for any lateral movement patterns that may indicate ongoing compromise
7. Consider implementing network segmentation to limit exposure if SMBv1 cannot be immediately disabled
','["https://learn.microsoft.com/en-us/windows-server/storage/file-server/troubleshoot/detect-enable-and-disable-smbv1-v2-v3","https://attack.mitre.org/techniques/T1210/"]','equals("log.eventId", "3000") && equals("log.providerName", "Microsoft-Windows-SMBServer") && contains("log.message", "SMB1")','2026-04-01 18:00:00.954',true,true,'origin',NULL,'[]','["origin.host","origin.ip"]'), (1391,'Windows Remote Management (WinRM) Abuse',3,3,2,'Lateral Movement','T1021.006 - Remote Services: Windows Remote Management','Detects potential abuse of Windows Remote Management (WinRM) for lateral movement. Monitors for successful logon events (4624) with network logon type 3 combined with privilege escalation (4672) and WinRM-related process activity, indicating remote command execution via WinRM.

Next Steps:
1. Investigate the source IP address and verify if it''s an authorized administrative workstation
2. Review the target user account for any signs of compromise or unusual privilege usage
3. Examine recent PowerShell execution logs and command history on the target system
4. Check for concurrent suspicious activities from the same source IP across other systems
5. Verify if the WinRM connection aligns with scheduled maintenance or authorized administrative tasks
6. Review network traffic patterns between source and target systems for data exfiltration indicators
7. Validate the legitimacy of any processes spawned through the WinRM session
8. Consider implementing additional monitoring for WinRM usage if this represents unexpected activity
','["https://jpcertcc.github.io/ToolAnalysisResultSheet/details/WinRM.htm","https://attack.mitre.org/techniques/T1021/006/"]','equals("log.eventCode", "4624") && equals("log.eventDataLogonType", "3") && exists("log.eventDataProcessName") && (contains("log.eventDataProcessName", "wsmprovhost.exe") || contains("log.eventDataProcessName", "winrshost.exe") || contains("log.eventDataProcessName", "powershell.exe"))','2026-04-01 18:00:03.807',true,true,'origin',NULL,'[]','["target.user","origin.host","origin.ip"]');

